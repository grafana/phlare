package model

import (
	"container/heap"
	"fmt"
	"io"
	"sync"

	"github.com/bcmills/unsafeslice"
	dvarint "github.com/dennwc/varint"
	"github.com/pyroscope-io/pyroscope/pkg/util/varint"
	"github.com/xlab/treeprint"
)

type Tree struct {
	root []*node
}

func emptyTree() *Tree {
	return &Tree{}
}

func newTree(stacks []stacktraces) *Tree {
	t := emptyTree()
	for _, stack := range stacks {
		if stack.value == 0 {
			continue
		}
		if t == nil {
			t = stackToTree(stack)
			continue
		}
		MergeTree(t, stackToTree(stack))
	}
	return t
}

type stacktraces struct {
	locations []string
	value     int64
}

func (t *Tree) add(name string, self, total int64) *node {
	new := &node{
		name:  name,
		self:  self,
		total: total,
	}
	t.root = append(t.root, new)
	return new
}

func stackToTree(stack stacktraces) *Tree {
	t := emptyTree()
	if len(stack.locations) == 0 {
		return t
	}
	current := &node{
		self:  stack.value,
		total: stack.value,
		name:  stack.locations[0],
	}
	if len(stack.locations) == 1 {
		t.root = append(t.root, current)
		return t
	}
	remaining := stack.locations[1:]
	for len(remaining) > 0 {

		location := remaining[0]
		name := location
		remaining = remaining[1:]

		// This pack node with the same name as the next location
		// Disable for now but we might want to introduce it if we find it useful.
		// for len(remaining) != 0 {
		// 	if remaining[0].function == name {
		// 		remaining = remaining[1:]
		// 		continue
		// 	}
		// 	break
		// }

		parent := &node{
			children: []*node{current},
			total:    current.total,
			name:     name,
		}
		current.parent = parent
		current = parent
	}
	t.root = []*node{current}
	return t
}

func (t Tree) String() string {
	type branch struct {
		nodes []*node
		treeprint.Tree
	}
	tree := treeprint.New()
	for _, n := range t.root {
		b := tree.AddBranch(fmt.Sprintf("%s: self %d total %d", n.name, n.self, n.total))
		remaining := append([]*branch{}, &branch{nodes: n.children, Tree: b})
		for len(remaining) > 0 {
			current := remaining[0]
			remaining = remaining[1:]
			for _, n := range current.nodes {
				if len(n.children) > 0 {
					remaining = append(remaining, &branch{nodes: n.children, Tree: current.Tree.AddBranch(fmt.Sprintf("%s: self %d total %d", n.name, n.self, n.total))})
				} else {
					current.Tree.AddNode(fmt.Sprintf("%s: self %d total %d", n.name, n.self, n.total))
				}
			}
		}
	}
	return tree.String()
}

func MergeTree(dst, src *Tree) {
	// walk src and insert src's nodes into dst
	for _, rootNode := range src.root {
		parent, found, toMerge := findNodeOrParent(dst.root, rootNode)
		if found == nil {
			if parent == nil {
				dst.root = append(dst.root, toMerge)
				continue
			}
			toMerge.parent = parent
			parent.children = append(parent.children, toMerge)
			for p := parent; p != nil; p = p.parent {
				p.total = p.total + toMerge.total
			}
			continue
		}
		found.total = found.total + toMerge.self
		found.self = found.self + toMerge.self
		for p := found.parent; p != nil; p = p.parent {
			p.total = p.total + toMerge.total
		}
	}
}

type node struct {
	parent      *node
	children    []*node
	self, total int64
	name        string
}

func (n *node) String() string {
	return fmt.Sprintf("{%s: self %d total %d}", n.name, n.self, n.total)
}

func (n *node) Add(name string, self, total int64) *node {
	new := &node{
		parent: n,
		name:   name,
		self:   self,
		total:  total,
	}
	n.children = append(n.children, new)
	return new
}

// Walks into root nodes to find a node, return the latest common parent visited.
func findNodeOrParent(root []*node, new *node) (parent, found, toMerge *node) {
	current := new
	var lastParent *node
	remaining := root
	for len(remaining) > 0 {
		n := remaining[0]
		remaining = remaining[1:]
		// we found the common parent so we go down
		if n.name == current.name {
			// we reach the end of the new path to find.
			if len(current.children) == 0 {
				return lastParent, n, current
			}
			lastParent = n
			remaining = n.children
			current = current.children[0]
			continue
		}
	}

	return lastParent, nil, current
}

// minValue returns the minimum "total" value a node in a tree has to have to show up in
// the resulting flamegraph
func (t *Tree) minValue(maxNodes int64) int64 {
	if maxNodes == -1 { // -1 means show all nodes
		return 0
	}
	nodes := t.root

	mh := &minHeap{}
	heap.Init(mh)

	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]
		number := node.total

		if mh.Len() < int(maxNodes) {
			heap.Push(mh, number)
		} else {
			if number > (*mh)[0] {
				heap.Pop(mh)
				heap.Push(mh, number)
				nodes = append(node.children, nodes...)
			}
		}
	}

	if mh.Len() < int(maxNodes) {
		return 0
	}

	return (*mh)[0]
}

