package storegateway

import (
	"github.com/grafana/phlare/pkg/phlaredb/block"
)

type bucketBlock struct {
	Meta *block.Meta
}

func NewBucketBlock(meta *block.Meta) *bucketBlock {
	return &bucketBlock{
		Meta: meta,
	}
}

func (b *bucketBlock) Close() error {
	return nil
}
