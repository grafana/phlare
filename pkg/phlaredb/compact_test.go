package phlaredb

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	_ "net/http/pprof"

	"github.com/grafana/phlare/pkg/objstore/client"
	"github.com/grafana/phlare/pkg/objstore/providers/filesystem"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	"github.com/prometheus/common/model"
)

func TestCompact(t *testing.T) {
	go func() {
		t.Log(http.ListenAndServe("localhost:6060", nil))
	}()
	metasMap, err := block.ListBlock("./compactor_testdata/", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("total blocks:", len(metasMap))
	var minTime, maxTime model.Time
	for i := range metasMap {
		if minTime == 0 || metasMap[i].MinTime < minTime {
			minTime = metasMap[i].MinTime
		}

		if maxTime == 0 || metasMap[i].MaxTime > maxTime {
			maxTime = metasMap[i].MaxTime
		}
	}

	ctx := context.Background()
	bkt, err := client.NewBucket(ctx, client.Config{
		StorageBackendConfig: client.StorageBackendConfig{
			Backend: client.Filesystem,
			Filesystem: filesystem.Config{
				Directory: "./compactor_testdata/",
			},
		},
	}, "local")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll("./compactor_testdata/compacted/"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll("./compactor_testdata/compacted/", 0o755); err != nil {
		t.Fatal(err)
	}
	blocks := make([]*singleBlockQuerier, len(metasMap))
	j := 0
	for i := range metasMap {
		blocks[j] = NewSingleBlockQuerierFromMeta(ctx, bkt, metasMap[i])
		j++
	}
	if err := Compact(ctx, blocks, "./compactor_testdata/compacted/"); err != nil {
		t.Fatal(err)
	}
}
