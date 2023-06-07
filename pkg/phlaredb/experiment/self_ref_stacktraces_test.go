package experiment

import (
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/segmentio/parquet-go"
	"github.com/stretchr/testify/require"

	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
)

var (
	sb strings.Builder
)

func key(l []uint64) string {
	sb.Reset()
	for _, v := range l {
		sb.WriteString(strconv.FormatUint(v, 10))
	}
	return sb.String()
}

type SelfRefStacktrace struct {
	Parent     uint64 `parquet:",delta"`
	LocationID uint64 `parquet:",delta"`
}

type SelfRefStacktraces struct {
	stacktraces []SelfRefStacktrace
	m           map[string]uint64
}

func newStacktraces(l int) SelfRefStacktraces {
	return SelfRefStacktraces{
		stacktraces: make([]SelfRefStacktrace, 1, l), // the first element is the root
		m:           make(map[string]uint64),
	}
}

// add a stacktrace
func (s *SelfRefStacktraces) add(v []uint64) uint64 {
	if len(v) == 0 {
		return 0
	}

	k := key(v)
	pos, ok := s.m[k]
	if ok {
		return pos
	}

	parent := uint64(0)
	if len(v) > 1 {
		parent = s.add(v[1:])
	}

	s.stacktraces = append(s.stacktraces, SelfRefStacktrace{
		Parent:     parent,
		LocationID: v[0],
	})
	pos = uint64(len(s.stacktraces) - 1)
	s.m[k] = pos
	return pos
}

func TestSelfRefStacktraces(t *testing.T) {
	beforePath := "/home/christian/parquet-notebook/stacktraces.parquet"
	afterPath := "/home/christian/parquet-notebook/stacktraces2.parquet"
	f, err := os.Open(beforePath)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	reader := parquet.NewGenericReader[schemav1.Stacktrace](f)

	t.Logf("original has %d rows", reader.NumRows())

	s := newStacktraces(int(reader.NumRows()))

	var slice = make([]schemav1.Stacktrace, 100_000)
	for {
		n, err := reader.Read(slice)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		t.Logf("read %d rows", n)

		for _, c := range slice[:n] {
			s.add(c.LocationIDs)
		}
	}

	// read all rows and insert into new structure

	require.NoError(t, parquet.WriteFile[SelfRefStacktrace](afterPath, s.stacktraces))

}
