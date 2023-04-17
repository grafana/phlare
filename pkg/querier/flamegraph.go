package querier

import (
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/samber/lo"

	querierv1 "github.com/grafana/phlare/api/gen/proto/go/querier/v1"
	typesv1 "github.com/grafana/phlare/api/gen/proto/go/types/v1"
)

type stackNode struct {
	xOffset int
	level   int
	node    *node
}

func NewFlameGraph(t *tree) *querierv1.FlameGraph {
	var total, max int64
	for _, node := range t.root {
		total += node.total
	}
	names := []string{}
	nameLocationCache := map[string]int{}
	res := []*Stack[int64]{}
	defer func() {
		for _, stack := range res {
			stackIntPool.Put(stack)
		}
	}()

	stack := stackNodePool.Get().(*Stack[stackNode])
	defer stackNodePool.Put(stack)
	stack.Reset()
	stack.Push(stackNode{xOffset: 0, level: 0, node: &node{children: t.root, total: total}})

	for {
		current, hasMoreNodes := stack.Pop()
		if !hasMoreNodes {
			break
		}
		if current.node.self > max {
			max = current.node.self
		}
		var i int
		var ok bool
		name := current.node.name
		if i, ok = nameLocationCache[name]; !ok {
			i = len(names)
			if i == 0 {
				name = "total"
			}
			nameLocationCache[name] = i
			names = append(names, name)
		}

		if current.level == len(res) {
			s := stackIntPool.Get().(*Stack[int64])
			s.Reset()
			res = append(res, s)
		}

		// i+0 = x offset
		// i+1 = total
		// i+2 = self
		// i+3 = index in names array
		level := res[current.level]
		level.Push(int64(i))
		level.Push((current.node.self))
		level.Push((current.node.total))
		level.Push(int64(current.xOffset))
		current.xOffset += int(current.node.self)

		for _, child := range current.node.children {
			stack.Push(stackNode{xOffset: current.xOffset, level: current.level + 1, node: child})
			current.xOffset += int(child.total)
		}
	}
	result := make([][]int64, len(res))
	for i := range result {
		result[i] = res[i].Slice()
	}
	// delta encode xoffsets
	for _, l := range result {
		prev := int64(0)
		for i := 0; i < len(l); i += 4 {
			l[i] -= prev
			prev += l[i] + l[i+1]
		}
	}
	levels := make([]*querierv1.Level, len(result))
	for i := range levels {
		levels[i] = &querierv1.Level{
			Values: result[i],
		}
	}

	return &querierv1.FlameGraph{
		Names:   names,
		Levels:  levels,
		Total:   total,
		MaxSelf: max,
	}
}

type FlamegraphFormat string

const (
	FormatSingle FlamegraphFormat = "single"
	FormatDiff   FlamegraphFormat = "double"
)

// ExportToFlamebearer exports the flamegraph to a Flamebearer struct.
func ExportToFlamebearer(fg *querierv1.FlameGraph, profileType *typesv1.ProfileType, format FlamegraphFormat) *flamebearer.FlamebearerProfile {
	unit := metadata.Units(profileType.SampleUnit)
	sampleRate := uint32(100)

	switch profileType.SampleType {
	case "inuse_objects", "alloc_objects", "goroutine", "samples":
		unit = metadata.ObjectsUnits
	case "cpu":
		unit = metadata.SamplesUnits
		sampleRate = uint32(100000000)

	}
	levels := make([][]int, len(fg.Levels))
	for i := range levels {
		levels[i] = lo.Map(fg.Levels[i].Values, func(v int64, i int) int { return int(v) })
	}
	return &flamebearer.FlamebearerProfile{
		Version: 1,
		FlamebearerProfileV1: flamebearer.FlamebearerProfileV1{
			Flamebearer: flamebearer.FlamebearerV1{
				Names:    fg.Names,
				NumTicks: int(fg.Total),
				MaxSelf:  int(fg.MaxSelf),
				Levels:   levels,
			},
			Metadata: flamebearer.FlamebearerMetadataV1{
				Format:     string(format),
				Units:      unit,
				Name:       profileType.SampleType,
				SampleRate: sampleRate,
			},
		},
	}
}
