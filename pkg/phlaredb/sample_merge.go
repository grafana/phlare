package phlaredb

import (
	"context"
	"sort"

	"github.com/google/pprof/profile"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/common/model"
	"github.com/samber/lo"

	googlev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	ingestv1 "github.com/grafana/phlare/api/gen/proto/go/ingester/v1"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	"github.com/grafana/phlare/pkg/iter"
	phlaremodel "github.com/grafana/phlare/pkg/model"
	query "github.com/grafana/phlare/pkg/phlaredb/query"
)

// MergeStacktraceOptions is the options to use when merging stacktraces
type MergeStacktraceOptions struct {
	// HideOptions is the options to use when hiding stacktraces
	Hide HideOptions
}

type HideOptions struct {
	// Fraction hides stracktraces below/above of the profile's total * fraction
	Fraction float32
	// FractionStrategy is the strategy to use when Hiding the stacktraces
	Strategy HideStrategy
}

type HideStrategy int

const (
	// Only include the stacktraces where the value is above the total's fraction.
	TOP HideStrategy = 0
	// Only include the stacktraces where the value is below the total's fraction.
	BOTTOM HideStrategy = 1
	others              = "others"
)

func hideStacktraces[T any, ID comparable](stacktraceAggrByID map[ID]T, total int64, getValue func(T) int64, opt HideOptions) int64 {
	if opt.Fraction == 0 {
		return 0
	}
	if opt.Fraction < 0 || opt.Fraction > 1 {
		return 0
	}
	// calculate threshold
	threshold := int64(float32(total) * opt.Fraction)

	// hide stacktraces
	var hidden int64
	for id, sample := range stacktraceAggrByID {
		v := getValue(sample)
		if opt.Strategy == TOP && v < threshold {
			hidden += v
			delete(stacktraceAggrByID, id)
		}
		if opt.Strategy == BOTTOM && v > threshold {
			hidden += v
			delete(stacktraceAggrByID, id)
		}
	}
	return hidden
}

func (b *singleBlockQuerier) MergeByStacktraces(ctx context.Context, rows iter.Iterator[Profile], opt MergeStacktraceOptions) (*ingestv1.MergeProfilesStacktracesResult, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "MergeByStacktraces - Block")
	defer sp.Finish()
	// clone the rows to be able to iterate over them twice
	multiRows, err := iter.CloneN(rows, 2)
	if err != nil {
		return nil, err
	}
	it := query.NewMultiRepeatedPageIterator(
		repeatedColumnIter(ctx, b.profiles.file, "Samples.list.element.StacktraceID", multiRows[0]),
		repeatedColumnIter(ctx, b.profiles.file, "Samples.list.element.Value", multiRows[1]),
	)
	defer it.Close()

	stacktraceAggrValues := map[int64]*ingestv1.StacktraceSample{}
	total := int64(0)
	for it.Next() {
		values := it.At().Values
		for i := 0; i < len(values[0]); i++ {
			id, value := values[0][i].Int64(), values[1][i].Int64()
			total += value
			sample, ok := stacktraceAggrValues[id]
			if ok {
				sample.Value += value
				continue
			}
			stacktraceAggrValues[id] = &ingestv1.StacktraceSample{
				Value: value,
			}
		}
	}
	hidden := hideStacktraces(stacktraceAggrValues, total, func(s *ingestv1.StacktraceSample) int64 { return s.Value }, opt.Hide)
	result, err := b.resolveSymbols(ctx, stacktraceAggrValues)
	if err != nil {
		return nil, err
	}
	AddHiddenNodeToStacktraces(result, hidden)
	return result, nil
}

func AddHiddenNodeToStacktraces(result *ingestv1.MergeProfilesStacktracesResult, hidden int64) {
	if hidden == 0 {
		return
	}
	result.FunctionNames = append(result.FunctionNames, others)
	result.Stacktraces = append(result.Stacktraces, &ingestv1.StacktraceSample{
		Value:       hidden,
		FunctionIds: []int32{int32(len(result.FunctionNames) - 1)},
	})
}

func AddHiddenNodeToPprof(result *profile.Profile, hidden int64) {
	if hidden == 0 {
		return
	}
	hiddenFunction := &profile.Function{
		Name: others,
	}
	hiddenLocation := &profile.Location{
		Mapping: result.Mapping[0],
		Line: []profile.Line{
			{Function: hiddenFunction},
		},
	}
	result.Sample = append(result.Sample, &profile.Sample{
		Value: []int64{hidden},
		Location: []*profile.Location{
			hiddenLocation,
		},
	})
	result.Location = append(result.Location, hiddenLocation)
	hiddenLocation.ID = uint64(len(result.Location) - 1)
	result.Function = append(result.Function, hiddenFunction)
	hiddenFunction.ID = uint64(len(result.Function) - 1)
}

