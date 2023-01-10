package phlaredb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonv1 "github.com/grafana/phlare/pkg/gen/common/v1"
	profilev1 "github.com/grafana/phlare/pkg/gen/google/v1"
	ingestv1 "github.com/grafana/phlare/pkg/gen/ingester/v1"
	phlaremodel "github.com/grafana/phlare/pkg/model"
	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/grafana/phlare/pkg/pprof"
)

func newTestHead(t testing.TB) *testHead {
	dataPath := t.TempDir()
	reg := prometheus.NewPedanticRegistry()
	ctx := phlarecontext.WithRegistry(context.Background(), reg)
	head, err := NewHead(ctx, Config{DataPath: dataPath})
	require.NoError(t, err)
	return &testHead{Head: head, t: t, reg: reg}
}

type testHead struct {
	*Head
	t   testing.TB
	reg *prometheus.Registry
}

func (t *testHead) Flush(ctx context.Context) error {
	defer func() {
		t.t.Logf("flushing head of block %v", t.Head.meta.ULID)
	}()
	return t.Head.Flush(ctx)
}

func parseProfile(t testing.TB, path string) *profilev1.Profile {
	p, err := pprof.OpenFile(path)
	require.NoError(t, err, "failed opening profile: ", path)
	return p.Profile
}

var valueTypeStrings = []string{"unit", "type"}

func newValueType() *profilev1.ValueType {
	return &profilev1.ValueType{
		Unit: 1,
		Type: 2,
	}
}

func newProfileFoo() *profilev1.Profile {
	baseTable := append([]string{""}, valueTypeStrings...)
	baseTableLen := int64(len(baseTable)) + 0
	return &profilev1.Profile{
		Function: []*profilev1.Function{
			{
				Id:   1,
				Name: baseTableLen + 0,
			},
			{
				Id:   2,
				Name: baseTableLen + 1,
			},
		},
		Location: []*profilev1.Location{
			{
				Id:        1,
				MappingId: 1,
				Address:   0x1337,
			},
			{
				Id:        2,
				MappingId: 1,
				Address:   0x1338,
			},
		},
		Mapping: []*profilev1.Mapping{
			{Id: 1, Filename: baseTableLen + 2},
		},
		StringTable: append(baseTable, []string{
			"func_a",
			"func_b",
			"my-foo-binary",
		}...),
		TimeNanos:  123456,
		PeriodType: newValueType(),
		SampleType: []*profilev1.ValueType{newValueType()},
		Sample: []*profilev1.Sample{
			{
				Value:      []int64{0o123},
				LocationId: []uint64{1},
			},
			{
				Value:      []int64{1234},
				LocationId: []uint64{1, 2},
			},
		},
	}
}

func newProfileBar() *profilev1.Profile {
	baseTable := append([]string{""}, valueTypeStrings...)
	baseTableLen := int64(len(baseTable)) + 0
	return &profilev1.Profile{
		Function: []*profilev1.Function{
			{
				Id:   10,
				Name: baseTableLen + 1,
			},
			{
				Id:   21,
				Name: baseTableLen + 0,
			},
		},
		Location: []*profilev1.Location{
			{
				Id:        113,
				MappingId: 1,
				Address:   0x1337,
				Line: []*profilev1.Line{
					{FunctionId: 10, Line: 1},
				},
			},
		},
		Mapping: []*profilev1.Mapping{
			{Id: 1, Filename: baseTableLen + 2},
		},
		StringTable: append(baseTable, []string{
			"func_b",
			"func_a",
			"my-bar-binary",
		}...),
		TimeNanos:  123456,
		PeriodType: newValueType(),
		SampleType: []*profilev1.ValueType{newValueType()},
		Sample: []*profilev1.Sample{
			{
				Value:      []int64{2345},
				LocationId: []uint64{113},
			},
		},
	}
}

func newProfileBaz() *profilev1.Profile {
	return &profilev1.Profile{
		Function: []*profilev1.Function{
			{
				Id:   25,
				Name: 1,
			},
		},
		StringTable: []string{
			"",
			"func_c",
		},
	}
}

func TestHeadMetrics(t *testing.T) {
	head := newTestHead(t)
	require.NoError(t, head.Ingest(context.Background(), newProfileFoo(), uuid.New()))
	require.NoError(t, head.Ingest(context.Background(), newProfileBar(), uuid.New()))
	require.NoError(t, head.Ingest(context.Background(), newProfileBaz(), uuid.New()))
	time.Sleep(time.Second)
	require.NoError(t, testutil.GatherAndCompare(head.reg,
		strings.NewReader(`
# HELP phlare_head_ingested_sample_values_total Number of sample values ingested into the head per profile type.
# TYPE phlare_head_ingested_sample_values_total counter
phlare_head_ingested_sample_values_total{profile_name=""} 3
# HELP phlare_head_profiles_created_total Total number of profiles created in the head
# TYPE phlare_head_profiles_created_total counter
phlare_head_profiles_created_total{profile_name=""} 2
# HELP phlare_head_received_sample_values_total Number of sample values received into the head per profile type.
# TYPE phlare_head_received_sample_values_total counter
phlare_head_received_sample_values_total{profile_name=""} 3

# HELP phlare_head_size_bytes Size of a particular in memory store within the head phlaredb block.
# TYPE phlare_head_size_bytes gauge
phlare_head_size_bytes{type="functions"} 240
phlare_head_size_bytes{type="locations"} 344
phlare_head_size_bytes{type="mappings"} 192
phlare_head_size_bytes{type="profiles"} 416
phlare_head_size_bytes{type="stacktraces"} 104
phlare_head_size_bytes{type="strings"} 52

`),
		"phlare_head_received_sample_values_total",
		"phlare_head_profiles_created_total",
		"phlare_head_ingested_sample_values_total",
		"phlare_head_size_bytes",
	))
}

