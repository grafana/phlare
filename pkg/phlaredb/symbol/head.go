package symbol

import (
	"fmt"

	profilev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
)

type idConversionTable map[int64]int64

// nolint unused
func (t idConversionTable) rewrite(idx *int64) {
	pos := *idx
	var ok bool
	*idx, ok = t[pos]
	if !ok {
		panic(fmt.Sprintf("unable to rewrite index %d", pos))
	}
}

// nolint unused
func (t idConversionTable) rewriteUint64(idx *uint64) {
	pos := *idx
	v, ok := t[int64(pos)]
	if !ok {
		panic(fmt.Sprintf("unable to rewrite index %d", pos))
	}
	*idx = uint64(v)
}

type Models interface {
	*schemav1.Profile | *schemav1.Stacktrace | *profilev1.Location | *profilev1.Mapping | *profilev1.Function | string | *schemav1.StoredString
}

func emptyRewriter() *rewriter {
	return &rewriter{
		strings: []int64{0},
	}
}

// rewriter contains slices to rewrite the per profile reference into per head references.
type rewriter struct {
	strings stringConversionTable
	// nolint unused
	functions idConversionTable
	// nolint unused
	mappings idConversionTable
	// nolint unused
	locations   idConversionTable
	stacktraces idConversionTable
}

type storeHelper[M Models] interface {
	// some Models contain their own IDs within the struct, this allows to set them and keep track of the preexisting ID. It should return the oldID that is supposed to be rewritten.
	setID(existingSliceID uint64, newID uint64, element M) uint64

	// size returns a (rough estimation) of the size of a single element M
	size(M) uint64

	// clone copies parts that are not optimally sized from protobuf parsing
	clone(M) M

	rewrite(*rewriter, M) error
}

type Helper[M Models, K comparable] interface {
	storeHelper[M]
	key(M) K
	addToRewriter(*rewriter, idConversionTable)
}
