package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/cache"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	"github.com/oklog/ulid"
	"github.com/samber/lo"
	"github.com/thanos-io/objstore"
	"github.com/thanos-io/objstore/providers/gcs"
	"golang.org/x/sync/errgroup"
	yaml "gopkg.in/yaml.v3"
)

var ulidEntropy = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateULID() ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(time.Now()), ulidEntropy)
}

const orderOfSplit = 10 // how many bytes of the ulid id are used

// orderOfSplit is the number of bytes of the ulid id used for the split. The duration of the split is:
// 0: 1114y
// 1: 34.8y
// 2: 1y
// 3: 12.4d
// 4: 9h19m
// TODO: To needs to be adapted based on the MaxBlockDuration.
func blockPrefixesFromTo(from, to time.Time, orderOfSplit uint8) (prefixes []string, err error) {
	var id ulid.ULID

	if orderOfSplit < 0 || orderOfSplit > 9 {
		return nil, fmt.Errorf("order of split must be between 0 and 9")
	}

	byteShift := (9 - orderOfSplit) * 5

	ms := uint64(from.UnixMilli()) >> byteShift
	ms = ms << byteShift
	for ms <= uint64(to.UnixMilli()) {
		if err := id.SetTime(ms); err != nil {
			return nil, err
		}
		prefixes = append(prefixes, id.String()[:orderOfSplit+1])

		ms = ms >> byteShift
		ms += 1
		ms = ms << byteShift
	}

	return prefixes, nil
}

type BucketQuerier struct {
	bucketCacheCfg cache.BackendConfig
	bucketCache    cache.Cache

	gcsConfig gcs.Config
}

const tenantID = "27821"

func NewBucketQuerier() *BucketQuerier {
	bq := &BucketQuerier{}
	bq.gcsConfig.Bucket = "ops-us-east-0-profiles-ops-001-data"

	bq.bucketCacheCfg.Backend = "memcached"
	bq.bucketCacheCfg.Memcached.Addresses = "localhost:11211"
	bq.bucketCacheCfg.Memcached.MaxAsyncConcurrency = 16

	c, err := cache.CreateClient("bucket-cache", bq.bucketCacheCfg, logger, nil)
	if err != nil {
		panic(err)
	}
	bq.bucketCache = c

	return bq
}

var logger = log.NewLogfmtLogger(os.Stdout)

func main() {
	if err := NewBucketQuerier().run(); err != nil {
		level.Error(logger).Log("msg", "failed to run", "err", err)
		os.Exit(1)
	}
}

func (bq *BucketQuerier) run() error {

	to := time.Now()
	from := to.Add(-time.Hour * 24 * 7)

	blockPrefixes, err := blockPrefixesFromTo(from, to, 4)
	if err != nil {
		return err
	}

	level.Info(logger).Log("msg", "block prefixes to query", "prefixes", fmt.Sprintf("%#v", blockPrefixes))

	// Thanos currently doesn't support passing the config as is, but expects a YAML,
	// so we're going to serialize it.
	serialized, err := yaml.Marshal(bq.gcsConfig)
	if err != nil {
		return err
	}

	logger := log.NewLogfmtLogger(os.Stdout)

	ctx := context.Background()
	client, err := gcs.NewBucket(ctx, logger, serialized, "test")
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	pathPrefix := tenantID + "/phlaredb/"

	// ask cache first
	var (
		cachePrefixes = make(map[string]int, len(blockPrefixes))
		metas         = make([][]*block.Meta, len(blockPrefixes))
	)
	for idx := range blockPrefixes {
		cachePrefixes["blocklist/"+tenantID+"/"+blockPrefixes[idx]] = idx
	}

	result := bq.bucketCache.Fetch(ctx, lo.Keys(cachePrefixes))
	for k, v := range result {
		if err := json.Unmarshal(v, &metas[cachePrefixes[k]]); err != nil {
			return err
		}
		delete(cachePrefixes, k)
	}

	// now only use remaining block prefixes
	for _, idx := range lo.Values(cachePrefixes) {
		var idx = idx
		g.Go(func() error {
			blockPrefix := blockPrefixes[idx]
			level.Info(logger).Log("msg", "listing", "prefix", blockPrefix)
			err = client.Iter(ctx, pathPrefix+blockPrefix, func(name string) error {
				level.Info(logger).Log("msg", "found block", "id", name)

				// now read metadata
				r, err := client.Get(ctx, name+"meta.json")
				if err != nil {
					return err
				}

				m, err := block.Read(r)
				if err != nil {
					return err
				}
				metas[idx] = append(metas[idx], m)

				if err := r.Close(); err != nil {
					return err
				}

				return nil
			}, objstore.WithoutApendingDirDelim)
			if err != nil {
				return err
			}

			metasBytes, err := json.Marshal(metas[idx])
			if err != nil {
				return err
			}

			// store in cache
			bq.bucketCache.Store(ctx, map[string][]byte{
				"blocklist/" + tenantID + "/" + blockPrefix: metasBytes,
			}, time.Hour)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	for idx, metas := range metas {
		var samples, series uint64
		for _, m := range metas {
			samples += m.Stats.NumSamples
			series += m.Stats.NumSeries
		}
		level.Info(logger).Log(
			"msg", "found block",
			"prefix", blockPrefixes[idx],
			"block_count", len(metas),
			"sample_count", samples,
			"series_count", series,
		)

	}

	return nil

}
