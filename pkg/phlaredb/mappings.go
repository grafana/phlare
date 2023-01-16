package phlaredb

import (
	"context"

	profilev1 "github.com/grafana/phlare/pkg/gen/google/v1"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

type mappingsHelper struct{}

type mappingsKey struct {
	MemoryStart     uint64
	MemoryLimit     uint64
	FileOffset      uint64
	Filename        int64 // Index into string table
	BuildId         int64 //nolint // Index into string table
	HasFunctions    bool
	HasFilenames    bool
	HasLineNumbers  bool
	HasInlineFrames bool
}

func newMappingsStore(phlarectx context.Context, cfg *ParquetConfig) *deduplicatingStore[profilev1.Mapping, mappingsKey, *schemav1.MappingPersister] {
	return newDeduplicatingStore[profilev1.Mapping, mappingsKey, *schemav1.MappingPersister](phlarectx, cfg, &mappingsHelper{})
}

func (*mappingsHelper) key(m *profilev1.Mapping) mappingsKey {
	return mappingsKey{
		MemoryStart:     m.MemoryStart,
		MemoryLimit:     m.MemoryLimit,
		FileOffset:      m.FileOffset,
		Filename:        m.Filename,
		BuildId:         m.BuildId,
		HasFunctions:    m.HasFunctions,
		HasFilenames:    m.HasFilenames,
		HasLineNumbers:  m.HasFunctions,
		HasInlineFrames: m.HasInlineFrames,
	}
}

func (*mappingsHelper) addToRewriter(r *rewriter, elemRewriter idConversionTable) {
	r.mappings = elemRewriter
}

func (*mappingsHelper) rewrite(r *rewriter, m *profilev1.Mapping) error {
	r.strings.rewrite(&m.Filename)
	r.strings.rewrite(&m.BuildId)
	return nil
}

func (*mappingsHelper) setID(_, newID uint64, m *profilev1.Mapping) uint64 {
	var oldID = m.Id
	m.Id = newID
	return oldID
}

func (*mappingsHelper) clone(m *profilev1.Mapping) *profilev1.Mapping {
	return m
}
