package phlaredb

import schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"

type stringConversionTable []int64

func (t stringConversionTable) rewrite(idx *int64) {
	originalValue := int(*idx)
	newValue := t[originalValue]
	*idx = newValue
}

type stringsHelper struct{}

func (*stringsHelper) key(s *schemav1.StoredString) string {
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

func (*stringsHelper) rewrite(*rewriter, *schemav1.StoredString) error {
	return nil
}

func (*stringsHelper) size(s *schemav1.StoredString) uint64 {
	return uint64(len(s.String)) + 8
}

func (*stringsHelper) setID(oldID, newID uint64, s *schemav1.StoredString) uint64 {
	return oldID
}

func (*stringsHelper) clone(s *schemav1.StoredString) *schemav1.StoredString {
	return &schemav1.StoredString{
		ID:     s.ID,
		String: s.String,
	}
}
