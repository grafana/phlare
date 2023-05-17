package storegateway

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/pkg/errors"

	phlareobjstore "github.com/grafana/phlare/pkg/objstore"
	"github.com/grafana/phlare/pkg/phlaredb/block"
)

type lazyBlock struct {
	meta   *block.Meta
	logger log.Logger
	bucker phlareobjstore.Bucket
}

func OpenFromDisk(dir string, meta *block.Meta, bucket phlareobjstore.Bucket, logger log.Logger) (*lazyBlock, error) {
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
		bucker: bucket,
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
