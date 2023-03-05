package pprof

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	pushv1 "github.com/grafana/phlare/api/gen/proto/go/push/v1"
	"github.com/grafana/phlare/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
)

func Test_SendOneOff(t *testing.T) {
	client := pushv1connect.NewPusherServiceClient(http.DefaultClient, "http://localhost:4100/")

	data, err := os.ReadFile("testdata/heap")
	require.NoError(t, err)

	req := connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{
				Labels: []*typesv1.LabelPair{
					{Name: labels.MetricName, Value: "memory"},
					{Name: "foo", Value: "bar"},
				},
				Samples: []*pushv1.RawSample{
					{RawProfile: data},
				},
			},
		},
		// todo: add profile
	})
	// todo: only numerical orgs id are supported
	// req.Header().Add("X-Scope-OrgID", "foo")
	basic := base64.StdEncoding.EncodeToString([]byte("1218" + ":" + "REDACTED"))
	req.Header().Set("Authorization", "Basic "+basic)

	_, err = client.Push(context.Background(), req)
	require.NoError(t, err)
}
