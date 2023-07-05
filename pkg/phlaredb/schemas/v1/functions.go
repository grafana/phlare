package v1

import "github.com/segmentio/parquet-go"

var functionsSchema = parquet.SchemaOf(new(InMemoryFunction))

type FunctionPersister struct{}

func (*FunctionPersister) Name() string { return "functions" }

func (*FunctionPersister) Schema() *parquet.Schema { return functionsSchema }

func (*FunctionPersister) SortingColumns() parquet.SortingOption { return parquet.SortingColumns() }

func (*FunctionPersister) Deconstruct(row parquet.Row, _ uint64, l *InMemoryFunction) parquet.Row {
	row = functionsSchema.Deconstruct(row, l)
	return row
}

func (*FunctionPersister) Reconstruct(row parquet.Row) (uint64, *InMemoryFunction, error) {
	var function InMemoryFunction
	if err := functionsSchema.Reconstruct(&function, row); err != nil {
		return 0, nil, err
	}
	return 0, &function, nil
}

type InMemoryFunction struct {
	// Unique nonzero id for the function.
	Id uint64
	// Name of the function, in human-readable form if available.
	Name uint32
	// Name of the function, as identified by the system.
	// For instance, it can be a C++ mangled name.
	SystemName uint32
	// Source file containing the function.
	Filename uint32
	// Line number in source file.
	StartLine uint32
}
