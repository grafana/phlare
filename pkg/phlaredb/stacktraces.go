// nolint unused
package phlaredb

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	"github.com/grafana/phlare/pkg/phlaredb/block"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	schemav2 "github.com/grafana/phlare/pkg/phlaredb/schemas/v2"
	"github.com/grafana/phlare/pkg/util/build"
	art "github.com/plar/go-adaptive-radix-tree"
	"github.com/segmentio/parquet-go"
)

const (
	stacktraceSize = uint64(unsafe.Sizeof(schemav2.Stacktrace{}))
)

type genericTable[P any] struct {
	persister schemav1.Persister[P]
	file      *os.File
	cfg       *ParquetConfig
	metrics   *headMetrics
	writer    *parquet.GenericWriter[P]
}

func (t *genericTable[P]) Init(path string, cfg *ParquetConfig, metrics *headMetrics) error {
	t.cfg = cfg
	t.metrics = metrics
	file, err := os.OpenFile(filepath.Join(path, t.persister.Name()+block.ParquetSuffix), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	t.file = file

	// TODO: Reuse parquet.Writer beyond life time of the head.
	t.writer = parquet.NewGenericWriter[P](file, t.persister.Schema(),
		parquet.ColumnPageBuffers(parquet.NewFileBufferPool(os.TempDir(), "phlaredb-parquet-buffers*")),
		parquet.CreatedBy("github.com/grafana/phlare/", build.Version, build.Revision),
	)
	return nil
}

type Stacktrace struct {
	MappingID   uint64
	LocationIDs []uint64
}

// this takes the location ids and creates a byte slice in reverse order (the last location ID is the root of the stacktrace)
// TODO: Do this with unsafe might save us some allocations
func locationSliceToByteSlice(v []uint64) []byte {
	r := make([]byte, 8*len(v))
	for pos := range v {
		binary.LittleEndian.PutUint64(r[8*pos:], v[len(v)-1-pos])
	}
	return r
}

// this takes the byte slice and creates the location IDs in reverse order
// TODO: Do this with unsafe might save us some allocations
func byteSliceToLocationSlice(v []byte) []uint64 {
	if len(v)%8 != 0 {
		panic("byte slice is not a multiple of 8")
	}
	r := make([]uint64, len(v)/8)
	for pos := range r {
		r[len(r)-pos-1] = binary.LittleEndian.Uint64(v[8*pos : 8*pos+8])
	}
	return r
}

type idKeyspace struct {
	lock  sync.RWMutex
	nodes []art.Node
}

type stacktraceStore struct {
	m           *sync.Map
	idKeyspaces []*idKeyspace // we shard the id keyspaces by mappingID supplied

	*genericTable[*schemav2.Stacktrace]
}

func newStacktracesStore(mappingShards int) *stacktraceStore {
	s := &stacktraceStore{
		m:           &sync.Map{},
		idKeyspaces: make([]*idKeyspace, mappingShards),
		genericTable: &genericTable[*schemav2.Stacktrace]{
			persister: &schemav2.StacktracePersister{},
		},
	}

	for idx := range s.idKeyspaces {
		s.idKeyspaces[idx] = &idKeyspace{}
	}

	s.idKeyspaces[0].nodes = append(s.idKeyspaces[0].nodes, nil)

	return s
}

func (s *stacktraceStore) idKeyspaceShardSize() uint64 {
	maxUint64 := ^uint64(0)
	return maxUint64 / uint64(len(s.idKeyspaces))
}

func (s *stacktraceStore) allocateID(mappingID uint64) uint64 {
	// find out which keyspace to use
	idx := mappingID % uint64(len(s.idKeyspaces))
	ks := s.idKeyspaces[idx]

	// aquire keyspace lock
	ks.lock.Lock()
	defer ks.lock.Unlock()

	// increment the keyspace size
	s.idKeyspaces[idx].nodes = append(s.idKeyspaces[idx].nodes, nil)

	return s.idKeyspaceShardSize()*uint64(idx) + uint64(len(s.idKeyspaces[idx].nodes)-1)
}

func (s *stacktraceStore) slice() []*schemav2.StoredStacktrace {
	panic("TODO")
	/*r := make([]*schemav2.StoredStacktrace, 0, s.tree.Size())
	s.tree.ForEach(func(n art.Node) bool {
		r = append(r, &schemav2.StoredStacktrace{
			ID:          *(n.Value().(*uint64)),
			LocationIDs: byteSliceToLocationSlice(n.Key()),
		})
		return true
	})
	return r
	*/
}

func (s *stacktraceStore) getLocationIDs(stacktraceID uint64) (locationIDs []uint64) {
	// find out which keyspace to use
	ksIdx := stacktraceID / s.idKeyspaceShardSize()
	ks := s.idKeyspaces[ksIdx]

	ks.lock.RLock()
	defer ks.lock.RUnlock()

	key := ks.nodes[stacktraceID%s.idKeyspaceShardSize()].Key()
	if key == nil {
		panic("could not find key")
	}

	return byteSliceToLocationSlice(key)

}

// add adds a stacktrace to the store, if a stacktrace contains locations from multiple mappingIDs, choose the mapping ID that is most common throughout the stacktrace.
func (s *stacktraceStore) add(mappingID uint64, locationIDs []uint64) (stacktraceID uint64) {
	key := art.Key(locationSliceToByteSlice(locationIDs))

	if id, found := s.tree.Search(key); found {
		return *(id.(*uint64))
	}

	// allocate a new ID within the shard determined by mappingID
	id := s.allocateID(mappingID)

	oldValue, updated := s.tree.Insert(key, art.Value(&id))
	if updated {
		s.tree.Insert(key, oldValue)
		return *(oldValue.(*uint64))
	}

	// find node again
	var node art.Node
	s.tree.ForEachPrefix(key, func(n art.Node) bool {
		val := n.Value()
		if val == nil {
			return true
		}
		if *(val.(*uint64)) == id {
			node = n
			return false
		}
		return true
	})
	if node == nil {
		panic("could not find node")
	}

	// link up node in the keyspace
	ksIdx := id / s.idKeyspaceShardSize()
	ks := s.idKeyspaces[ksIdx]
	ks.lock.Lock()
	defer ks.lock.Unlock()

	ks.nodes[id%s.idKeyspaceShardSize()] = node

	return stacktraceID
}

func (s *stacktraceStore) Name() string {
	return "stacktraces"
}

func (s *stacktraceStore) Size() uint64 {
	// TODO: This is not very good assumption
	return uint64(s.tree.Size()) * stacktraceSize
}

func (s *stacktraceStore) MemorySize() uint64 {
	// TODO: This is not very good assumption
	return uint64(s.tree.Size()) * stacktraceSize
}

func (s *stacktraceStore) Flush(ctx context.Context) (numRows uint64, numRowGroups uint64, err error) {
	panic("TODO")
}

func (s *stacktraceStore) Close() error {
	return nil
}

func (s *stacktraceStore) ingest(ctx context.Context, elems []*Stacktrace, rewriter *rewriter) error {
	var (
		rewritingMap = make(map[int64]int64)
	)

	// TODO: rewrite location ids

	for idBefore, elem := range elems {
		idAfter := s.add(elem.MappingID, elem.LocationIDs)
		rewritingMap[int64(idBefore)] = int64(idAfter)
	}

	// add rewrite information to struct
	rewriter.stacktraces = rewritingMap

	return nil
}

type stacktracesKey struct {
	Parent     uint64
	LocationID uint64
}

type stacktracesHelper struct{}

func (*stacktracesHelper) key(s *schemav2.Stacktrace) stacktracesKey {
	return stacktracesKey{
		Parent:     s.Parent,
		LocationID: s.LocationID,
	}
}

func (*stacktracesHelper) addToRewriter(r *rewriter, m idConversionTable) {
	r.stacktraces = m
}

func (*stacktracesHelper) rewrite(r *rewriter, s *schemav2.Stacktrace) error {
	r.locations.rewriteUint64(&s.LocationID)
	return nil
}

func (*stacktracesHelper) setID(oldID, newID uint64, s *schemav2.Stacktrace) uint64 {
	return oldID
}

func (*stacktracesHelper) size(s *schemav2.Stacktrace) uint64 {
	return stacktraceSize
}

func (*stacktracesHelper) clone(s *schemav2.Stacktrace) *schemav2.Stacktrace {
	return &schemav2.Stacktrace{
		Parent:     s.Parent,
		LocationID: s.LocationID,
	}
}
