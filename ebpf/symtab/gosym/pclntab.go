// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// copied from here
// https://github.com/golang/go/blob/go1.20.5/src/debug/gosym/pclntab.go
// modified go12Funcs function to be exported return a FlatFuncIndex instead of []Func
// added FuncNameOffset to export the funcnametabOffset

/*
 * Line tables
 */

package gosym

import (
	"encoding/binary"
	"sync"
)

// version of the pclntab
type version int

const (
	verUnknown version = iota
	ver11
	ver12
	ver116
	ver118
	ver120
)

// A LineTable is a data structure mapping program counters to line numbers.
//
// In Go 1.1 and earlier, each function (represented by a Func) had its own LineTable,
// and the line number corresponded to a numbering of all source lines in the
// program, across all files. That absolute line number would then have to be
// converted separately to a file name and line number within the file.
//
// In Go 1.2, the format of the data changed so that there is a single LineTable
// for the entire program, shared by all Funcs, and there are no absolute line
// numbers, just line numbers within specific files.
//
// For the most part, LineTable's methods should be treated as an internal
// detail of the package; callers should use the methods on Table instead.
type LineTable struct {
	Data []byte
	PC   uint64
	Line int

	// This mutex is used to keep parsing of pclntab synchronous.
	mu sync.Mutex

	// Contains the version of the pclntab section.
	version version

	// Go 1.2/1.16/1.18 state
	binary      binary.ByteOrder
	quantum     uint32
	ptrsize     uint32
	textStart   uint64 // address of runtime.text symbol (1.18+)
	funcnametab []byte
	cutab       []byte
	funcdata    []byte
	functab     []byte
	nfunctab    uint32
	filetab     []byte
	pctab       []byte // points to the pctables.
	nfiletab    uint32

	funcnametabOffset uint64
	failed            bool
}

// NewLineTable returns a new PC/line table
// corresponding to the encoded data.
// Text must be the start address of the
// corresponding text segment.
func NewLineTable(data []byte, text uint64) *LineTable {
	return &LineTable{
		Data: data,
		PC:   text,
		Line: 0,
	}
}

// Go 1.2 symbol table format.
// See golang.org/s/go12symtab.
//
// A general note about the methods here: rather than try to avoid
// index out of bounds errors, we trust Go to detect them, and then
// we recover from the panics and treat them as indicative of a malformed
// or incomplete table.
//
// The methods called by symtab.go, which begin with "go12" prefixes,
// are expected to have that recovery logic.

// IsGo12 reports whether this is a Go 1.2 (or later) symbol table.
func (t *LineTable) IsGo12() bool {
	t.parsePclnTab()
	return t.version >= ver12
}

const (
	go12magic  = 0xfffffffb
	go116magic = 0xfffffffa
	go118magic = 0xfffffff0
	go120magic = 0xfffffff1
)

// uintptr returns the pointer-sized value encoded at b.
// The pointer size is dictated by the table being read.
func (t *LineTable) uintptr(b []byte) uint64 {
	if t.ptrsize == 4 {
		return uint64(t.binary.Uint32(b))
	}
	return t.binary.Uint64(b)
}

// parsePclnTab parses the pclntab, setting the version.
func (t *LineTable) parsePclnTab() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.version != verUnknown {
		return
	}

	// Note that during this function, setting the version is the last thing we do.
	// If we set the version too early, and parsing failed (likely as a panic on
	// slice lookups), we'd have a mistaken version.
	//
	// Error paths through this code will default the version to 1.1.
	t.version = ver11

	if !disableRecover {
		defer func() {
			// If we panic parsing, assume it's a Go 1.1 pclntab.
			if r := recover(); r != nil {
				t.failed = true
			}
		}()
	}

	// Check header: 4-byte magic, two zeros, pc quantum, pointer size.
	if len(t.Data) < 16 || t.Data[4] != 0 || t.Data[5] != 0 ||
		(t.Data[6] != 1 && t.Data[6] != 2 && t.Data[6] != 4) || // pc quantum
		(t.Data[7] != 4 && t.Data[7] != 8) { // pointer size

		return
	}

	var possibleVersion version
	leMagic := binary.LittleEndian.Uint32(t.Data)
	beMagic := binary.BigEndian.Uint32(t.Data)
	switch {
	case leMagic == go12magic:
		t.binary, possibleVersion = binary.LittleEndian, ver12
	case beMagic == go12magic:
		t.binary, possibleVersion = binary.BigEndian, ver12
	case leMagic == go116magic:
		t.binary, possibleVersion = binary.LittleEndian, ver116
	case beMagic == go116magic:
		t.binary, possibleVersion = binary.BigEndian, ver116
	case leMagic == go118magic:
		t.binary, possibleVersion = binary.LittleEndian, ver118
	case beMagic == go118magic:
		t.binary, possibleVersion = binary.BigEndian, ver118
	case leMagic == go120magic:
		t.binary, possibleVersion = binary.LittleEndian, ver120
	case beMagic == go120magic:
		t.binary, possibleVersion = binary.BigEndian, ver120
	default:
		return
	}
	t.version = possibleVersion

	// quantum and ptrSize are the same between 1.2, 1.16, and 1.18
	t.quantum = uint32(t.Data[6])
	t.ptrsize = uint32(t.Data[7])

	offset := func(word uint32) uint64 {
		return t.uintptr(t.Data[8+word*t.ptrsize:])
	}
	data := func(word uint32) []byte {
		return t.Data[offset(word):]
	}

	switch possibleVersion {
	case ver118, ver120:
		t.nfunctab = uint32(offset(0))
		t.nfiletab = uint32(offset(1))
		t.textStart = t.PC // use the start PC instead of reading from the table, which may be unrelocated
		t.funcnametab = data(3)
		t.funcnametabOffset = offset(3)
		t.cutab = data(4)
		t.filetab = data(5)
		t.pctab = data(6)
		t.funcdata = data(7)
		t.functab = data(7)
		functabsize := (int(t.nfunctab)*2 + 1) * t.functabFieldSize()
		t.functab = t.functab[:functabsize]
	case ver116:
		t.nfunctab = uint32(offset(0))
		t.nfiletab = uint32(offset(1))
		t.funcnametab = data(2)
		t.funcnametabOffset = offset(2)
		t.cutab = data(3)
		t.filetab = data(4)
		t.pctab = data(5)
		t.funcdata = data(6)
		t.functab = data(6)
		functabsize := (int(t.nfunctab)*2 + 1) * t.functabFieldSize()
		t.functab = t.functab[:functabsize]
	case ver12:
		t.nfunctab = uint32(t.uintptr(t.Data[8:]))
		t.funcdata = t.Data
		t.funcnametab = t.Data
		t.functab = t.Data[8+t.ptrsize:]
		t.pctab = t.Data
		functabsize := (int(t.nfunctab)*2 + 1) * t.functabFieldSize()
		fileoff := t.binary.Uint32(t.functab[functabsize:])
		t.functab = t.functab[:functabsize]
		t.filetab = t.Data[fileoff:]
		t.nfiletab = t.binary.Uint32(t.filetab)
		t.filetab = t.filetab[:t.nfiletab*4]
	default:
		panic("unreachable")
	}
}

