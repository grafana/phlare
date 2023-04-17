package querier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/gogo/status"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/pyroscope-io/pyroscope/pkg/util/attime"
	"google.golang.org/grpc/codes"

	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	phlaremodel "github.com/grafana/phlare/pkg/model"
)

// LabelValuesHandler only returns the label values for the given label name.
// This is mostly for fulfilling the pyroscope API and won't be used in the future.
// /label-values?label=__name__
func (q *Querier) LabelValuesHandler(w http.ResponseWriter, req *http.Request) {
	label := req.URL.Query().Get("label")
	if label == "" {
		http.Error(w, "label parameter is required", http.StatusBadRequest)
		return
	}
	var (
		res []string
		err error
	)

	if label == "__name__" {
		response, err := q.ProfileTypes(req.Context(), connect.NewRequest(&querierv1.ProfileTypesRequest{}))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, t := range response.Msg.ProfileTypes {
			res = append(res, t.ID)
		}
	} else {
		response, err := q.LabelValues(req.Context(), connect.NewRequest(&querierv1.LabelValuesRequest{}))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res = response.Msg.Names
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (q *Querier) RenderDiffHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Left
	leftSelectParams, leftProfileType, err := parseSelectProfilesRequest(renderRequestFieldNames{
		query: "leftQuery",
		from:  "leftFrom",
		until: "leftUntil",
	}, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rightSelectParams, rightProfileType, err := parseSelectProfilesRequest(renderRequestFieldNames{
		query: "rightQuery",
		from:  "rightFrom",
		until: "rightUntil",
	}, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: check profile types?
	if leftProfileType.ID != rightProfileType.ID {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := q.Diff(req.Context(), connect.NewRequest(leftSelectParams), connect.NewRequest(rightSelectParams))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ExportToFlamebearer(res.Msg.Flamegraph, leftProfileType, FormatDiff)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (q *Querier) RenderHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	selectParams, profileType, err := parseSelectProfilesRequest(renderRequestFieldNames{}, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := q.SelectMergeStacktraces(req.Context(), connect.NewRequest(selectParams))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ExportToFlamebearer(res.Msg.Flamegraph, profileType, FormatSingle)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type renderRequestFieldNames struct {
	query string
	from  string
	until string
}

// render/render?format=json&from=now-12h&until=now&query=pyroscope.server.cpu
func parseSelectProfilesRequest(fieldNames renderRequestFieldNames, req *http.Request) (*querierv1.SelectMergeStacktracesRequest, *typesv1.ProfileType, error) {
	if fieldNames == (renderRequestFieldNames{}) {
		fieldNames = renderRequestFieldNames{
			query: "query",
			from:  "from",
			until: "until",
		}
	}
	selector, ptype, err := parseQuery(fieldNames.query, req)
	if err != nil {
		return nil, nil, err
	}

	v := req.URL.Query()

	// parse time using pyroscope's attime parser
	start := model.TimeFromUnixNano(attime.Parse(v.Get(fieldNames.from)).UnixNano())
	end := model.TimeFromUnixNano(attime.Parse(v.Get(fieldNames.until)).UnixNano())
	//	start := model.TimeFromUnixNano(attime.Parse(v.Get("from")).UnixNano())
	//	end := model.TimeFromUnixNano(attime.Parse(v.Get("until")).UnixNano())

	return &querierv1.SelectMergeStacktracesRequest{
		Start:         int64(start),
		End:           int64(end),
		LabelSelector: selector,
		ProfileTypeID: ptype.ID,
	}, ptype, nil
}

func parseQuery(fieldName string, req *http.Request) (string, *typesv1.ProfileType, error) {
	q := req.Form.Get(fieldName)
	if q == "" {
		return "", nil, fmt.Errorf("'%s' is required", fieldName)
	}

	parsedSelector, err := parser.ParseMetricSelector(q)
	if err != nil {
		return "", nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse '%s'", fieldName))
	}

	sel := make([]*labels.Matcher, 0, len(parsedSelector))
	var nameLabel *labels.Matcher
	for _, matcher := range parsedSelector {
		if matcher.Name == labels.MetricName {
			nameLabel = matcher
		} else {
			sel = append(sel, matcher)
		}
	}
	if nameLabel == nil {
		return "", nil, status.Error(codes.InvalidArgument, fmt.Sprintf("'%s' must contain a profile-type selection", fieldName))
	}

	profileSelector, err := phlaremodel.ParseProfileTypeSelector(nameLabel.Value)
	if err != nil {
		return "", nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse '%s'", fieldName))
	}
	return convertMatchersToString(sel), profileSelector, nil
}

func convertMatchersToString(matchers []*labels.Matcher) string {
	out := strings.Builder{}
	out.WriteRune('{')

	for idx, m := range matchers {
		if idx > 0 {
			out.WriteRune(',')
		}

		out.WriteString(m.String())
	}

	out.WriteRune('}')
	return out.String()
}
