package phlaredb

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	// profilev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	profilev1 "github.com/grafana/phlare/api/gen/proto/go/google/v1"
	"github.com/grafana/phlare/pkg/phlaredb/block"
	schemav1 "github.com/grafana/phlare/pkg/phlaredb/schemas/v1"
	"github.com/pkg/errors"
	"github.com/segmentio/parquet-go"
	"golang.org/x/sync/errgroup"
	// "github.com/apache/arrow/go/v12/arrow"
)

func Compact(ctx context.Context, blocks []*singleBlockQuerier, dstDir string) error {
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].meta.ULID.Compare(blocks[j].meta.ULID) < 0
	})

	out := block.NewMeta()
	out.Source = block.CompactorSource
	out.Version = block.MetaVersion1
	// compute min/max time out of all blocks
	// may be this should be done at the end.
	for _, b := range blocks {
		if b.meta.MinTime < out.MinTime {
			out.MinTime = b.meta.MinTime
		}
		if b.meta.MaxTime > out.MaxTime {
			out.MaxTime = b.meta.MaxTime
		}
	}
	blocks = blocks[:50]
	for _, b := range blocks {
		fmt.Println(b.meta.ULID)
	}
	// arrow.
	// merge all symbols and keep track of the rewritten offsets
	sbls, err := newInMemorySymbols(ctx, blocks, dstDir)
	if err != nil {
		return err
	}
	fmt.Println("total stacktraces", len(sbls.stacktraces.slice))
	return nil
}

// type profileWriter struct {
// 	profileStore
// }

type inMemorySymbols struct {
	strings     deduplicatingSlice[string, string, *stringsHelper, *schemav1.StringPersister]
	mappings    deduplicatingSlice[*profilev1.Mapping, mappingsKey, *mappingsHelper, *schemav1.MappingPersister]
	functions   deduplicatingSlice[*profilev1.Function, functionsKey, *functionsHelper, *schemav1.FunctionPersister]
	locations   deduplicatingSlice[*profilev1.Location, locationsKey, *locationsHelper, *schemav1.LocationPersister]
	stacktraces deduplicatingSlice[*schemav1.Stacktrace, stacktracesKey, *stacktracesHelper, *schemav1.StacktracePersister] // a stacktrace is a slice of location ids
}

