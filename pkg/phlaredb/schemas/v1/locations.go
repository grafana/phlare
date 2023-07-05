package v1

import "github.com/segmentio/parquet-go"

var locationsSchema = parquet.SchemaOf(new(InMemoryLocation))

type LocationPersister struct{}

func (*LocationPersister) Name() string { return "locations" }

func (*LocationPersister) Schema() *parquet.Schema { return locationsSchema }

func (*LocationPersister) SortingColumns() parquet.SortingOption { return parquet.SortingColumns() }

func (*LocationPersister) Deconstruct(row parquet.Row, _ uint64, l *InMemoryLocation) parquet.Row {
	// TODO: Preserve proto fields order
	// 	Idx Name      Type
	// 	0   Id        uint64
	// 	1   MappingId uint64
	// 	2   Address   uint64
	// 	3   Line      []*Line
	// 	4   IsFolded  bool
	row = locationsSchema.Deconstruct(row, l)
	return row
}

func (*LocationPersister) Reconstruct(row parquet.Row) (uint64, *InMemoryLocation, error) {
	var location InMemoryLocation
	if err := locationsSchema.Reconstruct(&location, row); err != nil {
		return 0, nil, err
	}
	return 0, &location, nil
}

type InMemoryLocation struct {
	// Unique nonzero id for the location.  A profile could use
	// instruction addresses or any integer sequence as ids.
	Id uint64
	// The instruction address for this location, if available.  It
	// should be within [Mapping.memory_start...Mapping.memory_limit]
	// for the corresponding mapping. A non-leaf address may be in the
	// middle of a call instruction. It is up to display tools to find
	// the beginning of the instruction if necessary.
	Address uint64
	// The id of the corresponding profile.Mapping for this location.
	// It can be unset if the mapping is unknown or not applicable for
	// this profile type.
	MappingId uint32
	// Provides an indication that multiple symbols map to this location's
	// address, for example due to identical code folding by the linker. In that
	// case the line information above represents one of the multiple
	// symbols. This field must be recomputed when the symbolization state of the
	// profile changes.
	IsFolded bool
	// Multiple line indicates this location has inlined functions,
	// where the last entry represents the caller into which the
	// preceding entries were inlined.
	//
	// E.g., if memcpy() is inlined into printf:
	//
	//	line[0].function_name == "memcpy"
	//	line[1].function_name == "printf"
	Line []InMemoryLine
}

type InMemoryLine struct {
	// The id of the corresponding profile.Function for this line.
	FunctionId uint32
	// Line number in source code.
	Line int32
}
