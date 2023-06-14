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

// https://en.wikipedia.org/wiki/List_of_file_signatures
var symdbMagic = [4]byte{'s', 'y', 'm', '1'}

var castagnoli = crc32.MakeTable(crc32.Castagnoli)

var (
	ErrInvalidSize  = fmt.Errorf("invalid size")
	ErrInvalidCRC   = fmt.Errorf("invalid CRC")
	ErrInvalidMagic = fmt.Errorf("invalid magic number")
)

type Header struct {
	Magic    [4]byte
	Version  uint32
	Reserved [20]byte // Reserved for future use; padding to 32.
	CRC      uint32   // CRC of the header.
}

func (h *Header) MarshalBinary() ([]byte, error) {
	b := make([]byte, headerSize)
	copy(b[0:4], h.Magic[:])
	binary.LittleEndian.PutUint32(b[4:8], h.Version)
	binary.LittleEndian.PutUint32(b[headerSize-4:], crc32.Checksum(b[:headerSize-4], castagnoli))
	return b, nil
}

func (h *Header) UnmarshalBinary(b []byte) error {
	if len(b) != headerSize {
		return ErrInvalidSize
	}
	if h.CRC = binary.LittleEndian.Uint32(b[headerSize-4:]); h.CRC != crc32.Checksum(b[:headerSize-4], castagnoli) {
		return ErrInvalidCRC
	}
	if copy(h.Magic[:], b[0:4]); !bytes.Equal(h.Magic[:], symdbMagic[:]) {
		return ErrInvalidMagic
	}
	h.Version = binary.LittleEndian.Uint32(b[4:8])
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

func (toc *TOC) MarshalBinary() ([]byte, error) {
	b := make([]byte, len(toc.Entries)*tocEntrySize+tocSizeAlignment)
	for i := range toc.Entries {
		toc.Entries[i].marshal(b[i*tocEntrySize:])
	}
	binary.LittleEndian.PutUint32(b[len(b)-4:], crc32.Checksum(b[:len(b)-4], castagnoli))
	return b, nil
}

func (toc *TOC) UnmarshalBinary(b []byte) error {
	s := len(b)
	entriesSize := s - tocSizeAlignment
	if entriesSize < tocEntrySize || entriesSize%tocEntrySize > 0 {
		return ErrInvalidSize
	}
	if toc.CRC = binary.LittleEndian.Uint32(b[s-4:]); toc.CRC != crc32.Checksum(b[:s-4], castagnoli) {
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
	binary.LittleEndian.PutUint64(b[0:8], uint64(h.Size))
	binary.LittleEndian.PutUint64(b[8:16], uint64(h.Offset))
	return
}

func (h *TOCEntry) unmarshal(b []byte) {
	h.Size = int64(binary.LittleEndian.Uint64(b[0:8]))
	h.Offset = int64(binary.LittleEndian.Uint64(b[8:16]))
	return
}
