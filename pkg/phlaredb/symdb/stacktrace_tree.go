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
	// Auxiliary members only needed for insertion.
	i  int32 // Index of the node in the stacktraces.	// TODO: Get rid of the self-reference.
	fc int32 // First child index.
	ns int32 // Next sibling index.

	p   int32 // Parent index.
	ref int32 // Reference the to stack frame data.
}

func newStacktraceTree(size int) *stacktraceTree {
	t := stacktraceTree{nodes: make([]node, 0, size)}
	t.newNode(-1, 0)
	return &t
}

func (t *stacktraceTree) newNode(parent int32, ref int32) node {
	n := node{
		ref: ref,
		i:   int32(len(t.nodes)),
		p:   parent,
		fc:  -1,
		ns:  -1,
	}
	t.nodes = append(t.nodes, n)
	return n
}

func (t *stacktraceTree) insert(refs []int32) (id int32, ok bool) {
	var (
		i int32
		j int
		n node
	)

	// TODO(kolesnikovae): Optimize – avoid copying of nodes.
	for j < len(refs) {
		r := refs[j]
		if i < 0 {
			x := t.newNode(n.i, r)
			n.fc = x.i
			t.nodes[n.i] = n
			n = x
		} else {
			n = t.nodes[i]
		}

		switch {
		case n.ref == r:
			t.nodes[n.i] = n
			i = n.fc
			j++
			continue
		case n.i == 0:
			i = n.fc
			continue
		case n.ns < 0:
			x := t.newNode(n.p, r)
			n.ns = x.i
			t.nodes[n.i] = n
		}

		i = n.ns
	}

	return n.i, ok
}

func (t *stacktraceTree) resolve(dst []int32, id int32) []int32 {
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
