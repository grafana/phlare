package ingester

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/grafana/dskit/kv"
	"github.com/grafana/dskit/ring"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	schemav1 "github.com/grafana/fire/pkg/firedb/schemas/v1"
	commonv1 "github.com/grafana/fire/pkg/gen/common/v1"
	profilev1 "github.com/grafana/fire/pkg/gen/google/v1"
	ingesterv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	"github.com/grafana/fire/pkg/gen/ingester/v1/ingesterv1connect"
	"github.com/grafana/fire/pkg/iterator"
	firemodel "github.com/grafana/fire/pkg/model"
	"github.com/grafana/fire/pkg/testhelper"
)

var _ DB = (*mockDB)(nil)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockDB) Ingest(ctx context.Context, p *profilev1.Profile, id uuid.UUID, externalLabels ...*commonv1.LabelPair) error {
	args := m.Called(ctx, p, id, externalLabels)
	return args.Error(0)
}

func (m *mockDB) SelectProfiles(ctx context.Context, req *ingestv1.SelectProfilesRequest) (iterator.Interface[firemodel.Profile], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(iterator.Interface[firemodel.Profile]), args.Error(1)
}

func (m *mockDB) ProfileTypes(ctx context.Context, req *connect.Request[ingestv1.ProfileTypesRequest]) (*connect.Response[ingestv1.ProfileTypesResponse], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*connect.Response[ingestv1.ProfileTypesResponse]), args.Error(1)
}

func (m *mockDB) LabelValues(ctx context.Context, req *connect.Request[ingestv1.LabelValuesRequest]) (*connect.Response[ingestv1.LabelValuesResponse], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*connect.Response[ingestv1.LabelValuesResponse]), args.Error(1)
}

func (m *mockDB) Series(ctx context.Context, req *connect.Request[ingestv1.SeriesRequest]) (*connect.Response[ingestv1.SeriesResponse], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*connect.Response[ingestv1.SeriesResponse]), args.Error(1)
}

func TestSelectProfiles(t *testing.T) {
	fooLabels := firemodel.NewLabelsBuilder().Set("label", "foo").Labels()
	barLabels := firemodel.NewLabelsBuilder().Set("label", "bar").Labels()
	for _, tt := range []struct {
		name      string
		batchSize int
		in        []firemodel.Profile
		expected  []*ingesterv1.SelectProfilesResponse
	}{
		{
			"empty",
			10,
			[]firemodel.Profile{},
			[]*ingesterv1.SelectProfilesResponse{},
		},
		{
			"batch exactly full",
			2,
			[]firemodel.Profile{
				{Labels: fooLabels, Profile: &schemav1.Profile{TimeNanos: 1, ID: uuid.UUID{1}}, Fingerprint: model.Fingerprint(fooLabels.Hash())},
				{Labels: barLabels, Profile: &schemav1.Profile{TimeNanos: 2, ID: uuid.UUID{2}}, Fingerprint: model.Fingerprint(barLabels.Hash())},
			},
			[]*ingesterv1.SelectProfilesResponse{
				{
					Profiles:  []*ingestv1.Profile{{ID: "01000000-0000-0000-0000-000000000000", Timestamp: 1, LabelsetIndex: 0}, {ID: "02000000-0000-0000-0000-000000000000", Timestamp: 2, LabelsetIndex: 1}},
					Labelsets: []*ingestv1.Labels{{Labels: fooLabels}, {Labels: barLabels}},
				},
			},
		},
		{
			"batch not full",
			3,
			[]firemodel.Profile{
				{Labels: fooLabels, Profile: &schemav1.Profile{TimeNanos: 1, ID: uuid.UUID{1}}, Fingerprint: model.Fingerprint(fooLabels.Hash())},
				{Labels: barLabels, Profile: &schemav1.Profile{TimeNanos: 2, ID: uuid.UUID{2}}, Fingerprint: model.Fingerprint(barLabels.Hash())},
			},
			[]*ingesterv1.SelectProfilesResponse{
				{
					Profiles:  []*ingestv1.Profile{{ID: "01000000-0000-0000-0000-000000000000", Timestamp: 1, LabelsetIndex: 0}, {ID: "02000000-0000-0000-0000-000000000000", Timestamp: 2, LabelsetIndex: 1}},
					Labelsets: []*ingestv1.Labels{{Labels: fooLabels}, {Labels: barLabels}},
				},
			},
		},
		{
			"mutiple batches",
			1,
			[]firemodel.Profile{
				{Labels: fooLabels, Profile: &schemav1.Profile{TimeNanos: 1, ID: uuid.UUID{1}}, Fingerprint: model.Fingerprint(fooLabels.Hash())},
				{Labels: barLabels, Profile: &schemav1.Profile{TimeNanos: 2, ID: uuid.UUID{2}}, Fingerprint: model.Fingerprint(barLabels.Hash())},
			},
			[]*ingesterv1.SelectProfilesResponse{
				{
					Profiles:  []*ingestv1.Profile{{ID: "01000000-0000-0000-0000-000000000000", Timestamp: 1, LabelsetIndex: 0}},
					Labelsets: []*ingestv1.Labels{{Labels: fooLabels}},
				},
				{
					Profiles:  []*ingestv1.Profile{{ID: "02000000-0000-0000-0000-000000000000", Timestamp: 2, LabelsetIndex: 0}},
					Labelsets: []*ingestv1.Labels{{Labels: barLabels}},
				},
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			db := &mockDB{}
			selectProfilesBatchSize = tt.batchSize

			db.On("SelectProfiles", mock.Anything, mock.Anything).Return(iterator.NewSliceIterator(tt.in), nil)
			ing, err := New(Config{
				LifecyclerConfig: ring.LifecyclerConfig{Addr: "foo", RingConfig: ring.Config{
					KVStore: kv.Config{Store: "inmemory"},
				}},
			}, log.NewNopLogger(), nil, db)
			require.NoError(t, err)

			mux := http.NewServeMux()
			mux.Handle(
				ingesterv1connect.NewIngesterServiceHandler(ing),
			)
			server := httptest.NewUnstartedServer(mux)
			server.EnableHTTP2 = true
			server.StartTLS()
			defer server.Close()

			httpClient := server.Client()
			client := ingesterv1connect.NewIngesterServiceClient(httpClient, server.URL, connect.WithGRPC())

			stream, err := client.SelectProfiles(context.Background(), connect.NewRequest(&ingesterv1.SelectProfilesRequest{}))
			require.NoError(t, err)

			responses := []*ingesterv1.SelectProfilesResponse{}

			for stream.Receive() {
				responses = append(responses, testhelper.CloneProto(t, stream.Msg()))
			}
			testhelper.EqualProto(t, tt.expected, responses)
		})
	}
}

