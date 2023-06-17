package symdb

import (
	"encoding/binary"
	"io"

	"github.com/dennwc/varint"
)

const defaultStacktraceTreeSize = 1 << 10

type stacktraceTree struct {
	nodes []node
}

type node struct {
	// TODO: Get rid of the self-reference.
	i int // Index of the node in the stacktraces.

	// Auxiliary members only needed for insertion.
	fc int32 // First child index.
	ns int32 // Next sibling index.

	p   int32 // Parent index.
	ref int32 // Reference the to stack frame data.
}

func newStacktraceTree(size int) *stacktraceTree {
	t := stacktraceTree{nodes: make([]node, 0, size)}
	t.newNode(sentinel, 0)
	return &t
}

const sentinel = -1

func (t *stacktraceTree) len() uint32 { return uint32(len(t.nodes)) }

func (t *stacktraceTree) newNode(parent int32, ref int32) node {
	n := node{
		ref: ref,
		i:   len(t.nodes),
		p:   parent,
		fc:  sentinel,
		ns:  sentinel,
	}
	t.nodes = append(t.nodes, n)
	return n
}

func (t *stacktraceTree) insert(refs []uint64) (id uint32) {
	var (
		i int32
		n node
	)

	// TODO(kolesnikovae): Optimize:
	//   Avoid copying of nodes.
	//   Avoid type conversions.
	//   Avoid redundant accesses to the node slice.
	for j := len(refs) - 1; j >= 0; {
		r := int32(refs[j])
		if i == sentinel {
			x := t.newNode(int32(n.i), r)
			n.fc = int32(x.i)
			t.nodes[n.i] = n
			n = x
		} else {
			n = t.nodes[i]
		}

		switch {
		case n.ref == r:
			t.nodes[n.i] = n
			i = n.fc
			j--
			continue
		case n.p == sentinel: // case n.i == 0:
			i = n.fc
			continue
		case n.ns == sentinel:
			x := t.newNode(n.p, r)
			n.ns = int32(x.i)
			t.nodes[n.i] = n
		}

		i = n.ns
	}

	return uint32(n.i)
}

func (t *stacktraceTree) resolve(dst []int32, id uint32) []int32 {
	dst = dst[:0]
	if id >= uint32(len(t.nodes)) {
		return dst
	}
	n := t.nodes[id]
	for n.p >= 0 {
		dst = append(dst, n.ref)
		n = t.nodes[n.p]
	}
	return dst
}

func (t *stacktraceTree) merge(*stacktraceTree) {
	// TODO(kolesnikovae): implement.
}

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

type parentPointerTree struct {
	nodes []pptNode
}

type pptNode struct {
	p   int32 // Parent index.
	ref int32 // Reference the to stack frame data.
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

func (t *parentPointerTree) resolve(dst []int32, id int32) []int32 {
	if id >= int32(len(t.nodes)) {
		return dst
	}
	dst = dst[:0]
	n := t.nodes[id]
	for n.p >= 0 {
		dst = append(dst, n.ref)
		n = t.nodes[n.p]
	}
	return dst
}
