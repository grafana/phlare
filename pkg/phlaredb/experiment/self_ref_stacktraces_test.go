package experiment

import (
	"encoding/binary"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	art "github.com/plar/go-adaptive-radix-tree"
	"github.com/segmentio/parquet-go"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

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
	tree  art.Tree
	count []*atomic.Uint64
}

func newStacktraces(keyspaceShards int) SelfRefStacktraces {
	s := SelfRefStacktraces{
		tree:  art.New(),
		count: atomic.NewUint64(0),
	}
}

type stacktraceID uint64

func (id *stacktraceID) EqualTo(v interface{}) bool {
	return *id == *(v.(*stacktraceID))
}

// this takes the location ids and creates a byte slice in reverse order (the last location ID is the root of the stacktrace)
func uint64SliceToByteSlice(v []uint64) []byte {
	r := make([]byte, 8*len(v))
	for pos := range v {
		binary.LittleEndian.PutUint64(r[8*pos:], v[len(v)-1-pos])
	}
	return r
}

// add a stacktrace
func (s *SelfRefStacktraces) add(v []uint64) uint64 {
	key := art.Key(uint64SliceToByteSlice(v))
	if id, found := s.tree.Search(key); found {
		return uint64(*(id.(*stacktraceID)))
	}

	id := stacktraceID(s.count.Inc())
	s.tree.Insert(art.Key(uint64SliceToByteSlice(v)), art.Value(&id))
	return uint64(id)
}

func TestSelfRefStacktraces(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	beforePath := "/home/christian/parquet-notebook/stacktraces.parquet"
	//afterPath := "/home/christian/parquet-notebook/stacktraces_self_ref.parquet"
	f, err := os.Open(beforePath)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	p

	reader := parquet.NewGenericReader[schemav1.Stacktrace](f)

	t.Logf("original has %d rows", reader.NumRows())

	s := newStacktraces(16)

	var slice = make([]schemav1.Stacktrace, 100_000)
	for {
		n, err := reader.Read(slice)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		t.Logf("read %d rows", n)

		for _, c := range slice[:n] {
			_ = s.add(c.LocationIDs)
		}
	}

	s.tree.ForEach

	t.Logf("read all into radix tree size=%d", s.tree.Size())

	/*	s.tree.ForEach(func(n art.Node) (cont bool) {
			t.Logf("key=%s value=%v", n.Key(), *(n.Value().(*stacktraceID)))

			return true
		})
	*/

	time.Sleep(time.Hour)

	/*
		// read all rows and insert into new structure

		require.NoError(t, parquet.WriteFile[SelfRefStacktrace](afterPath, s.stacktraces))
	*/

}