// func Test_selectMerge(t *testing.T) {
// 	cfg := defaultIngesterTestConfig(t)
// 	profileStore, err := profilestore.New(log.NewNopLogger(), nil, trace.NewNoopTracerProvider(), defaultProfileStoreTestConfig(t))
// 	require.NoError(t, err)

// 	d, err := New(cfg, log.NewNopLogger(), nil, profileStore)
// 	require.NoError(t, err)
// 	resp, err := d.Push(context.Background(), connect.NewRequest(&pushv1.PushRequest{
// 		Series: []*pushv1.RawProfileSeries{
// 			{
// 				Labels: []*commonv1.LabelPair{
// 					{Name: "__name__", Value: "memory"},
// 				},
// 				Samples: []*pushv1.RawSample{
// 					{
// 						RawProfile: generateProfile(
// 							t, "inuse_space", "bytes", "space", "bytes", time.Now().Add(-1*time.Minute),
// 							[]int64{1, 1},
// 							[][]string{
// 								{"bar", "foo"},
// 								{"buzz", "foo"},
// 							},
// 						),
// 					},
// 				},
// 			},
// 		},
// 	}))

// 	require.NoError(t, err)
// 	require.NotNil(t, resp)
// 	f, err := d.selectMerge(context.Background(), profileQuery{
// 		name:       "memory",
// 		sampleType: "inuse_space",
// 		sampleUnit: "bytes",
// 		periodType: "space",
// 		periodUnit: "bytes",
// 	}, 0, int64(model.Latest))
// 	require.NoError(t, err)

// 	// aggregate plan have no guarantee of order so we sort the results
// 	sort.Strings(f.Flamebearer.Names)

