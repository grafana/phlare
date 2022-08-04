package querier

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ingesterv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	"github.com/grafana/fire/pkg/gen/ingester/v1/ingesterv1connect"
	pushv1 "github.com/grafana/fire/pkg/gen/push/v1"
	"github.com/grafana/fire/pkg/iterator"
	firemodel "github.com/grafana/fire/pkg/model"
	"github.com/grafana/fire/pkg/testhelper"
)

var (
	fooLabels = firemodel.NewLabelsBuilder().Set("label", "foo").Labels()
	barLabels = firemodel.NewLabelsBuilder().Set("label", "bar").Labels()
)

func generateProfiles(t *testing.T, start, end int, lbs firemodel.Labels, ingAddr string) []ProfileWithLabels {
	t.Helper()
	res := make([]ProfileWithLabels, 0, end-start)
	for i := start; i <= end; i++ {
		res = append(res, generateProfile(t, i, lbs, ingAddr))
	}
	return res
}

func generateProfile(t *testing.T, i int, lbs firemodel.Labels, ingAddr string) ProfileWithLabels {
	t.Helper()
	return ProfileWithLabels{
		Labels:       lbs,
		IngesterAddr: ingAddr,
		Profile: &ingesterv1.Profile{
			ID:        fmt.Sprintf("%d", i) + ingAddr,
			Timestamp: int64(i),
		},
	}
}

func TestDedupe(t *testing.T) {
	for _, tt := range []struct {
		name     string
		in       [][]ProfileWithLabels
		expected []ProfileWithLabels
	}{
		{
			"empty",
			[][]ProfileWithLabels{},
			nil,
		},
		{
			"single",
			[][]ProfileWithLabels{
				generateProfiles(t, 0, 1, fooLabels, "foo"),
			},
			generateProfiles(t, 0, 1, fooLabels, "foo"),
		},
		{
			"no duplicates",
			[][]ProfileWithLabels{
				generateProfiles(t, 0, 5, fooLabels, "foo"),
				generateProfiles(t, 6, 10, fooLabels, "foo"),
			},
			generateProfiles(t, 0, 10, fooLabels, "foo"),
		},
		{
			"different labels",
			[][]ProfileWithLabels{
				generateProfiles(t, 0, 1, fooLabels, "foo"),
				generateProfiles(t, 0, 1, barLabels, "bar"),
			},
			[]ProfileWithLabels{
				generateProfile(t, 0, barLabels, "bar"),
				generateProfile(t, 0, fooLabels, "foo"),
				generateProfile(t, 1, barLabels, "bar"),
				generateProfile(t, 1, fooLabels, "foo"),
			},
		},
		{
			"same labels",
			[][]ProfileWithLabels{
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
				generateProfiles(t, 0, 10, fooLabels, "foo"),
			},
			generateProfiles(t, 0, 10, fooLabels, "foo"),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var in []iterator.Interface[ProfileWithLabels]
			for _, profiles := range tt.in {
				in = append(in, iterator.NewSliceIterator(profiles))
			}
			actual, err := iterator.Slice(NewDedupeProfileIterator(in))
			require.NoError(t, err)
			testhelper.EqualProto(t, tt.expected, actual)
		})
	}
}

var _ ingesterv1connect.IngesterServiceHandler = &ingesterMock{}

type ingesterMock struct {
	mock.Mock
}

func (i *ingesterMock) Push(ctx context.Context, req *connect.Request[pushv1.PushRequest]) (*connect.Response[pushv1.PushResponse], error) {
	args := i.Called(ctx, req)
	return args.Get(0).(*connect.Response[pushv1.PushResponse]), args.Error(1)
}

func (i *ingesterMock) LabelValues(ctx context.Context, req *connect.Request[ingesterv1.LabelValuesRequest]) (*connect.Response[ingesterv1.LabelValuesResponse], error) {
	args := i.Called(ctx, req)
	return args.Get(0).(*connect.Response[ingesterv1.LabelValuesResponse]), args.Error(1)
}

func (i *ingesterMock) ProfileTypes(ctx context.Context, req *connect.Request[ingesterv1.ProfileTypesRequest]) (*connect.Response[ingesterv1.ProfileTypesResponse], error) {
	args := i.Called(ctx, req)
	return args.Get(0).(*connect.Response[ingesterv1.ProfileTypesResponse]), args.Error(1)
}

func (i *ingesterMock) Flush(ctx context.Context, req *connect.Request[ingesterv1.FlushRequest]) (*connect.Response[ingesterv1.FlushResponse], error) {
	args := i.Called(ctx, req)
	return args.Get(0).(*connect.Response[ingesterv1.FlushResponse]), args.Error(1)
}

