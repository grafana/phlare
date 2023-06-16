package symdb

import "github.com/grafana/phlare/pkg/objstore"

var (
	_ MappingReader      = (*mappingFileReader)(nil)
	_ StacktraceResolver = (*stacktraceResolverFile)(nil)
)

type Reader struct{}

func Open(objstore.BucketReader) (*Reader, error) {
	// NOTE(kolesnikovae): We could accept fs.FS and implement it with
	//  the BucketReader, but it brings no actual value other than a
	//  cleaner signature.
	return new(Reader), nil
}

func (r *Reader) Close() error { return nil }

func (r *Reader) MappingReader(mappingName uint64) MappingReader {
	return new(mappingFileReader)
}

type mappingFileReader struct{}

func (r *mappingFileReader) StacktraceResolver() StacktraceResolver {
	return new(stacktraceResolverFile)
}

type stacktraceResolverFile struct{}

func (r *stacktraceResolverFile) ResolveStacktraces(StacktraceInserter, []uint32) {}

func (r *stacktraceResolverFile) Release() {}
