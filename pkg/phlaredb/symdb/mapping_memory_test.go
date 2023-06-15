package symdb

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func Test_StacktraceAppender_shards(t *testing.T) {
	t.Run("WithMaxStacktraceTreeNodesPerChunk", func(t *testing.T) {
		db := NewSymDB()
		db.config.MaxStacktraceTreeNodesPerChunk = 5

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
		assert.Equal(t, []uint32{3, 2, 4, 7}, sids)

		a.AppendStacktrace(sids, []*schemav1.Stacktrace{
			{LocationIDs: []uint64{3, 2, 1}},
			{LocationIDs: []uint64{2, 1}},
			{LocationIDs: []uint64{4, 3, 2, 1}},
		})
		// Same input. Note that len(sids) > len(schemav1.Stacktrace)
		assert.Equal(t, []uint32{3, 2, 4, 7}, sids)

		a.AppendStacktrace(sids, []*schemav1.Stacktrace{
			{LocationIDs: []uint64{5, 2, 1}},
		})
		assert.Equal(t, []uint32{9, 2, 4, 7}, sids)

		require.Len(t, db.mappings, 1)
		m := db.mappings[0]
		require.Len(t, m.stacktraceChunks, 2)

		c1 := m.stacktraceChunks[0]
		assert.Equal(t, uint32(0), c1.stid)
		assert.Equal(t, uint32(5), c1.tree.len())

		c2 := m.stacktraceChunks[1]
		assert.Equal(t, uint32(5), c2.stid)
		assert.Equal(t, uint32(5), c2.tree.len())
	})

	t.Run("WithoutMaxStacktraceTreeNodesPerChunk", func(t *testing.T) {
		db := NewSymDB()
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
		assert.Equal(t, []uint32{3, 2, 4, 5}, sids)

		a.AppendStacktrace(sids, []*schemav1.Stacktrace{
			{LocationIDs: []uint64{3, 2, 1}},
			{LocationIDs: []uint64{2, 1}},
			{LocationIDs: []uint64{4, 3, 2, 1}},
		})
		// Same input. Note that len(sids) > len(schemav1.Stacktrace)
		assert.Equal(t, []uint32{3, 2, 4, 5}, sids)

		a.AppendStacktrace(sids, []*schemav1.Stacktrace{
			{LocationIDs: []uint64{5, 2, 1}},
		})
		assert.Equal(t, []uint32{6, 2, 4, 5}, sids)

		require.Len(t, db.mappings, 1)
		m := db.mappings[0]
		require.Len(t, m.stacktraceChunks, 1)

		c1 := m.stacktraceChunks[0]
		assert.Equal(t, uint32(0), c1.stid)
		assert.Equal(t, uint32(7), c1.tree.len())
	})
}

func Test_hashLocations(t *testing.T) {
	t.Run("hashLocations is thread safe", func(t *testing.T) {
		b := []uint64{123, 234, 345, 456, 567}
		h := hashLocations(b)
		const N, M = 10, 10 << 10
		var wg sync.WaitGroup
		wg.Add(N)
		for i := 0; i < N; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < M; j++ {
					if hashLocations(b) != h {
						panic("hash mismatch")
					}
				}
			}()
		}
		wg.Wait()
	})
}
