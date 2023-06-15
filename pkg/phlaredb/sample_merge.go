package phlaredb

import (
	"context"
	"sort"

	"github.com/google/pprof/profile"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/common/model"
	"github.com/samber/lo"
	"github.com/segmentio/parquet-go"

	googlev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	ingestv1 "github.com/grafana/phlare/api/gen/proto/go/ingester/v1"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	"github.com/grafana/phlare/pkg/iter"
	phlaremodel "github.com/grafana/phlare/pkg/model"
	"github.com/grafana/phlare/pkg/phlaredb/query"
)

func (b *singleBlockQuerier) MergeByStacktraces(ctx context.Context, rows iter.Iterator[Profile]) (*ingestv1.MergeProfilesStacktracesResult, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "MergeByStacktraces - Block")
	defer sp.Finish()

	stacktraceAggrValues := make(stacktraceSampleMap)
	if err := mergeByStacktraces(ctx, b.profiles.file, rows, stacktraceAggrValues); err != nil {
		return nil, err
	}

	// TODO: Truncate insignificant stacks.
	return b.resolveSymbols(ctx, stacktraceAggrValues)
}

func (b *singleBlockQuerier) MergePprof(ctx context.Context, rows iter.Iterator[Profile]) (*profile.Profile, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "MergeByStacktraces - Block")
	defer sp.Finish()

	stacktraceAggrValues := make(profileSampleMap)
	if err := mergeByStacktraces(ctx, b.profiles.file, rows, stacktraceAggrValues); err != nil {
		return nil, err
	}

	return b.resolvePprofSymbols(ctx, stacktraceAggrValues)
}

