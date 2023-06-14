package symdb

import (
	"io"

	"github.com/grafana/phlare/pkg/iter"
)

var (
	_ IndexReader        = (*indexFileReader)(nil)
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

func (r *Reader) IndexReader(partitionID uint64) IndexReader {
	return new(indexFileReader)
}

type indexFileReader struct{}

func (r *indexFileReader) StacktraceResolver() StacktraceResolver {
	return new(stacktraceResolverFile)
}

type stacktraceResolverFile struct{}

func (r *stacktraceResolverFile) ResolveStacktraces(StacktraceInserter, iter.Iterator[int32]) {

}

func (r *stacktraceResolverFile) Release() {

}
