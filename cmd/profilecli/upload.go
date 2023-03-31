package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log/level"
	gprofile "github.com/google/pprof/profile"
	"github.com/grafana/dskit/runutil"
	"github.com/k0kubun/pp/v3"
	"github.com/klauspost/compress/gzip"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"

	"github.com/grafana/phlare/api/gen/proto/go/push/v1/pushv1connect"
	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
)

func (c *phlareClient) pusherClient() pushv1connect.PusherServiceClient {
	return pushv1connect.NewPusherServiceClient(
		c.httpClient(),
		c.URL,
	)
}

type uploadParams struct {
	*phlareClient
	extraLabels map[string]string
}

func addUploadParams(cmd flagger) *uploadParams {
	var (
		params = &uploadParams{}
	)
	params.phlareClient = addPhlareClient(cmd)

	cmd.Flag("extra-labels", "Add additional labels to the profile(s)").Default("").StringMapVar(&params.extraLabels)
	return params
}

func upload(ctx context.Context, params *queryParams, outputFlag string) (err error) {
	from, to, err := params.parseFromTo()
	if err != nil {
		return err
	}

	level.Info(logger).Log("msg", "query aggregated profile from profile store", "url", params.URL, "from", from, "to", to, "query", params.Query, "type", params.ProfileType)

	qc := params.phlareClient.queryClient()

	resp, err := qc.SelectMergeProfile(ctx, connect.NewRequest(&querierv1.SelectMergeProfileRequest{
		ProfileTypeID: params.ProfileType,
		Start:         from.UnixMilli(),
		End:           to.UnixMilli(),
		LabelSelector: params.Query,
	}))

	if err != nil {
		return errors.Wrap(err, "failed to query")
	}

	mypp := pp.New()
	mypp.SetColoringEnabled(isatty.IsTerminal(os.Stdout.Fd()))
	mypp.SetExportedOnly(true)

	if outputFlag == outputConsole {
		buf, err := resp.Msg.MarshalVT()
		if err != nil {
			return errors.Wrap(err, "failed to marshal protobuf")
		}

		p, err := gprofile.Parse(bytes.NewReader(buf))
		if err != nil {
			return errors.Wrap(err, "failed to parse profile")
		}

		fmt.Fprintln(output(ctx), p.String())
		return nil

	}

	if outputFlag == outputRaw {
		mypp.Print(resp.Msg)
		return nil
	}

	if strings.HasPrefix(outputFlag, outputPprof) {
		filePath := strings.TrimPrefix(outputFlag, outputPprof)
		if filePath == "" {
			return errors.New("no file path specified after pprof=")
		}
		buf, err := resp.Msg.MarshalVT()
		if err != nil {
			return errors.Wrap(err, "failed to marshal protobuf")
		}

		// open new file, fail when the file already exists
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to create pprof file")
		}
		defer runutil.CloseWithErrCapture(&err, f, "failed to close pprof file")

		gzipWriter := gzip.NewWriter(f)
		defer runutil.CloseWithErrCapture(&err, gzipWriter, "failed to close pprof gzip writer")

		if _, err := io.Copy(gzipWriter, bytes.NewReader(buf)); err != nil {
			return errors.Wrap(err, "failed to write pprof")
		}

		return nil
	}

	return errors.Errorf("unknown output %s", outputFlag)
}
