package v2

import (
	"github.com/segmentio/parquet-go"

	phlareparquet "github.com/grafana/phlare/pkg/parquet"
)

var stacktracesSchema = parquet.NewSchema("Stacktrace", phlareparquet.Group{
	phlareparquet.NewGroupField("Parent", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
	phlareparquet.NewGroupField("LocationID", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
})

type Stacktrace struct {
	Parent     uint64 `parquet:",delta"`
	LocationID uint64 `parquet:",delta"`
}

type StoredStacktrace struct {
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
	panic("TODO")
	return row
}

func (*StacktracePersister) Reconstruct(row parquet.Row) (id uint64, s *Stacktrace, err error) {
	panic("TODO")
	return 0, nil, nil
}
