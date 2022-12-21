package query

import (
	"strings"

	pq "github.com/segmentio/parquet-go"
)

type Source interface {
	// Name() returns the name of the table.
	Name() string
	Root() *pq.Column
	// Columns() returns the columns defintions.
	// Columns() []*pq.Column
	// RowGroups() returns the current available row groups.
	RowGroups() []pq.RowGroup
	NumRows() int64
	Size() int64
}

func GetColumnIndexByPath(source Source, s string) (index, depth int) {
	colSelector := strings.Split(s, ".")
	n := source.Root()
	for len(colSelector) > 0 {
		n = n.Column(colSelector[0])
		if n == nil {
			return -1, -1
		}

		colSelector = colSelector[1:]
		depth++
	}

	return n.Index(), depth
}

func HasColumn(source Source, s string) bool {
	index, _ := GetColumnIndexByPath(source, s)
	return index >= 0
}
