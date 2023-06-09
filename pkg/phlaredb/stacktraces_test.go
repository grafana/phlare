package phlaredb

func (s *stacktraceStore) sliceFromMap() []*Stacktrace {
	var loc []uint64

	s.mLock.RLock()
	defer s.mLock.RUnlock()
	r := make([]*Stacktrace, 0, len(s.m))
	for key, id := range s.m {
		r = append(r, &Stacktrace{
			ID:          id,
			LocationIDs: stacktraceKeyToLocationSlice(key, loc),
		})
	}
	return r
}

func (s *stacktraceStore) slice() []*Stacktrace {
	var loc []uint64
	length := 0
	for _, ks := range s.idKeyspaces {
		ks.lock.RLock()
		length += len(ks.nodes)
		ks.lock.RUnlock()
	}

	r := make([]*Stacktrace, 0, length)
	for ksIdx, ks := range s.idKeyspaces {
		ks.lock.RLock()
		for idx, node := range ks.nodes {
			r = append(r, &Stacktrace{
				ID:          uint64(ksIdx)*s.idKeyspaceShardSize() + uint64(idx),
				LocationIDs: stacktraceKeyToLocationSlice(node, loc),
			})
		}
		ks.lock.RUnlock()
	}
	return r
}
