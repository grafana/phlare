package storegateway

import (
	"context"
	"flag"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/mimir/pkg/storegateway"
	phlareobjstore "github.com/grafana/phlare/pkg/objstore"
	"github.com/grafana/phlare/pkg/phlaredb/bucket"
	"github.com/prometheus/client_golang/prometheus"
)

type BucketStoreConfig struct {
	SyncDir      string        `yaml:"sync_dir"`
	SyncInterval time.Duration `yaml:"sync_interval" category:"advanced"`
}

// RegisterFlags registers the BucketStore flags
func (cfg *BucketStoreConfig) RegisterFlags(f *flag.FlagSet, logger log.Logger) {
	// cfg.IndexCache.RegisterFlagsWithPrefix(f, "blocks-storage.bucket-store.index-cache.")
	// cfg.ChunksCache.RegisterFlagsWithPrefix(f, "blocks-storage.bucket-store.chunks-cache.", logger)
	// cfg.MetadataCache.RegisterFlagsWithPrefix(f, "blocks-storage.bucket-store.metadata-cache.")
	// cfg.BucketIndex.RegisterFlagsWithPrefix(f, "blocks-storage.bucket-store.bucket-index.")
	// cfg.IndexHeader.RegisterFlagsWithPrefix(f, "blocks-storage.bucket-store.index-header.")

	f.StringVar(&cfg.SyncDir, "blocks-storage.bucket-store.sync-dir", "./tsdb-sync/", "Directory to store synchronized TSDB index headers. This directory is not required to be persisted between restarts, but it's highly recommended in order to improve the store-gateway startup time.")
	f.DurationVar(&cfg.SyncInterval, "blocks-storage.bucket-store.sync-interval", 15*time.Minute, "How frequently to scan the bucket, or to refresh the bucket index (if enabled), in order to look for changes (new blocks shipped by ingesters and blocks deleted by retention or compaction).")
	// f.Uint64Var(&cfg.MaxChunkPoolBytes, "blocks-storage.bucket-store.max-chunk-pool-bytes", uint64(2*units.Gibibyte), "Max size - in bytes - of a chunks pool, used to reduce memory allocations. The pool is shared across all tenants. 0 to disable the limit.")
	// f.IntVar(&cfg.ChunkPoolMinBucketSizeBytes, "blocks-storage.bucket-store.chunk-pool-min-bucket-size-bytes", ChunkPoolDefaultMinBucketSize, "Size - in bytes - of the smallest chunks pool bucket.")
	// f.IntVar(&cfg.ChunkPoolMaxBucketSizeBytes, "blocks-storage.bucket-store.chunk-pool-max-bucket-size-bytes", ChunkPoolDefaultMaxBucketSize, "Size - in bytes - of the largest chunks pool bucket.")
	// f.Uint64Var(&cfg.SeriesHashCacheMaxBytes, "blocks-storage.bucket-store.series-hash-cache-max-size-bytes", uint64(1*units.Gibibyte), "Max size - in bytes - of the in-memory series hash cache. The cache is shared across all tenants and it's used only when query sharding is enabled.")
	// f.IntVar(&cfg.MaxConcurrent, "blocks-storage.bucket-store.max-concurrent", 100, "Max number of concurrent queries to execute against the long-term storage. The limit is shared across all tenants.")
	// f.IntVar(&cfg.TenantSyncConcurrency, "blocks-storage.bucket-store.tenant-sync-concurrency", 10, "Maximum number of concurrent tenants synching blocks.")
	// f.IntVar(&cfg.BlockSyncConcurrency, "blocks-storage.bucket-store.block-sync-concurrency", 20, "Maximum number of concurrent blocks synching per tenant.")
	// f.IntVar(&cfg.MetaSyncConcurrency, "blocks-storage.bucket-store.meta-sync-concurrency", 20, "Number of Go routines to use when syncing block meta files from object storage per tenant.")
	// f.DurationVar(&cfg.DeprecatedConsistencyDelay, consistencyDelayFlag, 0, "Minimum age of a block before it's being read. Set it to safe value (e.g 30m) if your object storage is eventually consistent. GCS and S3 are (roughly) strongly consistent.")
	// f.DurationVar(&cfg.IgnoreDeletionMarksDelay, "blocks-storage.bucket-store.ignore-deletion-marks-delay", time.Hour*1, "Duration after which the blocks marked for deletion will be filtered out while fetching blocks. "+
	// 	"The idea of ignore-deletion-marks-delay is to ignore blocks that are marked for deletion with some delay. This ensures store can still serve blocks that are meant to be deleted but do not have a replacement yet.")
	// f.DurationVar(&cfg.IgnoreBlocksWithin, "blocks-storage.bucket-store.ignore-blocks-within", 10*time.Hour, "Blocks with minimum time within this duration are ignored, and not loaded by store-gateway. Useful when used together with -querier.query-store-after to prevent loading young blocks, because there are usually many of them (depending on number of ingesters) and they are not yet compacted. Negative values or 0 disable the filter.")
	// f.IntVar(&cfg.PostingOffsetsInMemSampling, "blocks-storage.bucket-store.posting-offsets-in-mem-sampling", DefaultPostingOffsetInMemorySampling, "Controls what is the ratio of postings offsets that the store will hold in memory.")
	// f.BoolVar(&cfg.IndexHeaderLazyLoadingEnabled, "blocks-storage.bucket-store.index-header-lazy-loading-enabled", true, "If enabled, store-gateway will lazy load an index-header only once required by a query.")
	// f.DurationVar(&cfg.IndexHeaderLazyLoadingIdleTimeout, "blocks-storage.bucket-store.index-header-lazy-loading-idle-timeout", 60*time.Minute, "If index-header lazy loading is enabled and this setting is > 0, the store-gateway will offload unused index-headers after 'idle timeout' inactivity.")
	// f.Uint64Var(&cfg.PartitionerMaxGapBytes, "blocks-storage.bucket-store.partitioner-max-gap-bytes", DefaultPartitionerMaxGapSize, "Max size - in bytes - of a gap for which the partitioner aggregates together two bucket GET object requests.")
	// f.IntVar(&cfg.StreamingBatchSize, "blocks-storage.bucket-store.batch-series-size", 5000, "This option controls how many series to fetch per batch. The batch size must be greater than 0.")
	// f.IntVar(&cfg.ChunkRangesPerSeries, "blocks-storage.bucket-store.fine-grained-chunks-caching-ranges-per-series", 1, "This option controls into how many ranges the chunks of each series from each block are split. This value is effectively the number of chunks cache items per series per block when -blocks-storage.bucket-store.chunks-cache.fine-grained-chunks-caching-enabled is enabled.")
	// f.StringVar(&cfg.SeriesSelectionStrategyName, "blocks-storage.bucket-store.series-selection-strategy", AllPostingsStrategy, "This option controls the strategy to selection of series and deferring application of matchers. A more aggressive strategy will fetch less posting lists at the cost of more series. This is useful when querying large blocks in which many series share the same label name and value. Supported values (most aggressive to least aggressive): "+strings.Join(validSeriesSelectionStrategies, ", ")+".")
}

