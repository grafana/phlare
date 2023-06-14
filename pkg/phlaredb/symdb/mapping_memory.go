package symdb

import (
	"hash/maphash"
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/grafana/phlare/pkg/iter"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

var (
	_ MappingReader = (*inMemoryMapping)(nil)
	_ MappingWriter = (*inMemoryMapping)(nil)

	_ StacktraceAppender = (*stacktraceAppender)(nil)
	_ StacktraceResolver = (*stacktraceResolverMemory)(nil)
)

type inMemoryMapping struct {
	maxStacksPerChunk int32
	// maxStackDepth int32

	// Stack traces originating from the mapping (binary):
	// their bottom frames (roots) refer to this mapping.
	stacktraceMutex    sync.RWMutex
	stacktraceHashToID map[uint64]int32
	stacktraceChunks   []*stacktraceChunk
}

func (b *inMemoryMapping) StacktraceAppender() StacktraceAppender {
	b.stacktraceMutex.RLock()
	// Assuming there is at least one chunk.
	c := b.stacktraceChunks[len(b.stacktraceChunks)-1]
	b.stacktraceMutex.RUnlock()
	return &stacktraceAppender{
		maxStacks: b.maxStacksPerChunk,
		mapping:   b,
		chunk:     c,
	}
}

func (b *inMemoryMapping) StacktraceResolver() StacktraceResolver {
	return new(stacktraceResolverMemory)
}

func (b *inMemoryMapping) newStacktraceChunk(stid int32) *stacktraceChunk {
	s := &stacktraceChunk{
		tree: newStacktraceTree(defaultStacktraceTreeSize),
		stid: stid,
	}
	b.stacktraceChunks = append(b.stacktraceChunks, s)
	return s
}

type stacktraceChunk struct {
	m      sync.Mutex // Write-intensive lock.
	stid   int32      // Initial stack trace ID.
	stacks int32      // Number of stacks in the tree.
	tree   *stacktraceTree
}

func (s *stacktraceChunk) WriteTo(dst io.Writer) (int64, error) {
	return s.tree.WriteTo(dst)
}

type stacktraceAppender struct {
	mapping     *inMemoryMapping
	chunk       *stacktraceChunk
	maxStacks   int32
	releaseOnce sync.Once
}

func (a *stacktraceAppender) AppendStacktrace(dst []int32, s []*schemav1.Stacktrace) {
	if len(s) == 0 {
		return
	}

	var (
		id     int32
		found  bool
		misses int32
	)

	a.mapping.stacktraceMutex.RLock()
	for i, x := range s {
		if dst[i], found = a.mapping.stacktraceHashToID[hashLocations(x.LocationIDs)]; !found {
			misses++
		}
	}
	a.mapping.stacktraceMutex.RUnlock()
	if misses == 0 {
		return
	}

	a.mapping.stacktraceMutex.Lock()
	defer a.mapping.stacktraceMutex.Unlock()
	for i, v := range dst {
		if v != 0 {
			// Already resolved. ID 0 is reserved
			// as it is the tree root.
			continue
		}
		if a.chunk.stacks == a.maxStacks {
			a.chunk = a.mapping.newStacktraceChunk(a.chunk.stid + a.chunk.stacks)
		}
		// tree insertion is idempotent,
		// we don't need to check the map.
		x := s[i].LocationIDs
		h := hashLocations(x)
		id = a.chunk.tree.insert(x) + a.chunk.stid
		dst[i] = id
		a.mapping.stacktraceHashToID[h] = id
		a.chunk.stacks++
	}
}

func (a *stacktraceAppender) Release() {}

type stacktraceResolverMemory struct {
	mapping     *inMemoryMapping
	releaseOnce sync.Once
}

func (r *stacktraceResolverMemory) ResolveStacktraces(StacktraceInserter, iter.Iterator[int32]) {

}

func (r *stacktraceResolverMemory) Release() {
	r.releaseOnce.Do(func() {})
}

var seed = maphash.MakeSeed()

func hash(b []byte) uint64 { return maphash.Bytes(seed, b) }

func hashLocations(s []uint64) uint64 {
	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Len = len(s) * 8
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&s[0]))
	return hash(b)
}
