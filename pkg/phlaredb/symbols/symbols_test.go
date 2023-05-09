package symbols

import (
	"fmt"
	"testing"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/memory"
)

func Test_foo(t *testing.T) {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "ID", Type: arrow.PrimitiveTypes.Uint64},
		{Name: "Functions", Type: arrow.ListOf(
			arrow.StructOf(
				arrow.Field{
					Name: "Names",
					Type: &arrow.DictionaryType{IndexType: &arrow.Int64Type{}, ValueType: arrow.BinaryTypes.String},
				},
				arrow.Field{
					Name:     "Line",
					Type:     arrow.ListOf(arrow.PrimitiveTypes.Int16),
					Nullable: true,
				},
			),
		)},
	}, nil)

	// array.NewDictionaryArray(arrow.BinaryTypes.String, arrow.Array)
	b := array.NewRecordBuilder(pool, schema)
	defer b.Release()

	b.Field(0).(*array.Uint64Builder).AppendValues([]uint64{1, 2, 3, 4, 5, 6}, nil)

	rec := b.NewRecord()
	defer rec.Release()

	for i, col := range rec.Columns() {
		fmt.Printf("column[%d] %q: %v\n", i, rec.ColumnName(i), col)
	}
}
