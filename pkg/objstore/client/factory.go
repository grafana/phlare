package client

import (
	"context"

	"github.com/thanos-io/objstore"

	phlareobj "github.com/grafana/phlare/pkg/objstore"
	"github.com/grafana/phlare/pkg/objstore/providers/azure"
	"github.com/grafana/phlare/pkg/objstore/providers/cos"
	"github.com/grafana/phlare/pkg/objstore/providers/filesystem"
	"github.com/grafana/phlare/pkg/objstore/providers/gcs"
	"github.com/grafana/phlare/pkg/objstore/providers/s3"
	"github.com/grafana/phlare/pkg/objstore/providers/swift"
	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
)

// NewBucket creates a new bucket client based on the configured backend
func NewBucket(ctx context.Context, cfg Config, name string) (phlareobj.Bucket, error) {
	var (
		backendClient objstore.Bucket
		err           error
	)
	logger := phlarecontext.Logger(ctx)
	reg := phlarecontext.Registry(ctx)

	switch cfg.Backend {
	case S3:
		backendClient, err = s3.NewBucketClient(cfg.S3, name, logger)
	case GCS:
		backendClient, err = gcs.NewBucketClient(ctx, cfg.GCS, name, logger)
	case Azure:
		backendClient, err = azure.NewBucketClient(cfg.Azure, name, logger)
	case Swift:
		backendClient, err = swift.NewBucketClient(cfg.Swift, name, logger)
	case COS:
		backendClient, err = cos.NewBucketClient(cfg.COS, name, logger)
	case Filesystem:
		backendClient, err = filesystem.NewBucket(cfg.Filesystem.Directory)
	default:
		return nil, ErrUnsupportedStorageBackend
	}

	if err != nil {
		return nil, err
	}

	// Wrap the client with any provided middleware
	for _, wrap := range cfg.Middlewares {
		backendClient, err = wrap(backendClient)
		if err != nil {
			return nil, err
		}
	}
	bkt := phlareobj.NewBucket(objstore.NewTracingBucket(objstore.BucketWithMetrics(name, backendClient, reg)))

	if cfg.StoragePrefix != "" {
		bkt = phlareobj.NewPrefixedBucket(bkt, cfg.StoragePrefix)
	}
	return bkt, nil
}
