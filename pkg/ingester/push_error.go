package ingester

import (
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/samber/lo"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/proto"

	"github.com/grafana/phlare/pkg/validation"
)

type pushError struct {
	err  error
	size int
	id   string
}

func (e pushError) Reason() validation.Reason {
	return validation.ReasonOf(e.err)
}

func (e *pushError) Error() string {
	return fmt.Sprintf("profile ID %s: %s", e.id, e.err)
}

func (e pushError) Details() proto.Message {
	switch e.Reason() {
	case validation.OutOfOrder:
		return &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "timestamp", Description: e.Error()},
			},
		}
	case validation.SeriesLimit:
		return &errdetails.QuotaFailure{
			Violations: []*errdetails.QuotaFailure_Violation{
				{
					Subject:     "series",
					Description: e.Error(),
				},
			},
		}
	case validation.InvalidProfile:
		return &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{Field: "profile", Description: e.Error()},
			},
		}
	default:
		return &errdetails.ErrorInfo{
			Reason: string(e.Reason()),
			Metadata: map[string]string{
				"err": e.Error(),
			},
		}
	}
}

type multiPushError []pushError

func (e multiPushError) Err() error {
	if len(e) == 0 {
		return nil
	}
	connectErr := connect.NewError(codeFromErrs(e), errors.New("error while ingesting profiles see responses details"))
	for _, errDetail := range e {
		details, err := connect.NewErrorDetail(errDetail.Details())
		if err != nil {
			return err
		}
		connectErr.AddDetail(details)
	}
	return connectErr
}

func (e *multiPushError) Add(err error, id string, size int) {
	if err == nil {
		return
	}
	*e = append(*e, pushError{
		err:  err,
		id:   id,
		size: size,
	})
}

func codeFromErrs(errs []pushError) connect.Code {
	countByReason := lo.CountValuesBy(errs, func(err pushError) validation.Reason { return err.Reason() })
	if countByReason[validation.Unknown] > 0 {
		return connect.CodeInternal
	}
	if countByReason[validation.SeriesLimit] > 0 {
		return connect.CodeResourceExhausted
	}
	return connect.CodeInvalidArgument
}
