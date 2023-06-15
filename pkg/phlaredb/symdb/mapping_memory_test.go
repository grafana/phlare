package symdb

import (
	"sync"
	"testing"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_StacktraceAppender_shards(t *testing.T) {
	db := NewSymDB()
	db.config = Config{MaxStacksPerChunk: 2}

	w := db.MappingWriter(0)
	a := w.StacktraceAppender()
	defer a.Release()

	sids := make([]int32, 4)
	a.AppendStacktrace(sids, []*schemav1.Stacktrace{
		{LocationIDs: []uint64{3, 2, 1}},
		{LocationIDs: []uint64{2, 1}},
		{LocationIDs: []uint64{4, 3, 2, 1}},
		{LocationIDs: []uint64{3, 1}},
	})
	assert.Equal(t, []int32{3, 2, 6, 7}, sids)

	a.AppendStacktrace(sids, []*schemav1.Stacktrace{
		{LocationIDs: []uint64{3, 2, 1}},
		{LocationIDs: []uint64{2, 1}},
		{LocationIDs: []uint64{4, 3, 2, 1}},
	})
	// Same input. Note that len(sids) > len(schemav1.Stacktrace)
	assert.Equal(t, []int32{3, 2, 6, 7}, sids)

	a.AppendStacktrace(sids, []*schemav1.Stacktrace{
		{LocationIDs: []uint64{5, 2, 1}},
	})
	assert.Equal(t, []int32{7, 2, 6, 7}, sids)

	require.Len(t, db.mappings, 1)
	m := db.mappings[0]
	require.Len(t, m.stacktraceChunks, 3)

	c1 := m.stacktraceChunks[0]
	assert.Equal(t, int32(0), c1.stid)
	assert.Equal(t, int32(2), c1.stacks)

	c2 := m.stacktraceChunks[1]
	assert.Equal(t, int32(2), c2.stid)
	assert.Equal(t, int32(2), c2.stacks)

	c3 := m.stacktraceChunks[2]
	assert.Equal(t, int32(4), c3.stid)
	assert.Equal(t, int32(1), c3.stacks)
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