package querier

import (
	"bytes"
	"sort"

	"github.com/cespare/xxhash/v2"
	"github.com/samber/lo"

	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
)

type profileWithSymbols struct {
	profile *ingestv1.Profile
	symbols []string
}

// dedupeProfiles dedupes profiles responses by timestamp and labels.
// It expects profiles from each response to be sorted by timestamp and labels already.
func dedupeProfiles(responses []responseFromIngesters[*ingestv1.SelectProfilesResponse]) []profileWithSymbols {
	// type tuple struct {
	// 	profile *ingestv1.Profile
	// 	responseFromIngesters[*ingestv1.SelectProfilesResponse]
	// }
	// var (
	// 	responsesHeap = newProfilesResponseHeap(responses)
	// 	deduped       []profileWithSymbols
	// 	tuples        = make([]tuple, 0, len(responses))
	// )

	// heap.Init(responsesHeap)

	// for responsesHeap.Len() > 0 || len(tuples) > 0 {
	// 	if responsesHeap.Len() > 0 {
	// 		current := heap.Pop(responsesHeap).(responseFromIngesters[*ingestv1.SelectProfilesResponse])
	// 		if len(current.response.Profiles) == 0 {
	// 			continue
	// 		}
	// 		// add the top profile to the tuple list if the current profile is equal the previous one.
	// 		if len(tuples) == 0 || current.response.Profiles[0].ID == tuples[len(tuples)-1].profile.ID {
	// 			tuples = append(tuples, tuple{
	// 				profile:               current.response.Profiles[0],
	// 				responseFromIngesters: current,
	// 			})
	// 			current.response.Profiles = current.response.Profiles[1:]
	// 			continue
	// 		}
	// 		// the current profile is different.
	// 		heap.Push(responsesHeap, current)
	// 	}
	// 	// if the heap is empty and we don't have tuples we're done.
	// 	if len(tuples) == 0 {
	// 		continue
	// 	}
	// 	// no duplicate found just a single profile.
	// 	if len(tuples) == 1 {
	// 		deduped = append(deduped, profileWithSymbols{profile: tuples[0].profile, symbols: tuples[0].responseFromIngesters.response.FunctionNames})
	// 		if len(tuples[0].response.Profiles) > 0 {
	// 			heap.Push(responsesHeap, tuples[0].responseFromIngesters)
	// 		}
	// 		tuples = tuples[:0]
	// 		continue
	// 	}

	// 	// we have a duplicate let's select first profile from the tuple list.
	// 	first := tuples[0]

	// 	deduped = append(deduped, profileWithSymbols{profile: first.profile, symbols: first.responseFromIngesters.response.FunctionNames})
	// 	if len(first.response.Profiles) > 0 {
	// 		heap.Push(responsesHeap, first.responseFromIngesters)
	// 	}
	// 	for _, t := range tuples[1:] {
	// 		if len(t.response.Profiles) > 0 {
	// 			heap.Push(responsesHeap, t.responseFromIngesters)
	// 			continue
	// 		}
	// 	}
	// 	tuples = tuples[:0]

	// }
	// return deduped
	return nil
}

type stacktraces struct {
	locations []string
	value     int64
}

// Merge stacktraces from multiple ingesters.
func mergeStacktraces(profiles []profileWithSymbols) []stacktraces {
	stacktracesByID := map[uint64]*stacktraces{}
	// buf := bytes.NewBuffer(make([]byte, 0, 4096))

	// for _, profile := range profiles {
	// 	// for _, st := range profile.profile.Stacktraces {
	// 	// 	fns := make([]string, len(st.FunctionIds))
	// 	// 	for i, fnID := range st.FunctionIds {
	// 	// 		fns[i] = profile.symbols[fnID]
	// 	// 	}
	// 	// 	id := stacktraceID(buf, fns)
	// 	// 	stacktrace, ok := stacktracesByID[id]
	// 	// 	if !ok {
	// 	// 		stacktrace = &stacktraces{
	// 	// 			locations: fns,
	// 	// 		}
	// 	// 		stacktracesByID[id] = stacktrace
	// 	// 	}
	// 	// 	stacktrace.value += st.Value
	// 	// }
	// }
	ids := lo.Keys(stacktracesByID)
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	result := make([]stacktraces, len(stacktracesByID))
	for pos, id := range ids {
		result[pos] = *stacktracesByID[id]
	}

	return result
}

func stacktraceID(buf *bytes.Buffer, names []string) uint64 {
	buf.Reset()
	for _, name := range names {
		buf.WriteString(name)
	}
	return xxhash.Sum64(buf.Bytes())
}
