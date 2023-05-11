package ingester

import (
	"errors"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"

	"github.com/grafana/phlare/pkg/validation"
)

func Test_multiPushError_Err(t *testing.T) {
	for _, tt := range []struct {
		name      string
		generator func() multiPushError
		status    connect.Code
		details   []*connect.ErrorDetail
	}{
		{
			name:      "empty",
			generator: func() multiPushError { return nil },
		},
		{
			name: "single empty",
			generator: func() multiPushError {
				var e multiPushError
				e.Add(nil, "", 0)
				return e
			},
		},
		{
			name: "nil error",
			generator: func() multiPushError {
				var e multiPushError
				e.Add(nil, "", 0)
				return e
			},
		},
		{
			name: "invalid profile & out of order",
			generator: func() multiPushError {
				var e multiPushError
				e.Add(validation.NewErrorf(validation.InvalidProfile, "invalid ID: %s", errors.New("foo")), "foo", 1)
				e.Add(validation.NewErrorf(validation.OutOfOrder, "out of order"), "bar", 2)
				return e
			},
			status: connect.CodeInvalidArgument,
			details: []*connect.ErrorDetail{
				mustNewErrorDetails(&errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{
						{Field: "profile", Description: (&pushError{err: validation.NewErrorf(validation.InvalidProfile, "invalid ID: %s", errors.New("foo")), id: "foo", size: 1}).Error()},
					},
				}),
				mustNewErrorDetails(&errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{
						{Field: "timestamp", Description: (&pushError{err: validation.NewErrorf(validation.OutOfOrder, "out of order"), id: "bar", size: 2}).Error()},
					},
				}),
			},
		},
		{
			name: "throttle overrides bad request",
			generator: func() multiPushError {
				var e multiPushError
				e.Add(validation.NewErrorf(validation.InvalidProfile, "invalid ID: %s", errors.New("foo")), "foo", 1)
				e.Add(validation.NewErrorf(validation.SeriesLimit, "limit reached"), "bar", 2)
				return e
			},
			status: connect.CodeResourceExhausted,
			details: []*connect.ErrorDetail{
				mustNewErrorDetails(&errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{
						{Field: "profile", Description: (&pushError{err: validation.NewErrorf(validation.InvalidProfile, "invalid ID: %s", errors.New("foo")), id: "foo", size: 1}).Error()},
					},
				}),
				mustNewErrorDetails(&errdetails.QuotaFailure{
					Violations: []*errdetails.QuotaFailure_Violation{
						{Subject: "series", Description: (&pushError{err: validation.NewErrorf(validation.SeriesLimit, "limit reached"), id: "bar", size: 2}).Error()},
					},
				}),
			},
		},
		{
			name: "unknown overrides all",
			generator: func() multiPushError {
				var e multiPushError
				e.Add(validation.NewErrorf(validation.InvalidProfile, "invalid ID: %s", errors.New("foo")), "foo", 1)
				e.Add(validation.NewErrorf(validation.SeriesLimit, "limit reached"), "bar", 2)
				e.Add(errors.New("bad"), "buzz", 3)
				return e
			},
			status: connect.CodeInternal,
			details: []*connect.ErrorDetail{
				mustNewErrorDetails(&errdetails.BadRequest{
					FieldViolations: []*errdetails.BadRequest_FieldViolation{
						{Field: "profile", Description: (&pushError{err: validation.NewErrorf(validation.InvalidProfile, "invalid ID: %s", errors.New("foo")), id: "foo", size: 1}).Error()},
					},
				}),
				mustNewErrorDetails(&errdetails.QuotaFailure{
					Violations: []*errdetails.QuotaFailure_Violation{
						{Subject: "series", Description: (&pushError{err: validation.NewErrorf(validation.SeriesLimit, "limit reached"), id: "bar", size: 2}).Error()},
					},
				}),
				mustNewErrorDetails(&errdetails.ErrorInfo{
					Reason: string(validation.Unknown),
					// TODO: Adding more metadata to the error is difficult to test because of how go maps are serialized in random order.
					Metadata: map[string]string{
						"err": (&pushError{err: errors.New("bad"), id: "buzz", size: 3}).Error(),
					},
				}),
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err, _ := tt.generator().Err().(*connect.Error)
			var (
				gotStatus  connect.Code
				gotDetails []*connect.ErrorDetail
			)
			if err != nil {
				gotStatus = err.Code()
				gotDetails = err.Details()
			}
			require.Equal(t, tt.status, gotStatus, "missing status want %d got %d", tt.status, gotStatus)
			require.EqualValues(t, tt.details, gotDetails)
		})
	}
}

func mustNewErrorDetails(m proto.Message) *connect.ErrorDetail {
	d, err := connect.NewErrorDetail(m)
	if err != nil {
		panic(err)
	}
	return d
}