// FlatFuncIndex
// Entry contains a sorted array of function entry address, which is ued for binary search.
// Name contains offsets into funcnametab, which is located in the .gopclntab section.
type FlatFuncIndex struct {
	Entry PCIndex
	Name  []uint32
	End   uint64
}

// Go12Funcs returns a slice of Funcs derived from the Go 1.2+ pcln table.
func (t *LineTable) Go12Funcs() (res FlatFuncIndex) {
	// Assume it is malformed and return nil on error.
	if !disableRecover {
		defer func() {
			err := recover()
			if err != nil {
				res = FlatFuncIndex{}
			}
		}()
	}

	ft := t.funcTab()
	nfunc := ft.Count()
	res = FlatFuncIndex{
		Entry: NewPCIndex(nfunc),
		Name:  make([]uint32, nfunc),
	}
	for i := 0; i < nfunc; i++ {
		entry := ft.pc(i)
		res.Entry.Set(i, entry)
		res.End = ft.pc(i + 1)
		info := t.funcData(uint32(i))
		res.Name[i] = info.nameOff()
	}
	return
}

// functabFieldSize returns the size in bytes of a single functab field.
func (t *LineTable) functabFieldSize() int {
	if t.version >= ver118 {
		return 4
	}
	return int(t.ptrsize)
}

// funcTab returns t's funcTab.
func (t *LineTable) funcTab() funcTab {
	return funcTab{LineTable: t, sz: t.functabFieldSize()}
}

// funcTab is memory corresponding to a slice of functab structs, followed by an invalid PC.
// A functab struct is a PC and a func offset.
type funcTab struct {
	*LineTable
	sz int // cached result of t.functabFieldSize
}

// Count returns the number of func entries in f.
func (f funcTab) Count() int {
	return int(f.nfunctab)
}

// pc returns the PC of the i'th func in f.
func (f funcTab) pc(i int) uint64 {
	u := f.uint(f.functab[2*i*f.sz:])
	if f.version >= ver118 {
		u += f.textStart
	}
	return u
}

// funcOff returns the funcdata offset of the i'th func in f.
func (f funcTab) funcOff(i int) uint64 {
	return f.uint(f.functab[(2*i+1)*f.sz:])
}

// uint returns the uint stored at b.
func (f funcTab) uint(b []byte) uint64 {
	if f.sz == 4 {
		return uint64(f.binary.Uint32(b))
	}
	return f.binary.Uint64(b)
}

// funcData is memory corresponding to an _func struct.
type funcData struct {
	t    *LineTable // LineTable this data is a part of
	data []byte     // raw memory for the function
}

// funcData returns the ith funcData in t.functab.
func (t *LineTable) funcData(i uint32) funcData {
	data := t.funcdata[t.funcTab().funcOff(int(i)):]
	return funcData{t: t, data: data}
}

// IsZero reports whether f is the zero value.
func (f funcData) IsZero() bool {
	return f.t == nil && f.data == nil
}

func (f funcData) nameOff() uint32 { return f.field(1) }

// field returns the nth field of the _func struct.
// It panics if n == 0 or n > 9; for n == 0, call f.entryPC.
// Most callers should use a named field accessor (just above).
func (f funcData) field(n uint32) uint32 {
	if n == 0 || n > 9 {
		panic("bad funcdata field")
	}
	// In Go 1.18, the first field of _func changed
	// from a uintptr entry PC to a uint32 entry offset.
	sz0 := f.t.ptrsize
	if f.t.version >= ver118 {
		sz0 = 4
	}
	off := sz0 + (n-1)*4 // subsequent fields are 4 bytes each
	data := f.data[off:]
	return f.t.binary.Uint32(data)
}

func (t *LineTable) IsFailed() bool {
	return t.failed
}

func (t *LineTable) FuncNameOffset() uint64 {
	return t.funcnametabOffset
}

// disableRecover causes this package not to swallow panics.
// This is useful when making changes.
const disableRecover = false