func (b *singleBlockQuerier) resolvePprofSymbols(ctx context.Context, stacktraceAggrByID map[int32]*profile.Sample) (*profile.Profile, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "ResolvePprofSymbols - Block")
	defer sp.Finish()

	// gather stacktraces
	stacktraceIDs := lo.Keys(stacktraceAggrByID)
	locationsIdsByStacktraceID := map[int64][]uint64{}

	sort.Slice(stacktraceIDs, func(i, j int) bool {
		return stacktraceIDs[i] < stacktraceIDs[j]
	})

	locationIDs := newUniqueIDs[struct{}]()
	stacktraces := repeatedColumnIter(ctx, b.stacktraces.file, "LocationIDs.list.element", iter.NewSliceIterator(stacktraceIDs))
	defer stacktraces.Close()

	for stacktraces.Next() {
		s := stacktraces.At()
		locationsIdsByStacktraceID[int64(s.Row)] = make([]uint64, len(s.Values))
		for i, locationID := range s.Values {
			locID := locationID.Uint64()
			locationIDs[int64(locID)] = struct{}{}
			locationsIdsByStacktraceID[int64(s.Row)][i] = locID
		}

	}
	if err := stacktraces.Err(); err != nil {
		return nil, err
	}
	sp.LogFields(otlog.Int("stacktraces", len(stacktraceIDs)))

	// gather locations
	var (
		functionIDs         = newUniqueIDs[struct{}]()
		mappingIDs          = newUniqueIDs[lo.Tuple2[*profile.Mapping, *googlev1.Mapping]]()
		locations           = b.locations.retrieveRows(ctx, locationIDs.iterator())
		locationModelsByIds = map[uint64]*profile.Location{}
		functionModelsByIds = map[uint64]*profile.Function{}
	)
	for locations.Next() {
		s := locations.At()
		m, ok := mappingIDs[int64(s.Result.MappingId)]
		if !ok {
			m = lo.T2(&profile.Mapping{
				ID: s.Result.MappingId,
			}, &googlev1.Mapping{
				Id: s.Result.MappingId,
			})
			mappingIDs[int64(s.Result.MappingId)] = m
		}
		loc := &profile.Location{
			ID:       s.Result.Id,
			Address:  s.Result.Address,
			IsFolded: s.Result.IsFolded,
			Mapping:  m.A,
		}
		for _, line := range s.Result.Line {
			functionIDs[int64(line.FunctionId)] = struct{}{}
			fn, ok := functionModelsByIds[line.FunctionId]
			if !ok {
				fn = &profile.Function{
					ID: line.FunctionId,
				}
				functionModelsByIds[line.FunctionId] = fn
			}

			loc.Line = append(loc.Line, profile.Line{
				Line:     line.Line,
				Function: fn,
			})
		}
		locationModelsByIds[uint64(s.RowNum)] = loc
	}
	if err := locations.Err(); err != nil {
		return nil, err
	}

	// gather functions
	var (
		stringsIds    = newUniqueIDs[int64]()
		functions     = b.functions.retrieveRows(ctx, functionIDs.iterator())
		functionsById = map[int64]*googlev1.Function{}
	)
	for functions.Next() {
		s := functions.At()
		functionsById[int64(s.Result.Id)] = &googlev1.Function{
			Id:         s.Result.Id,
			Name:       s.Result.Name,
			SystemName: s.Result.SystemName,
			Filename:   s.Result.Filename,
			StartLine:  s.Result.StartLine,
		}
		stringsIds[s.Result.Name] = 0
		stringsIds[s.Result.Filename] = 0
		stringsIds[s.Result.SystemName] = 0
	}
	if err := functions.Err(); err != nil {
		return nil, err
	}
	// gather mapping
	mapping := b.mappings.retrieveRows(ctx, mappingIDs.iterator())
	for mapping.Next() {
		cur := mapping.At()
		m := mappingIDs[int64(cur.Result.Id)]
		m.B.Filename = cur.Result.Filename
		m.B.BuildId = cur.Result.BuildId
		m.A.Start = cur.Result.MemoryStart
		m.A.Limit = cur.Result.MemoryLimit
		m.A.Offset = cur.Result.FileOffset
		m.A.HasFunctions = cur.Result.HasFunctions
		m.A.HasFilenames = cur.Result.HasFilenames
		m.A.HasLineNumbers = cur.Result.HasLineNumbers
		m.A.HasInlineFrames = cur.Result.HasInlineFrames

		stringsIds[cur.Result.Filename] = 0
		stringsIds[cur.Result.BuildId] = 0
	}
	// gather strings
	var (
		names   = make([]string, len(stringsIds))
		strings = b.strings.retrieveRows(ctx, stringsIds.iterator())
		idx     = int64(0)
	)
	for strings.Next() {
		s := strings.At()
		names[idx] = s.Result.String
		stringsIds[s.RowNum] = idx
		idx++
	}
	if err := strings.Err(); err != nil {
		return nil, err
	}

	for _, model := range mappingIDs {
		model.A.File = names[stringsIds[model.B.Filename]]
		model.A.BuildID = names[stringsIds[model.B.BuildId]]
	}

	mappingResult := make([]*profile.Mapping, 0, len(mappingIDs))
	for _, model := range mappingIDs {
		mappingResult = append(mappingResult, model.A)
	}

	for id, model := range stacktraceAggrByID {
		locsId := locationsIdsByStacktraceID[int64(id)]
		model.Location = make([]*profile.Location, len(locsId))
		for i, locId := range locsId {
			model.Location[i] = locationModelsByIds[locId]
		}
		// todo labels.
	}

	for id, model := range functionModelsByIds {
		fn := functionsById[int64(id)]
		model.Name = names[stringsIds[fn.Name]]
		model.Filename = names[stringsIds[fn.Filename]]
		model.SystemName = names[stringsIds[fn.SystemName]]
		model.StartLine = fn.StartLine
	}
	result := &profile.Profile{
		Sample:   lo.Values(stacktraceAggrByID),
		Location: lo.Values(locationModelsByIds),
		Function: lo.Values(functionModelsByIds),
		Mapping:  mappingResult,
	}
	normalizeProfileIds(result)

	return result, nil
}

