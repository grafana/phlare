package symbols

import (
	"fmt"
	"testing"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/memory"
)

func Test_foo(t *testing.T) {
	pool := memory.NewGoAllocator()
	dictBuilder := array.NewDictionaryBuilder(pool, &arrow.DictionaryType{IndexType: arrow.PrimitiveTypes.Int32, ValueType: arrow.BinaryTypes.String})
	defer dictBuilder.Release()

	// dictBuilder.AppendValueFromString(string)

	dictBuilder.NewDictionaryArray()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "ID", Type: arrow.PrimitiveTypes.Uint64},
		{Name: "Functions", Type: arrow.ListOf(
			arrow.StructOf(
				arrow.Field{
					Name: "Name",
					Type: arrow.RunEndEncodedOf(arrow.PrimitiveTypes.Int16, arrow.BinaryTypes.String),
				},
				arrow.Field{
					Name: "FileName",
					Type: arrow.RunEndEncodedOf(arrow.PrimitiveTypes.Int16, arrow.BinaryTypes.String),
				},
				arrow.Field{
					Name:     "Line",
					Type:     arrow.PrimitiveTypes.Uint64,
					Nullable: true,
				},
			),
		)},
	}, nil)

	// array.NewDictionaryArray(arrow.BinaryTypes.String, arrow.Array)
	b := array.NewRecordBuilder(pool, schema)
	defer b.Release()

	b.Field(0).(*array.Uint64Builder).AppendValues([]uint64{1, 2, 3, 4, 5, 6}, nil)

	functionsBuilder := b.Field(1).(*array.ListBuilder).ValueBuilder().(*array.StructBuilder)
	nameBuilder := functionsBuilder.FieldBuilder(0).(*array.RunEndEncodedBuilder)
	filenameBuilder := functionsBuilder.FieldBuilder(1).(*array.RunEndEncodedBuilder)

	lineBuilder := functionsBuilder.FieldBuilder(2).(*array.Uint64Builder)

	for i := 0; i < 6; i++ {
		b.Field(1).(*array.ListBuilder).Append(true)

		functionsBuilder.Reserve(6)

		functionsBuilder.Append(true)

		nameBuilder.AppendValueFromString("foo1")
		filenameBuilder.AppendValueFromString("foo1.bar")
		lineBuilder.Append(100)

		functionsBuilder.Append(true)

		nameBuilder.AppendValueFromString("foo1")
		filenameBuilder.AppendValueFromString("foo1.bar")
		lineBuilder.Append(10)

		functionsBuilder.Append(true)

		nameBuilder.AppendValueFromString("foo3")
		filenameBuilder.AppendValueFromString("foo3.bar")
		lineBuilder.Append(10)

		functionsBuilder.Append(true)

		nameBuilder.AppendValueFromString("foo3")
		filenameBuilder.AppendValueFromString("foo3.bar")
		lineBuilder.Append(10)

		functionsBuilder.Append(true)

		nameBuilder.AppendValueFromString("foo3")
		filenameBuilder.AppendValueFromString("foo3.bar")
		lineBuilder.Append(10)

		functionsBuilder.Append(true)

		nameBuilder.AppendValueFromString("fo")
		filenameBuilder.AppendValueFromString("fo.bar")
		lineBuilder.Append(10)
	}
	rec := b.NewRecord()
	defer rec.Release()

	for i, col := range rec.Columns() {
		fmt.Printf("column[%d] %q: %v\n", i, rec.ColumnName(i), col)
	}
}