func TestHeadIngestFunctions(t *testing.T) {
	head := newTestHead(t)

	require.NoError(t, head.Ingest(context.Background(), newProfileFoo(), uuid.New()))
	require.NoError(t, head.Ingest(context.Background(), newProfileBar(), uuid.New()))
	require.NoError(t, head.Ingest(context.Background(), newProfileBaz(), uuid.New()))

	require.Equal(t, int64(3), head.functions.buffer.NumRows())
	helper := &functionsHelper{}
	assert.Equal(t, functionsKey{Name: 3}, helper.key(head.functions.GetRowNum(0)))
	assert.Equal(t, functionsKey{Name: 4}, helper.key(head.functions.GetRowNum(1)))
	assert.Equal(t, functionsKey{Name: 7}, helper.key(head.functions.GetRowNum(2)))
}

func TestHeadIngestStrings(t *testing.T) {
	ctx := context.Background()
	head := newTestHead(t)

	r := &rewriter{}
	require.NoError(t, head.strings.ingest(ctx, schemav1.StringsFromStringSlice(newProfileFoo().StringTable), r))
	require.Equal(t, schemav1.StringsFromStringSlice([]string{"", "unit", "type", "func_a", "func_b", "my-foo-binary"}), head.strings.Slice())
	require.Equal(t, schemav1.StringsFromStringSlice([]string{"", "unit", "type", "func_a", "func_b", "my-foo-binary"}), head.strings.Slice())
	require.Equal(t, stringConversionTable{0, 1, 2, 3, 4, 5}, r.strings)

	r = &rewriter{}
	require.NoError(t, head.strings.ingest(ctx, schemav1.StringsFromStringSlice(newProfileBar().StringTable), r))
	require.Equal(t, schemav1.StringsFromStringSlice([]string{"", "unit", "type", "func_a", "func_b", "my-foo-binary", "my-bar-binary"}), head.strings.Slice())
	require.Equal(t, stringConversionTable{0, 1, 2, 4, 3, 6}, r.strings)

	r = &rewriter{}
	require.NoError(t, head.strings.ingest(ctx, schemav1.StringsFromStringSlice(newProfileBaz().StringTable), r))
	require.Equal(t, schemav1.StringsFromStringSlice([]string{"", "unit", "type", "func_a", "func_b", "my-foo-binary", "my-bar-binary", "func_c"}), head.strings.Slice())
	require.Equal(t, stringConversionTable{0, 7}, r.strings)
}

func TestHeadIngestStacktraces(t *testing.T) {
	ctx := context.Background()
	head := newTestHead(t)

	require.NoError(t, head.Ingest(ctx, newProfileFoo(), uuid.New()))
	require.NoError(t, head.Ingest(ctx, newProfileBar(), uuid.New()))
	require.NoError(t, head.Ingest(ctx, newProfileBar(), uuid.New()))

	// expect 2 mappings
	require.Equal(t, uint64(2), head.mappings.NumRows())
	assert.Equal(t, &schemav1.String{String: "my-foo-binary"}, head.strings.GetRowNum(uint64(head.mappings.GetRowNum(0).Filename)))
	assert.Equal(t, &schemav1.String{String: "my-bar-binary"}, head.strings.GetRowNum(uint64(head.mappings.GetRowNum(1).Filename)))

	// expect 3 stacktraces
	require.Equal(t, uint64(3), head.stacktraces.NumRows())

	// expect 3 profiles
	require.Equal(t, uint64(3), head.profiles.NumRows())

	var samples []uint64
	for _, profile := range head.profiles.Slice() {
		for _, sample := range profile.Samples {
			samples = append(samples, sample.StacktraceID)
		}
	}
	// expect 4 samples, 3 of which distinct
	require.Equal(t, []uint64{0, 1, 2, 2}, samples)
}

func TestHeadLabelValues(t *testing.T) {
	head := newTestHead(t)
	require.NoError(t, head.Ingest(context.Background(), newProfileFoo(), uuid.New(), &commonv1.LabelPair{Name: "job", Value: "foo"}, &commonv1.LabelPair{Name: "namespace", Value: "phlare"}))
	require.NoError(t, head.Ingest(context.Background(), newProfileBar(), uuid.New(), &commonv1.LabelPair{Name: "job", Value: "bar"}, &commonv1.LabelPair{Name: "namespace", Value: "phlare"}))

	res, err := head.LabelValues(context.Background(), connect.NewRequest(&ingestv1.LabelValuesRequest{Name: "cluster"}))
	require.NoError(t, err)
	require.Equal(t, []string{}, res.Msg.Names)

	res, err = head.LabelValues(context.Background(), connect.NewRequest(&ingestv1.LabelValuesRequest{Name: "job"}))
	require.NoError(t, err)
	require.Equal(t, []string{"bar", "foo"}, res.Msg.Names)
}

