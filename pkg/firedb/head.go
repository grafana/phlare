package firedb

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/oklog/ulid"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"google.golang.org/grpc/codes"

	schemav1 "github.com/grafana/fire/pkg/firedb/schemas/v1"
	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	profilev1 "github.com/grafana/fire/pkg/gen/google/v1"
	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	firemodel "github.com/grafana/fire/pkg/model"
)

var ulidEntropy = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateULID() ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(time.Now()), ulidEntropy)
}

func copySlice[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)
	return out
}

type idConversionTable map[int64]int64

func (t idConversionTable) rewrite(idx *int64) {
	pos := *idx
	var ok bool
	*idx, ok = t[pos]
	if !ok {
		panic(fmt.Sprintf("unable to rewrite index %d", pos))
	}
}

func (t idConversionTable) rewriteUint64(idx *uint64) {
	pos := *idx
	v, ok := t[int64(pos)]
	if !ok {
		panic(fmt.Sprintf("unable to rewrite index %d", pos))
	}
	*idx = uint64(v)
}

type Models interface {
	*schemav1.Profile | *schemav1.Stacktrace | *profilev1.Location | *profilev1.Mapping | *profilev1.Function | string
}

// rewriter contains slices to rewrite the per profile reference into per head references.
type rewriter struct {
	strings     stringConversionTable
	functions   idConversionTable
	mappings    idConversionTable
	locations   idConversionTable
	stacktraces idConversionTable
}

type Helper[M Models, K comparable] interface {
	key(M) K
	addToRewriter(*rewriter, idConversionTable)
	rewrite(*rewriter, M) error
	// some Models contain their own IDs within the struct, this allows to set them and keep track of the preexisting ID. It should return the oldID that is supposed to be rewritten.
	setID(existingSliceID uint64, newID uint64, element M) uint64

	// size returns a (rough estimation) of the size of a single element M
	size(M) uint64

	// clone copies parts that are not optimally sized from protobuf parsing
	clone(M) M
}

type Table interface {
	Name() string
	Size() uint64
	Init(path string) error
	Flush() (int, error)
	Close() error
}

func HeadWithRegistry(reg prometheus.Registerer) HeadOption {
	return func(h *Head) {
		h.metrics = newHeadMetrics(reg).setHead(h)
	}
}

func headWithMetrics(m *headMetrics) HeadOption {
	return func(h *Head) {
		h.metrics = m.setHead(h)
	}
}

func HeadWithLogger(l log.Logger) HeadOption {
	return func(h *Head) {
		h.logger = l
	}
}

type HeadOption func(*Head)

type Head struct {
	logger    log.Logger
	metrics   *headMetrics
	ulid      ulid.ULID
	blockPath string

	index           *profilesIndex
	strings         deduplicatingSlice[string, string, *stringsHelper, *schemav1.StringPersister]
	mappings        deduplicatingSlice[*profilev1.Mapping, mappingsKey, *mappingsHelper, *schemav1.MappingPersister]
	functions       deduplicatingSlice[*profilev1.Function, functionsKey, *functionsHelper, *schemav1.FunctionPersister]
	locations       deduplicatingSlice[*profilev1.Location, locationsKey, *locationsHelper, *schemav1.LocationPersister]
	stacktraces     deduplicatingSlice[*schemav1.Stacktrace, stacktracesKey, *stacktracesHelper, *schemav1.StacktracePersister] // a stacktrace is a slice of location ids
	profiles        deduplicatingSlice[*schemav1.Profile, profilesKey, *profilesHelper, *schemav1.ProfilePersister]
	tables          []Table
	pprofLabelCache labelCache
}

func NewHead(dataPath string, opts ...HeadOption) (*Head, error) {
	h := &Head{
		ulid: generateULID(),
	}
	h.blockPath = filepath.Join(dataPath, "head", h.ulid.String())

	// execute options
	for _, o := range opts {
		o(h)
	}

	// setup fall backs
	if h.logger == nil {
		h.logger = log.NewNopLogger()
	}
	if h.metrics == nil {
		h.metrics = newHeadMetrics(nil).setHead(h)
	}

	if err := os.MkdirAll(h.blockPath, 0o755); err != nil {
		return nil, err
	}

	h.tables = []Table{
		&h.strings,
		&h.mappings,
		&h.functions,
		&h.locations,
		&h.stacktraces,
		&h.profiles,
	}
	for _, t := range h.tables {
		if err := t.Init(h.blockPath); err != nil {
			return nil, err
		}
	}

	index, err := newProfileIndex(32, h.metrics)
	if err != nil {
		return nil, err
	}
	h.index = index

	h.pprofLabelCache.init()
	return h, nil
}

