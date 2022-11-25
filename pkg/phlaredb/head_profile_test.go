package phlaredb

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/stretchr/testify/require"
)

func newProfile(idx uint64) *schemav1.Profile {
	return &schemav1.Profile{
		ID:          uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", idx)),
		SeriesIndex: uint32(idx),
		TimeNanos:   int64(idx),
	}
}

func newProfiles(count uint64) []*schemav1.Profile {
	profiles := make([]*schemav1.Profile, 0, count)
	for i := uint64(0); i < count; i++ {
		profiles = append(profiles, newProfile(i))
	}
	return profiles
}

func TestProfileHead(t *testing.T) {
	rewriter := &rewriter{}
	rewriter.strings = []int64{0}

	h := newProfileHead()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, h.run(ctx))

	for _, profile := range newProfiles(1000) {
		require.NoError(t, h.ingest(context.TODO(), []*schemav1.Profile{profile}, rewriter))
	}

	// TODO: Check completeness
	// TODO: Check data
}

func BenchmarkProfileHead(b *testing.B) {
	h := newProfileHead()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(b, h.run(ctx))

	rewriter := &rewriter{}
	rewriter.strings = []int64{0}

	profiles := newProfiles(1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, h.ingest(context.TODO(), profiles, rewriter))
		b.Logf("buffer length %d", h.buffer.Len())
	}
}