// 	require.Equal(t, []string{"bar", "buzz", "foo", "total"}, f.Flamebearer.Names)
// 	require.Equal(t, flamebearer.FlamebearerMetadataV1{
// 		Format:     "single",
// 		Units:      "bytes",
// 		Name:       "inuse_space",
// 		SampleRate: 100,
// 	}, f.Metadata)
// 	require.Equal(t, 2, f.Flamebearer.NumTicks)
// 	require.Equal(t, 1, f.Flamebearer.MaxSelf)
// 	require.Equal(t, []int{0, 2, 0, 0}, f.Flamebearer.Levels[0])
// 	require.Equal(t, []int{0, 2, 0, 1}, f.Flamebearer.Levels[1])
// 	require.Equal(t, []int{0, 1, 1}, f.Flamebearer.Levels[2][:3])
// 	require.Equal(t, []int{0, 1, 1}, f.Flamebearer.Levels[2][4:7])
// 	require.True(t, f.Flamebearer.Levels[2][3] == 3 || f.Flamebearer.Levels[2][3] == 2)
// 	require.True(t, f.Flamebearer.Levels[2][7] == 3 || f.Flamebearer.Levels[2][7] == 2)
// 	require.NoError(
// 		t,
// 		profileStore.Close(),
// 	)
// }

/*
func Test_QueryMetadata(t *testing.T) {
	cfg := defaultIngesterTestConfig(t)
	logger := log.NewLogfmtLogger(os.Stdout)

	profileStore, err := profilestore.New(logger, nil, trace.NewNoopTracerProvider(), defaultProfileStoreTestConfig(t))
	require.NoError(t, err)

	d, err := New(cfg, log.NewLogfmtLogger(os.Stdout), nil, profileStore)
	require.NoError(t, err)

	rawProfile := testProfile(t)
	resp, err := d.Push(context.Background(), connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{
				Labels: []*commonv1.LabelPair{
					{Name: "__name__", Value: "memory"},
					{Name: "cluster", Value: "us-central1"},
				},
				Samples: []*pushv1.RawSample{
					{
						RawProfile: rawProfile,
					},
				},
			},
			{
				Labels: []*commonv1.LabelPair{
					{Name: "__name__", Value: "memory"},
					{Name: "cluster", Value: "us-east1"},
				},
				Samples: []*pushv1.RawSample{
					{
						RawProfile: rawProfile,
					},
				},
			},
		},
	}))
	require.NoError(t, err)
	require.NotNil(t, resp)

	clusterRes, err := d.LabelValues(context.Background(), connect.NewRequest(&ingestv1.LabelValuesRequest{Name: "cluster"}))
	require.NoError(t, err)
	require.Equal(t, []string{"us-central1", "us-east1"}, clusterRes.Msg.Names)
	typeRes, err := d.ProfileTypes(context.Background(), connect.NewRequest(&ingestv1.ProfileTypesRequest{}))
	require.NoError(t, err)
	expectedTypes := []string{
		"memory:inuse_space:bytes:space:bytes",
		"memory:inuse_objects:count:space:bytes",
		"memory:alloc_space:bytes:space:bytes",
		"memory:alloc_objects:count:space:bytes",
	}
	sort.Strings(expectedTypes)
	sort.Strings(typeRes.Msg.Names)
	require.Equal(t, expectedTypes, typeRes.Msg.Names)
}
*/

