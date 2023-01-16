package phlaredb

import (
	"context"

	profilev1 "github.com/grafana/phlare/pkg/gen/google/v1"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func newFunctionsStore(phlarectx context.Context, cfg *ParquetConfig) *deduplicatingStore[profilev1.Function, functionsKey, *schemav1.FunctionPersister] {
	return newDeduplicatingStore[profilev1.Function, functionsKey, *schemav1.FunctionPersister](phlarectx, cfg, &functionsHelper{})
}

type functionsKey struct {
	Name       int64
	SystemName int64
	Filename   int64
	StartLine  int64
}

type functionsHelper struct{}

func (*functionsHelper) key(f *profilev1.Function) functionsKey {
	return functionsKey{
		Name:       f.Name,
		SystemName: f.SystemName,
		Filename:   f.Filename,
		StartLine:  f.StartLine,
	}
}

func (*functionsHelper) addToRewriter(r *rewriter, elemRewriter idConversionTable) {
	r.functions = elemRewriter
}

func (*functionsHelper) rewrite(r *rewriter, f *profilev1.Function) error {
	r.strings.rewrite(&f.Filename)
	r.strings.rewrite(&f.Name)
	r.strings.rewrite(&f.SystemName)
	return nil
}

func (*functionsHelper) setID(_, newID uint64, f *profilev1.Function) uint64 {
	var oldID = f.Id
	f.Id = newID
	return oldID
}

func (*functionsHelper) clone(f *profilev1.Function) *profilev1.Function {
	return f
}
