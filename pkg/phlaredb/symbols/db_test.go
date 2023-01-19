package symbols

import (
	"os"
	"testing"

	"github.com/dustin/go-humanize"
	"github.com/segmentio/parquet-go"
	"github.com/stretchr/testify/require"

	profilev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	"github.com/grafana/phlare/pkg/pprof"
)

func Test_Store_Symbols(t *testing.T) {
	store := NewStore()
	prof := parseProfile(t, "../testdata/heap")
	for i := 0; i < 100; i++ {
		err := store.Ingest(prof)
		require.NoError(t, err)
	}
	tmp := t.TempDir()
	t.Log(humanize.Bytes(uint64(store.Size())))
	err := store.Flush(tmp + "/symbols.parquet")
	require.NoError(t, err)
	f, err := os.OpenFile(tmp+"/symbols.parquet", os.O_RDONLY, 0o644)
	require.NoError(t, err)
	info, err := f.Stat()
	require.NoError(t, err)

	pq, err := parquet.OpenFile(f, info.Size())
	require.NoError(t, err)
	t.Log(pq.Metadata())
}

func parseProfile(t testing.TB, path string) *profilev1.Profile {
	p, err := pprof.OpenFile(path)
	require.NoError(t, err, "failed opening profile: ", path)
	return p.Profile
}