func (b *singleBlockQuerier) resolveSymbols(ctx context.Context, stacktraceAggrByID map[int32]*ingestv1.StacktraceSample) (*ingestv1.MergeProfilesStacktracesResult, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "ResolveSymbols - Block")
	defer sp.Finish()
	locationsByStacktraceID := map[int64][]uint64{}

	// gather stacktraces
	sp.LogFields(otlog.String("msg", "gather stacktraces"))
	stacktraceIDs := lo.Keys(stacktraceAggrByID)
	sort.Slice(stacktraceIDs, func(i, j int) bool {
		return stacktraceIDs[i] < stacktraceIDs[j]
	})

	locationIDs := newUniqueIDs[struct{}]()
	stacktraces := repeatedColumnIter(ctx, b.stacktraces.file, "LocationIDs.list.element", iter.NewSliceIterator(stacktraceIDs))
	defer stacktraces.Close()

	for stacktraces.Next() {
		s := stacktraces.At()

		_, ok := locationsByStacktraceID[int64(s.Row)]
		if !ok {
			locationsByStacktraceID[int64(s.Row)] = make([]uint64, len(s.Values))
			for i, locationID := range s.Values {
				locID := locationID.Uint64()
				locationIDs[int64(locID)] = struct{}{}
				locationsByStacktraceID[int64(s.Row)][i] = locID
			}
			continue
		}
		for _, locationID := range s.Values {
			locID := locationID.Uint64()
			locationIDs[int64(locID)] = struct{}{}
			locationsByStacktraceID[int64(s.Row)] = append(locationsByStacktraceID[int64(s.Row)], locID)
		}
	}
	if err := stacktraces.Err(); err != nil {
		return nil, err
	}
	sp.LogFields(otlog.Int("stacktraces", len(stacktraceIDs)), otlog.Int("locationIDs", len(locationIDs)))

	// gather locations
	sp.LogFields(otlog.String("msg", "gather locations"))
	var (
		locationIDsByFunctionID = newUniqueIDs[[]int64]()
		locations               = b.locations.retrieveRows(ctx, locationIDs.iterator())
	)
	for locations.Next() {
		s := locations.At()

		for _, line := range s.Result.Line {
			locationIDsByFunctionID[int64(line.FunctionId)] = append(locationIDsByFunctionID[int64(line.FunctionId)], s.RowNum)
		}
	}
	if err := locations.Err(); err != nil {
		return nil, err
	}
	sp.LogFields(otlog.Int("functions", len(locationIDsByFunctionID)))

	// gather functions
	sp.LogFields(otlog.String("msg", "gather functions"))
	var (
		functionIDsByStringID = newUniqueIDs[[]int64]()
		functions             = b.functions.retrieveRows(ctx, locationIDsByFunctionID.iterator())
	)
	for functions.Next() {
		s := functions.At()

		functionIDsByStringID[s.Result.Name] = append(functionIDsByStringID[s.Result.Name], s.RowNum)
	}
	if err := functions.Err(); err != nil {
		return nil, err
	}

	// gather strings
	sp.LogFields(otlog.String("msg", "gather strings"))
	var (
		names   = make([]string, len(functionIDsByStringID))
		idSlice = make([][]int64, len(functionIDsByStringID))
		strings = b.strings.retrieveRows(ctx, functionIDsByStringID.iterator())
		idx     = 0
	)
	for strings.Next() {
		s := strings.At()
		names[idx] = s.Result.String
		idSlice[idx] = []int64{s.RowNum}
		idx++
	}
	if err := strings.Err(); err != nil {
		return nil, err
	}

	sp.LogFields(otlog.String("msg", "build MergeProfilesStacktracesResult"))
	// idSlice contains stringIDs and gets rewritten into functionIDs
	for nameID := range idSlice {
		var functionIDs []int64
		for _, stringID := range idSlice[nameID] {
			functionIDs = append(functionIDs, functionIDsByStringID[stringID]...)
		}
		idSlice[nameID] = functionIDs
	}

	// idSlice contains functionIDs and gets rewritten into locationIDs
	for nameID := range idSlice {
		var locationIDs []int64
		for _, functionID := range idSlice[nameID] {
			locationIDs = append(locationIDs, locationIDsByFunctionID[functionID]...)
		}
		idSlice[nameID] = locationIDs
	}

	// write a map locationID two nameID
	nameIDbyLocationID := make(map[int64]int32)
	for nameID := range idSlice {
		for _, locationID := range idSlice[nameID] {
			nameIDbyLocationID[locationID] = int32(nameID)
		}
	}

	// write correct string ID into each sample
	for stacktraceID, samples := range stacktraceAggrByID {
		locationIDs := locationsByStacktraceID[int64(stacktraceID)]

		functionIDs := make([]int32, len(locationIDs))
		for idx := range functionIDs {
			functionIDs[idx] = nameIDbyLocationID[int64(locationIDs[idx])]
		}
		samples.FunctionIds = functionIDs
	}

	return &ingestv1.MergeProfilesStacktracesResult{
		Stacktraces:   lo.Values(stacktraceAggrByID),
		FunctionNames: names,
	}, nil
}

func (b *singleBlockQuerier) MergeByLabels(ctx context.Context, rows iter.Iterator[Profile], by ...string) ([]*typesv1.Series, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "MergeByLabels - Block")
	defer sp.Finish()

	m := make(seriesByLabels)
	if err := mergeByLabels(ctx, b.profiles.file, rows, m, by...); err != nil {
		return nil, err
	}
	return m.normalize(), nil
}

type Source interface {
	Schema() *parquet.Schema
	RowGroups() []parquet.RowGroup
}

type profileSampleByMapping map[uint64]profileSampleMap

func (m profileSampleByMapping) add(mapping uint64, key int32, value int64) {
	if _, ok := m[mapping]; !ok {
		m[mapping] = make(profileSampleMap)
	}
	m[mapping].add(key, value)
}

func (m profileSampleByMapping) ForEach(f func(mapping uint64, samples profileSampleMap)) {
	for mapping, samples := range m {
		f(mapping, samples)
	}
}

