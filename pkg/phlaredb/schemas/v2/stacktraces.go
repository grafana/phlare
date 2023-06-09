package v2

import (
	"github.com/segmentio/parquet-go"

	phlareparquet "github.com/grafana/phlare/pkg/parquet"
)

var stacktracesSchema = parquet.NewSchema("Stacktrace", phlareparquet.Group{
	phlareparquet.NewGroupField("ID", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
	phlareparquet.NewGroupField("ParentID", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
	phlareparquet.NewGroupField("LocationID", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
})

type Stacktrace struct {
	ID         uint64 `parquet:",delta"`
	ParentID   uint64 `parquet:",delta"`
	LocationID uint64 `parquet:",delta"`
}

type StacktracePersister struct{}

func (*StacktracePersister) Name() string {
	return "stacktraces"
}

func (*StacktracePersister) Schema() *parquet.Schema {
	return stacktracesSchema
}

func (*StacktracePersister) SortingColumns() parquet.SortingOption {
	return parquet.SortingColumns(
		parquet.Ascending("ID"),
	)
}

func (*StacktracePersister) Deconstruct(row parquet.Row, id uint64, s *Stacktrace) parquet.Row {
	return stacktracesSchema.Deconstruct(row, s)
}

func (*StacktracePersister) Reconstruct(row parquet.Row) (id uint64, s *Stacktrace, err error) {
	var stored Stacktrace
	if err := stacktracesSchema.Reconstruct(&stored, row); err != nil {
		return 0, nil, err
	}

	return s.ID, &stored, nil

	panic("TODO")
	return 0, nil, nil
}
