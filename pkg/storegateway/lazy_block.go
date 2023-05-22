package storegateway

import (
	"context"
	"os"
	"path/filepath"

	"github.com/dgraph-io/ristretto"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"

	"github.com/grafana/phlare/pkg/phlaredb"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	"github.com/grafana/phlare/pkg/util"
)

type lazyBlock struct {
	meta   *block.Meta
	logger log.Logger
}

func OpenFromDisk(dir string, meta *block.Meta, logger log.Logger) (*lazyBlock, error) {
	blockLocalPath := filepath.Join(dir, meta.ULID.String())
	// add the dir if it doesn't exist
	if _, err := os.Stat(blockLocalPath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, errors.Wrap(err, "create dir")
		}
	}
	metaPath := filepath.Join(dir, block.MetaFilename)
	if _, err := os.Stat(metaPath); errors.Is(err, os.ErrNotExist) {
		// add meta.json if it does not exist
		if _, err := meta.WriteToFile(logger, dir); err != nil {
			return nil, errors.Wrap(err, "write meta.json")
		}
	} else {
		// read meta.json if it exists and validate it
		if diskMeta, _, err := block.MetaFromDir(dir); err != nil {
			if meta.String() != diskMeta.String() {
				return nil, errors.Wrap(err, "meta.json does not match")
			}
			return nil, errors.Wrap(err, "read meta.json")
		}
	}

	return &lazyBlock{
		meta:   meta,
		logger: logger,
	}, nil
}

func (b *lazyBlock) Load(ctx context.Context) error {
	// load the block from the object store
	// reads strings, functions, and series
	return nil
}

func (b *lazyBlock) Close() error {
	return nil
}

type Block interface {
	phlaredb.Querier
	Open(context.Context) error
	Close() error
}

type blocksCache struct {
	lru        *ristretto.Cache
	openBlocks func(context.Context, []*block.Meta, string) ([]Block, error)
}

func NewBlocksCache(openBlocks func(context.Context, []*block.Meta, string) ([]Block, error)) (*blocksCache, error) {
	// 500 block max for now of cost 1 which is around 20MB per block so 10GB of memory
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 2500,
		MaxCost:     500,
		BufferItems: 64,
		OnEvict: func(item *ristretto.Item) {
			err := item.Value.(Block).Close()
			if err != nil {
				level.Error(util.Logger).Log("msg", "block cache on-evict close", "err", err)
			}
		},
	})
	if err != nil {
		return nil, err
	}
	return &blocksCache{
		lru:        c,
		openBlocks: openBlocks,
	}, nil
}

func cacheKey(meta *block.Meta, tenantID string) string {
	return tenantID + "/" + meta.ULID.String()
}

func (b *blocksCache) Get(ctx context.Context, metas []*block.Meta, tenantID string) ([]Block, error) {
	result := make([]Block, len(metas))
	missingIdx := []int{}

	for i, m := range metas {
		if block, ok := b.lru.Get(cacheKey(m, tenantID)); ok {
			result[i] = block.(Block)
			continue
		}
		missingIdx = append(missingIdx, i)
	}

	if len(missingIdx) == 0 {
		return result, nil
	}
	missing := make([]*block.Meta, len(missingIdx))
	for i, idx := range missingIdx {
		missing[i] = metas[idx]
	}
	missingBlocks, err := b.openBlocks(ctx, missing, tenantID)
	if err != nil {
		return nil, err
	}
	for i, block := range missingBlocks {
		result[missingIdx[i]] = block
		// todo cost... by measuring the size of the block

		if ok := b.lru.Set(cacheKey(missing[i], tenantID), block, 1); !ok {
			// block is rejected by the cache, closing should be done by the caller.
			level.Debug(util.Logger).Log("msg", "block cache rejected", "block", missing[i].ULID.String())
			continue
		}
		// the block is cached so we don't want to close it, eviction will close it.
		result[missingIdx[i]] = noopCloseBlock{block}
	}

	return result, nil
}

type noopCloseBlock struct {
	Block
}

func (n noopCloseBlock) Close() error {
	return nil
}
