package phlaredb

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/runutil"
	"github.com/pkg/errors"
	"github.com/segmentio/parquet-go"

	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/grafana/phlare/pkg/util/build"
)

var int64SlicePool = &sync.Pool{
	New: func() interface{} {
		return make([]int64, 0)
	},
}

var defaultParquetConfig = &ParquetConfig{
	MaxBufferRowCount: 100_000,
	MaxRowGroupBytes:  128 * 1024 * 1024,
	MaxBlockBytes:     10 * 128 * 1024 * 1024,
}

type store[M Models, P schemav1.Persister[*M]] struct {
	logger log.Logger
	cfg    *ParquetConfig
	helper storeHelper[M]

	buffer       *parquet.GenericBuffer[*M]
	writer       *parquet.GenericWriter[M]
	appendCh     chan *appendElems[M]
	appendCloser sync.Once

	// filter is a hook which allows to check if an element should be added or
	// if it is already existing. This is not executed concurrently.
	filter func(*appendElems[M])

	// updateIndex is a hook which will be called synchronously once a new elemet gets added at the position pos
	updateIndex func(elem *M, pos uint64)

	wg        sync.WaitGroup
	lock      sync.RWMutex
	persister P

	path               string
	rowsFlushed        uint64
	rowGroupBoundaries []uint64
	rowGroups          []parquet.RowGroup
	// buffer size
	bufferByteLastBoundary uint64
}

func newStore[M Models, P schemav1.Persister[*M]](phlarectx context.Context, cfg *ParquetConfig, helper storeHelper[M]) *store[M, P] {
	var s = &store[M, P]{
		logger: phlarecontext.Logger(phlarectx),
		cfg:    cfg,
		helper: helper,

		// initialize hooks with noop methods
		filter:      func(*appendElems[M]) {},
		updateIndex: func(*M, uint64) {},
	}

	// initialize the buffer
	s.buffer = parquet.NewGenericBuffer[*M](
		s.persister.Schema(),
		parquet.SortingRowGroupConfig(s.persister.SortingColumns()),
		parquet.ColumnBufferCapacity(s.cfg.MaxBufferRowCount),
	)

	// Initialize writer on /dev/null
	// TODO: Reuse parquet.Writer beyond life time of the head.
	s.writer = parquet.NewGenericWriter[M](io.Discard, s.persister.Schema(),
		parquet.ColumnPageBuffers(parquet.NewFileBufferPool(os.TempDir(), "phlaredb-parquet-buffers*")),
		parquet.CreatedBy("github.com/grafana/phlare/", build.Version, build.Revision),
	)

	return s
}

func (s *store[M, P]) Name() string {
	return s.persister.Name()
}

func (s *store[M, P]) Size() uint64 {
	return uint64(s.buffer.Size())
}

func (s *store[M, P]) Reset(path string) error {
	// close previous iteration
	if err := s.Close(); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.path = path

	s.appendCh = make(chan *appendElems[M], 32)

	s.rowsFlushed = 0
	s.rowGroupBoundaries = s.rowGroupBoundaries[:]
	s.buffer.Reset()

	// start goroutine for ingest
	s.wg.Add(1)
	go s.appendLoop(s.appendCh)

	return nil
}

func (s *store[M, P]) Close() error {
	// ask appendCh to close
	s.lock.Lock()
	if s.appendCh != nil {
		close(s.appendCh)
		s.appendCh = nil
	}
	s.lock.Unlock()

	s.wg.Wait()

	return nil
}

func (s *store[M, P]) offsetFromPath(p string) uint64 {
	p = filepath.Base(p)
	p = strings.TrimPrefix(p, s.persister.Name()+".")
	p = strings.TrimSuffix(p, block.ParquetSuffix)

	v, err := strconv.ParseUint(p, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func copyRowGroupsFromFile(path string, writer parquet.RowGroupWriter) error {
	sourceFile, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "opening row groups segment file %s", path)
	}
	defer runutil.CloseWithErrCapture(&err, sourceFile, "closing row groups segment file %s", path)

	stats, err := sourceFile.Stat()
	if err != nil {
		return errors.Wrapf(err, "getting stat of row groups segment file %s", path)
	}

	sourceParquet, err := parquet.OpenFile(sourceFile, stats.Size())
	if err != nil {
		return errors.Wrapf(err, "reading parquet of row groups segment file %s", path)
	}

	for pos, rg := range sourceParquet.RowGroups() {
		_, err := writer.WriteRowGroup(rg)
		if err != nil {
			return errors.Wrapf(err, "writing row group %d of row groups segment file %s", pos, path)
		}

	}

	sourceParquet.RowGroups()
	return nil
}

