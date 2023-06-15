package symdb

import (
	"io"
)

var (
	_ MappingReader      = (*mappingFileReader)(nil)
	_ StacktraceResolver = (*stacktraceResolverFile)(nil)
)

type File interface {
	RangeReader(offset, size int64) (io.ReadCloser, error)
}

type Reader struct {
	f File

	header Header
	toc    TOC
}

func OpenFile(f File) (*Reader, error) {
	return new(Reader), nil
}

func (r *Reader) MappingReader(mappingName uint64) MappingReader {
	return new(mappingFileReader)
}

type mappingFileReader struct{}

func (r *mappingFileReader) StacktraceResolver() StacktraceResolver {
	return new(stacktraceResolverFile)
}

type stacktraceResolverFile struct{}

func (r *stacktraceResolverFile) ResolveStacktraces(StacktraceInserter, []uint32) {

}

func (r *stacktraceResolverFile) Release() {

}
