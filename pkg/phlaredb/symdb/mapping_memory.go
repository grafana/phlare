package symdb

import (
	"io"
	"sync"

	"github.com/grafana/phlare/pkg/iter"
)

var (
	_ MappingReader = (*inMemoryMapping)(nil)
	_ MappingWriter = (*inMemoryMapping)(nil)

	_ StacktraceAppender = (*stacktraceAppender)(nil)
	_ StacktraceResolver = (*stacktraceResolverMemory)(nil)
)

type inMemoryMapping struct {
	m sync.RWMutex

	maxStacksPerChunk int32
	// maxStackDepth int32

	// Stack traces originating from the mapping (binary):
	// their bottom frames (roots) refer to this mapping.
	stacktraceChunks []*stacktraceChunk
}

func (b *inMemoryMapping) StacktraceAppender() StacktraceAppender {
	b.m.RLock()
	// Assuming there is at least one chunk.
	c := b.stacktraceChunks[len(b.stacktraceChunks)-1]
	b.m.RUnlock()
	a := stacktraceAppender{
		index:     b,
		maxStacks: b.maxStacksPerChunk,
	}
	// We lock the chunk until it released
	// by caller with appender.Release
	c.m.Lock()
	a.useChunk(c)
	return &a
}

func (b *inMemoryMapping) StacktraceResolver() StacktraceResolver {
	return new(stacktraceResolverMemory)
}

func (b *inMemoryMapping) newStacktraceChunk(stid int32) *stacktraceChunk {
	t := newStacktraceTree(defaultStacktraceTreeSize)
	b.m.Lock()
	s := &stacktraceChunk{
		tree: t,
		stid: stid,
	}
	b.stacktraceChunks = append(b.stacktraceChunks, s)
	b.m.Unlock()
	return s
}

type stacktraceChunk struct {
	m      sync.Mutex // Write-intensive lock.
	stid   int32      // Initial stack trace ID.
	stacks int32      // Number of stacks in the tree.
	tree   *stacktraceTree
}

func (s *stacktraceChunk) LookupStacktrace(locations []int32, id int32) []int32 {
	return s.tree.resolve(locations, id-s.stid)
}

type stacktraceAppender struct {
	index       *inMemoryMapping
	chunk       *stacktraceChunk
	stid        int32
	stacks      int32
	maxStacks   int32
	releaseOnce sync.Once
}

func (a *stacktraceAppender) AppendStacktrace(locations []int32) int32 {
	id, ok := a.chunk.tree.insert(locations)
	if ok {
		if a.stacks++; a.maxStacks > 0 && a.stacks == a.maxStacks {
			s := a.index.newStacktraceChunk(a.stid + a.stacks)
			a.useChunk(s)
		}
	}
	return a.stid + id
}

func (a *stacktraceAppender) Release() {
	a.releaseOnce.Do(a.releaseChunk)
}

func (a *stacktraceAppender) useChunk(s *stacktraceChunk) {
	if a.chunk != nil {
		a.releaseChunk()
	}
	a.chunk = s
	a.stid = s.stid
	a.stacks = s.stacks
}

func (a *stacktraceAppender) releaseChunk() {
	// Update chunk stats.
	a.chunk.stacks = a.stacks
	a.chunk.m.Unlock()
}

func (s *stacktraceChunk) WriteTo(dst io.Writer) (int64, error) {
	return s.tree.WriteTo(dst)
}

type stacktraceResolverMemory struct {
	mapping     *inMemoryMapping
	releaseOnce sync.Once
}

func (r *stacktraceResolverMemory) ResolveStacktraces(StacktraceInserter, iter.Iterator[int32]) {

}

func (r *stacktraceResolverMemory) Release() {
	r.releaseOnce.Do(func() {})
}
