package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log/level"
	"github.com/k0kubun/pp/v3"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	"gopkg.in/alecthomas/kingpin.v2"

	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	"github.com/grafana/phlare/api/gen/proto/go/querier/v1/querierv1connect"
)

const (
	outputConsole = "console"
	outputPprof   = "pprof="
)

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	// try if it is a relative time
	d, rerr := parseRelativeTime(s)
	if rerr == nil {
		return time.Now().Add(-d), nil
	}

	// if not return first error
	return time.Time{}, err

}

func parseRelativeTime(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "now" {
		return 0, nil
	}
	s = strings.TrimPrefix(s, "now-")

	d, err := model.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return time.Duration(d), nil
}

type queryParams struct {
	URL         string
	From        string
	To          string
	ProfileType string
	Query       string
}

func (p *queryParams) parseFromTo() (from time.Time, to time.Time, err error) {
	from, err = parseTime(p.From)
	if err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(err, "failed to parse from")
	}
	to, err = parseTime(p.To)
	if err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(err, "failed to parse to")
	}

	if to.Before(from) {
		return time.Time{}, time.Time{}, errors.Wrap(err, "from cannot be after")
	}

	return from, to, nil
}

func (p *queryParams) client() querierv1connect.QuerierServiceClient {
	return querierv1connect.NewQuerierServiceClient(
		http.DefaultClient,
		p.URL,
	)
}

type flagger interface {
	Flag(name, help string) *kingpin.FlagClause
}

func addQueryParams(queryCmd flagger) *queryParams {
	params := &queryParams{}
	queryCmd.Flag("url", "URL of the profile store.").Default("http://localhost:4100").StringVar(&params.URL)
	queryCmd.Flag("from", "Beginning of the query.").Default("now-1h").StringVar(&params.From)
	queryCmd.Flag("to", "End of the query.").Default("now").StringVar(&params.To)
	queryCmd.Flag("profile-type", "Profile type to query.").Default("process_cpu:cpu:nanoseconds:cpu:nanoseconds").StringVar(&params.ProfileType)
	queryCmd.Flag("query", "Label selector to query.").Default("{}").StringVar(&params.Query)
	return params
}

func queryMerge(ctx context.Context, params *queryParams, output string) error {
	from, to, err := params.parseFromTo()
	if err != nil {
		return err
	}

	level.Info(logger).Log("msg", "query aggregated profile from profile store", "url", params.URL, "from", from, "to", to, "query", params.Query, "type", params.ProfileType)

	qc := params.client()

	resp, err := qc.SelectMergeProfile(ctx, connect.NewRequest(&querierv1.SelectMergeProfileRequest{
		ProfileTypeID: params.ProfileType,
		Start:         from.UnixMilli(),
		End:           to.UnixMilli(),
		LabelSelector: params.Query,
	}))

	if err != nil {
		return errors.Wrap(err, "failed to query")
	}

	if output == outputConsole {
		mypp := pp.New()
		mypp.SetColoringEnabled(isatty.IsTerminal(os.Stdout.Fd()))
		mypp.SetExportedOnly(true)
		mypp.Print(resp.Msg)
		return nil
	}
	if strings.HasPrefix(output, outputPprof) {
		filePath := strings.TrimPrefix(output, outputPprof)
		buf, err := resp.Msg.MarshalVT()
		if err != nil {
			return errors.Wrap(err, "failed to marshal protobuf")
		}

		os.WriteFile(filePath, buf, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write pprof")
		}

		return nil
	}

	return errors.Errorf("unknown output %s", output)
}