func (h *Head) convertSamples(ctx context.Context, r *rewriter, in []*profilev1.Sample) ([]*schemav1.Sample, error) {
	var (
		out         = make([]*schemav1.Sample, len(in))
		stacktraces = make([]*schemav1.Stacktrace, len(in))
	)

	for pos := range in {
		// populate samples
		out[pos] = &schemav1.Sample{
			Values: copySlice(in[pos].Value),
			Labels: h.pprofLabelCache.rewriteLabels(r.strings, in[pos].Label),
		}

		// build full stack traces
		stacktraces[pos] = &schemav1.Stacktrace{
			// no copySlice necessary at this point,stacktracesHelper.clone
			// will copy it, if it is required to be retained.
			LocationIDs: in[pos].LocationId,
		}
	}

	// ingest stacktraces
	if err := h.stacktraces.ingest(ctx, stacktraces, r); err != nil {
		return nil, err
	}

	// reference stacktraces
	for pos := range out {
		out[pos].StacktraceID = uint64(r.stacktraces[int64(pos)])
	}

	return out, nil
}

func (h *Head) Ingest(ctx context.Context, p *profilev1.Profile, id uuid.UUID, externalLabels ...*commonv1.LabelPair) error {
	// build label set per sample type before references are rewritten
	var (
		sb                                             strings.Builder
		lbls                                           = firemodel.NewLabelsBuilder(externalLabels)
		sampleType, sampleUnit, periodType, periodUnit string
		metricName                                     = firemodel.Labels(externalLabels).Get(model.MetricNameLabel)
	)

	// set common labels
	if p.PeriodType != nil {
		periodType = p.StringTable[p.PeriodType.Type]
		lbls.Set(firemodel.LabelNamePeriodType, periodType)
		periodUnit = p.StringTable[p.PeriodType.Unit]
		lbls.Set(firemodel.LabelNamePeriodUnit, periodUnit)
	}

	seriesRefs := make([]model.Fingerprint, len(p.SampleType))
	profilesLabels := make([]firemodel.Labels, len(p.SampleType))
	for pos := range p.SampleType {
		sampleType = p.StringTable[p.SampleType[pos].Type]
		lbls.Set(firemodel.LabelNameType, sampleType)
		sampleUnit = p.StringTable[p.SampleType[pos].Unit]
		lbls.Set(firemodel.LabelNameUnit, sampleUnit)

		sb.Reset()
		_, _ = sb.WriteString(metricName)
		_, _ = sb.WriteRune(':')
		_, _ = sb.WriteString(sampleType)
		_, _ = sb.WriteRune(':')
		_, _ = sb.WriteString(sampleUnit)
		_, _ = sb.WriteRune(':')
		_, _ = sb.WriteString(periodType)
		_, _ = sb.WriteRune(':')
		_, _ = sb.WriteString(periodUnit)
		t := sb.String()
		lbls.Set(firemodel.LabelNameProfileType, t)
		lbs := lbls.Labels().Clone()
		profilesLabels[pos] = lbs
		seriesRefs[pos] = model.Fingerprint(lbs.Hash())
	}

	// create a rewriter state
	rewrites := &rewriter{}

	if err := h.strings.ingest(ctx, p.StringTable, rewrites); err != nil {
		return err
	}

	if err := h.mappings.ingest(ctx, p.Mapping, rewrites); err != nil {
		return err
	}

	if err := h.functions.ingest(ctx, p.Function, rewrites); err != nil {
		return err
	}

	if err := h.locations.ingest(ctx, p.Location, rewrites); err != nil {
		return err
	}

	samples, err := h.convertSamples(ctx, rewrites, p.Sample)
	if err != nil {
		return err
	}

	profile := &schemav1.Profile{
		ID:                id,
		SeriesRefs:        seriesRefs,
		Samples:           samples,
		DropFrames:        p.DropFrames,
		KeepFrames:        p.KeepFrames,
		TimeNanos:         p.TimeNanos,
		DurationNanos:     p.DurationNanos,
		Comments:          copySlice(p.Comment),
		DefaultSampleType: p.DefaultSampleType,
	}

	if err := h.profiles.ingest(ctx, []*schemav1.Profile{profile}, rewrites); err != nil {
		return err
	}
	h.index.Add(profile, profilesLabels, metricName)

	h.metrics.sampleValuesIngested.WithLabelValues(metricName).Add(float64(len(p.Sample) * len(profilesLabels)))

	return nil
}