/*
func Test_selectProfiles(t *testing.T) {
	cfg := defaultIngesterTestConfig(t)
	logger := log.NewLogfmtLogger(os.Stdout)
	storeCfg := defaultProfileStoreTestConfig(t)
	profileStore, err := profilestore.New(logger, nil, trace.NewNoopTracerProvider(), storeCfg)
	require.NoError(t, err)

	d, err := New(cfg, log.NewLogfmtLogger(os.Stdout), nil, profileStore)
	require.NoError(t, err)

	resp, err := d.Push(context.Background(), connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{
				Labels: []*commonv1.LabelPair{
					{Name: "__name__", Value: "memory"},
					{Name: "cluster", Value: "us-central1"},
					{Name: "foo", Value: "bar"},
				},
				Samples: []*pushv1.RawSample{
					{
						RawProfile: generateProfile(
							t, "inuse_space", "bytes", "space", "bytes", time.Unix(1, 0),
							[]int64{1, 2},
							[][]string{
								{"foo", "bar", "buzz"},
								{"buzz", "baz", "foo"},
							},
						),
					},
				},
			},
			{
				Labels: []*commonv1.LabelPair{
					{Name: "__name__", Value: "memory"},
					{Name: "cluster", Value: "us-east1"},
				},
				Samples: []*pushv1.RawSample{
					{
						RawProfile: generateProfile(
							t, "inuse_space", "bytes", "space", "bytes", time.Unix(2, 0),
							[]int64{4, 5, 6},
							[][]string{
								{"foo", "bar", "buzz"},
								{"buzz", "baz", "foo"},
								{"1", "2", "3"},
							},
						),
					},
				},
			},
		},
	}))
	require.NoError(t, err)
	require.NotNil(t, resp)

	res, err := d.SelectProfiles(context.Background(), connect.NewRequest(&ingestv1.SelectProfilesRequest{
		LabelSelector: `{cluster=~".*"}`,
		Type: &ingestv1.ProfileType{
			Name:       "memory",
			SampleType: "inuse_space",
			SampleUnit: "bytes",
			PeriodType: "space",
			PeriodUnit: "bytes",
		},
		Start: 0,
		End:   int64(model.Latest),
	}))
	require.NoError(t, err)
	sort.Slice(res.Msg.Profiles, func(i, j int) bool {
		return res.Msg.Profiles[i].Timestamp < res.Msg.Profiles[j].Timestamp
	})
	require.Equal(t, 2, len(res.Msg.Profiles))
	require.Equal(t, 2, len(res.Msg.Profiles[0].Labels))
	require.Equal(t, 1, len(res.Msg.Profiles[1].Labels))

	require.Equal(t, "cluster", res.Msg.Profiles[0].Labels[0].Name)
	require.Equal(t, "us-central1", res.Msg.Profiles[0].Labels[0].Value)
	require.Equal(t, "foo", res.Msg.Profiles[0].Labels[1].Name)
	require.Equal(t, "bar", res.Msg.Profiles[0].Labels[1].Value)
	require.Equal(t, "cluster", res.Msg.Profiles[1].Labels[0].Name)
	require.Equal(t, "us-east1", res.Msg.Profiles[1].Labels[0].Value)

	require.Equal(t, 2, len(res.Msg.Profiles[0].Stacktraces))
	require.Equal(t, 3, len(res.Msg.Profiles[1].Stacktraces))

	stackTracesID := [][]byte{}
	for _, p := range res.Msg.Profiles {
		for _, s := range p.Stacktraces {
			stackTracesID = append(stackTracesID, s.ID)
		}
	}

	symbolsReponse, err := d.SymbolizeStacktraces(context.Background(), connect.NewRequest(&ingestv1.SymbolizeStacktraceRequest{Ids: stackTracesID}))
	require.NoError(t, err)

	var stacktraces []string
	for _, p := range symbolsReponse.Msg.Locations {
		stracktrace := strings.Builder{}
		for j, l := range p.Ids {
			if j > 0 {
				stracktrace.WriteString("|")
			}
			stracktrace.WriteString(symbolsReponse.Msg.FunctionNames[l])

		}
		stacktraces = append(stacktraces, stracktrace.String())

	}
	sort.Strings(stacktraces)
	require.Equal(t, []string{"1|2|3", "buzz|baz|foo", "buzz|baz|foo", "foo|bar|buzz", "foo|bar|buzz"}, stacktraces)
	require.Equal(t, 5, len(symbolsReponse.Msg.Locations))
}

func generateProfile(
	t *testing.T,
	sampleType, sampleUnit, periodType, periodUnit string,
	ts time.Time,
	values []int64,
	locations [][]string,
) []byte {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	mapping := &profile.Mapping{
		ID: 1,
	}
	functionMap := map[string]uint64{}
	locMap := map[string]*profile.Location{}
	fns := []*profile.Function{}
	locs := []*profile.Location{}
	id := uint64(1)
	for _, location := range locations {
		for _, function := range location {
			if _, ok := functionMap[function]; !ok {
				functionMap[function] = id
				fn := &profile.Function{
					ID:        id,
					Name:      function,
					StartLine: 1,
				}
				fns = append(fns, fn)
				loc := &profile.Location{
					ID:      id,
					Address: 0,
					Mapping: mapping,
					Line: []profile.Line{
						{Function: fn, Line: 1},
					},
				}
				locMap[function] = loc
				locs = append(locs, loc)
				id++
			}
		}
	}
	var samples []*profile.Sample
	for i, loc := range locations {
		s := &profile.Sample{
			Value: []int64{values[i]},
		}
		samples = append(samples, s)
		for _, function := range loc {
			s.Location = append(s.Location, locMap[function])
		}
	}
	p := &profile.Profile{
		SampleType: []*profile.ValueType{
			{Type: sampleType, Unit: sampleUnit},
		},
		PeriodType: &profile.ValueType{
			Type: periodType,
			Unit: periodUnit,
		},
		DurationNanos: 0,
		Period:        3,
		TimeNanos:     ts.UnixNano(),
		Sample:        samples,
		Mapping: []*profile.Mapping{
			mapping,
		},
		Function: fns,
		Location: locs,
	}
	require.NoError(t, p.Write(buf))
	return buf.Bytes()
}
*/
