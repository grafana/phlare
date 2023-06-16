package symdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func Test_Writer_OpenIndexFile(t *testing.T) {
	dir := filepath.Join("testdata", DefaultDirName)
	db := NewSymDB(&Config{
		Dir: dir,
		Stacktraces: StacktracesConfig{
			MaxNodesPerChunk: 5,
		},
	})

	w := db.MappingWriter(0)
	a := w.StacktraceAppender()
	defer a.Release()

	sids := make([]uint32, 4)
	a.AppendStacktrace(sids, []*schemav1.Stacktrace{
		{LocationIDs: []uint64{3, 2, 1}},
		{LocationIDs: []uint64{2, 1}},
		{LocationIDs: []uint64{4, 3, 2, 1}},
		{LocationIDs: []uint64{3, 1}},
	})
	a.AppendStacktrace(sids, []*schemav1.Stacktrace{
		{LocationIDs: []uint64{5, 2, 1}},
	})

	r := db.MappingReader(0).StacktraceResolver()
	dst := new(mockStacktraceInserter)
	dst.On("InsertStacktrace", uint32(2), []int32{2, 1})
	dst.On("InsertStacktrace", uint32(3), []int32{3, 2, 1})
	dst.On("InsertStacktrace", uint32(4), []int32{4, 3, 2, 1})
	dst.On("InsertStacktrace", uint32(7), []int32{3, 1})
	dst.On("InsertStacktrace", uint32(9), []int32{5, 2, 1})
	r.ResolveStacktraces(dst, []uint32{2, 3, 4, 7, 9})

	require.NoError(t, db.Flush())

	b, err := os.ReadFile(filepath.Join(dir, IndexFileName))
	require.NoError(t, err)

	idx, err := OpenIndexFile(b)
	require.NoError(t, err)
	assert.Len(t, idx.StacktraceChunkHeaders.Entries, 2)

	// TODO(kolesnikovae): Validate the index and headers.

	t.Log(pretty.Sprint(idx))
}