func (i *ingesterMock) SelectProfiles(ctx context.Context, req *connect.Request[ingesterv1.SelectProfilesRequest], stream *connect.ServerStream[ingesterv1.SelectProfilesResponse]) error {
	args := i.Called(ctx, req, stream)
	return args.Error(0)
}

func (i *ingesterMock) SelectStacktraceSamples(ctx context.Context, stream *connect.ClientStream[ingesterv1.SelectStacktraceSamplesRequest]) (*connect.Response[ingesterv1.SelectStacktraceSamplesResponse], error) {
	args := i.Called(ctx, stream)
	return args.Get(0).(*connect.Response[ingesterv1.SelectStacktraceSamplesResponse]), args.Error(1)
}

func TestStreamingIterator(t *testing.T) {
	for _, tt := range []struct {
		name     string
		streamFn func(stream *connect.ServerStream[ingesterv1.SelectProfilesResponse])
		expected []ProfileWithLabels
	}{
		{
			name:     "empty",
			streamFn: func(stream *connect.ServerStream[ingesterv1.SelectProfilesResponse]) {},
			expected: nil,
		},
		{
			name: "one batch",
			streamFn: func(stream *connect.ServerStream[ingesterv1.SelectProfilesResponse]) {
				_ = stream.Send(&ingesterv1.SelectProfilesResponse{
					Profiles: []*ingesterv1.Profile{
						{ID: "1foo", LabelsetIndex: 0, Timestamp: 1, TotalValue: 0},
					},
					Labelsets: []*ingesterv1.Labels{{Labels: fooLabels}},
				})
			},
			expected: []ProfileWithLabels{
				generateProfile(t, 1, fooLabels, "foo"),
			},
		},
		{
			name: "two batch",
			streamFn: func(stream *connect.ServerStream[ingesterv1.SelectProfilesResponse]) {
				_ = stream.Send(&ingesterv1.SelectProfilesResponse{
					Profiles: []*ingesterv1.Profile{
						{ID: "1foo", LabelsetIndex: 0, Timestamp: 1, TotalValue: 0},
					},
					Labelsets: []*ingesterv1.Labels{{Labels: fooLabels}},
				})
				_ = stream.Send(&ingesterv1.SelectProfilesResponse{
					Profiles: []*ingesterv1.Profile{
						{ID: "2foo", LabelsetIndex: 0, Timestamp: 2, TotalValue: 0},
					},
					Labelsets: []*ingesterv1.Labels{{Labels: fooLabels}},
				})
				_ = stream.Send(&ingesterv1.SelectProfilesResponse{
					Profiles: []*ingesterv1.Profile{
						{ID: "3foo", LabelsetIndex: 0, Timestamp: 3, TotalValue: 0},
					},
					Labelsets: []*ingesterv1.Labels{{Labels: fooLabels}},
				})
			},
			expected: generateProfiles(t, 1, 3, fooLabels, "foo"),
		},
		{
			name: "two batch multiple profiles",
			streamFn: func(stream *connect.ServerStream[ingesterv1.SelectProfilesResponse]) {
				_ = stream.Send(&ingesterv1.SelectProfilesResponse{
					Profiles: []*ingesterv1.Profile{
						{ID: "1foo", LabelsetIndex: 0, Timestamp: 1, TotalValue: 0},
						{ID: "2foo", LabelsetIndex: 0, Timestamp: 2, TotalValue: 0},
					},
					Labelsets: []*ingesterv1.Labels{{Labels: fooLabels}},
				})
				_ = stream.Send(&ingesterv1.SelectProfilesResponse{
					Profiles: []*ingesterv1.Profile{
						{ID: "3foo", LabelsetIndex: 0, Timestamp: 3, TotalValue: 0},
						{ID: "4foo", LabelsetIndex: 0, Timestamp: 4, TotalValue: 0},
					},
					Labelsets: []*ingesterv1.Labels{{Labels: fooLabels}},
				})
			},
			expected: generateProfiles(t, 1, 4, fooLabels, "foo"),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ing := &ingesterMock{}
			ing.On("SelectProfiles", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
				tt.streamFn(args.Get(2).(*connect.ServerStream[ingesterv1.SelectProfilesResponse]))
			}).Return(nil)
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

			it := NewStreamsProfileIterator([]responseFromIngesters[*connect.ServerStreamForClient[ingesterv1.SelectProfilesResponse]]{
				{
					addr:     "foo",
					response: stream,
				},
			})
			actual, err := iterator.Slice(it)
			require.NoError(t, err)
			testhelper.EqualProto(t, tt.expected, actual)
		})
	}
}
