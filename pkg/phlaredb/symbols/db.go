package symbols

import (
	"io"
	"os"
	"sort"

	googlev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	phlareparquet "github.com/grafana/phlare/pkg/parquet"
	"github.com/segmentio/parquet-go"
)

const (
	stacktraceIDFieldName = "ID"
)

var stracktraceSchema = parquet.NewSchema("stacktraces", phlareparquet.Group{
	// todo bloom filter on ID parquet.BloomFilters(filters ...parquet.BloomFilterColumn)
	phlareparquet.NewGroupField(stacktraceIDFieldName, parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
	phlareparquet.NewGroupField("Locations", parquet.List(phlareparquet.Group{
		phlareparquet.NewGroupField("Address", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
		phlareparquet.NewGroupField("IsFolded", parquet.Leaf(parquet.BooleanType)),
		phlareparquet.NewGroupField("Mapping", phlareparquet.Group{
			phlareparquet.NewGroupField("MemoryStart", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
			phlareparquet.NewGroupField("MemoryLimit", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
			phlareparquet.NewGroupField("FileOffset", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
			phlareparquet.NewGroupField("Filename", parquet.Encoded(parquet.String(), &parquet.RLEDictionary)),
			phlareparquet.NewGroupField("BuildID", parquet.Encoded(parquet.String(), &parquet.RLEDictionary)),
			phlareparquet.NewGroupField("HasFunctions", parquet.Leaf(parquet.BooleanType)),
			phlareparquet.NewGroupField("HasFilenames", parquet.Leaf(parquet.BooleanType)),
			phlareparquet.NewGroupField("HasLineNumbers", parquet.Leaf(parquet.BooleanType)),
			phlareparquet.NewGroupField("HasInlineFrames", parquet.Leaf(parquet.BooleanType)),
		}),
		phlareparquet.NewGroupField("Functions", parquet.List(phlareparquet.Group{
			phlareparquet.NewGroupField("Line", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
			phlareparquet.NewGroupField("Name", parquet.Encoded(parquet.String(), &parquet.RLEDictionary)),
			phlareparquet.NewGroupField("SystemName", parquet.Encoded(parquet.String(), &parquet.RLEDictionary)),
			phlareparquet.NewGroupField("Filename", parquet.Encoded(parquet.String(), &parquet.RLEDictionary)),
			phlareparquet.NewGroupField("StartLine", parquet.Encoded(parquet.Uint(64), &parquet.DeltaBinaryPacked)),
		})),
	})),
})

type Stacktrace struct {
	ID        uint64     // unique global id
	Locations []Location `parquet:",list"`
}

type Location struct {
	Mapping   Mapping
	Address   uint64
	Functions []Function `parquet:",list"`
	IsFolded  bool
}

type Mapping struct {
	MemoryStart     uint64
	MemoryLimit     uint64
	FileOffset      uint64
	Filename        string
	BuildID         string
	HasFunctions    bool
	HasFilenames    bool
	HasLineNumbers  bool
	HasInlineFrames bool
}

type Function struct {
	Line       int64
	Name       string
	SystemName string
	Filename   string
	StartLine  int64
}

// todo rename buffer ?
type store struct {
	buffer        *parquet.GenericBuffer[Stacktrace]
	reader        *parquet.GenericReader[Stacktrace]
	columnIDIndex int
}

func NewStore() *store {
	buffer := parquet.NewGenericBuffer[Stacktrace](stracktraceSchema)

	return &store{
		buffer:        buffer,
		reader:        parquet.NewGenericRowGroupReader[Stacktrace](buffer, stracktraceSchema),
		columnIDIndex: getColumnIDIndex(),
	}
}

// todo return new stacktrace id
func (s *store) Ingest(prof *googlev1.Profile) error {
	rows := make([]Stacktrace, 0, len(prof.Sample))
	for _, sample := range prof.Sample {
		st := Stacktrace{
			ID:        1, // todo compute id
			Locations: make([]Location, 0, len(sample.LocationId)),
		}
		for _, locationID := range sample.LocationId {
			location := prof.Location[locationID-1]
			mapping := prof.Mapping[location.MappingId-1]
			newLocation := Location{
				Mapping: Mapping{
					MemoryStart:     mapping.MemoryStart,
					MemoryLimit:     mapping.MemoryLimit,
					FileOffset:      mapping.FileOffset,
					Filename:        prof.StringTable[mapping.Filename],
					BuildID:         prof.StringTable[mapping.BuildId],
					HasFunctions:    mapping.HasFunctions,
					HasFilenames:    mapping.HasFilenames,
					HasLineNumbers:  mapping.HasLineNumbers,
					HasInlineFrames: mapping.HasInlineFrames,
				},
				Address:   location.Address,
				IsFolded:  location.IsFolded,
				Functions: make([]Function, 0, len(location.Line)),
			}
			// convert functions
			for _, line := range location.Line {
				function := prof.Function[line.FunctionId-1]
				newLocation.Functions = append(newLocation.Functions, Function{
					Line:       line.Line,
					Name:       prof.StringTable[function.Name],
					SystemName: prof.StringTable[function.SystemName],
					Filename:   prof.StringTable[function.Filename],
					StartLine:  function.StartLine,
				})
			}
			st.Locations = append(st.Locations, newLocation)
		}
		rows = append(rows, st)
	}
	// todo search for existing stacktraces and remove then from the rows
	_, err := s.buffer.Write(rows)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) Size() int64 {
	return s.buffer.Size()
}

func (s *store) Flush(filename string) error {
	sort.Sort(s.buffer)
	defer s.buffer.Reset()

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := parquet.NewWriter(f) // todo add sorting and bloom filter
	parquet.CopyRows(writer, s.buffer.Rows())

	return writer.Close()
}

func (s *store) GetStacktrace(id uint64) (Stacktrace, bool, error) {
	var (
		IDColumn = s.buffer.ColumnBuffers()[s.columnIDIndex]
		pages    = IDColumn.Pages()
		values   = make([]parquet.Value, 1024)
	)
	defer pages.Close()
	var (
		err  error
		page parquet.Page
	)
	var rowNum int64
	for {
		page, err = pages.ReadPage()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Stacktrace{}, false, err
		}
		reader := page.Values()
		for {
			n, err := reader.ReadValues(values)
			if err == io.EOF {
				break
			}
			if err != nil {
				return Stacktrace{}, false, err
			}
			for _, value := range values[:n] {
				rowNum++
				if value.Uint64() == id {
					// s.reader.
				}
			}
		}

	}
	return Stacktrace{}, false, nil
}

// getStacktraceRowNum returns stracktrace row numbers for the given ids
func (s *store) getStacktraceRowNum(ids []uint64) ([]int64, error) {
	var (
		IDColumn = s.buffer.ColumnBuffers()[s.columnIDIndex]
		pages    = IDColumn.Pages()
		result   = make([]int64, len(ids))
		rowNum   int64
	)
	defer pages.Close()
	ScanPages(pages, func(v parquet.Value) bool {
		value := v.Uint64()
		for i, id := range ids {
			if value == id {
				// todo stop early if we found all ids
				result[i] = rowNum
			}
		}
		rowNum++
		return true
	})
	return result, nil
}

func ScanPages(pages parquet.Pages, f func(parquet.Value) bool) error {
	for {
		page, err := pages.ReadPage()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := forAllValues(page.Values(), f); err != nil {
			return err
		}
	}
	return nil
}

func forAllValues(reader parquet.ValueReader, f func(parquet.Value) bool) error {
	values := make([]parquet.Value, 1024) // todo pooling
	for {
		n, err := reader.ReadValues(values)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		for _, value := range values[:n] {
			if !f(value) {
				return nil
			}
		}
	}
	return nil
}

func getColumnIDIndex() int {
	leaf, found := stracktraceSchema.Lookup(stacktraceIDFieldName)
	if !found {
		panic("could not find stacktrace id field")
	}
	return leaf.ColumnIndex
}
