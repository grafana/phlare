package symdb

import "github.com/grafana/phlare/pkg/iter"

// Mapping is a binary that is part of the program during the profile
// collection. https://github.com/google/pprof/blob/main/proto/README.md
// Currently, we maintain a Mapping for all the version of a binary.

type MappingWriter interface {
	// StacktraceAppender provides exclusive write access
	// to the stack traces of the mapping.
	//
	// StacktraceAppender.Release must be called in order
	// to dispose the object and release the lock.
	// Released resolver must not be used.
	StacktraceAppender() StacktraceAppender
}

type MappingReader interface {
	// StacktraceResolver provides non-exclusive read
	// access to the stack traces of the mapping.
	//
	// StacktraceResolver.Release must be called in order
	// to dispose the object and release the lock.
	// Released resolver must not be used.
	StacktraceResolver() StacktraceResolver
}

type StacktraceAppender interface {
	AppendStacktrace(locations []int32) int32
	Release()
}

type StacktraceResolver interface {
	// ResolveStacktraces resolves locations for each stack trace
	// and inserts it to the StacktraceInserter provided.
	// The iterator implementation must ensure ascending order.
	ResolveStacktraces(dst StacktraceInserter, stacktraces iter.Iterator[int32])
	Release()
}

// StacktraceInserter accepts resolved locations for a given stack trace.
// The top stack trace frame is at locations[0].
// Locations slice must not be retained by implementation.
type StacktraceInserter interface {
	InsertStacktrace(stacktraceID int32, locations []int32)
}