func (s *store[M, P]) joinRowGroupSegments(path string) (numRows uint64, numRowGroups uint64, err error) {
	// Short cut if this is only a single written rowgroup
	// find row group segments
	// TODO: Use the boundary slice to find the files
	rowGroups, err := filepath.Glob(filepath.Join(
		s.path,
		fmt.Sprintf(
			"%s.*%s",
			s.persister.Name(),
			block.ParquetSuffix,
		),
	))

	// sort row groups by offset
	sort.Slice(rowGroups, func(i, j int) bool {
		return s.offsetFromPath(rowGroups[i]) < s.offsetFromPath(rowGroups[j])
	})

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, 0, err
	}
	defer runutil.CloseWithErrCapture(&err, file, "failed to close rowGroup file")

	s.writer.Reset(file)

	for _, rg := range rowGroups {
		if err := copyRowGroupsFromFile(rg, s.writer); err != nil {
			return 0, 0, err
		}

		if err := os.Remove(rg); err != nil {
			return 0, 0, err
		}
	}

	if err := s.writer.Close(); err != nil {
		return 0, 0, err
	}

	level.Debug(s.logger).Log("msg", "aggregated row group segment into block", "path", path, "segments", len(rowGroups))

	return uint64(s.rowsFlushed), uint64(len(s.rowGroupBoundaries)), nil
}

func (s *store[M, P]) Flush() (numRows uint64, numRowGroups uint64, err error) {
	// close ingest loop
	if err := s.Close(); err != nil {
		return 0, 0, err
	}

	path := filepath.Join(
		s.path,
		s.persister.Name()+block.ParquetSuffix,
	)

	// if flushing of indivdiual row groups is enabled, join them up again
	if s.hasFlushRowGroupsToDisk() {
		return s.joinRowGroupSegments(path)
	}

	if _, err := s.cutRowGroup(); err != nil {
		return 0, 0, err
	}

	return s.writeRowGroups(path, s.rowGroups)

}

// TODO: Remove me, bad idea
func (s *store[M, P]) Slice() []*M {
	var (
		mPtr   = make([]*M, s.buffer.Len())
		mReal  = make([]M, s.buffer.Len())
		rows   = make([]parquet.Row, s.buffer.Len())
		reader = s.buffer.Rows()
	)
	defer reader.Close()

	if _, err := reader.ReadRows(rows); err != nil {
		panic(err)
	}

	for pos := range rows {
		reader.Schema().Reconstruct(&mReal[pos], rows[pos])
		mPtr[pos] = &mReal[pos]
	}

	return mPtr
}

func (s *store[M, P]) GetRowNum(rowNum uint64) *M {
	var (
		m      M
		row    = make([]parquet.Row, 1)
		reader = s.buffer.Rows()
	)
	defer reader.Close()

	if err := reader.SeekToRow(int64(rowNum)); err != nil {
		panic(err)
	}

	if _, err := reader.ReadRows(row); err != nil {
		panic(err)
	}

	if err := reader.Schema().Reconstruct(&m, row[0]); err != nil {
		panic(err)
	}

	return &m
}

type appendElems[M Models] struct {
	elems        []*M
	rewritingMap map[int64]int64
	originalID   []int64
	done         chan struct{}
	err          error
}

