package v1

import (
	"github.com/segmentio/parquet-go"

	phlareparquet "github.com/grafana/phlare/pkg/parquet"
)

var stacktracesSchema = parquet.NewSchema("Stacktrace", phlareparquet.Group{
	phlareparquet.NewGroupField("ID", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
	phlareparquet.NewGroupField("LocationIDs", parquet.List(parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked))),
})

type Stacktrace struct {
	ID          uint64   `parquet:",delta"`
	LocationIDs []uint64 `parquet:",list"`
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
		parquet.Ascending("LocationIDs", "list", "element"),
	)
}

func (*StacktracePersister) Deconstruct(row parquet.Row, id uint64, s *Stacktrace) parquet.Row {
	return stacktracesSchema.Deconstruct(row, &s)
}

func (*StacktracePersister) Reconstruct(row parquet.Row) (id uint64, s *Stacktrace, err error) {
	var stored Stacktrace
	if err := stacktracesSchema.Reconstruct(&stored, row); err != nil {
		return 0, nil, err
	}
	return stored.ID, &stored, nil
}