func (m profileSampleByMapping) StacktraceSamples() []*profile.Sample {
	var result []*profile.Sample
	for _, samples := range m {
		result = append(result, lo.Values(samples)...)
	}
	return result
}

type profileSampleMap map[int32]*profile.Sample

func (m profileSampleMap) add(key int32, value int64) {
	if _, ok := m[key]; ok {
		m[key].Value[0] += value
		return
	}
	m[key] = &profile.Sample{
		Value: []int64{value},
	}
}

func (m profileSampleMap) Ids() iter.Iterator[int32] {
	return iter.NewSliceIterator(lo.Keys(m))
}

type stacktracesByMapping map[uint64]stacktraceSampleMap

func (m stacktracesByMapping) add(mapping uint64, key int32, value int64) {
	if _, ok := m[mapping]; !ok {
		m[mapping] = make(stacktraceSampleMap)
	}
	m[mapping].add(key, value)
}

func (m stacktracesByMapping) ForEach(f func(mapping uint64, samples stacktraceSampleMap)) {
	for mapping, samples := range m {
		f(mapping, samples)
	}
}

func (m stacktracesByMapping) StacktraceSamples() []*ingestv1.StacktraceSample {
	var result []*ingestv1.StacktraceSample
	for _, samples := range m {
		result = append(result, lo.Values(samples)...)
	}
	return result
}

type stacktraceSampleMap map[int32]*ingestv1.StacktraceSample

func (m stacktraceSampleMap) add(key int32, value int64) {
	if _, ok := m[key]; ok {
		m[key].Value += value
		return
	}
	m[key] = &ingestv1.StacktraceSample{
		Value: value,
	}
}

func (m stacktraceSampleMap) Ids() iter.Iterator[int32] {
	return iter.NewSliceIterator(lo.Keys(m))
}

type mapAdder interface {
	add(key int32, value int64)
}

func mergeByStacktraces(ctx context.Context, profileSource Source, rows iter.Iterator[Profile], m mapAdder) error {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "mergeByStacktraces")
	defer sp.Finish()
	// clone the rows to be able to iterate over them twice
	multiRows, err := iter.CloneN(rows, 2)
	if err != nil {
		return err
	}
	it := query.NewMultiRepeatedPageIterator(
		repeatedColumnIter(ctx, profileSource, "Samples.list.element.StacktraceID", multiRows[0]),
		repeatedColumnIter(ctx, profileSource, "Samples.list.element.Value", multiRows[1]),
	)
	defer it.Close()

	for it.Next() {
		values := it.At().Values
		for i := 0; i < len(values[0]); i++ {
			m.add(int32(values[0][i].Int64()), values[1][i].Int64())
		}
	}
	return nil
}

type seriesByLabels map[string]*typesv1.Series

func (m seriesByLabels) normalize() []*typesv1.Series {
	result := lo.Values(m)
	sort.Slice(result, func(i, j int) bool {
		return phlaremodel.CompareLabelPairs(result[i].Labels, result[j].Labels) < 0
	})
	// we have to sort the points in each series because labels reduction may have changed the order
	for _, s := range result {
		sort.Slice(s.Points, func(i, j int) bool {
			return s.Points[i].Timestamp < s.Points[j].Timestamp
		})
	}
	return result
}

func mergeByLabels(ctx context.Context, profileSource Source, rows iter.Iterator[Profile], m seriesByLabels, by ...string) error {
	it := repeatedColumnIter(ctx, profileSource, "Samples.list.element.Value", rows)

	defer it.Close()

	labelsByFingerprint := map[model.Fingerprint]string{}
	labelBuf := make([]byte, 0, 1024)

	for it.Next() {
		values := it.At()
		p := values.Row
		var total int64
		for _, e := range values.Values {
			total += e.Int64()
		}
		labelsByString, ok := labelsByFingerprint[p.Fingerprint()]
		if !ok {
			labelBuf = p.Labels().BytesWithLabels(labelBuf, by...)
			labelsByString = string(labelBuf)
			labelsByFingerprint[p.Fingerprint()] = labelsByString
			if _, ok := m[labelsByString]; !ok {
				m[labelsByString] = &typesv1.Series{
					Labels: p.Labels().WithLabels(by...),
					Points: []*typesv1.Point{
						{
							Timestamp: int64(p.Timestamp()),
							Value:     float64(total),
						},
					},
				}
				continue
			}
		}
		series := m[labelsByString]
		series.Points = append(series.Points, &typesv1.Point{
			Timestamp: int64(p.Timestamp()),
			Value:     float64(total),
		})
	}
	return it.Err()
}