// Validate the config.
func (cfg *BucketStoreConfig) Validate(logger log.Logger) error {
	// if cfg.StreamingBatchSize <= 0 {
	// 	return errInvalidStreamingBatchSize
	// }
	// if err := cfg.IndexCache.Validate(); err != nil {
	// 	return errors.Wrap(err, "index-cache configuration")
	// }
	// if err := cfg.ChunksCache.Validate(); err != nil {
	// 	return errors.Wrap(err, "chunks-cache configuration")
	// }
	// if err := cfg.MetadataCache.Validate(); err != nil {
	// 	return errors.Wrap(err, "metadata-cache configuration")
	// }
	// if cfg.DeprecatedConsistencyDelay > 0 {
	// 	util.WarnDeprecatedConfig(consistencyDelayFlag, logger)
	// }
	// if !util.StringsContain(validSeriesSelectionStrategies, cfg.SeriesSelectionStrategyName) {
	// 	return errors.New("invalid series-selection-strategy, set one of " + strings.Join(validSeriesSelectionStrategies, ", "))
	// }
	return nil
}

type BucketStores struct {
	storageBucket phlareobjstore.Bucket
}

func NewBucketStores(shardingStrategy storegateway.ShardingStrategy, storageBucket phlareobjstore.Bucket, limits Limits, logger log.Logger, reg prometheus.Registerer) (*BucketStores, error) {
	m := &BucketStores{
		storageBucket: storageBucket,
	}
	return m, nil
}

func (bs *BucketStores) SyncBlocks(ctx context.Context) error {
	return nil
}

func (bs *BucketStores) InitialSync(ctx context.Context) error {
	return nil
}

func (bs *BucketStores) scanUsers(ctx context.Context) ([]string, error) {
	return bucket.ListUsers(ctx, bs.storageBucket)
}
