package storegateway

import (
	"context"
	"os"
	"path"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/mimir/pkg/storegateway"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"github.com/thanos-io/thanos/pkg/block/metadata"

	phlareobjstore "github.com/grafana/phlare/pkg/objstore"
	"github.com/grafana/phlare/pkg/phlaredb/block"
)

type BucketStore struct {
	bucket            phlareobjstore.Bucket
	tenantID, syncDir string

	logger log.Logger

	blocksMx sync.RWMutex
	blocks   map[ulid.ULID]*bucketBlock
}

func NewBucketStore(bucket phlareobjstore.Bucket, tenantID string, syncDir string, filters []BlockMetaFilter, logger log.Logger) (*BucketStore, error) {
	s := &BucketStore{
		bucket:   phlareobjstore.BucketWithPrefix(bucket, tenantID+"/phlaredb"),
		tenantID: tenantID,
		syncDir:  syncDir,
		logger:   logger,
	}

	if err := os.MkdirAll(syncDir, 0o750); err != nil {
		return nil, errors.Wrap(err, "create dir")
	}

	return s, nil
}

func (b *BucketStore) InitialSync(ctx context.Context) error {
	if err := b.SyncBlocks(ctx); err != nil {
		return errors.Wrap(err, "sync block")
	}

	fis, err := os.ReadDir(b.syncDir)
	if err != nil {
		return errors.Wrap(err, "read dir")
	}
	names := make([]string, 0, len(fis))
	for _, fi := range fis {
		names = append(names, fi.Name())
	}
	for _, n := range names {
		id, ok := block.IsBlockDir(n)
		if !ok {
			continue
		}
		if b := b.getBlock(id); b != nil {
			continue
		}

		// No such block loaded, remove the local dir.
		if err := os.RemoveAll(path.Join(b.syncDir, id.String())); err != nil {
			level.Warn(b.logger).Log("msg", "failed to remove block which is not needed", "err", err)
		}
	}

	return nil
}

func (s *BucketStore) getBlock(id ulid.ULID) *bucketBlock {
	s.blocksMx.RLock()
	defer s.blocksMx.RUnlock()
	return s.blocks[id]
}

func (s *BucketStore) SyncBlocks(ctx context.Context) error {
	metas, _, metaFetchErr := s.fetcher.Fetch(ctx)
	// For partial view allow adding new blocks at least.
	if metaFetchErr != nil && metas == nil {
		return metaFetchErr
	}

	var wg sync.WaitGroup
	blockc := make(chan *metadata.Meta)

	for i := 0; i < s.blockSyncConcurrency; i++ {
		wg.Add(1)
		go func() {
			for meta := range blockc {
				if err := s.addBlock(ctx, meta); err != nil {
					continue
				}
			}
			wg.Done()
		}()
	}

	for id, meta := range metas {
		if b := s.getBlock(id); b != nil {
			continue
		}
		select {
		case <-ctx.Done():
		case blockc <- meta:
		}
	}

	close(blockc)
	wg.Wait()

	if metaFetchErr != nil {
		return metaFetchErr
	}

	// Drop all blocks that are no longer present in the bucket.
	for id := range s.blocks {
		if _, ok := metas[id]; ok {
			continue
		}
		if err := s.removeBlock(id); err != nil {
			level.Warn(s.logger).Log("msg", "drop of outdated block failed", "block", id, "err", err)
		}
		level.Info(s.logger).Log("msg", "dropped outdated block", "block", id)
	}

	return nil
}

func (b *BucketStore) Stats() storegateway.BucketStoreStats {
	return storegateway.BucketStoreStats{}
}

// RemoveBlocksAndClose remove all blocks from local disk and releases all resources associated with the BucketStore.
func (s *BucketStore) RemoveBlocksAndClose() error {
	// err := s.removeAllBlocks()

	// // Release other resources even if it failed to close some blocks.
	// s.indexReaderPool.Close()

	// return err
	return nil
}
