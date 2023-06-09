// nolint unused
package phlaredb

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"unsafe"

	"github.com/grafana/phlare/pkg/phlaredb/block"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	schemav2 "github.com/grafana/phlare/pkg/phlaredb/schemas/v2"
	"github.com/grafana/phlare/pkg/util/build"
	"github.com/pkg/errors"
	"github.com/segmentio/parquet-go"
	"go.uber.org/atomic"
)

const (
	stacktraceSize = uint64(unsafe.Sizeof(schemav2.Stacktrace{}))
)

type genericTable[P any] struct {
	persister schemav1.Persister[P]

	file    *os.File
	cfg     *ParquetConfig
	metrics *headMetrics
	writer  *parquet.GenericWriter[P]
	buffer  *parquet.GenericBuffer[P]

	rowsFlushed int
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

func (t *genericTable[P]) Close() error {
	if err := t.writer.Close(); err != nil {
		return errors.Wrap(err, "closing parquet writer")
	}

	if err := t.file.Close(); err != nil {
		return errors.Wrap(err, "closing parquet file")
	}

	return nil
}

func (t *genericTable[P]) maxRowsPerRowGroup(slice []P, size uint64) int {
	// with empty slice we need to return early
	if len(slice) == 0 {
		return 1
	}

	var (
		// average size per row in memory
		bytesPerRow = size / uint64(len(slice))

		// how many rows per RG with average size are fitting in the maxRowGroupBytes, ensure that we at least flush 1 row
		maxRows = t.cfg.MaxRowGroupBytes / bytesPerRow
	)

	if maxRows <= 0 {
		return 1
	}

	return int(maxRows)
}

func (t *genericTable[P]) flush(ctx context.Context, slice []P, size uint64) (numRows uint64, numRowGroups uint64, err error) {
	// TODO: lock or make sure you call it locked
	if t.buffer == nil {
		t.buffer = parquet.NewGenericBuffer[P](t.persister.Schema(),
			parquet.SortingRowGroupConfig(t.persister.SortingColumns()),
			parquet.ColumnBufferCapacity(t.cfg.MaxBufferRowCount),
		)
	}

	var (
		maxRows = t.maxRowsPerRowGroup(slice, size)

		rowGroupsFlushed int
		rowsFlushed      int
	)

	for {
		// how many rows of the head still in need of flushing
		rowsToFlush := len(slice) - t.rowsFlushed

		if rowsToFlush == 0 {
			break
		}

		// cap max row group size by bytes
		if rowsToFlush > maxRows {
			rowsToFlush = maxRows
		}
		// cap max row group size by buffer
		if rowsToFlush > t.cfg.MaxBufferRowCount {
			rowsToFlush = t.cfg.MaxBufferRowCount
		}

		rows := make([]parquet.Row, rowsToFlush)
		var slicePos int
		for pos := range rows {
			slicePos = pos + t.rowsFlushed
			rows[pos] = t.persister.Deconstruct(rows[pos], uint64(slicePos), slice[slicePos])
		}

		t.buffer.Reset()
		if _, err := t.buffer.WriteRows(rows); err != nil {
			return 0, 0, err
		}

		sort.Sort(t.buffer)

		if _, err = t.writer.WriteRowGroup(t.buffer); err != nil {
			return 0, 0, err
		}

		t.rowsFlushed += rowsToFlush
		rowsFlushed += rowsToFlush
		rowGroupsFlushed++
	}

	return uint64(rowsFlushed), uint64(rowGroupsFlushed), nil
}

type Stacktrace struct {
	ID          uint64
	MappingID   uint64
	LocationIDs []uint64
}

type stacktraceKey string

// this takes the location ids and creates a byte slice in reverse order (the last location ID is the root of the stacktrace)
// TODO: Do this with unsafe might save us some allocations
func locationSliceToStacktraceKey(v []uint64) stacktraceKey {
	r := make([]byte, 8*len(v))
	for pos := range v {
		binary.LittleEndian.PutUint64(r[8*pos:], v[len(v)-1-pos])
	}
	return stacktraceKey(r)
}

// this takes the byte slice and creates the location IDs in reverse order
// TODO: Do this with unsafe might save us some allocations
func stacktraceKeyToLocationSlice(key stacktraceKey, loc []uint64) []uint64 {
	v := []byte(key)
	if len(v)%8 != 0 {
		panic("byte slice is not a multiple of 8")
	}
	if cap(loc) < len(v)/8 {
		loc = make([]uint64, len(v)/8)
	} else {
		loc = loc[:len(v)/8]
	}
	for pos := range loc {
		loc[len(loc)-pos-1] = binary.LittleEndian.Uint64(v[8*pos : 8*pos+8])
	}
	return loc
}

type idKeyspace struct {
	lock  sync.RWMutex
	nodes []stacktraceKey
}

type stacktraceStore struct {
	m     map[stacktraceKey]uint64 // maps the key (derived from location IDs) to the stacktrace ID
	mLock sync.RWMutex             // protects m only

	idKeyspaces []*idKeyspace // we shard the id keyspaces by mappingID supplied

	keySize *atomic.Uint64

	*genericTable[*schemav2.Stacktrace]
}

func newStacktracesStore(mappingShards int) *stacktraceStore {
	s := &stacktraceStore{
		m:           make(map[stacktraceKey]uint64),
		idKeyspaces: make([]*idKeyspace, mappingShards),
		keySize:     atomic.NewUint64(0),
		genericTable: &genericTable[*schemav2.Stacktrace]{
			persister: &schemav2.StacktracePersister{},
		},
	}

	for idx := range s.idKeyspaces {
		s.idKeyspaces[idx] = &idKeyspace{}
	}

	// add the empty stacktrace by default
	s.idKeyspaces[0].nodes = append(s.idKeyspaces[0].nodes, "")
	s.m[""] = 0

	return s
}

func (s *stacktraceStore) idKeyspaceShardSize() uint64 {
	maxUint64 := ^uint64(0)
	return maxUint64 / uint64(len(s.idKeyspaces))
}

func (s *stacktraceStore) getLocationIDs(stacktraceID uint64, locationIDs []uint64) []uint64 {
	// find out which keyspace to use
	ksIdx := stacktraceID / s.idKeyspaceShardSize()
	ks := s.idKeyspaces[ksIdx]

	ks.lock.RLock()
	defer ks.lock.RUnlock()

	return stacktraceKeyToLocationSlice(ks.nodes[stacktraceID%s.idKeyspaceShardSize()], locationIDs)

}

// add adds a stacktrace to the store, if a stacktrace contains locations from multiple mappingIDs, choose the mapping ID that is most common throughout the stacktrace.
func (s *stacktraceStore) add(mappingID uint64, locationIDs []uint64) (stacktraceID uint64) {
	// not existing, so find out which keyspace to use based on mapping ID
	ksIdx := int(mappingID) % len(s.idKeyspaces)
	return s.addWithKSIndex(ksIdx, locationIDs)
}

// add adds a stacktrace to the store, if a stacktrace contains locations from multiple mappingIDs, choose the mapping ID that is most common throughout the stacktrace.
func (s *stacktraceStore) addWithKSIndex(ksIdx int, locationIDs []uint64) (stacktraceID uint64) {
	key := locationSliceToStacktraceKey(locationIDs)

	// check if stacktrace already exists
	s.mLock.RLock()
	existingID, ok := s.m[key]
	s.mLock.RUnlock()
	if ok {
		return existingID
	}

	ks := s.idKeyspaces[ksIdx]

	// now acquire the map write lock
	s.mLock.Lock()
	defer s.mLock.Unlock()

	// ensure it has not been created in the meantime
	existingID, ok = s.m[key]
	if ok {
		return existingID
	}

	// aquire keyspace lock
	ks.lock.Lock()
	defer ks.lock.Unlock()

	// add key to id space and map
	s.idKeyspaces[ksIdx].nodes = append(s.idKeyspaces[ksIdx].nodes, key)
	newID := s.idKeyspaceShardSize()*uint64(ksIdx) + uint64(len(s.idKeyspaces[ksIdx].nodes)-1)
	s.m[key] = newID
	s.keySize.Add(uint64(len(key)))

	return newID
}

func (s *stacktraceStore) Name() string {
	return "stacktraces"
}

func (s *stacktraceStore) Size() uint64 {
	return s.MemorySize()
}

func (s *stacktraceStore) MemorySize() uint64 {
	totalSize := s.keySize.Load()
	s.mLock.RLock()
	totalSize += uint64(len(s.m)) * (8 + 8)
	s.mLock.RUnlock()
	return totalSize
}

// Flush goes through all shards and retrieves all stacktraces in there and writes them into sel
func (s *stacktraceStore) Flush(ctx context.Context) (numRows uint64, numRowGroups uint64, err error) {
	var (
		slice []*schemav2.Stacktrace
		loc   []uint64
	)

	// now go through all shards and ensure the parent stacktraces are inserted into the same shard, if they do not exists elsewhere.
	for ksIdx, ks := range s.idKeyspaces {
		// gather length
		ksLen := len(ks.nodes)
		fmt.Printf("call flush with slice ks=%d ksLen=%d\n", ksIdx, ksLen)
		if cap(slice) < 5*ksLen { // assume a stack depth of on average 5
			slice = make([]*schemav2.Stacktrace, ksLen, 5*ksLen)
		} else {
			slice = slice[:ksLen]
		}

		insertIdx := uint64(ksLen)
		for idx := 0; idx < ksLen; idx++ {
			loc = stacktraceKeyToLocationSlice(ks.nodes[idx], loc)

			if len(loc) == 0 {
				continue
			}

			parentID := uint64(0)
			for pos := len(loc) - 1; pos >= 0; pos-- {
				locationID := loc[pos]

				if pos == 0 {
					if slice[idx] == nil {
						slice[idx] = &schemav2.Stacktrace{}
					}
					slice[idx].ID = uint64(idx) + uint64(ksIdx)*s.idKeyspaceShardSize()
					slice[idx].LocationID = locationID
					slice[idx].ParentID = parentID
					break
				}

				key := locationSliceToStacktraceKey(loc[pos:])
				id, found := s.m[key]
				if !found {

					id = insertIdx + uint64(ksIdx)*s.idKeyspaceShardSize()
					insertIdx++

					slice = append(slice, &schemav2.Stacktrace{
						ID:         id,
						LocationID: locationID,
						ParentID:   parentID,
					})
					s.m[key] = uint64(insertIdx)

					// TODO: Search local map and ADD if not here
				}

				parentID = id

			}
		}

		if len(slice) == 0 {
			continue
		}
		fmt.Printf("call flush with slice %+#v\n", slice)
		nR, nRG, err := s.genericTable.flush(ctx, slice, s.Size())
		if err != nil {
			return 0, 0, err
		}
		numRows += nR
		numRowGroups += nRG
	}

	return numRows, numRowGroups, err
}

func (s *stacktraceStore) ingest(ctx context.Context, elems []*Stacktrace, rewriter *rewriter) error {
	var (
		rewritingMap = make(map[int64]int64)
	)

	// rewrite location ids
	for idxStacktrace := range elems {
		for idxLocation := range elems[idxStacktrace].LocationIDs {
			rewriter.locations.rewriteUint64(&elems[idxStacktrace].LocationIDs[idxLocation])
		}
	}

	// append stacktraces if necessary and build rewrite table
	for idBefore, elem := range elems {
		idAfter := s.add(elem.MappingID, elem.LocationIDs)
		rewritingMap[int64(idBefore)] = int64(idAfter)
	}

	// add rewrite information to struct
	rewriter.stacktraces = rewritingMap

	return nil
}
