package symdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"unsafe"
)

// The file contains version-agnostic wire format implementation of
// the symbols database file.
//
// Layout (two-pass write):
//
// [Header] Header defines the format version and denotes the content type.
//
// [TOC]    Table of contents. Its entries refer to the Data section.
//          It is of a fixed size for a given version (number of entries).
//
// [Data]   Data is an arbitrary structured section. The exact structure is
//          defined by the TOC and Header (version, flags, etc).

const headerSize = int(unsafe.Sizeof(Header{}))

const FormatV1 = 1

const (
	// TOC entries are version-specific.
	tocEntryStacktraceChunkHeaders = iota
	tocEntryStacktraceChunkData
	tocEntries
)

// https://en.wikipedia.org/wiki/List_of_file_signatures
var symdbMagic = [4]byte{'s', 'y', 'm', '1'}

var castagnoli = crc32.MakeTable(crc32.Castagnoli)

var (
	ErrInvalidSize    = &FormatError{fmt.Errorf("invalid size")}
	ErrInvalidCRC     = &FormatError{fmt.Errorf("invalid CRC")}
	ErrInvalidMagic   = &FormatError{fmt.Errorf("invalid magic number")}
	ErrUnknownVersion = &FormatError{fmt.Errorf("unknown version")}
)

type FormatError struct{ err error }

func (e *FormatError) Error() string {
	return e.err.Error()
}

type Header struct {
	Magic    [4]byte
	Version  uint32
	Reserved [20]byte // Reserved for future use; padding to 32.
	CRC      uint32   // CRC of the header.
}

func (h *Header) MarshalBinary() ([]byte, error) {
	b := make([]byte, headerSize)
	copy(b[0:4], h.Magic[:])
	binary.BigEndian.PutUint32(b[4:8], h.Version)
	binary.BigEndian.PutUint32(b[headerSize-4:], crc32.Checksum(b[:headerSize-4], castagnoli))
	return b, nil
}

func (h *Header) UnmarshalBinary(b []byte) error {
	if len(b) != headerSize {
		return ErrInvalidSize
	}
	if h.CRC = binary.BigEndian.Uint32(b[headerSize-4:]); h.CRC != crc32.Checksum(b[:headerSize-4], castagnoli) {
		return ErrInvalidCRC
	}
	if copy(h.Magic[:], b[0:4]); !bytes.Equal(h.Magic[:], symdbMagic[:]) {
		return ErrInvalidMagic
	}
	h.Version = binary.BigEndian.Uint32(b[4:8])
	return nil
}

// Table of contents.

const (
	tocEntrySize     = int(unsafe.Sizeof(TOCEntry{}))
	tocSizeAlignment = 16 // Reserved + CRC.
)

type TOC struct {
	Entries  []TOCEntry
	Reserved [12]byte
	CRC      uint32
}

type TOCEntry struct {
	Offset int64
	Size   int64
}

func (toc *TOC) Size() int {
	return tocEntrySize*tocEntries + tocSizeAlignment
}

func (toc *TOC) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(toc.Entries)*tocEntrySize+tocSizeAlignment)
	for i := range toc.Entries {
		toc.Entries[i].marshal(b[i*tocEntrySize:])
	}
	binary.BigEndian.PutUint32(b[len(b)-4:], crc32.Checksum(b[:len(b)-4], castagnoli))
	return b, nil
}

func (toc *TOC) UnmarshalBinary(b []byte) error {
	s := len(b)
	entriesSize := s - tocSizeAlignment
	if entriesSize < tocEntrySize || entriesSize%tocEntrySize > 0 {
		return ErrInvalidSize
	}
	if toc.CRC = binary.BigEndian.Uint32(b[s-4:]); toc.CRC != crc32.Checksum(b[:s-4], castagnoli) {
		return ErrInvalidCRC
	}
	toc.Entries = make([]TOCEntry, entriesSize/tocEntrySize)
	for i := range toc.Entries {
		off := i * tocEntrySize
		toc.Entries[i].unmarshal(b[off : off+tocEntrySize])
	}
	return nil
}

func (h *TOCEntry) marshal(b []byte) {
	binary.BigEndian.PutUint64(b[0:8], uint64(h.Size))
	binary.BigEndian.PutUint64(b[8:16], uint64(h.Offset))
}

