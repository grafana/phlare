package main

import (
	"log"
	"os"

	"github.com/klauspost/compress/gzip"

	profilev1 "github.com/grafana/fire/pkg/gen/google/v1"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var p = &profilev1.Profile{
		SampleType: []*profilev1.ValueType{{
			Unit: 1,
			Type: 2,
		}},
		Sample: []*profilev1.Sample{
			{
				LocationId: []uint64{
					1,
				},
				Value: []int64{123},
				// TODO: Label
			},
			{
				LocationId: []uint64{
					3,
					2,
					1,
				},
				Value: []int64{80},
				Label: []*profilev1.Label{{
					Key:     6,
					Num:     4,
					NumUnit: 10,
				}},
			},
			{
				LocationId: []uint64{
					4,
					2,
					1,
				},
				Value: []int64{35},
				Label: []*profilev1.Label{{
					Key:     6,
					Num:     5,
					NumUnit: 10,
				}},
			},
			{
				LocationId: []uint64{
					3,
					2,
					1,
				},
				Value: []int64{96},
				Label: []*profilev1.Label{{
					Key:     6,
					Num:     6,
					NumUnit: 10,
				}},
			},
		},
		Location: []*profilev1.Location{
			{
				Id:        1,
				MappingId: 1,
				Address:   0x1,
				Line: []*profilev1.Line{
					{FunctionId: 1, Line: 1},
				},
			},
			{
				Id:        2,
				MappingId: 1,
				Address:   0x2,
				Line: []*profilev1.Line{
					{FunctionId: 2, Line: 2},
				},
			},
			{
				Id:        3,
				MappingId: 1,
				Address:   0x3,
				Line: []*profilev1.Line{
					{FunctionId: 3, Line: 3},
				},
			},
			{
				Id:        4,
				MappingId: 1,
				Address:   0x4,
				Line: []*profilev1.Line{
					{FunctionId: 4, Line: 4},
				},
			},
		},
		Mapping: []*profilev1.Mapping{
			{
				Id:             1,
				MemoryStart:    0x0,
				MemoryLimit:    0xf,
				Filename:       5,
				HasFunctions:   true,
				HasFilenames:   true,
				HasLineNumbers: true,
			},
		},
		Function: []*profilev1.Function{
			{Id: 1, Name: 3},
			{Id: 2, Name: 4},
			{Id: 3, Name: 7},
			{Id: 4, Name: 8},
		},
		StringTable: []string{
			"",
			"dollars",
			"expedidature",
			"loki-in-paris",
			"drinks",
			"my-sample-binary",
			"bytes",
			"beer",
			"wine",
			"food",
			"ml",
		},
	}

	data, err := p.MarshalVT()
	if err != nil {
		return err
	}

	w := gzip.NewWriter(os.Stdout)
	if _, err := w.Write(data); err != nil {
		return err
	}

	return w.Close()
}
