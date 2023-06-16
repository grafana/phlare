package symdb

import (
	"bufio"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
)

const (
	DefaultFileName = "symbols.symdb" // "index.symdb"
	// TODO(kolesnikovae): Should we maybe store stacktraces file separately?
	tmpStacktracesFileName = "stacktraces.symdb.tmp"
)

type Writer struct {
	dir string

	header Header
	toc    TOC
	sch    StacktraceChunkHeaders
	scd    *fileWriter

	crc hash.Hash32
}

func NewWriter(dir string) *Writer {
	return &Writer{
		dir: dir,
		toc: TOC{
			Entries: make([]TOCEntry, tocEntries),
		},
		header: Header{
			Magic:   symdbMagic,
			Version: FormatV1,
		},
	}
}

func (w *Writer) writeStacktraceChunk(c *stacktraceChunk) (err error) {
	if w.scd == nil {
		p := filepath.Join(w.dir, tmpStacktracesFileName)
		if w.scd, err = newFileWriter(p); err != nil {
			return err
		}
	}
	h := StacktraceChunkHeader{
		Offset:             w.scd.off,
		Size:               0, // Set later.
		MappingName:        c.mapping.name,
		Stacktraces:        0, // TODO
		StacktraceNodes:    c.tree.len(),
		StacktraceMaxDepth: 0, // TODO
		StacktraceMaxNodes: c.mapping.maxNodesPerChunk,
		CRC:                0, // Set later.
	}
	w.crc.Reset()
	if h.Size, err = c.WriteTo(io.MultiWriter(w.crc, w.scd)); err != nil {
		return fmt.Errorf("writing stacktrace chunk data: %w", err)
	}
	h.CRC = w.crc.Sum32()
	w.sch.Entries = append(w.sch.Entries, h)
	return nil
}

func (w *Writer) WriteTo(dst io.Writer) (n int64, err error) {
	var m int
	header, _ := w.header.MarshalBinary()
	if m, err = dst.Write(header); err != nil {
		return n + int64(m), fmt.Errorf("header write: %w", err)
	}

	// Make sure all the entries have proper offsets.
	w.toc.Entries[tocEntryStacktraceChunkHeaders] = TOCEntry{
		Offset: w.dataOffset(),
		Size:   w.sch.Size(),
	}
	w.toc.Entries[tocEntryStacktraceChunkData] = TOCEntry{
		Offset: w.stacktraceChunkHeaderOffset(),
		Size:   w.scd.off,
	}
	w.shiftStacktraceChunkHeaderOffsets()
	toc, _ := w.toc.MarshalBinary()
	if m, err = dst.Write(toc); err != nil {
		return n + int64(m), fmt.Errorf("toc write: %w", err)
	}

	sch, _ := w.sch.MarshalBinary()
	if m, err = dst.Write(sch); err != nil {
		return n + int64(m), fmt.Errorf("stacktrace chunk headers: %w", err)
	}

	err = w.rewriteStacktraceChunks(dst)
	return n, err
}

func (w *Writer) dataOffset() int64 {
	return int64(headerSize + tocEntrySize*tocEntries)
}

func (w *Writer) stacktraceChunkHeaderOffset() int64 {
	return w.dataOffset() + w.scd.off
}

func (w *Writer) shiftStacktraceChunkHeaderOffsets() {
	offset := w.stacktraceChunkHeaderOffset()
	for i := range w.sch.Entries {
		w.sch.Entries[i].Offset += offset
	}
}

func (w *Writer) rewriteStacktraceChunks(dst io.Writer) (err error) {
	if _, err = w.scd.WriteTo(dst); err != nil {
		return fmt.Errorf("flushing stacktrace chunk data: %w", err)
	}
	if err = w.scd.remove(); err != nil {
		return fmt.Errorf("removing flushing stacktrace chunk data: %w", err)
	}
	return nil
}

type fileWriter struct {
	name string
	buf  *bufio.Writer
	f    *os.File
	off  int64
}

func newFileWriter(name string) (*fileWriter, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	w := fileWriter{
		f: f,
		// There is no particular reason to use
		// a buffer larger than the default 4K.
		buf:  bufio.NewWriterSize(f, 4<<10),
		name: name,
	}
	return &w, nil
}

func (f *fileWriter) Write(p []byte) (n int, err error) {
	n, err = f.buf.Write(p)
	f.off += int64(n)
	return n, err
}

func (f *fileWriter) sync() (err error) {
	if err = f.buf.Flush(); err != nil {
		return err
	}
	return f.f.Sync()
}

func (f *fileWriter) Close() (err error) {
	if err = f.sync(); err != nil {
		return err
	}
	return f.f.Close()
}

func (f *fileWriter) WriteTo(w io.Writer) (n int64, err error) {
	if err = f.sync(); err != nil {
		return n, err
	}
	if _, err = f.f.Seek(0, io.SeekStart); err != nil {
		return n, err
	}
	// We expect that w implements io.ReaderFrom, thus avoiding
	// the buffer allocation (potentially).
	return io.Copy(w, f.f)
}

func (f *fileWriter) ReadFrom(r io.Reader) (n int64, err error) {
	// Ensure disk and memory states are synchronised, before
	// writing to the file directly.
	if err = f.sync(); err != nil {
		return n, err
	}
	// os.File does satisfy the io.ReadFrom interface, however
	// in most cases (all but Linux), it uses io.Copy internally,
	// and the buffer of the default size is allocated either way.
	if n, err = f.f.ReadFrom(r); err != nil {
		return n, err
	}
	f.off += n
	return n, nil
}

func (f *fileWriter) remove() (err error) {
	if err = f.Close(); err != nil {
		return err
	}
	return os.RemoveAll(f.name)
}
