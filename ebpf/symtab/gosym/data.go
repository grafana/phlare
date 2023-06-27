package gosym

import (
	bufra "github.com/avvmoto/buf-readerat"
	"io"
	"os"
)

type PCLNData interface {
	ReadAt(data []byte, offset int) error
}

type MemPCLNData struct {
	Data []byte
}

func (m MemPCLNData) ReadAt(data []byte, offset int) error {
	copy(data, m.Data[offset:])
	return nil
}

type FilePCLNData struct {
	file io.ReaderAt
	//buf       []byte
	//bufOffset int
	offset int
}

func NewFilePCLNData(f *os.File, offset int) *FilePCLNData {
	return &FilePCLNData{
		file: bufra.NewBufReaderAt(f, 0x1000),
		//file:   f,
		offset: offset,
		//buf:    make([]byte, 16*0x1000),
	}
}

func (f *FilePCLNData) ReadAt(data []byte, offset int) error {
	n, err := f.file.ReadAt(data, int64(offset+f.offset))
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.EOF
	}
	return nil
}

type BufferedReaderAt struct {
	f       io.ReaderAt
	buf     []byte
	bufSize int
	offset  int
}

func (b *BufferedReaderAt) ReadAt(p []byte, ifrom int64) (n int, err error) {
	from := int(ifrom)
	to := from + len(p)
	if b.offset != -1 && from >= b.offset && to <= b.offset+len(b.buf) {
		//todo EOF case
		copy(p, b.buf[from-b.offset:])
		return len(p), nil
	}
	newBufferOffset := from / len(b.buf) * len(b.buf)
	newBufferEnd := newBufferOffset + len(b.buf)
	if to > newBufferEnd {
		// reading cross buffer boundary, not interested
		return b.f.ReadAt(p, ifrom)
	}
	b.buf = b.buf[:cap(b.buf)]
	n, err = b.f.ReadAt(b.buf, int64(newBufferOffset))
	b.buf = b.buf[:n]
	b.offset = newBufferOffset
	if err != nil {
		return 0, err
	}
	if b.offset != -1 && from >= b.offset && to <= b.offset+len(b.buf) {
		//todo EOF case
		copy(p, b.buf[from-b.offset:])
		return len(p), nil
	}
	panic("hui")
	return 0, nil
}
