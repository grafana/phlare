package phlaredb

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/segmentio/parquet-go"

	"github.com/grafana/phlare/pkg/phlaredb/block"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/grafana/phlare/pkg/util/build"
)

type deduplicatingSliceSharded[M Models, SK, K comparable, H Helper[M, K], P schemav1.Persister[M]] struct {
	shardsLock sync.RWMutex
	shards     map[SK]*deduplicatingSlice[M, K, H, P]

	persister P
	helper    H

	file    *os.File
	cfg     *ParquetConfig
	metrics *headMetrics
	writer  *parquet.GenericWriter[P]

	sm map[SK]rowRange
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) Name() string {
	return s.persister.Name()
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) MemorySize() uint64 {
	return s.size()
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) Size() uint64 {
	return s.size()
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) size() (x uint64) {
	s.shardsLock.RLock()
	for _, sh := range s.shards {
		x += sh.size.Load()
	}
	s.shardsLock.RUnlock()
	return x
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) Init(path string, cfg *ParquetConfig, metrics *headMetrics) error {
	s.cfg = cfg
	s.metrics = metrics
	file, err := os.OpenFile(filepath.Join(path, s.persister.Name()+block.ParquetSuffix), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	s.file = file

	// TODO: Reuse parquet.Writer beyond life time of the head.
	s.writer = parquet.NewGenericWriter[P](file, s.persister.Schema(),
		parquet.ColumnPageBuffers(parquet.NewFileBufferPool(os.TempDir(), "phlaredb-parquet-buffers*")),
		parquet.CreatedBy("github.com/grafana/phlare/", build.Version, build.Revision),
	)

	s.shards = make(map[SK]*deduplicatingSlice[M, K, H, P])
	s.sm = make(map[SK]rowRange, len(s.shards))
	return nil
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) Close() error {
	if err := s.writer.Close(); err != nil {
		return errors.Wrap(err, "closing parquet writer")
	}
	if err := s.file.Close(); err != nil {
		return errors.Wrap(err, "closing parquet file")
	}
	return nil
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) Flush(ctx context.Context) (numRows uint64, numRowGroups uint64, err error) {
	s.shardsLock.RLock()
	defer s.shardsLock.RUnlock()
	for sk, sh := range s.shards {
		r, g, err := sh.Flush(ctx)
		if err != nil {
			return numRows, numRowGroups, err
		}
		s.sm[sk] = rowRange{
			rowNum: int64(numRows),
			length: int(r),
		}
		numRows += r
		numRowGroups += g
	}
	return numRows, numRowGroups, nil
}

const defaultShardSize = 1 << 10

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) ingest(ctx context.Context, sk SK, elems []M, rewriter *rewriter) error {
	return s.shard(sk).ingest(ctx, elems, rewriter)
}

func (s *deduplicatingSliceSharded[M, SK, K, H, P]) shard(sk SK) *deduplicatingSlice[M, K, H, P] {
	s.shardsLock.RLock()
	shard, ok := s.shards[sk]
	if ok {
		s.shardsLock.RUnlock()
		return shard
	}
	s.shardsLock.RUnlock()
	s.shardsLock.Lock()
	shard, ok = s.shards[sk]
	if ok {
		s.shardsLock.Unlock()
		return shard
	}
	shard = &deduplicatingSlice[M, K, H, P]{
		slice:     make([]M, 0, defaultShardSize),
		lookup:    make(map[K]int64, defaultShardSize),
		persister: s.persister,
		helper:    s.helper,
		file:      s.file,
		cfg:       s.cfg,
		metrics:   s.metrics,
		writer:    s.writer,
	}
	s.shards[sk] = shard
	s.shardsLock.Unlock()
	return shard
}
