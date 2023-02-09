package validation

import (
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/model"

	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
	phlaremodel "github.com/grafana/phlare/pkg/model"
)

type Reason string

const (
	ReasonLabel string = "reason"
	// InvalidLabels is a reason for discarding profiles which have labels that are invalid.
	InvalidLabels Reason = "invalid_labels"
	// MissingLabels is a reason for discarding profiles which have no labels.
	MissingLabels Reason = "missing_labels"
	// RateLimited is one of the values for the reason to discard samples.
	RateLimited Reason = "rate_limited"
	// OutOfOrder is a reason for discarding profiles when Phlare doesn't accept out
	// of order profiles.
	OutOfOrder Reason = "out_of_order"
	// MaxLabelNamesPerSeries is a reason for discarding a request which has too many label names
	MaxLabelNamesPerSeries Reason = "max_label_names_per_series"
	// LabelNameTooLong is a reason for discarding a request which has a label name too long
	LabelNameTooLong Reason = "label_name_too_long"
	// LabelValueTooLong is a reason for discarding a request which has a label value too long
	LabelValueTooLong Reason = "label_value_too_long"
	// DuplicateLabelNames is a reason for discarding a request which has duplicate label names
	DuplicateLabelNames Reason = "duplicate_label_names"

	MissingLabelsErrorMsg          = "error at least one label pair is required per profile"
	InvalidLabelsErrorMsg          = "invalid labels '%s' with error: %s"
	MaxLabelNamesPerSeriesErrorMsg = "profile series '%s' has %d label names; limit %d"
	LabelNameTooLongErrorMsg       = "profile with labels '%s' has label name too long: '%s'"
	LabelValueTooLongErrorMsg      = "profile with labels '%s' has label value too long: '%s'"
	DuplicateLabelNamesErrorMsg    = "profile with labels '%s' has duplicate label name: '%s'"
)

var (
	// DiscardedBytes is a metric of the total discarded bytes, by reason.
	DiscardedBytes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "phlare",
			Name:      "discarded_bytes_total",
			Help:      "The total number of bytes that were discarded.",
		},
		[]string{ReasonLabel, "tenant"},
	)

	// DiscardedProfiles is a metric of the number of discarded profiles, by reason.
	DiscardedProfiles = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "phlare",
			Name:      "discarded_samples_total",
			Help:      "The total number of samples that were discarded.",
		},
		[]string{ReasonLabel, "tenant"},
	)
)

type LabelValidationLimits interface {
	MaxLabelNameLength(userID string) int
	MaxLabelValueLength(userID string) int
	MaxLabelNamesPerSeries(userID string) int
}

// ValidateLabels validates the labels of a profile.
func ValidateLabels(limits LabelValidationLimits, userID string, ls []*typesv1.LabelPair) error {
	if len(ls) == 0 {
		return NewErrorf(MissingLabels, MissingLabelsErrorMsg)
	}
	sort.Sort(phlaremodel.Labels(ls))
	numLabelNames := len(ls)
	maxLabels := limits.MaxLabelNamesPerSeries(userID)
	if numLabelNames > maxLabels {
		return NewErrorf(MaxLabelNamesPerSeries, MaxLabelNamesPerSeriesErrorMsg, phlaremodel.LabelPairsString(ls), numLabelNames, maxLabels)
	}
	nameValue := phlaremodel.Labels(ls).Get(model.MetricNameLabel)
	if !model.IsValidMetricName(model.LabelValue(nameValue)) {
		return NewErrorf(InvalidLabels, InvalidLabelsErrorMsg, phlaremodel.LabelPairsString(ls), "invalid metric name")
	}
	lastLabelName := ""

	for _, l := range ls {
		if len(l.Name) > limits.MaxLabelNameLength(userID) {
			return NewErrorf(LabelNameTooLong, LabelNameTooLongErrorMsg, phlaremodel.LabelPairsString(ls), l.Name)
		} else if len(l.Value) > limits.MaxLabelValueLength(userID) {
			return NewErrorf(LabelValueTooLong, LabelValueTooLongErrorMsg, phlaremodel.LabelPairsString(ls), l.Value)
		} else if !model.LabelName(l.Name).IsValid() {
			return NewErrorf(InvalidLabels, InvalidLabelsErrorMsg, phlaremodel.LabelPairsString(ls), "invalid label name '"+l.Name+"'")
		} else if !model.LabelValue(l.Value).IsValid() {
			return NewErrorf(InvalidLabels, InvalidLabelsErrorMsg, phlaremodel.LabelPairsString(ls), "invalid label value '"+l.Value+"'")
		} else if cmp := strings.Compare(lastLabelName, l.Name); cmp == 0 {
			return NewErrorf(DuplicateLabelNames, DuplicateLabelNamesErrorMsg, phlaremodel.LabelPairsString(ls), l.Name)
		}
		lastLabelName = l.Name
	}

	return nil
}

type Error struct {
	Reason Reason
	msg    string
}

func (e *Error) Error() string {
	return e.msg
}

func NewErrorf(reason Reason, msg string, args ...interface{}) *Error {
	return &Error{
		Reason: reason,
		msg:    msg,
	}
}

func ReasonOf(err error) Reason {
	var validationErr *Error
	ok := errors.As(err, &validationErr)
	if !ok {
		return "unknown"
	}
	return validationErr.Reason
}
