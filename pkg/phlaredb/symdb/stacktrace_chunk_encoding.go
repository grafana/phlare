package symdb

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/dennwc/varint"
)

func (t *stacktraceTree) WriteTo(dst io.Writer) (int64, error) {
	var m int64
	var prev node
	// TODO: Should we use optimized group varint encoding to increase decoding speed?
	b := make([]byte, 2*binary.MaxVarintLen64)
	for _, n := range t.nodes {
		v := n.p - prev.p // Delta ZigZag
		x := binary.PutUvarint(b, uint64((v<<1)^(v>>31)))
		x += binary.PutUvarint(b[x:], uint64(n.ref))
		a, err := dst.Write(b[:x])
		if err != nil {
			return m, err
		}
		m += int64(a)
		prev = n
	}
	return m, nil
}

func (t *parentPointerTree) unmarshal(b []byte) {
	var prev, cur pptNode
	for n := 0; n < len(b); {
		v, m := varint.Uvarint(b[n:])
		x := int32(v)
		cur.p = (x>>1 ^ ((x << 31) >> 31)) + prev.p
		n += m
		v, m = varint.Uvarint(b[n:])
		cur.ref = int32(v) // TODO(kolesnikovae): Experiment with encoding.
		n += m
		prev = cur
		t.nodes = append(t.nodes, cur)
	}
}

func (t *parentPointerTree) unmarshalBuffered(b *bufio.Reader) error {
	panic("implement me")
}
