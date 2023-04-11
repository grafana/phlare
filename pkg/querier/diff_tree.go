package querier

import (
	"bytes"

	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	"github.com/pyroscope-io/pyroscope/pkg/structs/cappedarr"
)

// DiffTree generates the FlameGraph struct from 2 trees.
// They must be the response trees from CombineTree (i.e. all children nodes
// must be the same length). The Flamebearer struct returned from this function
// is different to the one returned from Tree.FlamebearerStruct(). It has the
// following structure:
//
//	i+0 = x offset, left  tree
//	i+1 = total   , left  tree
//	i+2 = self    , left  tree
//	i+3 = x offset, right tree
//	i+4 = total   , right tree
//	i+5 = self    , right tree
//	i+6 = index in the names array
func DiffTree(left, right *tree, maxNodes int) *querierv1.FlameGraph {
	leftTree, rightTree := combineTree(left, right)

	totalLeft := addTotalRoot(leftTree)
	totalRight := addTotalRoot(rightTree)

	res := &querierv1.FlameGraph{
		Names:   []string{},
		Levels:  []*querierv1.Level{},
		Total:   totalLeft + totalRight,
		MaxSelf: 0,
	}

	leftNodes, xLeftOffsets := leftTree.root, []int64{0}
	rghtNodes, xRghtOffsets := rightTree.root, []int64{0}
	levels := []int{0}
	// TODO: dangerous conversion
	minVal := int64(combineMinValues(leftTree, rightTree, maxNodes))
	nameLocationCache := map[string]int{}

	for len(leftNodes) > 0 {
		left, rght := leftNodes[0], rghtNodes[0]
		leftNodes, rghtNodes = leftNodes[1:], rghtNodes[1:]
		xLeftOffset, xRghtOffset := xLeftOffsets[0], xRghtOffsets[0]
		xLeftOffsets, xRghtOffsets = xLeftOffsets[1:], xRghtOffsets[1:]

		level := levels[0]
		levels = levels[1:]

		// both left.Name and rght.Name must be the same
		name := string(left.name)
		if left.total >= minVal || rght.total >= minVal || name == "other" {
			i, ok := nameLocationCache[name]
			if !ok {
				i = len(res.Names)
				nameLocationCache[name] = i
				if i == 0 {
					name = "total"
				}

				res.Names = append(res.Names, name)
			}

			if level == len(res.Levels) {
				res.Levels = append(res.Levels, &querierv1.Level{})
			}
			res.MaxSelf = max(res.MaxSelf, left.self)
			res.MaxSelf = max(res.MaxSelf, rght.self)

			// i+0 = x offset, left  tree
			// i+1 = total   , left  tree
			// i+2 = self    , left  tree
			// i+3 = x offset, right tree
			// i+4 = total   , right tree
			// i+5 = self    , right tree
			// i+6 = index in the names array
			values := []int64{
				xLeftOffset, left.total, left.self,
				xRghtOffset, rght.total, rght.self,
				int64(i),
			}

			res.Levels[level].Values = append(values, res.Levels[level].Values...)
			xLeftOffset += left.self
			xRghtOffset += rght.self
			otherLeftTotal, otherRghtTotal := int64(0), int64(0)

			// both left and right must have the same number of children nodes
			for ni := range left.children {
				leftNode, rghtNode := left.children[ni], rght.children[ni]
				if leftNode.total >= minVal || rghtNode.total >= minVal {
					levels = prependInt(levels, level+1)
					xLeftOffsets = prependInt64(xLeftOffsets, xLeftOffset)
					xRghtOffsets = prependInt64(xRghtOffsets, xRghtOffset)
					leftNodes = prependTreeNode(leftNodes, leftNode)
					rghtNodes = prependTreeNode(rghtNodes, rghtNode)
					xLeftOffset += leftNode.total
					xRghtOffset += rghtNode.total
				} else {
					otherLeftTotal += leftNode.total
					otherRghtTotal += rghtNode.total
				}
			}
			if otherLeftTotal != 0 || otherRghtTotal != 0 {
				levels = prependInt(levels, level+1)
				{
					leftNode := &node{
						name:  "other",
						total: otherLeftTotal,
						self:  otherLeftTotal,
					}
					xLeftOffsets = prependInt64(xLeftOffsets, xLeftOffset)
					leftNodes = prependTreeNode(leftNodes, leftNode)
				}
				{
					rghtNode := &node{
						name:  "other",
						total: otherRghtTotal,
						self:  otherRghtTotal,
					}
					xRghtOffsets = prependInt64(xRghtOffsets, xRghtOffset)
					rghtNodes = prependTreeNode(rghtNodes, rghtNode)
				}
			}
		}
	}

	deltaEncoding(res.Levels, 0, 7)
	deltaEncoding(res.Levels, 3, 7)

	return res
}

