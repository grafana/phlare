package phlaredb

import (
	"context"
	"encoding/binary"

	"github.com/cespare/xxhash/v2"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func newStacktracesStore(phlarectx context.Context, cfg *ParquetConfig) *deduplicatingStore[schemav1.Stacktrace, stacktracesKey, *schemav1.StacktracePersister] {
	return newDeduplicatingStore[schemav1.Stacktrace, stacktracesKey, *schemav1.StacktracePersister](phlarectx, cfg, &stacktracesHelper{})
}

type stacktracesKey uint64

type stacktracesHelper struct{}

func (*stacktracesHelper) key(s *schemav1.Stacktrace) stacktracesKey {
	var (
		h = xxhash.New()
		b = make([]byte, 8)
	)

	for pos := range s.LocationIDs {
		binary.LittleEndian.PutUint64(b, s.LocationIDs[pos])
		if _, err := h.Write(b); err != nil {
			panic("unable to write hash")
		}
	}

	// TODO: Those hashes might as well collide, let's defer handling collisions till we need to do it
	return stacktracesKey(h.Sum64())
}

func (*stacktracesHelper) addToRewriter(r *rewriter, m idConversionTable) {
	r.stacktraces = m
}

func (*stacktracesHelper) rewrite(r *rewriter, s *schemav1.Stacktrace) error {
	for pos := range s.LocationIDs {
		r.locations.rewriteUint64(&s.LocationIDs[pos])
	}
	return nil
}

func (*stacktracesHelper) setID(oldID, newID uint64, s *schemav1.Stacktrace) uint64 {
	return oldID
}

func (*stacktracesHelper) clone(s *schemav1.Stacktrace) *schemav1.Stacktrace {
	s.LocationIDs = copySlice(s.LocationIDs)
	return s
}