// minHeap is a custom min-heap data structure that stores integers.
type minHeap []int64

// Len returns the number of elements in the min-heap.
func (h minHeap) Len() int { return len(h) }

// Less returns true if the element at index i is less than the element at index j.
// This method is used by the container/heap package to maintain the min-heap property.
func (h minHeap) Less(i, j int) bool { return h[i] < h[j] }

// Swap exchanges the elements at index i and index j.
// This method is used by the container/heap package to reorganize the min-heap during its operations.
func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push adds an element (x) to the min-heap.
// This method is used by the container/heap package to grow the min-heap.
func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(int64))
}

// Pop removes and returns the smallest element (minimum) from the min-heap.
// This method is used by the container/heap package to shrink the min-heap.
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

const lostDuringSerializationName = "other"

// MarshalTruncate writes tree byte representation to the writer provider,
// the number of nodes is limited to maxNodes. The function modifies
// the tree: truncated nodes are removed from the tree.
func (t *Tree) MarshalTruncate(w io.Writer, maxNodes int64) (err error) {
	if len(t.root) == 0 {
		return nil
	}
	vw := varint.NewWriter()
	minVal := t.minValue(maxNodes)
	nodes := t.root
	// Virtual root node.
	n := &node{children: t.root}
	for len(nodes) > 0 {
		if _, _ = vw.Write(w, uint64(len(n.name))); err != nil {
			return err
		}
		if _, _ = w.Write(unsafeslice.OfString(n.name)); err != nil {
			return err
		}
		if _, err = vw.Write(w, uint64(n.self)); err != nil {
			return err
		}
		children := n.children
		n.children = n.children[:0]

		var other int64
		for _, cn := range children {
			isOtherNode := cn.name == lostDuringSerializationName
			if cn.total >= minVal || isOtherNode {
				n.children = append(n.children, cn)
			} else {
				other += cn.total
			}
		}
		if other > 0 {
			n.children = append(n.children, &node{
				parent: n,
				self:   other,
				total:  other,
				name:   lostDuringSerializationName,
			})
		}

		if len(n.children) > 0 {
			nodes = append(n.children, nodes...)
		} else {
			n.children = nil // Just to make it eligible for GC.
		}
		if _, err = vw.Write(w, uint64(len(n.children))); err != nil {
			return err
		}
		n, nodes = nodes[0], nodes[1:]
	}

	return nil
}

var errMalformedTreeBytes = fmt.Errorf("malformed tree bytes")

const estimateBytesPerNode = 16 // Chosen empirically.

func UnmarshalTree(b []byte) (*Tree, error) {
	t := new(Tree)
	if len(b) < 2 {
		return t, nil
	}

	parents := make([]*node, 1, len(b)/estimateBytesPerNode)
	// Virtual root node.
	root := new(node)
	parents[0] = root
	var parent *node
	var offset int

	for len(parents) > 0 {
		parent, parents = parents[len(parents)-1], parents[:len(parents)-1]
		nameLen, o := dvarint.Uvarint(b[offset:])
		if o < 0 {
			return nil, errMalformedTreeBytes
		}
		offset += o
		name := unsafeslice.AsString(b[offset : offset+int(nameLen)])
		offset += int(nameLen)
		value, o := dvarint.Uvarint(b[offset:])
		if o < 0 {
			return nil, errMalformedTreeBytes
		}
		offset += o
		childrenLen, o := dvarint.Uvarint(b[offset:])
		if o < 0 {
			return nil, errMalformedTreeBytes
		}
		offset += o

		n := &node{
			parent:   parent,
			children: make([]*node, 0, childrenLen),
			self:     int64(value),
			name:     name,
		}

		pn := n
		for pn.parent != nil {
			pn.total += n.self
			pn = pn.parent
		}

		parent.children = append(parent.children, n)
		for i := uint64(0); i < childrenLen; i++ {
			parents = append(parents, n)
		}
	}

	// Virtual root.
	t.root = root.children[0].children

	return t, nil
}

type TreeMerger struct {
	mu sync.Mutex
	t  *Tree
}

func NewTreeMerger() *TreeMerger {
	return new(TreeMerger)
}

func (m *TreeMerger) MergeTreeBytes(b []byte) error {
	m.mu.Lock()
	t, err := UnmarshalTree(b)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	if m.t != nil {
		MergeTree(m.t, t)
	} else {
		m.t = t
	}
	m.mu.Unlock()
	return nil
}

func (m *TreeMerger) Tree() *Tree {
	if m.t == nil {
		return new(Tree)
	}
	return m.t
}
