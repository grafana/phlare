package symdb

import "sync"

type SymDB struct {
	m sync.RWMutex

	partitions map[uint64]*inMemoryIndex
}

type Stats struct {
	MemorySize uint64
	Partitions uint32
}

func NewSymDB() *SymDB {
	return &SymDB{partitions: make(map[uint64]*inMemoryIndex)}
}

func (s *SymDB) Stats() Stats {
	return Stats{}
}

func (s *SymDB) IndexWriter(partitionID uint64) IndexWriter {
	return s.partition(partitionID)
}

func (s *SymDB) IndexReader(partitionID uint64) IndexReader {
	return s.partition(partitionID)
}

func (s *SymDB) lookupPartition(partitionID uint64) (*inMemoryIndex, bool) {
	s.m.RLock()
	p, ok := s.partitions[partitionID]
	if ok {
		s.m.RUnlock()
		return p, true
	}
	s.m.RUnlock()
	return nil, false
}

func (s *SymDB) partition(partitionID uint64) *inMemoryIndex {
	p, ok := s.lookupPartition(partitionID)
	if ok {
		return p
	}
	s.m.Lock()
	if p, ok = s.partitions[partitionID]; ok {
		s.m.Unlock()
		return p
	}
	p = &inMemoryIndex{
		stacktraceChunks: []*stacktraceChunk{{
			tree: newStacktraceTree(defaultStacktraceTreeSize),
		}},
	}
	s.partitions[partitionID] = p
	s.m.Unlock()
	return p
}