// append loop is used to serialize the append and avoid locking
func (s *store[M, P]) appendLoop(ch chan *appendElems[M]) {
	defer s.wg.Done()

	defer func() {
		if _, err := s.cutRowGroup(); err != nil {
			level.Error(s.logger).Log("msg", "cut row group", "err", err)
		}
	}()

	for {
		select {
		case elems, open := <-ch:
			if !open {
				return
			}

			// run filter if set
			s.filter(elems)

			// all elements already exist
			if len(elems.elems) == 0 {
				close(elems.done)
				continue
			}

			// update previous and new IDs for all elements
			var (
				previousID uint64 // previous id of the element
				newID      uint64 // new store id after potential deduplication
			)
			for pos := range elems.elems {
				// set previous id of the element
				if len(elems.originalID) > 0 {
					// incase of a filter the original pos is noted down in the append structure
					previousID = uint64(elems.originalID[pos]) // TODO RENAME TO ID
				} else {
					previousID = uint64(pos)
				}
				newID = s.NumRows() + uint64(pos)

				// this updates a potential index
				s.updateIndex(elems.elems[pos], newID)

				// update element itself
				previousID = s.helper.setID(previousID, uint64(newID), elems.elems[pos])

				// update rewrite information
				elems.rewritingMap[int64(previousID)] = int64(newID)
			}

			// append rows to buffer
			_, err := s.buffer.Write(elems.elems)
			if err != nil {
				elems.err = err
				close(elems.done)
				continue
			}

			// close done channel
			close(elems.done)

			// check if row group is now considered as full

			if s.cfg.MaxRowGroupBytes > 0 && (uint64(s.buffer.Size())-s.bufferByteLastBoundary) >= s.cfg.MaxRowGroupBytes || // has too many bytes
				s.cfg.MaxBufferRowCount > 0 && int(s.buffer.NumRows()) >= s.cfg.MaxBufferRowCount { // has too many rows
				if _, err := s.cutRowGroup(); err != nil {
					level.Error(s.logger).Log("msg", "cut row group", "err", err)
				}
				continue
			}
		}
	}

}

// only flush to disk if explicitly enabled
func (s *store[M, P]) hasFlushRowGroupsToDisk() bool {
	if s.cfg != nil && s.cfg.FlushEachRowGroupToDisk != nil {
		return *s.cfg.FlushEachRowGroupToDisk
	}
	return false
}

// cutRowGroups gets called, when a patrticular row group has been finished, depending on the store configuration it will flush a rowGroup to disk or just mark it within the buffer
// TODO: writeRowGroups asynchronously
func (s *store[M, P]) cutRowGroup() (n uint64, err error) {
	// do nothing with empty buffer
	bufferRowNums := s.buffer.NumRows()
	if bufferRowNums == 0 {
		return 0, nil
	}

	// sort the buffer
	sort.Sort(s.buffer)

	// if not flushing row groups to disk, markdown boundary
	if !s.hasFlushRowGroupsToDisk() {
		s.rowGroups = append(s.rowGroups, s.buffer)
		// TODO: Recycle buffers using a buffer pool
		s.buffer = parquet.NewGenericBuffer[*M](
			s.persister.Schema(),
			parquet.SortingRowGroupConfig(s.persister.SortingColumns()),
			parquet.ColumnBufferCapacity(s.cfg.MaxBufferRowCount),
		)
		return 0, nil
	}

	path := filepath.Join(
		s.path,
		fmt.Sprintf("%s.%d%s", s.persister.Name(), s.rowsFlushed, block.ParquetSuffix),
	)

	n, _, err = s.writeRowGroups(path, []parquet.RowGroup{s.buffer})
	if err != nil {
		return n, errors.Wrap(err, "write row group segment to disk")
	}

	s.buffer.Reset()

	return n, nil
}

func (s *store[M, P]) writeRowGroups(path string, rowGroups []parquet.RowGroup) (n uint64, numRowGroups uint64, err error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, 0, err
	}
	defer runutil.CloseWithErrCapture(&err, file, "failed to close rowGroup file")
	s.writer.Reset(file)

	for rgN, rg := range rowGroups {
		level.Debug(s.logger).Log("msg", "writing row group", "path", path, "row_group_number", rgN, "rows", s.buffer.NumRows())

		nInt64, err := s.writer.WriteRowGroup(rg)
		if err != nil {
			return 0, 0, err
		}
		n += uint64(nInt64)
		numRowGroups += 1
	}

	if err := s.writer.Close(); err != nil {
		return 0, 0, err
	}

	s.rowsFlushed += n

	return n, numRowGroups, nil
}

func (s *store[M, P]) ingest(ctx context.Context, elems []*M, rewriter *rewriter) error {
	var appendElems = &appendElems[M]{
		rewritingMap: make(map[int64]int64),
		done:         make(chan struct{}),
		originalID:   make([]int64, 0, len(elems)),
		elems:        elems,
	}

	// rewrite elements
	for pos := range appendElems.elems {
		if err := s.helper.rewrite(rewriter, appendElems.elems[pos]); err != nil {
			return err
		}
	}

	// append to write channel
	s.appendCh <- appendElems

	<-appendElems.done

	if err := appendElems.err; err != nil {
		return err
	}

	// add rewrite information to struct
	s.helper.addToRewriter(rewriter, appendElems.rewritingMap)

	return nil
}

func (s *store[M, P]) NumRows() uint64 {
	return uint64(s.buffer.NumRows()) + s.rowsFlushed
}
