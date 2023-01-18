package query

import (
	"strings"

	pq "github.com/segmentio/parquet-go"
)

type Source interface {
	Name() string             // Name() returns the name of the table.
	Root() *pq.Column         // Root() returns the root column of the table including all column indexes
	Schema() *pq.Schema       // Schema() returns the schema of the table.
	RowGroups() []pq.RowGroup // RowGroups() returns the current available row groups.
	NumRows() int64
	Size() int64
}

func GetColumnIndexByPath(source Source, s string) (index, depth int) {
	colSelector := strings.Split(s, ".")
	leafColumn, found := source.Schema().Lookup(colSelector...)
	if !found {
		return -1, -1
	}
	return leafColumn.ColumnIndex, leafColumn.MaxDefinitionLevel
}

func HasColumn(source Source, s string) bool {
	index, _ := GetColumnIndexByPath(source, s)
	return index >= 0
}
