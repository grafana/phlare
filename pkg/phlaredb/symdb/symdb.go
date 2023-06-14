package symdb

import "sync"

type SymDB struct {
	config Config

	m        sync.RWMutex
	mappings map[uint64]*inMemoryMapping
}

type Config struct {
	MaxStacksPerChunk int32
}

type Stats struct {
	MemorySize uint64
	Mappings   uint32
}

func NewSymDB() *SymDB {
	return &SymDB{mappings: make(map[uint64]*inMemoryMapping)}
}

func (s *SymDB) Stats() Stats {
	return Stats{}
}

func (s *SymDB) MappingWriter(mappingName uint64) MappingWriter {
	return s.mapping(mappingName)
}

func (s *SymDB) MappingReader(mappingName uint64) MappingReader {
	return s.mapping(mappingName)
}

func (s *SymDB) lookupMapping(mappingName uint64) (*inMemoryMapping, bool) {
	s.m.RLock()
	p, ok := s.mappings[mappingName]
	if ok {
		s.m.RUnlock()
		return p, true
	}
	s.m.RUnlock()
	return nil, false
}

func (s *SymDB) mapping(mappingName uint64) *inMemoryMapping {
	p, ok := s.lookupMapping(mappingName)
	if ok {
		return p
	}
	s.m.Lock()
	if p, ok = s.mappings[mappingName]; ok {
		s.m.Unlock()
		return p
	}
	p = &inMemoryMapping{
		maxStacksPerChunk:  s.config.MaxStacksPerChunk,
		stacktraceHashToID: make(map[uint64]int32, defaultStacktraceTreeSize/2),
		stacktraceChunks: []*stacktraceChunk{{
			tree: newStacktraceTree(defaultStacktraceTreeSize),
		}},
	}
	s.mappings[mappingName] = p
	s.m.Unlock()
	return p
}
