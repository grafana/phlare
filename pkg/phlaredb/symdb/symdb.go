package symdb

import (
	"io"
	"os"
	"sync"
)

type SymDB struct {
	config *Config
	writer *Writer

	m        sync.RWMutex
	mappings map[uint64]*inMemoryMapping
}

type Config struct {
	Dir         string
	Stacktraces StacktracesConfig
}

type StacktracesConfig struct {
	MaxNodesPerChunk uint32
}

func DefaultConfig() *Config {
	return &Config{
		Dir: os.TempDir(),
		Stacktraces: StacktracesConfig{
			// A million of nodes ensures predictable
			// memory consumption, although causes a
			// small overhead.
			MaxNodesPerChunk: 1 << 20,
		},
	}
}

func (c *Config) WithDirectory(dir string) Config {
	c.Dir = dir
	return *c
}

type Stats struct {
	MemorySize uint64
	Mappings   uint32
}

func NewSymDB(c *Config) *SymDB {
	if c == nil {
		c = DefaultConfig()
	}
	return &SymDB{
		config:   c,
		writer:   NewWriter(c.Dir),
		mappings: make(map[uint64]*inMemoryMapping),
	}
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
		name:               mappingName,
		maxNodesPerChunk:   s.config.Stacktraces.MaxNodesPerChunk,
		stacktraceHashToID: make(map[uint64]uint32, defaultStacktraceTreeSize/2),
		stacktraceChunks: []*stacktraceChunk{{
			tree: newStacktraceTree(defaultStacktraceTreeSize),
		}},
	}
	s.mappings[mappingName] = p
	s.m.Unlock()
	return p
}

func (s *SymDB) WriteTo(dst io.Writer) (int64, error) {
	return s.writer.WriteTo(dst)
}

func (s *SymDB) Flush() error {
	return nil // TODO: write to the default file
}
