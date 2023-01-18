package phlaredb

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/go-kit/log"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	phlarecontext "github.com/grafana/phlare/pkg/phlare/context"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func testContext(t *testing.T) context.Context {
	logger := log.NewNopLogger()
	if testing.Verbose() {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	return phlarecontext.WithLogger(context.Background(), logger)
}

func stringWithID(id int) *schemav1.String {
	return &schemav1.String{String: "value" + strconv.FormatInt(int64(id), 10)}
}

func constantString(id int) *schemav1.String {
	return &schemav1.String{String: "constant"}
}

func TestDeduplicatingStore_Ingestion(t *testing.T) {
	var (
		ctx   = testContext(t)
		store = newStringsStore(ctx, defaultParquetConfig)
	)

	for _, tc := range []struct {
		name            string
		cfg             *ParquetConfig
		expectedNumRows uint64
		expectedNumRGs  uint64
		values          func(int) *schemav1.String
	}{
		{
			name:            "single row group",
			cfg:             defaultParquetConfig,
			expectedNumRGs:  1,
			expectedNumRows: 100,
			values:          stringWithID,
		},
		{
			name:            "single row group, same string value",
			cfg:             defaultParquetConfig,
			expectedNumRGs:  1,
			expectedNumRows: 1,
			values:          constantString,
		},
		{
			name:            "multiple row groups because of maximum size",
			cfg:             &ParquetConfig{MaxRowGroupBytes: 110, MaxBufferRowCount: 100000},
			expectedNumRGs:  10,
			expectedNumRows: 100,
			values:          stringWithID,
		},
		{
			name:            "multiple row groups because of maximum row num",
			cfg:             &ParquetConfig{MaxRowGroupBytes: 128000, MaxBufferRowCount: 10},
			expectedNumRGs:  10,
			expectedNumRows: 100,
			values:          stringWithID,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store.cfg = tc.cfg
			path := t.TempDir()
			store.Reset(path)

			for i := 0; i < 100; i++ {
				store.ingest(ctx, []*schemav1.String{tc.values(i)}, &rewriter{})
			}

			// ensure the correct number of files are created
			numRows, numRGs, err := store.Flush()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedNumRows, numRows)
			assert.Equal(t, tc.expectedNumRGs, numRGs)

			// list folder to ensure only aggregted block exists
			files, err := os.ReadDir(path)
			require.NoError(t, err)
			require.Equal(t, []string{"strings.parquet"}, lo.Map(files, func(e os.DirEntry, _ int) string {
				return e.Name()
			}))
		})
	}
}

func TestDeduplicatingStore_FlushRowGroupsToDisk(t *testing.T) {
	var (
		ctx   = testContext(t)
		store = newStringsStore(ctx, defaultParquetConfig)
	)

	for _, tc := range []struct {
		name    string
		enabled bool
	}{
		{
			name:    "do not flush to disk",
			enabled: false,
		},
		{
			name:    "flush to disk",
			enabled: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store.cfg = &ParquetConfig{
				MaxRowGroupBytes:        12800000,
				MaxBufferRowCount:       10,
				FlushEachRowGroupToDisk: &tc.enabled,
			}
			path := t.TempDir()
			store.Reset(path)

			for i := 0; i < 100; i++ {
				store.ingest(ctx, []*schemav1.String{stringWithID(i)}, &rewriter{})
			}

			// TODO List files here

			// ensure the correct number of files are created
			numRows, numRGs, err := store.Flush()

			require.NoError(t, err)
			assert.Equal(t, uint64(100), numRows)
			assert.Equal(t, uint64(10), numRGs)

			// list folder to ensure only aggregted block exists
			files, err := os.ReadDir(path)
			require.NoError(t, err)
			require.Equal(t, []string{"strings.parquet"}, lo.Map(files, func(e os.DirEntry, _ int) string {
				return e.Name()
			}))
		})
	}
}
