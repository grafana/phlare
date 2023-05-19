package model

import (
	"container/heap"
	"fmt"

	"github.com/xlab/treeprint"
)

type Tree struct {
	root []*node
}

func emptyTree() *Tree {
	return &Tree{}
}

func NewTree(stacks []stacktraces) *Tree {
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

func (t *Tree) Add(name string, self, total int64) *node {
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

func UnmarshalTree([]byte) (*Tree, error) {
	panic("implement me")
}

func (t *Tree) Marshal(maxNodes int64) []byte {
	panic("implement me")
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