// addTotalRoot updates the tree root with a 'total' node
func addTotalRoot(t *tree) int64 {
	var total int64
	for _, node := range t.root {
		total += node.total
	}

	t.root = []*node{{children: t.root, total: total, name: "total"}}
	return total
}

// combineTree aligns 2 trees by making them having the same structure with the
// same number of nodes
func combineTree(leftTree, rightTree *tree) (*tree, *tree) {
	leftNodes := leftTree.root
	rghtNodes := rightTree.root

	for len(leftNodes) > 0 {
		left, rght := leftNodes[0], rghtNodes[0]
		leftNodes, rghtNodes = leftNodes[1:], rghtNodes[1:]

		left.children, rght.children = combineNodes(left.children, rght.children)
		leftNodes = append(leftNodes, left.children...)
		rghtNodes = append(rghtNodes, rght.children...)
	}
	return leftTree, rightTree
}

// combineNodes makes 2 slices of nodes equal
// by filling with non existing nodes
// and sorting lexicographically
func combineNodes(leftNodes, rghtNodes []*node) ([]*node, []*node) {
	size := nextPow2(maxInt(len(leftNodes), len(rghtNodes)))
	leftResult := make([]*node, 0, size)
	rghtResult := make([]*node, 0, size)

	for len(leftNodes) != 0 && len(rghtNodes) != 0 {
		left, rght := leftNodes[0], rghtNodes[0]
		switch bytes.Compare([]byte(left.name), []byte(rght.name)) {
		case 0:
			leftResult = append(leftResult, left)
			rghtResult = append(rghtResult, rght)
			leftNodes, rghtNodes = leftNodes[1:], rghtNodes[1:]
		case -1:
			leftResult = append(leftResult, left)
			rghtResult = append(rghtResult, &node{name: left.name})
			leftNodes = leftNodes[1:]
		case 1:
			leftResult = append(leftResult, &node{name: rght.name})
			rghtResult = append(rghtResult, rght)
			rghtNodes = rghtNodes[1:]
		}
	}
	leftResult = append(leftResult, leftNodes...)
	rghtResult = append(rghtResult, rghtNodes...)
	for _, left := range leftNodes {
		rghtResult = append(rghtResult, &node{name: left.name})
	}
	for _, rght := range rghtNodes {
		leftResult = append(leftResult, &node{name: rght.name})
	}
	return leftResult, rghtResult
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func nextPow2(a int) int {
	a--
	a |= a >> 1
	a |= a >> 2
	a |= a >> 4
	a |= a >> 8
	a |= a >> 16
	a++
	return a
}

func combineMinValues(leftTree, rightTree *tree, maxNodes int) uint64 {
	c := cappedarr.New(maxNodes)
	combineIterateWithTotal(leftTree, rightTree, func(left uint64, right uint64) bool {
		return c.Push(maxUint64(left, right))
	})
	return c.MinValue()
}

// // iterate both trees, both trees must be returned from CombineTree
func combineIterateWithTotal(leftTree, rightTree *tree, cb func(uint64, uint64) bool) {
	leftNodes, rghtNodes := leftTree.root, rightTree.root
	i := 0
	for len(leftNodes) > 0 {
		leftNode, rghtNode := leftNodes[0], rghtNodes[0]
		leftNodes, rghtNodes = leftNodes[1:], rghtNodes[1:]
		i++

		// TODO: dangerous conversion
		if cb(uint64(leftNode.total), uint64(rghtNode.total)) {
			leftNodes = append(leftNode.children, leftNodes...)
			rghtNodes = append(rghtNode.children, rghtNodes...)
		}
	}
}

func deltaEncoding(levels []*querierv1.Level, start, step int) {
	for _, l := range levels {
		prev := int64(0)
		for i := start; i < len(l.Values); i += step {
			l.Values[i] -= prev
			prev += l.Values[i] + l.Values[i+1]
		}
	}
}

func prependInt(s []int, x int) []int {
	s = append(s, 0)
	copy(s[1:], s)
	s[0] = x
	return s
}

func prependInt64(s []int64, x int64) []int64 {
	s = append(s, 0)
	copy(s[1:], s)
	s[0] = x
	return s
}

func prependTreeNode(s []*node, x *node) []*node {
	s = append(s, nil)
	copy(s[1:], s)
	s[0] = x
	return s
}
