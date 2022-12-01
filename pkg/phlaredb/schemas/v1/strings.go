package v1

import (
	"github.com/segmentio/parquet-go"

	phlareparquet "github.com/grafana/phlare/pkg/parquet"
)

var stringsSchema = parquet.NewSchema("String", phlareparquet.Group{
	phlareparquet.NewGroupField("ID", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
	phlareparquet.NewGroupField("String", parquet.Encoded(parquet.String(), &parquet.RLEDictionary)),
})

type String struct {
	ID     uint64 `parquet:",delta"`
	String string `parquet:",dict"`
}

type StringPersister struct{}

func StringsFromStringSlice(strings []string) []*String {
	sl := make([]String, len(strings))
	slp := make([]*String, len(strings))
	for i, s := range strings {
		sl[i].String = s
		slp[i] = &sl[i]
	}
	return slp
}

func (*StringPersister) Name() string {
	return "strings"
}

func (*StringPersister) Schema() *parquet.Schema {
	return stringsSchema
}

func (*StringPersister) SortingColumns() parquet.SortingOption {
	return parquet.SortingColumns(
		parquet.Ascending("ID"),
		parquet.Ascending("String"),
	)
}

func (*StringPersister) Deconstruct(row parquet.Row, id uint64, s *String) parquet.Row {
	var stored String
	stored.ID = id
	stored.String = s.String
	row = stringsSchema.Deconstruct(row, &stored)
	return row
}

func (*StringPersister) Reconstruct(row parquet.Row) (id uint64, s *String, err error) {
	if err := stringsSchema.Reconstruct(&s, row); err != nil {
		return 0, nil, err
	}
	return s.ID, s, nil
}
