package symdb

import (
	"hash/maphash"
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

var (
	_ MappingReader = (*inMemoryMapping)(nil)
	_ MappingWriter = (*inMemoryMapping)(nil)

	_ StacktraceAppender = (*stacktraceAppender)(nil)
	_ StacktraceResolver = (*stacktraceResolverMemory)(nil)
)

type inMemoryMapping struct {
	name uint64

	maxNodesPerChunk uint32
	// maxStackDepth uint32

	// Stack traces originating from the mapping (binary):
	// their bottom frames (roots) refer to this mapping.
	stacktraceMutex    sync.RWMutex
	stacktraceHashToID map[uint64]uint32
	stacktraceChunks   []*stacktraceChunk
	// Headers of already written stack trace chunks.
	stacktraceChunkHeaders []StacktraceChunkHeader
}

func (b *inMemoryMapping) StacktraceAppender() StacktraceAppender {
	b.stacktraceMutex.RLock()
	// Assuming there is at least one chunk.
	c := b.stacktraceChunks[len(b.stacktraceChunks)-1]
	b.stacktraceMutex.RUnlock()
	return &stacktraceAppender{
		mapping: b,
		chunk:   c,
	}
}

func (b *inMemoryMapping) StacktraceResolver() StacktraceResolver {
	return &stacktraceResolverMemory{
		mapping: b,
	}
}

// stacktraceChunkForInsert returns a chunk for insertion:
// if the existing one has capacity, or a new one, if the former is full.
// Must be called with the stracktraces mutex write lock held.
func (b *inMemoryMapping) stacktraceChunkForInsert() *stacktraceChunk {
	c := b.stacktraceChunks[len(b.stacktraceChunks)-1]
	if n := c.tree.len(); b.maxNodesPerChunk > 0 && n >= b.maxNodesPerChunk {
		c = &stacktraceChunk{
			mapping: b,
			tree:    newStacktraceTree(defaultStacktraceTreeSize),
			stid:    c.stid + b.maxNodesPerChunk,
		}
		b.stacktraceChunks = append(b.stacktraceChunks, c)
	}
	return c
}

type stacktraceChunk struct {
	mapping *inMemoryMapping
	m       sync.Mutex // It is a write-intensive lock.
	stid    uint32     // Initial stack trace ID.
	tree    *stacktraceTree
}

func (s *stacktraceChunk) WriteTo(dst io.Writer) (int64, error) {
	return s.tree.WriteTo(dst)
}

type stacktraceAppender struct {
	mapping     *inMemoryMapping
	chunk       *stacktraceChunk
	releaseOnce sync.Once
}

func (a *stacktraceAppender) AppendStacktrace(dst []uint32, s []*v1.Stacktrace) {
	if len(s) == 0 {
		return
	}

	var (
		id     uint32
		found  bool
		misses int
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
	m := int(a.mapping.maxNodesPerChunk)
	t, j := a.chunk.tree, a.chunk.stid
	for i, v := range dst {
		if v != 0 {
			// Already resolved. ID 0 is reserved
			// as it is the tree root.
			continue
		}
		x := s[i].LocationIDs
		if m > 0 && len(t.nodes)+len(x) >= m {
			// If we're close to the max nodes limit and can
			// potentially exceed it, we take the next chunk,
			// even if there are some space.
			a.chunk = a.mapping.stacktraceChunkForInsert()
			t, j = a.chunk.tree, a.chunk.stid
		}
		// Tree insertion is idempotent,
		// we don't need to check the map.
		id = t.insert(x) + j
		h := hashLocations(x) // TODO(kolesnikovae): Avoid rehashing.
		a.mapping.stacktraceHashToID[h] = id
		dst[i] = id
	}
}

func (a *stacktraceAppender) Release() {}

type stacktraceResolverMemory struct {
	mapping     *inMemoryMapping
	releaseOnce sync.Once
}

const defaultStacktraceDepth = 64

func (r *stacktraceResolverMemory) ResolveStacktraces(dst StacktraceInserter, stacktraces []uint32) {
	// We assume stacktraces is sorted in the ascending order.
	// First, we split it into ranges corresponding to the chunks.
	m := r.mapping.maxNodesPerChunk
	d := splitStacktraces(stacktraces, m)
	for _, x := range d {
		// TODO(kolesnikovae):
		// Each chunk should be resolved independently with
		// a limit on concurrency and memory consumption.
		c := r.mapping.stacktraceChunks[x.chunk]
		s := make([]int32, 0, defaultStacktraceDepth)
		// Restore the original stacktrace ID.
		off := x.chunk * m
		for _, sid := range x.ids {
			s = c.tree.resolve(s, sid)
			dst.InsertStacktrace(off+sid, s)
		}
	}
}

type stacktraceIDRange struct {
	chunk uint32
	ids   []uint32
}

// splitStacktraces splits the range of stack trace IDs by limit n into
// sub-ranges matching to the corresponding chunks and shifts the values
// accordingly. Note that the input s is modified in place.
//
// stack trace ID 0 is reserved and not expected at the input.
// stack trace ID % max_nodes == 0 is not expected as well.
func splitStacktraces(s []uint32, n uint32) []stacktraceIDRange {
	if s[len(s)-1] < n || n == 0 {
		// Fast path, just one chunk: the highest stack trace ID
		// is less than the chunk size, or the size is not limited.
		// It's expected that in most cases we'll end up here.
		return []stacktraceIDRange{{ids: s}}
	}

	var (
		loi int
		lov = (s[0] / n) * n // Lowest possible value for the current chunk.
		hiv = lov + n        // Highest possible value for the current chunk.
		p   uint32           // Previous value (to derive chunk index).
		// 16 chunks should be more than enough in most cases.
		cs = make([]stacktraceIDRange, 0, 16)
	)

	for i, v := range s {
		if v < hiv {
			// The stack belongs to the current chunk.
			s[i] -= lov
			p = v
			continue
		}
		lov = (v / n) * n
		hiv = lov + n
		s[i] -= lov
		cs = append(cs, stacktraceIDRange{
			chunk: p / n,
			ids:   s[loi:i],
		})
		loi = i
		p = v
	}

	if t := s[loi:]; len(t) > 0 {
		cs = append(cs, stacktraceIDRange{
			chunk: p / n,
			ids:   t,
		})
	}

	return cs
}

func (r *stacktraceResolverMemory) Release() {
	r.releaseOnce.Do(func() {})
}

var seed = maphash.MakeSeed()

func hashLocations(s []uint64) uint64 {
	var b []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Len = len(s) * 8
	hdr.Cap = hdr.Len
	hdr.Data = uintptr(unsafe.Pointer(&s[0]))
	return maphash.Bytes(seed, b)
}