func (b *singleBlockQuerier) MergePprof(ctx context.Context, rows iter.Iterator[Profile], opt MergeStacktraceOptions) (*profile.Profile, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "MergeByStacktraces - Block")
	defer sp.Finish()
	// clone the rows to be able to iterate over them twice
	multiRows, err := iter.CloneN(rows, 2)
	if err != nil {
		return nil, err
	}
	it := query.NewMultiRepeatedPageIterator(
		repeatedColumnIter(ctx, b.profiles.file, "Samples.list.element.StacktraceID", multiRows[0]),
		repeatedColumnIter(ctx, b.profiles.file, "Samples.list.element.Value", multiRows[1]),
	)
	defer it.Close()

	var (
		stacktraceAggrValues = map[int64]*profile.Sample{}
		total                = int64(0)
	)

	for it.Next() {
		values := it.At().Values
		for i := 0; i < len(values[0]); i++ {
			sample, ok := stacktraceAggrValues[values[0][i].Int64()]
			if ok {
				sample.Value[0] += values[1][i].Int64()
				continue
			}
			stacktraceAggrValues[values[0][i].Int64()] = &profile.Sample{
				Value: []int64{values[1][i].Int64()},
			}
		}
	}
	hidden := hideStacktraces(stacktraceAggrValues, total, func(s *profile.Sample) int64 { return s.Value[0] }, opt.Hide)
	result, err := b.resolvePprofSymbols(ctx, stacktraceAggrValues)
	if err != nil {
		return nil, err
	}
	AddHiddenNodeToPprof(result, hidden)
	return result, nil
}

func (b *singleBlockQuerier) resolvePprofSymbols(ctx context.Context, stacktraceAggrByID map[int64]*profile.Sample) (*profile.Profile, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "ResolvePprofSymbols - Block")
	defer sp.Finish()

	// gather stacktraces
	stacktraceIDs := lo.Keys(stacktraceAggrByID)
	locationsIdsByStacktraceID := map[int64][]uint64{}

	sort.Slice(stacktraceIDs, func(i, j int) bool {
		return stacktraceIDs[i] < stacktraceIDs[j]
	})

	var (
		locationIDs = newUniqueIDs[struct{}]()
		stacktraces = repeatedColumnIter(ctx, b.stacktraces.file, "LocationIDs.list.element", iter.NewSliceIterator(stacktraceIDs))
	)

	for stacktraces.Next() {
		s := stacktraces.At()
		locationsIdsByStacktraceID[s.Row] = make([]uint64, len(s.Values))
		for i, locationID := range s.Values {
			locID := locationID.Uint64()
			locationIDs[int64(locID)] = struct{}{}
			locationsIdsByStacktraceID[s.Row][i] = locID
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
		locsId := locationsIdsByStacktraceID[id]
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

func (b *singleBlockQuerier) resolveSymbols(ctx context.Context, stacktraceAggrByID map[int64]*ingestv1.StacktraceSample) (*ingestv1.MergeProfilesStacktracesResult, error) {
	sp, ctx := opentracing.StartSpanFromContext(ctx, "ResolveSymbols - Block")
	defer sp.Finish()
	locationsByStacktraceID := map[int64][]uint64{}

	// gather stacktraces
	stacktraceIDs := lo.Keys(stacktraceAggrByID)
	sort.Slice(stacktraceIDs, func(i, j int) bool {
		return stacktraceIDs[i] < stacktraceIDs[j]
	})

	var (
		locationIDs = newUniqueIDs[struct{}]()
		stacktraces = repeatedColumnIter(ctx, b.stacktraces.file, "LocationIDs.list.element", iter.NewSliceIterator(stacktraceIDs))
	)

	for stacktraces.Next() {
		s := stacktraces.At()

		_, ok := locationsByStacktraceID[s.Row]
		if !ok {
			locationsByStacktraceID[s.Row] = make([]uint64, len(s.Values))
			for i, locationID := range s.Values {
				locID := locationID.Uint64()
				locationIDs[int64(locID)] = struct{}{}
				locationsByStacktraceID[s.Row][i] = locID
			}
			continue
		}
		for _, locationID := range s.Values {
			locID := locationID.Uint64()
			locationIDs[int64(locID)] = struct{}{}
			locationsByStacktraceID[s.Row] = append(locationsByStacktraceID[s.Row], locID)
		}
	}
	if err := stacktraces.Err(); err != nil {
		return nil, err
	}
	sp.LogFields(otlog.Int("stacktraces", len(stacktraceIDs)))
	// gather locations
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

	// gather functions
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
		locationIDs := locationsByStacktraceID[stacktraceID]

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

	it := repeatedColumnIter(ctx, b.profiles.file, "Samples.list.element.Value", rows)

	defer it.Close()

	labelsByFingerprint := map[model.Fingerprint]string{}
	seriesByLabels := map[string]*typesv1.Series{}
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
			if _, ok := seriesByLabels[labelsByString]; !ok {
				seriesByLabels[labelsByString] = &typesv1.Series{
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
		series := seriesByLabels[labelsByString]
		series.Points = append(series.Points, &typesv1.Point{
			Timestamp: int64(p.Timestamp()),
			Value:     float64(total),
		})
	}

	result := lo.Values(seriesByLabels)
	sort.Slice(result, func(i, j int) bool {
		return phlaremodel.CompareLabelPairs(result[i].Labels, result[j].Labels) < 0
	})
	// we have to sort the points in each series because labels reduction may have changed the order
	for _, s := range result {
		sort.Slice(s.Points, func(i, j int) bool {
			return s.Points[i].Timestamp < s.Points[j].Timestamp
		})
	}
	return result, nil
}
