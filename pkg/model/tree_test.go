package model

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Tree(t *testing.T) {
	for _, tc := range []struct {
		name     string
		stacks   []stacktraces
		expected func() *Tree
	}{
		{
			"empty",
			[]stacktraces{},
			func() *Tree { return &Tree{} },
		},
		{
			"double node single stack",
			[]stacktraces{
				{
					locations: []string{"buz", "bar"},
					value:     1,
				},
				{
					locations: []string{"buz", "bar"},
					value:     1,
				},
			},
			func() *Tree {
				tr := emptyTree()
				tr.add("bar", 0, 2).Add("buz", 2, 2)
				return tr
			},
		},
		{
			"double node double stack",
			[]stacktraces{
				{
					locations: []string{"blip", "buz", "bar"},
					value:     1,
				},
				{
					locations: []string{"blap", "blop", "buz", "bar"},
					value:     2,
				},
			},
			func() *Tree {
				tr := emptyTree()
				buz := tr.add("bar", 0, 3).Add("buz", 0, 3)
				buz.Add("blip", 1, 1)
				buz.Add("blop", 0, 2).Add("blap", 2, 2)
				return tr
			},
		},
		{
			"multiple stacks and duplicates nodes",
			[]stacktraces{
				{
					locations: []string{"buz", "bar"},
					value:     1,
				},
				{
					locations: []string{"buz", "bar"},
					value:     1,
				},
				{
					locations: []string{"buz"},
					value:     1,
				},
				{
					locations: []string{"foo", "buz", "bar"},
					value:     1,
				},
				{
					locations: []string{"blop", "buz", "bar"},
					value:     2,
				},
				{
					locations: []string{"blip", "bar"},
					value:     4,
				},
			},
			func() *Tree {
				tr := emptyTree()

				bar := tr.add("bar", 0, 9)

				buz := bar.Add("buz", 2, 5)
				buz.Add("foo", 1, 1)
				buz.Add("blop", 2, 2)
				bar.Add("blip", 4, 4)

				tr.add("buz", 1, 1)
				return tr
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.expected()
			tr := newTree(tc.stacks)
			require.Equal(t, tr, expected, "tree should be equal got:%s\n expected:%s\n", tr.String(), expected)
		})
	}
}

func Test_TreeMarshalUnmarshal(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		expected := new(Tree)
		var buf bytes.Buffer
		require.NoError(t, expected.MarshalTruncate(&buf, -1))
		actual, err := UnmarshalTree(buf.Bytes())
		require.NoError(t, err)
		require.Equal(t, expected.String(), actual.String())
	})

	t.Run("non-empty tree", func(t *testing.T) {
		expected := newTree([]stacktraces{
			{locations: []string{"c", "b", "a"}, value: 1},
			{locations: []string{"c", "b", "a"}, value: 1},
			{locations: []string{"c1", "b", "a"}, value: 1},
			{locations: []string{"c", "b1", "a"}, value: 1},
			{locations: []string{"c1", "b1", "a"}, value: 1},
			{locations: []string{"c", "b", "a1"}, value: 1},
			{locations: []string{"c1", "b", "a1"}, value: 1},
			{locations: []string{"c", "b1", "a1"}, value: 1},
			{locations: []string{"c1", "b1", "a1"}, value: 1},
		})

		var buf bytes.Buffer
		require.NoError(t, expected.MarshalTruncate(&buf, -1))
		actual, err := UnmarshalTree(buf.Bytes())
		require.NoError(t, err)
		require.Equal(t, expected.String(), actual.String())
	})
}
