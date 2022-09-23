package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/grafana/fire/pkg/fire"
	"github.com/grafana/fire/pkg/firedb"
	"github.com/grafana/fire/pkg/objstore"
	"github.com/grafana/fire/pkg/objstore/client"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
)

type QueryConfig struct {
	Storage fire.StorageConfig `yaml:"storage"`
}

func pprof(ctx context.Context, since time.Duration, query string) error {
	logger := log.NewLogfmtLogger(os.Stdout)
	var queryConfig QueryConfig
	yamlFile, err := ioutil.ReadFile(cfg.query.config)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(yamlFile, &queryConfig); err != nil {
		return err
	}
	bucket, err := client.NewBucket(logger, []byte(queryConfig.Storage.BucketConfig), prometheus.DefaultRegisterer, "firetool")
	if err != nil {
		return err
	}
	prefixedBucket := objstore.BucketWithPrefix(bucket, "firedb")
	querier := firedb.NewBlockQuerier(logger, prefixedBucket)
	metas, err := querier.BlockMetas(ctx)
	if err != nil {
		return err
	}
	for _, meta := range metas {
		fmt.Println(meta)
	}
	return nil
}
