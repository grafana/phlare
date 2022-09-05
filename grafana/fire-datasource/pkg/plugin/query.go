package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	querierv1 "github.com/grafana/fire/pkg/gen/querier/v1"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
)

type queryModel struct {
	WithStreaming bool
	ProfileTypeID string `json:"profileTypeId"`
	LabelSelector string `json:"labelSelector"`
}

// query processes single Fire query transforming the response to data.Frame packaged in DataResponse
func (d *FireDatasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var qm queryModel
	response := backend.DataResponse{}

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		response.Error = err
		return response
	}

	log.DefaultLogger.Debug("Querying SelectMergeStacktraces()", "queryModel", qm)

	resp, err := d.client.SelectMergeStacktraces(ctx, makeRequest(qm, query))
	if err != nil {
		response.Error = err
		return response
	}
	profileFrame, err := profileToDataFrame(resp)
	if err != nil {
		response.Error = err
		return response
	}

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      "stream",
		}
		profileFrame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}

	seriesResp, err := d.client.SelectSeries(ctx, connect.NewRequest(&querierv1.SelectSeriesRequest{
		ProfileTypeID: qm.ProfileTypeID,
		LabelSelector: qm.LabelSelector,
		Start:         query.TimeRange.From.UnixMilli(),
		End:           query.TimeRange.To.UnixMilli(),
		Step:          query.Interval.Seconds(),
		// todo add one or more group bys
		GroupBy: []string{},
	}))
	if err != nil {
		response.Error = err
		return response
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, seriesToDataFrame(seriesResp, qm.ProfileTypeID))
	response.Frames = append(response.Frames, profileFrame)

	return response
}

func makeRequest(qm queryModel, query backend.DataQuery) *connect.Request[querierv1.SelectMergeStacktracesRequest] {
	return &connect.Request[querierv1.SelectMergeStacktracesRequest]{
		Msg: &querierv1.SelectMergeStacktracesRequest{
			ProfileTypeID: qm.ProfileTypeID,
			LabelSelector: qm.LabelSelector,
			Start:         query.TimeRange.From.UnixMilli(),
			End:           query.TimeRange.To.UnixMilli(),
		},
	}
}

type CustomMeta struct {
	Names   []string
	Total   int64
	MaxSelf int64
}

// profileToDataFrame turns fire profile response to data.Frame. At this point this transform is very simple, each
// level being encoded as json string and set as a single value in a single column. Reason for this is that each level
// can have variable number of values but in data.Frame each column needs to have the same number of values.
// In addition, Names, Total, MaxSelf is added to Meta.Custom which may not be the best practice so needs to be
// evaluated later on
func profileToDataFrame(resp *connect.Response[querierv1.SelectMergeStacktracesResponse]) (*data.Frame, error) {
	frame := data.NewFrame("profile")
	frame.Meta = &data.FrameMeta{PreferredVisualization: "profile"}

	levelsField := data.NewField("levels", nil, []string{})

	for index, level := range resp.Msg.Flamegraph.Levels {
		bytes, err := json.Marshal(level.Values)
		if err != nil {
			return nil, fmt.Errorf("error marshaling level %d with values %v: %v", index, level, err)
		}
		levelsField.Append(string(bytes))
	}
	frame.Fields = []*data.Field{levelsField}
	frame.Meta.Custom = CustomMeta{
		Names:   resp.Msg.Flamegraph.Names,
		Total:   resp.Msg.Flamegraph.Total,
		MaxSelf: resp.Msg.Flamegraph.MaxSelf,
	}
	return frame, nil
}

func seriesToDataFrame(seriesResp *connect.Response[querierv1.SelectSeriesResponse], profileTypeID string) *data.Frame {
	frame := data.NewFrame("series")
	frame.Meta = &data.FrameMeta{PreferredVisualization: "graph"}

	fields := data.Fields{}
	timeField := data.NewField("time", nil, []time.Time{})
	fields = append(fields, timeField)

	for index, series := range seriesResp.Msg.Series {
		label := ""
		if len(series.Labels) > 0 {
			label = series.Labels[0].Name
		} else {
			parts := strings.Split(profileTypeID, ":")
			if len(parts) == 5 {
				label = parts[3]
			}
		}
		valueField := data.NewField(label, nil, []float64{})

		for _, point := range series.Points {
			if index == 0 {
				timeField.Append(time.UnixMilli(point.T))
			}
			valueField.Append(point.V)
		}

		fields = append(fields, valueField)
	}

	frame.Fields = fields
	return frame
}