func (h *TOCEntry) unmarshal(b []byte) {
	h.Size = int64(binary.BigEndian.Uint64(b[0:8]))
	h.Offset = int64(binary.BigEndian.Uint64(b[8:16]))
}

// Types below define the Data section structure.
// Currently, the data section is as follows:
//
//   []StacaktraceChunkHeader
//   []StacaktraceChunkData

const (
	stacktraceChunkHeaderSize       = int(unsafe.Sizeof(StacktraceChunkHeader{}))
	stacktraceChunkHeadersAlignment = 32
)

type StacktraceChunkHeaders struct {
	Entries []StacktraceChunkHeader

	_   [28]byte // Reserved. Aligned to 32.
	CRC uint32
}

func (h *StacktraceChunkHeaders) Size() int64 {
	return int64(stacktraceChunkHeaderSize*len(h.Entries) + stacktraceChunkHeadersAlignment)
}

func (h *StacktraceChunkHeaders) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(h.Entries)*stacktraceChunkHeaderSize+stacktraceChunkHeadersAlignment)
	for i := range h.Entries {
		off := i * stacktraceChunkHeaderSize
		h.Entries[i].marshal(b[off : off+stacktraceChunkHeaderSize])
	}
	h.CRC = crc32.Checksum(b[stacktraceChunkHeaderSize-4:], castagnoli)
	binary.BigEndian.PutUint32(b[stacktraceChunkHeaderSize-4:], h.CRC)
	return b, nil
}

func (h *StacktraceChunkHeaders) UnmarshalBinary(b []byte) error {
	if s := len(b); s < stacktraceChunkHeadersAlignment || s%stacktraceChunkHeaderSize > 0 {
		return ErrInvalidSize
	}
	h.CRC = binary.BigEndian.Uint32(b[stacktraceChunkHeaderSize-4:])
	if crc32.Checksum(b[stacktraceChunkHeaderSize-4:], castagnoli) != h.CRC {
		return ErrInvalidCRC
	}
	h.Entries = make([]StacktraceChunkHeader, (len(b)-stacktraceChunkHeadersAlignment)/stacktraceChunkHeaderSize)
	for i := range h.Entries {
		off := i * stacktraceChunkHeaderSize
		h.Entries[i].unmarshal(b[off : off+stacktraceChunkHeaderSize])
	}
	return nil
}

type StacktraceChunkHeader struct {
	Offset int64 // Relative to the mapping offset.
	Size   int64

	MappingName uint64 // MappingName the chunk refers to.

	Stacktraces        uint32 // Number of unique stack traces in the chunk.
	StacktraceNodes    uint32 // Number of nodes in the stacktrace tree.
	StacktraceMaxDepth uint32 // Max stack trace depth in the tree.
	StacktraceMaxNodes uint32 // Max number of nodes at the time of the chunk creation.

	_   [20]byte // Padding. 64 bytes per chunk header.
	CRC uint32   // Checksum of the chunk data.
}

func (h *StacktraceChunkHeader) marshal(b []byte) {
	binary.BigEndian.PutUint64(b[0:8], uint64(h.Offset))
	binary.BigEndian.PutUint64(b[8:16], uint64(h.Size))
	binary.BigEndian.PutUint64(b[16:24], h.MappingName)
	binary.BigEndian.PutUint32(b[24:28], h.Stacktraces)
	binary.BigEndian.PutUint32(b[28:32], h.StacktraceNodes)
	binary.BigEndian.PutUint32(b[32:36], h.StacktraceMaxDepth)
	binary.BigEndian.PutUint32(b[36:40], h.StacktraceMaxNodes)
}

func (h *StacktraceChunkHeader) unmarshal(b []byte) {
	h.Offset = int64(binary.BigEndian.Uint64(b[0:8]))
	h.Size = int64(binary.BigEndian.Uint64(b[8:16]))
	h.MappingName = binary.BigEndian.Uint64(b[16:24])
	h.Stacktraces = binary.BigEndian.Uint32(b[24:28])
	h.StacktraceNodes = binary.BigEndian.Uint32(b[28:32])
	h.StacktraceMaxDepth = binary.BigEndian.Uint32(b[32:36])
	h.StacktraceMaxNodes = binary.BigEndian.Uint32(b[36:40])
}
