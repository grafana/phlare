package phlaredb

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/runutil"
	"github.com/pkg/errors"
	"github.com/segmentio/parquet-go"
	"go.uber.org/atomic"

	phlaremodel "github.com/grafana/phlare/pkg/model"
	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	"github.com/grafana/phlare/pkg/phlaredb/query"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/grafana/phlare/pkg/util/build"
)

type profileStore struct {
	size      atomic.Uint64
	totalSize atomic.Uint64
	// lock serializes writes to the slice. Consider using sync.Mutex.
	lock  sync.RWMutex
	slice []*schemav1.Profile

	metrics *headMetrics

	index *profilesIndex

	persister schemav1.Persister[*schemav1.Profile]
	helper    storeHelper[*schemav1.Profile]

	logger log.Logger
	cfg    *ParquetConfig

	writer *parquet.GenericWriter[*schemav1.Profile]

	path        string
	rowsLock    sync.RWMutex
	rowsFlushed uint64
	rowGroups   []*rowGroupOnDisk
}

func newProfileStore(phlarectx context.Context) *profileStore {
	s := &profileStore{
		logger:    phlarecontext.Logger(phlarectx),
		metrics:   contextHeadMetrics(phlarectx),
		persister: &schemav1.ProfilePersister{},
		helper:    &profilesHelper{},
	}

	// Initialize writer on /dev/null
	// TODO: Reuse parquet.Writer beyond life time of the head.
	s.writer = parquet.NewGenericWriter[*schemav1.Profile](io.Discard, s.persister.Schema(),
		parquet.ColumnPageBuffers(parquet.NewFileBufferPool(os.TempDir(), "phlaredb-parquet-buffers*")),
		parquet.CreatedBy("github.com/grafana/phlare/", build.Version, build.Revision),
	)

	return s
}

func (s *profileStore) Name() string {
	return s.persister.Name()
}

func (s *profileStore) Size() uint64 {
	return s.totalSize.Load()
}

func (s *profileStore) MemorySize() uint64 {
	return s.size.Load()
}

// resets the store
func (s *profileStore) Init(path string, cfg *ParquetConfig, metrics *headMetrics) (err error) {
	// close previous iteration
	if err := s.Close(); err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	// create index
	s.index, err = newProfileIndex(32, s.metrics)
	if err != nil {
		return err
	}

	s.path = path
	s.cfg = cfg
	s.metrics = metrics

	s.slice = s.slice[:0]

	s.rowsFlushed = 0

	return nil
}

func (s *profileStore) Close() error {
	return nil
}

func (s *profileStore) RowGroups() (rowGroups []parquet.RowGroup) {
	rowGroups = make([]parquet.RowGroup, len(s.rowGroups))
	for pos := range rowGroups {
		rowGroups[pos] = s.rowGroups[pos]
	}
	return rowGroups
}

func (s *profileStore) profileSort(x []*schemav1.Profile) func(i, j int) bool {
	return func(i, j int) bool {
		// first compare the labels, if they don't match return
		var (
			pI   = x[i]
			pJ   = x[j]
			lbsI = s.index.profilesPerFP[pI.SeriesFingerprint].lbs
			lbsJ = s.index.profilesPerFP[pJ.SeriesFingerprint].lbs
		)
		if cmp := phlaremodel.CompareLabelPairs(lbsI, lbsJ); cmp != 0 {
			return cmp < 0
		}

		// then compare timenanos, if they don't match return
		if pI.TimeNanos < pJ.TimeNanos {
			return true
		} else if pI.TimeNanos > pJ.TimeNanos {
			return false
		}

		// finally use ID as tie breaker
		return bytes.Compare(pI.ID[:], pJ.ID[:]) < 0
	}
}

func (s *profileStore) Flush(ctx context.Context) (numRows uint64, numRowGroups uint64, err error) {
	if err := s.Close(); err != nil {
		return 0, 0, err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if err = s.cutRowGroup(); err != nil {
		return 0, 0, err
	}

	indexPath := filepath.Join(
		s.path,
		block.IndexFilename,
	)

	rowRangerPerRG, err := s.index.writeTo(ctx, indexPath)
	if err != nil {
		return 0, 0, err
	}

	// TODO: Ensure there are no races.
	for idx, ranges := range rowRangerPerRG {
		s.rowGroups[idx].seriesIndexes = ranges
	}

	parquetPath := filepath.Join(
		s.path,
		s.persister.Name()+block.ParquetSuffix,
	)

	numRows, numRowGroups, err = s.writeRowGroups(parquetPath, s.RowGroups())
	if err != nil {
		return 0, 0, err
	}

	// TODO should not be closed now, but only after block querier is added to PhlareDB
	// cleanup row groups, which need cleaning
	for _, rg := range s.rowGroups {
		if err := rg.Close(); err != nil {
			return 0, 0, err
		}
	}
	s.rowGroups = s.rowGroups[:0]

	return numRows, numRowGroups, nil
}

func (s *profileStore) prepareFile(path string) (f *os.File, err error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	s.writer.Reset(file)

	return file, err
}

func (s *profileStore) empty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.slice) == 0
}

