package symdb

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func Test_StacktraceAppender_shards(t *testing.T) {
	t.Run("WithMaxStacktraceTreeNodesPerChunk", func(t *testing.T) {
		db := NewSymDB(&Config{
			Stacktraces: StacktracesConfig{
				MaxNodesPerChunk: 7,
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
		assert.Equal(t, []uint32{3, 2, 11, 16}, sids)

		a.AppendStacktrace(sids[:3], []*schemav1.Stacktrace{
			{LocationIDs: []uint64{3, 2, 1}},
			{LocationIDs: []uint64{2, 1}},
			{LocationIDs: []uint64{4, 3, 2, 1}},
		})
		// Same input. Note that len(sids) > len(schemav1.Stacktrace)
		assert.Equal(t, []uint32{3, 2, 11}, sids[:3])

		a.AppendStacktrace(sids[:1], []*schemav1.Stacktrace{
			{LocationIDs: []uint64{5, 2, 1}},
		})
		assert.Equal(t, []uint32{18}, sids[:1])

		require.Len(t, db.mappings, 1)
		m := db.mappings[0]
		require.Len(t, m.stacktraceChunks, 3)

		c1 := m.stacktraceChunks[0]
		assert.Equal(t, uint32(0), c1.stid)
		assert.Equal(t, uint32(4), c1.tree.len())

		c2 := m.stacktraceChunks[1]
		assert.Equal(t, uint32(7), c2.stid)
		assert.Equal(t, uint32(5), c2.tree.len())

		c3 := m.stacktraceChunks[2]
		assert.Equal(t, uint32(14), c3.stid)
		assert.Equal(t, uint32(5), c3.tree.len())
	})

	t.Run("WithoutMaxStacktraceTreeNodesPerChunk", func(t *testing.T) {
		db := NewSymDB(new(Config))
		w := db.MappingWriter(0)
		a := w.StacktraceAppender()
		defer a.Release()

		sids := make([]uint32, 5)
		a.AppendStacktrace(sids, []*schemav1.Stacktrace{
			{LocationIDs: []uint64{3, 2, 1}},
			{LocationIDs: []uint64{2, 1}},
			{LocationIDs: []uint64{4, 3, 2, 1}},
			{LocationIDs: []uint64{3, 1}},
			{LocationIDs: []uint64{5, 3, 2, 1}},
		})
		assert.Equal(t, []uint32{3, 2, 4, 5, 6}, sids)

		require.Len(t, db.mappings, 1)
		m := db.mappings[0]
		require.Len(t, m.stacktraceChunks, 1)

		c1 := m.stacktraceChunks[0]
		assert.Equal(t, uint32(0), c1.stid)
		assert.Equal(t, uint32(7), c1.tree.len())
	})
}

func Test_StacktraceResolver_stacktraces_split(t *testing.T) {
	type testCase struct {
		description string
		maxNodes    uint32
		stacktraces []uint32
		expected    []StacktracesRange
	}

	testCases := []testCase{
		{
			description: "no limit",
			stacktraces: []uint32{234, 1234, 2345},
			expected: []StacktracesRange{
				{ids: []uint32{234, 1234, 2345}},
			},
		},
		{
			description: "one chunk",
			maxNodes:    4,
			stacktraces: []uint32{1, 2, 3},
			expected: []StacktracesRange{
				{m: 4, chunk: 0, ids: []uint32{1, 2, 3}},
			},
		},
		{
			description: "one chunk shifted",
			maxNodes:    4,
			stacktraces: []uint32{401, 402},
			expected: []StacktracesRange{
				{m: 4, chunk: 100, ids: []uint32{1, 2}},
			},
		},
		{
			description: "multiple shards",
			maxNodes:    4,
			stacktraces: []uint32{1, 2, 5, 7, 11, 13, 14, 15, 17, 41, 42, 43, 83, 85, 86},
			//         : []uint32{1, 2, 1, 3,  3,  1,  2,  3,  1,  1,  2,  3,  3,  1,  2},
			//         : []uint32{0, 0, 1, 1,  2,  3,  3,  3,  4, 10, 10, 10, 20, 21, 21},
			expected: []StacktracesRange{
				{m: 4, chunk: 0, ids: []uint32{1, 2}},
				{m: 4, chunk: 1, ids: []uint32{1, 3}},
				{m: 4, chunk: 2, ids: []uint32{3}},
				{m: 4, chunk: 3, ids: []uint32{1, 2, 3}},
				{m: 4, chunk: 4, ids: []uint32{1}},
				{m: 4, chunk: 10, ids: []uint32{1, 2, 3}},
				{m: 4, chunk: 20, ids: []uint32{3}},
				{m: 4, chunk: 21, ids: []uint32{1, 2}},
			},
		},
		{
			description: "multiple shards exact",
			maxNodes:    4,
			stacktraces: []uint32{1, 2, 5, 7, 11, 13, 14, 15, 17, 41, 42, 43, 83, 85, 86, 87},
			expected: []StacktracesRange{
				{m: 4, chunk: 0, ids: []uint32{1, 2}},
				{m: 4, chunk: 1, ids: []uint32{1, 3}},
				{m: 4, chunk: 2, ids: []uint32{3}},
				{m: 4, chunk: 3, ids: []uint32{1, 2, 3}},
				{m: 4, chunk: 4, ids: []uint32{1}},
				{m: 4, chunk: 10, ids: []uint32{1, 2, 3}},
				{m: 4, chunk: 20, ids: []uint32{3}},
				{m: 4, chunk: 21, ids: []uint32{1, 2, 3}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, SplitStacktraces(tc.stacktraces, tc.maxNodes))
		})
	}
}

func Test_Stacktraces_append_resolve(t *testing.T) {
	t.Run("single chunk", func(t *testing.T) {
		db := NewSymDB(new(Config))
		w := db.MappingWriter(0)
		a := w.StacktraceAppender()
		defer a.Release()

		sids := make([]uint32, 5)
		a.AppendStacktrace(sids, []*schemav1.Stacktrace{
			{LocationIDs: []uint64{3, 2, 1}},
			{LocationIDs: []uint64{2, 1}},
			{LocationIDs: []uint64{4, 3, 2, 1}},
			{LocationIDs: []uint64{3, 1}},
			{LocationIDs: []uint64{5, 2, 1}},
		})

		r := db.MappingReader(0).StacktraceResolver()
		dst := new(mockStacktraceInserter)
		dst.On("InsertStacktrace", uint32(2), []int32{2, 1})
		dst.On("InsertStacktrace", uint32(3), []int32{3, 2, 1})
		dst.On("InsertStacktrace", uint32(4), []int32{4, 3, 2, 1})
		dst.On("InsertStacktrace", uint32(5), []int32{3, 1})
		dst.On("InsertStacktrace", uint32(6), []int32{5, 2, 1})
		r.ResolveStacktraces(dst, []uint32{2, 3, 4, 5, 6})
	})

	t.Run("multiple chunks", func(t *testing.T) {
		db := NewSymDB(&Config{
			Stacktraces: StacktracesConfig{
				MaxNodesPerChunk: 7,
			},
		})

		w := db.MappingWriter(0)
		a := w.StacktraceAppender()
		defer a.Release()

		stacktraces := []*schemav1.Stacktrace{ // ID, Chunk ID:
			{LocationIDs: []uint64{3, 2, 1}},        // 3  0
			{LocationIDs: []uint64{2, 1}},           // 2  0
			{LocationIDs: []uint64{4, 3, 2, 1}},     // 11 1
			{LocationIDs: []uint64{3, 1}},           // 16 2
			{LocationIDs: []uint64{5, 2, 1}},        // 18 2
			{LocationIDs: []uint64{13, 12, 11}},     // 24 3
			{LocationIDs: []uint64{12, 11}},         // 23 3
			{LocationIDs: []uint64{14, 13, 12, 11}}, // 32 4
			{LocationIDs: []uint64{13, 11}},         // 37 5
			{LocationIDs: []uint64{15, 12, 11}},     // 39 5
		}
		/*
			// TODO(kolesnikovae): Add test cases:
			// Invariants:
			//        0
			//      1
			//      1 0
			//    2
			//    2   0
			//    2 1
			//    2 1 0
			//  3
			//  3     0
			//  3   1
			//  3   1 0
			//  3 2
			//  3 2   0
			//  3 2 1
			//  3 2 1 0
		*/
		sids := make([]uint32, len(stacktraces))
		a.AppendStacktrace(sids, stacktraces)
		require.Len(t, db.mappings[0].stacktraceChunks, 6)

		t.Run("adjacent shards at beginning", func(t *testing.T) {
			r := db.MappingReader(0).StacktraceResolver()
			dst := new(mockStacktraceInserter)
			dst.On("InsertStacktrace", uint32(2), []int32{2, 1})
			dst.On("InsertStacktrace", uint32(3), []int32{3, 2, 1})
			dst.On("InsertStacktrace", uint32(11), []int32{4, 3, 2, 1})
			dst.On("InsertStacktrace", uint32(16), []int32{3, 1})
			dst.On("InsertStacktrace", uint32(18), []int32{5, 2, 1})
			r.ResolveStacktraces(dst, []uint32{2, 3, 11, 16, 18})
		})

		t.Run("adjacent shards at end", func(t *testing.T) {
			r := db.MappingReader(0).StacktraceResolver()
			dst := new(mockStacktraceInserter)
			dst.On("InsertStacktrace", uint32(23), []int32{12, 11})
			dst.On("InsertStacktrace", uint32(24), []int32{13, 12, 11})
			dst.On("InsertStacktrace", uint32(32), []int32{14, 13, 12, 11})
			dst.On("InsertStacktrace", uint32(37), []int32{13, 11})
			dst.On("InsertStacktrace", uint32(39), []int32{15, 12, 11})
			r.ResolveStacktraces(dst, []uint32{23, 24, 32, 37, 39})
		})

		t.Run("non-adjacent shards", func(t *testing.T) {
			r := db.MappingReader(0).StacktraceResolver()
			dst := new(mockStacktraceInserter)
			dst.On("InsertStacktrace", uint32(11), []int32{4, 3, 2, 1})
			dst.On("InsertStacktrace", uint32(32), []int32{14, 13, 12, 11})
			r.ResolveStacktraces(dst, []uint32{11, 32})
		})
	})
}

type mockStacktraceInserter struct{ mock.Mock }

func (m *mockStacktraceInserter) InsertStacktrace(stacktraceID uint32, locations []int32) {
	m.Called(stacktraceID, locations)
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
