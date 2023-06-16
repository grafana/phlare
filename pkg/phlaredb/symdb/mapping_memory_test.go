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
		expected    []stacktraceIDRange
	}

	testCases := []testCase{
		{
			description: "no limit",
			stacktraces: []uint32{234, 1234, 2345},
			expected: []stacktraceIDRange{
				{chunk: 0, ids: []uint32{234, 1234, 2345}},
			},
		},
		{
			description: "one chunk",
			maxNodes:    4,
			stacktraces: []uint32{1, 2, 3},
			expected: []stacktraceIDRange{
				{chunk: 0, ids: []uint32{1, 2, 3}},
			},
		},
		{
			description: "one chunk shifted",
			maxNodes:    4,
			stacktraces: []uint32{401, 402},
			expected: []stacktraceIDRange{
				{chunk: 100, ids: []uint32{1, 2}},
			},
		},
		{
			description: "multiple shards",
			maxNodes:    4, // denominator
			stacktraces: []uint32{1, 2, 5, 7, 11, 13, 14, 15, 17, 41, 42, 43, 83, 85, 86},
			// reminder: []uint32{1, 2, 1, 3,  3,  1,  2,  3,  1,  1,  2,  3,  3,  1,  2},
			//         : []uint32{0, 0, 1, 1,  2,  3,  3,  3,  4, 10, 10, 10, 20, 21, 21},
			expected: []stacktraceIDRange{
				{chunk: 0, ids: []uint32{1, 2}},
				{chunk: 1, ids: []uint32{1, 3}},
				{chunk: 2, ids: []uint32{3}},
				{chunk: 3, ids: []uint32{1, 2, 3}},
				{chunk: 4, ids: []uint32{1}},
				{chunk: 10, ids: []uint32{1, 2, 3}},
				{chunk: 20, ids: []uint32{3}},
				{chunk: 21, ids: []uint32{1, 2}},
			},
		},
		{
			description: "multiple shards exact",
			maxNodes:    4,
			stacktraces: []uint32{1, 2, 5, 7, 11, 13, 14, 15, 17, 41, 42, 43, 83, 85, 86, 87},
			expected: []stacktraceIDRange{
				{chunk: 0, ids: []uint32{1, 2}},
				{chunk: 1, ids: []uint32{1, 3}},
				{chunk: 2, ids: []uint32{3}},
				{chunk: 3, ids: []uint32{1, 2, 3}},
				{chunk: 4, ids: []uint32{1}},
				{chunk: 10, ids: []uint32{1, 2, 3}},
				{chunk: 20, ids: []uint32{3}},
				{chunk: 21, ids: []uint32{1, 2, 3}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, splitStacktraces(tc.stacktraces, tc.maxNodes))
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
	})

	t.Run("multiple chunks mixed", func(t *testing.T) {
		db := NewSymDB(&Config{
			Stacktraces: StacktracesConfig{
				MaxNodesPerChunk: 5,
			},
		})

		w := db.MappingWriter(0)
		a := w.StacktraceAppender()
		defer a.Release()

		stacktraces := []*schemav1.Stacktrace{ // ID, Chunk ID:
			{LocationIDs: []uint64{3, 2, 1}},        // 3  0
			{LocationIDs: []uint64{2, 1}},           // 2  0
			{LocationIDs: []uint64{4, 3, 2, 1}},     // 4  0
			{LocationIDs: []uint64{3, 1}},           // 7  1
			{LocationIDs: []uint64{5, 2, 1}},        // 9  1
			{LocationIDs: []uint64{13, 12, 11}},     // 13 2
			{LocationIDs: []uint64{12, 11}},         // 12 2
			{LocationIDs: []uint64{14, 13, 12, 11}}, // 14 2
			{LocationIDs: []uint64{13, 11}},         // 17 3
			{LocationIDs: []uint64{15, 12, 11}},     // 19 4
		}

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

		sids := make([]uint32, len(stacktraces))
		a.AppendStacktrace(sids, stacktraces)
		require.Len(t, db.mappings[0].stacktraceChunks, 4)

		t.Run("adjacent shards at beginning", func(t *testing.T) {
			r := db.MappingReader(0).StacktraceResolver()
			dst := new(mockStacktraceInserter)
			dst.On("InsertStacktrace", uint32(2), []int32{2, 1})
			dst.On("InsertStacktrace", uint32(3), []int32{3, 2, 1})
			dst.On("InsertStacktrace", uint32(4), []int32{4, 3, 2, 1})
			dst.On("InsertStacktrace", uint32(7), []int32{3, 1})
			dst.On("InsertStacktrace", uint32(9), []int32{5, 2, 1})
			r.ResolveStacktraces(dst, []uint32{2, 3, 4, 7, 9})
		})

		t.Run("adjacent shards at end", func(t *testing.T) {
			r := db.MappingReader(0).StacktraceResolver()
			dst := new(mockStacktraceInserter)
			dst.On("InsertStacktrace", uint32(12), []int32{12, 11})
			dst.On("InsertStacktrace", uint32(13), []int32{13, 12, 11})
			dst.On("InsertStacktrace", uint32(14), []int32{14, 13, 12, 11})
			dst.On("InsertStacktrace", uint32(17), []int32{13, 11})
			dst.On("InsertStacktrace", uint32(19), []int32{15, 12, 11})
			r.ResolveStacktraces(dst, []uint32{12, 13, 14, 17, 19})
		})

		t.Run("non-adjacent shards", func(t *testing.T) {
			r := db.MappingReader(0).StacktraceResolver()
			dst := new(mockStacktraceInserter)
			dst.On("InsertStacktrace", uint32(3), []int32{3, 2, 1})
			dst.On("InsertStacktrace", uint32(19), []int32{15, 12, 11})
			r.ResolveStacktraces(dst, []uint32{3, 19})
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
