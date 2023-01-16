package phlaredb

import (
	"context"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

func newDeduplicatingStore[M Models, K comparable, P schemav1.Persister[*M]](phlarectx context.Context, cfg *ParquetConfig, helper deduplicatingStoreHelper[M, K]) *deduplicatingStore[M, K, P] {
	baseStore := newStore[M, P](phlarectx, cfg, helper)

	store := &deduplicatingStore[M, K, P]{
		store:  baseStore,
		helper: helper,
		lookup: make(map[K]int64),
	}

	// set hooks to the baseStore to do the actual deduplication
	baseStore.filter = store.filterAlreadyExistingElems
	baseStore.updateIndex = store.setIndexByElem

	return store
}

type deduplicatingStore[M Models, K comparable, P schemav1.Persister[*M]] struct {
	*store[M, P]

	lookup map[K]int64
	helper deduplicatingStoreHelper[M, K]
}

func (s *deduplicatingStore[M, K, P]) filterAlreadyExistingElems(elems *appendElems[M]) {
	for pos := range elems.elems {
		k := s.helper.key(elems.elems[pos])
		if posSlice, exists := s.getIndex(k); exists {
			elems.rewritingMap[int64(s.helper.setID(uint64(pos), uint64(posSlice), elems.elems[pos]))] = posSlice
		} else {
			elems.elems[len(elems.originalPos)] = elems.elems[pos]
			elems.originalPos = append(elems.originalPos, int64(pos))
		}
	}

	// reset slice to only contain missing elements
	elems.elems = elems.elems[:len(elems.originalPos)]
}

func (s *deduplicatingStore[M, K, P]) Reset(path string) error {
	s.lock.Lock()
	s.lookup = make(map[K]int64)
	s.lock.Unlock()

	return s.store.Reset(path)

}

func (s *deduplicatingStore[M, K, P]) setIndex(key K, pos int64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.lookup[key] = pos
}

func (s *deduplicatingStore[M, K, P]) setIndexByElem(elem *M, pos uint64) {
	s.setIndex(s.helper.key(elem), int64(pos))
}

func (s *deduplicatingStore[M, K, P]) getIndex(key K) (int64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.lookup[key]
	return v, ok
}