func newInMemorySymbols(ctx context.Context, blocks []*singleBlockQuerier, path string) (*inMemorySymbols, error) {
	result := &inMemorySymbols{}
	metrics := newHeadMetrics(nil)
	if err := result.strings.Init(path, defaultParquetConfig, metrics); err != nil {
		return nil, err
	}
	if err := result.mappings.Init(path, defaultParquetConfig, metrics); err != nil {
		return nil, err
	}
	if err := result.functions.Init(path, defaultParquetConfig, metrics); err != nil {
		return nil, err
	}
	if err := result.locations.Init(path, defaultParquetConfig, metrics); err != nil {
		return nil, err
	}
	if err := result.stacktraces.Init(path, defaultParquetConfig, metrics); err != nil {
		return nil, err
	}
	conversionPerBlock := make([]idConversionTable, len(blocks))
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(1)
	for i, b := range blocks {
		b := b
		i := i
		g.Go(func() error {
			fmt.Println("processing block", b.meta.ULID)

			// todo We should only open symbols at this point
			if err := b.Open(ctx); err != nil {
				return err
			}
			rewriter := &rewriter{}

			stringsRows := b.strings.Rows()
			strings := make([]string, len(stringsRows))
			for i := range strings {
				strings[i] = stringsRows[i].String
			}
			if err := result.strings.ingest(ctx, strings, rewriter); err != nil {
				return err
			}
			if err := result.mappings.ingest(ctx, b.mappings.Rows(), rewriter); err != nil {
				return err
			}
			type mappingPerName struct {
				name     string
				mappings []*profilev1.Mapping
			}
			mappings := make([]*mappingPerName, 0, len(result.mappings.slice))
			mappingPerNameMap := make(map[string]*mappingPerName, len(result.mappings.slice))
			for _, m := range result.mappings.slice {
				m := m
				mpn, ok := mappingPerNameMap[result.strings.slice[m.Filename]]
				if !ok {
					mpn = &mappingPerName{name: result.strings.slice[m.Filename]}
					mappingPerNameMap[result.strings.slice[m.Filename]] = mpn
					mappings = append(mappings, mpn)
				}
				mpn.mappings = append(mpn.mappings, m)
			}
			for _, mpn := range mappings {
				fmt.Println("mapping", mpn.name)
				for _, m := range mpn.mappings {
					fmt.Println("  ", m)
				}
			}
			sort.Slice(mappings, func(i, j int) bool {
				return mappings[i].name < mappings[j].name
			})
			if err := result.functions.ingest(ctx, b.functions.Rows(), rewriter); err != nil {
				return err
			}
			locations := b.locations.Rows()
			if err := result.locations.ingest(ctx, locations, rewriter); err != nil {
				return err
			}
			locationsPerMapping := make(map[uint64]uint64, len(result.locations.slice))
			for _, l := range result.locations.slice {
				l := l
				locationsPerMapping[l.MappingId]++
			}
			totalUsed := uint64(0)
			for id, l := range locationsPerMapping {
				fmt.Println("mapping", result.strings.slice[result.mappings.slice[id].Filename], "locations", l)
				totalUsed += l
			}
			fmt.Println("total used locations", totalUsed)
			fmt.Println("total locations", len(result.locations.slice))
			batch := make([]*schemav1.Stacktrace, 4*1024)
			reader := parquet.NewGenericReader[*schemav1.Stacktrace](b.stacktraces.reader)
			stPerBlock := 0
			for {
				n, err := reader.Read(batch)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
				if n == 0 {
					break
				}
				stPerBlock += n
				for i := range batch[:n] {
					if err := result.stacktraces.ingest(ctx, []*schemav1.Stacktrace{batch[i]}, rewriter); err != nil {
						return err
					}
				}

			}
			fmt.Println("stacktraces per block", stPerBlock, "total :", len(result.stacktraces.slice), "block", i)
			conversionPerBlock[i] = rewriter.stacktraces
			stacktracePerMapping := make(map[uint64]uint64, len(result.stacktraces.slice))
			for id, s := range result.stacktraces.slice {
				if len(result.stacktraces.slice[id].LocationIDs) == 0 {
					continue
				}

				mapping := result.locations.slice[s.LocationIDs[0]].MappingId
				for _, l := range s.LocationIDs[1:] {
					if result.locations.slice[l].MappingId != mapping {
						panic("two mappings in the same stacktrace")
					}
				}
				stacktracePerMapping[mapping]++
			}
			for id, s := range stacktracePerMapping {
				fmt.Println("mapping", result.strings.slice[result.mappings.slice[id].Filename], "stacktraces", s, " % ", float64(s)/float64(len(result.stacktraces.slice))*100)
			}
			fmt.Println("----")
			printLokiStacktraces(result.strings.slice, result.stacktraces.slice, result.mappings.slice, result.locations.slice, result.functions.slice)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func printLokiStacktraces(stringsTable []string, stacktraces []*schemav1.Stacktrace, mappings []*profilev1.Mapping, locations []*profilev1.Location, functions []*profilev1.Function) {
	var st []string
	dupe := make(map[string][]int)
	var sb strings.Builder
	for sID, s := range stacktraces {
		if len(s.LocationIDs) == 0 {
			continue
		}
		// check if the mapping is log
		mapping := locations[s.LocationIDs[0]].MappingId
		if stringsTable[mappings[mapping].Filename] != "/usr/bin/enterprise-logs" {
			continue
		}
		sb.Reset()
		for _, l := range s.LocationIDs {
			line := locations[l].Line
			for _, f := range line {
				sb.WriteString(stringsTable[functions[f.FunctionId].Name])
				// sb.WriteString(":")
				// sb.WriteString(strconv.Itoa(int(f.Line)))
				sb.WriteString(" < ")
			}
		}
		sb.WriteByte('\n')
		key := sb.String()
		if len(dupe[key]) >= 1 {
			// fmt.Println("duplicate stacktrace ", sID, " ", key, " ", dupe[key])
			// stracktracesdupe := stacktraces[dupe[key][0]]
			// fmt.Println(s)
			// fmt.Println(stracktracesdupe)
			// for _, l := range stracktracesdupe.LocationIDs {
			// 	line := locations[l].Line
			// 	for _, f := range line {
			// 		fmt.Println("  ", stringsTable[functions[f.FunctionId].Name])
			// 	}
			// }
		}
		dupe[key] = append(dupe[key], sID)
		st = append(st, sb.String())
	}
	sort.Strings(st)
	totalDupe := int(0)
	for _, d := range dupe {
		if len(d) > 1 {
			totalDupe += len(d)
		}
	}
	fmt.Println("total stacktraces", len(st))
	fmt.Println("total dupe", totalDupe)
	fmt.Println("% of dupe", float64(totalDupe)/float64(len(st))*100)
	ioutil.WriteFile("stacktraces.txt", []byte(strings.Join(st, "\n")), 0o644)
	// for i := range st {
	// 	// if i+1 >= len(st) {
	// 	// 	break
	// 	// }
	// 	// fmt.Println(cmp.Diff(st[i], st[i+1]))
	// 	fmt.Println(st[i])
	// }
}