func TestHeadLabelNames(t *testing.T) {
	head := newTestHead(t)
	require.NoError(t, head.Ingest(context.Background(), newProfileFoo(), uuid.New(), &commonv1.LabelPair{Name: "job", Value: "foo"}, &commonv1.LabelPair{Name: "namespace", Value: "phlare"}))
	require.NoError(t, head.Ingest(context.Background(), newProfileBar(), uuid.New(), &commonv1.LabelPair{Name: "job", Value: "bar"}, &commonv1.LabelPair{Name: "namespace", Value: "phlare"}))

	res, err := head.LabelNames(context.Background(), connect.NewRequest(&ingestv1.LabelNamesRequest{}))
	require.NoError(t, err)
	require.Equal(t, []string{"__period_type__", "__period_unit__", "__profile_type__", "__type__", "__unit__", "job", "namespace"}, res.Msg.Names)
}

func TestHeadSeries(t *testing.T) {
	head := newTestHead(t)
	fooLabels := phlaremodel.NewLabelsBuilder(nil).Set("namespace", "phlare").Set("job", "foo").Labels()
	barLabels := phlaremodel.NewLabelsBuilder(nil).Set("namespace", "phlare").Set("job", "bar").Labels()
	require.NoError(t, head.Ingest(context.Background(), newProfileFoo(), uuid.New(), fooLabels...))
	require.NoError(t, head.Ingest(context.Background(), newProfileBar(), uuid.New(), barLabels...))

	expected := phlaremodel.NewLabelsBuilder(nil).
		Set("namespace", "phlare").
		Set("job", "foo").
		Set("__period_type__", "type").
		Set("__period_unit__", "unit").
		Set("__type__", "type").
		Set("__unit__", "unit").
		Set("__profile_type__", ":type:unit:type:unit").
		Labels()
	res, err := head.Series(context.Background(), connect.NewRequest(&ingestv1.SeriesRequest{Matchers: []string{`{job="foo"}`}}))
	require.NoError(t, err)
	require.Equal(t, []*commonv1.Labels{{Labels: expected}}, res.Msg.LabelsSet)
}

func TestHeadProfileTypes(t *testing.T) {
	head := newTestHead(t)
	require.NoError(t, head.Ingest(context.Background(), newProfileFoo(), uuid.New(), &commonv1.LabelPair{Name: "__name__", Value: "foo"}, &commonv1.LabelPair{Name: "job", Value: "foo"}, &commonv1.LabelPair{Name: "namespace", Value: "phlare"}))
	require.NoError(t, head.Ingest(context.Background(), newProfileBar(), uuid.New(), &commonv1.LabelPair{Name: "__name__", Value: "bar"}, &commonv1.LabelPair{Name: "namespace", Value: "phlare"}))

	res, err := head.ProfileTypes(context.Background(), connect.NewRequest(&ingestv1.ProfileTypesRequest{}))
	require.NoError(t, err)
	require.Equal(t, []*commonv1.ProfileType{
		mustParseProfileSelector(t, "bar:type:unit:type:unit"),
		mustParseProfileSelector(t, "foo:type:unit:type:unit"),
	}, res.Msg.ProfileTypes)
}

func mustParseProfileSelector(t testing.TB, selector string) *commonv1.ProfileType {
	ps, err := phlaremodel.ParseProfileTypeSelector(selector)
	require.NoError(t, err)
	return ps
}

func TestHeadIngestRealProfiles(t *testing.T) {
	profilePaths := []string{
		"testdata/heap",
		"testdata/profile",
		"testdata/profile_uncompressed",
		"testdata/profile_python",
		"testdata/profile_java",
	}

	head := newTestHead(t)
	ctx := context.Background()

	for pos := range profilePaths {
		path := profilePaths[pos]
		t.Run(path, func(t *testing.T) {
			profile := parseProfile(t, profilePaths[pos])
			require.NoError(t, head.Ingest(ctx, profile, uuid.New()))
		})
	}

	require.NoError(t, head.Flush(ctx))
	t.Logf("strings=%d samples=%d", head.strings.NumRows(), head.totalSamples.Load())
}

func BenchmarkHeadIngestProfiles(t *testing.B) {
	var (
		profilePaths = []string{
			"testdata/heap",
			"testdata/profile",
		}
		profileCount = 0
	)

	head := newTestHead(t)
	ctx := context.Background()

	t.ReportAllocs()

	for n := 0; n < t.N; n++ {
		for pos := range profilePaths {
			p := parseProfile(t, profilePaths[pos])
			require.NoError(t, head.Ingest(ctx, p, uuid.New()))
			profileCount++
		}
	}
}
