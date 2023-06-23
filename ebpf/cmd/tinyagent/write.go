package main

import (
	"context"
	"github.com/bufbuild/connect-go"
	pushv1 "github.com/grafana/phlare/api/gen/proto/go/push/v1"
	"github.com/prometheus/common/model"

	"github.com/grafana/phlare/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	commonconfig "github.com/prometheus/common/config"
	"github.com/prometheus/prometheus/model/labels"
	"strings"
)

type Write struct {
	externalLabels map[string]string
	client         pushv1connect.PusherServiceClient
}

func (w *Write) Append(ctx context.Context, lbs labels.Labels, samples []*RawSample) error {
	// todo(ctovena): we should probably pool the label pair arrays and label builder to avoid allocs.
	var (
		protoLabels  = make([]*typesv1.LabelPair, 0, len(lbs))
		protoSamples = make([]*pushv1.RawSample, 0, len(samples))
		lbsBuilder   = labels.NewBuilder(nil)
	)

	for _, label := range lbs {
		// only __name__ is required as a private label.
		if strings.HasPrefix(label.Name, model.ReservedLabelPrefix) && label.Name != labels.MetricName {
			continue
		}
		lbsBuilder.Set(label.Name, label.Value)
	}
	//for name, value := range f.config.ExternalLabels {
	//	lbsBuilder.Set(name, value)
	//}
	for _, l := range lbsBuilder.Labels() {
		protoLabels = append(protoLabels, &typesv1.LabelPair{
			Name:  l.Name,
			Value: l.Value,
		})
	}
	for _, sample := range samples {
		protoSamples = append(protoSamples, &pushv1.RawSample{
			RawProfile: sample.RawProfile,
		})
	}
	// push to all clients
	_, err := w.client.Push(ctx, connect.NewRequest(&pushv1.PushRequest{
		Series: []*pushv1.RawProfileSeries{
			{Labels: protoLabels, Samples: protoSamples},
		},
	}))
	return err
}

func (w *Write) Appender() Appender {
	return w
}

func NewWrite(endpointUrl string, username string, password string, passwordFile string) *Write {

	//basic_auth {
	//	password = null
	//	password_file = "/passwords/profiles-dev-001/write"
	//	username = "1218"
	//}
	//name = "profiles-dev-001"
	//url = "http://cortex-gw.fire-dev-001.svc.cluster.local.:80"

	config := commonconfig.DefaultHTTPClientConfig
	config.BasicAuth = &commonconfig.BasicAuth{
		Username:     username,
		Password:     commonconfig.Secret(password),
		PasswordFile: passwordFile,
	}
	httpClient, err := commonconfig.NewClientFromConfig(config, "tiny")
	if err != nil {
		panic(err)
	}

	client := pushv1connect.NewPusherServiceClient(httpClient, endpointUrl)

	return &Write{client: client}
}
