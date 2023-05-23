package storegateway

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"

	"github.com/grafana/phlare/pkg/phlaredb"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	"github.com/grafana/phlare/pkg/util"
)

type BlockCloser interface {
	phlaredb.Querier
	Close() error
}

type Block struct {
	BlockCloser
	meta   *block.Meta
	logger log.Logger
}

func (bs *BucketStore) createBlock(ctx context.Context, meta *block.Meta) (*Block, error) {
	blockLocalPath := filepath.Join(bs.syncDir, meta.ULID.String())
	// add the dir if it doesn't exist
	if _, err := os.Stat(blockLocalPath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(bs.syncDir, 0o750); err != nil {
			return nil, errors.Wrap(err, "create dir")
		}
	}
	metaPath := filepath.Join(bs.syncDir, block.MetaFilename)
	if _, err := os.Stat(metaPath); errors.Is(err, os.ErrNotExist) {
		// add meta.json if it does not exist
		if _, err := meta.WriteToFile(bs.logger, bs.syncDir); err != nil {
			return nil, errors.Wrap(err, "write meta.json")
		}
	} else {
		// read meta.json if it exists and validate it
		if diskMeta, _, err := block.MetaFromDir(bs.syncDir); err != nil {
			if meta.String() != diskMeta.String() {
				return nil, errors.Wrap(err, "meta.json does not match")
			}
			return nil, errors.Wrap(err, "read meta.json")
		}
	}

	blk := phlaredb.NewSingleBlockQuerierFromMeta(ctx, bs.bucket, meta)
	// Load the block into memory if it's within the last 24 hours.
	// Todo make this configurable
	if blk.InRange(model.Now().Add(-24*time.Hour), model.Now()) {
		go func() {
			if err := blk.Open(ctx); err != nil {
				level.Error(util.Logger).Log("msg", "open block", "err", err)
			}
		}()
	}
	return &Block{
		meta:        meta,
		logger:      bs.logger,
		BlockCloser: blk,
	}, nil
}