// cutRowGroups gets called, when a patrticular row group has been finished
// and it will flush it to disk. The caller of cutRowGroups should be holding
// the write lock.
//
// Writes are not allowed during cutting the rows, but readers are not blocked
// during the most of the time: only after when rows have written on disk, we
// lock them shortly (via rowsLock).
//
// TODO: write row groups asynchronously
func (s *profileStore) cutRowGroup() (err error) {
	// if cutRowGroup fails record it as failed segment
	defer func() {
		if err != nil {
			s.metrics.writtenProfileSegments.WithLabelValues("failed").Inc()
		}
	}()

	// do nothing with empty buffer
	bufferRowNums := len(s.slice)
	if bufferRowNums == 0 {
		return nil
	}

	path := filepath.Join(
		s.path,
		fmt.Sprintf("%s.%d%s", s.persister.Name(), s.rowsFlushed, block.ParquetSuffix),
	)

	f, err := s.prepareFile(path)
	if err != nil {
		return err
	}

	// order profiles properly
	// TODO: We don't need the copy? It seems reader don't access the slice
	//  but only use the index that has it's own lock.
	tmpSlice := make([]*schemav1.Profile, len(s.slice))
	copy(tmpSlice, s.slice) // Use of a pool is hardly reasonable here due to GC pressure.
	sort.Slice(tmpSlice, s.profileSort(tmpSlice))

	n, err := s.writer.Write(tmpSlice)
	if err != nil {
		return errors.Wrap(err, "write row group segments to disk")
	}

	if err := s.writer.Close(); err != nil {
		return errors.Wrap(err, "close row group segment writer")
	}

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "closing row group segment file")
	}
	s.metrics.writtenProfileSegments.WithLabelValues("success").Inc()

	// get row group segment size on disk
	if stat, err := f.Stat(); err == nil {
		s.metrics.writtenProfileSegmentsBytes.Observe(float64(stat.Size()))
	}

	rowGroup, err := newRowGroupOnDisk(path)
	if err != nil {
		return err
	}

	// We need to make the new on-disk rows available to readers simultaneously
	// with cutting the series from the index. Until this point, profiles can
	// be read from s.slice/s.index. This lock should not be held for long as
	// it only performs in-memory operations, although blocking readers.
	s.rowsLock.Lock()
	s.rowsFlushed += uint64(n)
	s.rowGroups = append(s.rowGroups, rowGroup)
	err = s.index.cutRowGroup(tmpSlice)
	for i := range s.slice {
		// don't retain profiles and samples in memory as re-slice.
		s.slice[i] = nil
	}
	// reset slice and metrics
	s.slice = s.slice[:0]
	s.size.Store(0)
	// After the lock is released, rows/profiles should be read from the disk.
	s.rowsLock.Unlock()
	if err != nil {
		return err
	}

	level.Debug(s.logger).Log("msg", "cut row group segment", "path", path, "numProfiles", n)
	s.metrics.sizeBytes.WithLabelValues(s.Name()).Set(0)
	return nil
}

func (s *profileStore) writeRowGroups(path string, rowGroups []parquet.RowGroup) (n uint64, numRowGroups uint64, err error) {
	fileCloser, err := s.prepareFile(path)
	if err != nil {
		return 0, 0, err
	}
	defer runutil.CloseWithErrCapture(&err, fileCloser, "closing parquet file")

	for rgN, rg := range rowGroups {
		level.Debug(s.logger).Log("msg", "writing row group", "path", path, "row_group_number", rgN, "rows", rg.NumRows())

		nInt64, err := s.writer.ReadRowsFrom(rg.Rows())
		if err != nil {
			return 0, 0, err
		}

		n += uint64(nInt64)
		numRowGroups += 1

		if err := s.writer.Flush(); err != nil {
			return 0, 0, err
		}
	}

	if err := s.writer.Close(); err != nil {
		return 0, 0, err
	}

	s.rowsFlushed += n

	return n, numRowGroups, nil
}