// LabelValues returns the possible label values for a given label name.
func (h *Head) LabelValues(ctx context.Context, req *connect.Request[ingestv1.LabelValuesRequest]) (*connect.Response[ingestv1.LabelValuesResponse], error) {
	values, err := h.index.ix.LabelValues(req.Msg.Name, nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.LabelValuesResponse{
		Names: values,
	}), nil
}

// ProfileTypes returns the possible profile types.
func (h *Head) ProfileTypes(ctx context.Context, req *connect.Request[ingestv1.ProfileTypesRequest]) (*connect.Response[ingestv1.ProfileTypesResponse], error) {
	values, err := h.index.ix.LabelValues(firemodel.LabelNameProfileType, nil)
	if err != nil {
		return nil, err
	}
	sort.Strings(values)

	profileTypes := make([]*commonv1.ProfileType, len(values))
	for i, v := range values {
		tp, err := firemodel.ParseProfileTypeSelector(v)
		if err != nil {
			return nil, err
		}
		profileTypes[i] = tp
	}

	return connect.NewResponse(&ingestv1.ProfileTypesResponse{
		ProfileTypes: profileTypes,
	}), nil
}

func (h *Head) SelectProfiles(ctx context.Context, req *ingestv1.SelectProfilesRequest, callback func(ts int64, lbs firemodel.Labels, _ model.Fingerprint, idx int, profile *schemav1.Profile) error) error {
	var (
		totalSamples   int64
		totalLocations int64
		totalProfiles  int64
	)
	// nolint:ineffassign
	// we might use ctx later.
	sp, ctx := opentracing.StartSpanFromContext(ctx, "Head - SelectProfiles")
	defer func() {
		sp.LogFields(
			otlog.Int64("total_samples", totalSamples),
			otlog.Int64("total_locations", totalLocations),
			otlog.Int64("total_profiles", totalProfiles),
		)
		sp.Finish()
	}()

	selectors, err := parser.ParseMetricSelector(req.LabelSelector)
	if err != nil {
		return status.Error(codes.InvalidArgument, "failed to label selector")
	}
	selectors = append(selectors, &labels.Matcher{
		Type:  labels.MatchEqual,
		Name:  firemodel.LabelNameProfileType,
		Value: req.Type.Name + ":" + req.Type.SampleType + ":" + req.Type.SampleUnit + ":" + req.Type.PeriodType + ":" + req.Type.PeriodUnit,
	})

	return h.index.forMatchingProfiles(selectors, func(lbs firemodel.Labels, fp model.Fingerprint, idx int, profile *schemav1.Profile) error {
		ts := int64(model.TimeFromUnixNano(profile.TimeNanos))
		// if the timestamp is not matching we skip this profile.
		if req.Start > ts || ts > req.End {
			return nil
		}
		totalProfiles++
		callback(ts, lbs, fp, idx, profile)
		return nil
	})
}

func (h *Head) Flush(ctx context.Context) error {
	if err := h.index.WriteTo(ctx, filepath.Join(h.blockPath, "index.tsdb")); err != nil {
		return errors.Wrap(err, "flushing of index")
	}

	for _, t := range h.tables {
		rowCount, err := t.Flush()
		if err != nil {
			return errors.Wrapf(err, "flushing of table %s", t.Name())
		}
		h.metrics.rowsWritten.WithLabelValues(t.Name()).Add(float64(rowCount))
	}

	for _, t := range h.tables {
		if err := t.Close(); err != nil {
			return errors.Wrapf(err, "closing of table %s", t.Name())
		}
	}
	level.Info(h.logger).Log("msg", "head successfully written to block", "block_path", h.blockPath)

	return nil
}
