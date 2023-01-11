package phlaredb

import (
	"context"
	"testing"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/stretchr/testify/require"
)

func TestDeduplicatingSlice_Ingest_Spit(t *testing.T) {

	var (
		store deduplicatingSlice[schemav1.String, string, *stringsHelper, *schemav1.StringPersister]
		ctx   = context.Background()
	)

	require.NoError(t, store.Init(t.TempDir(), &ParquetConfig{
		MaxBufferRowCount: 10,
		MaxRowGroupBytes:  25000,
	}))

	for i := 0; i < 100; i++ {
		store.ingest(ctx, []*schemav1.String{{String: "test"}}, &rewriter{})
	}
}