func (s *profileStore) ingest(_ context.Context, profiles []*schemav1.Profile, lbs phlaremodel.Labels, profileName string, rewriter *rewriter) error {
	// rewrite elements
	for pos := range profiles {
		if err := s.helper.rewrite(rewriter, profiles[pos]); err != nil {
			return err
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	for pos, p := range profiles {
		// check if row group is full
		if s.cfg.MaxBufferRowCount > 0 && len(s.slice) >= s.cfg.MaxBufferRowCount ||
			s.cfg.MaxRowGroupBytes > 0 && s.size.Load() >= s.cfg.MaxRowGroupBytes {
			if err := s.cutRowGroup(); err != nil {
				return err
			}
		}

		// add profile to the index
		s.index.Add(p, lbs, profileName)

		// increase size of stored data
		addedBytes := s.helper.size(profiles[pos])
		s.metrics.sizeBytes.WithLabelValues(s.Name()).Set(float64(s.size.Add(addedBytes)))
		s.totalSize.Add(addedBytes)

		// add to slice
		s.slice = append(s.slice, p)

	}

	return nil
}

type rowGroupOnDisk struct {
	parquet.RowGroup
	file          *os.File
	seriesIndexes rowRangesWithSeriesIndex
}

func newRowGroupOnDisk(path string) (*rowGroupOnDisk, error) {
	var (
		r   = &rowGroupOnDisk{}
		err error
	)

	// now open the row group file, so we are able to read the row group back in
	r.file, err = os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening row groups segment file %s", path)
	}

	stats, err := r.file.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "getting stat of row groups segment file %s", path)
	}

	segmentParquet, err := parquet.OpenFile(r.file, stats.Size())
	if err != nil {
		return nil, errors.Wrapf(err, "reading parquet of row groups segment file %s", path)
	}

	rowGroups := segmentParquet.RowGroups()
	if len(rowGroups) != 1 {
		return nil, errors.Wrapf(err, "segement file expected to have exactly one row group (actual %d)", len(rowGroups))
	}

	r.RowGroup = rowGroups[0]

	return r, nil
}

func (r *rowGroupOnDisk) RowGroups() []parquet.RowGroup {
	return []parquet.RowGroup{r.RowGroup}
}

func (r *rowGroupOnDisk) Rows() parquet.Rows {
	rows := r.RowGroup.Rows()
	if len(r.seriesIndexes) == 0 {
		return rows
	}

	return &seriesIDRowsRewriter{
		Rows:          rows,
		seriesIndexes: r.seriesIndexes,
	}
}

func (r *rowGroupOnDisk) Close() error {
	if err := r.file.Close(); err != nil {
		return err
	}

	if err := os.Remove(r.file.Name()); err != nil {
		return errors.Wrap(err, "deleting row group segment file")
	}

	return nil
}

func (r *rowGroupOnDisk) columnIter(ctx context.Context, columnName string, predicate query.Predicate, alias string) query.Iterator {
	column, found := r.RowGroup.Schema().Lookup(columnName)
	if !found {
		return query.NewErrIterator(fmt.Errorf("column '%s' not found in head row group segment '%s'", columnName, r.file.Name()))
	}
	return query.NewColumnIterator(ctx, []parquet.RowGroup{r.RowGroup}, column.ColumnIndex, columnName, 1000, predicate, alias)
}

type seriesIDRowsRewriter struct {
	parquet.Rows
	pos           int64
	seriesIndexes rowRangesWithSeriesIndex
}

func (r *seriesIDRowsRewriter) SeekToRow(pos int64) error {
	if err := r.Rows.SeekToRow(pos); err != nil {
		return err
	}
	r.pos += pos
	return nil
}

var colIdxSeriesIndex = func() int {
	p := &schemav1.ProfilePersister{}
	colIdx, found := p.Schema().Lookup("SeriesIndex")
	if !found {
		panic("column SeriesIndex not found")
	}
	return colIdx.ColumnIndex
}()

func (r *seriesIDRowsRewriter) ReadRows(rows []parquet.Row) (int, error) {
	n, err := r.Rows.ReadRows(rows)
	if err != nil {
		return n, err
	}

	for pos, row := range rows[:n] {
		// actual row num
		rowNum := r.pos + int64(pos)
		row[colIdxSeriesIndex] = parquet.ValueOf(r.seriesIndexes.getSeriesIndex(rowNum)).Level(0, 0, colIdxSeriesIndex)
	}

	r.pos += int64(n)

	return n, nil
}
