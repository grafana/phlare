package phlaredb

import (
	"context"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

type stringConversionTable []int64

func (t stringConversionTable) rewrite(idx *int64) {
	originalValue := int(*idx)
	newValue := t[originalValue]
	*idx = newValue
}

func newStringsStore(phlarectx context.Context, cfg *ParquetConfig) *deduplicatingStore[schemav1.String, string, *schemav1.StringPersister] {
	return newDeduplicatingStore[schemav1.String, string, *schemav1.StringPersister](phlarectx, cfg, &stringsHelper{})
}

type stringsHelper struct{}

func (*stringsHelper) key(s *schemav1.String) string {
	return s.String
}

func (*stringsHelper) addToRewriter(r *rewriter, m idConversionTable) {
	var maxID int64
	for id := range m {
		if id > maxID {
			maxID = id
		}
	}
	r.strings = make(stringConversionTable, maxID+1)

	for x, y := range m {
		r.strings[x] = y
	}
}

func (*stringsHelper) rewrite(*rewriter, *schemav1.String) error {
	return nil
}

func (*stringsHelper) size(s *schemav1.String) uint64 {
	return uint64(len(s.String)) + 8
}

func (*stringsHelper) setID(oldID, newID uint64, s *schemav1.String) uint64 {
	return oldID
}

func (*stringsHelper) clone(s *schemav1.String) *schemav1.String {
	return &schemav1.String{
		ID:     s.ID,
		String: s.String,
	}
}
