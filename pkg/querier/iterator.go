package querier

// This file implements iterator.Interface specifics to querier code.
// If you want to use for other types, we should move those to generics.

import (
	"container/heap"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/grafana/dskit/multierror"

	ingestv1 "github.com/grafana/fire/pkg/gen/ingester/v1"
	"github.com/grafana/fire/pkg/iterator"
	firemodel "github.com/grafana/fire/pkg/model"
)

var (
	_ = iterator.Interface[ProfileWithLabels]((*StreamProfileIterator)(nil))
	_ = iterator.Interface[ProfileWithLabels]((*DedupeProfileIterator)(nil))
)

type ProfileWithLabels struct {
	*ingestv1.Profile
	firemodel.Labels
	ingesterAddr string
}

func (p ProfileWithLabels) String() string {
	return fmt.Sprintf("id:%s ts:%d labels:%s ingester:%s", p.Profile.ID, p.Timestamp, p.Labels, p.ingesterAddr)
}

type StreamProfileIterator struct {
	stream       *connect.ServerStreamForClient[ingestv1.SelectProfilesResponse]
	current      *ingestv1.SelectProfilesResponse
	ingesterAddr string
}

func NewStreamProfileIterator(r responseFromIngesters[*connect.ServerStreamForClient[ingestv1.SelectProfilesResponse]]) iterator.Interface[ProfileWithLabels] {
	return &StreamProfileIterator{
		stream:       r.response,
		ingesterAddr: r.addr,
	}
}

func NewStreamsProfileIterator(r []responseFromIngesters[*connect.ServerStreamForClient[ingestv1.SelectProfilesResponse]]) iterator.Interface[ProfileWithLabels] {
	its := make([]iterator.Interface[ProfileWithLabels], 0, len(r))
	for _, r := range r {
		its = append(its, NewStreamProfileIterator(r))
	}
	return NewDedupeProfileIterator(its)
}

func (s *StreamProfileIterator) Next() bool {
	if s.current == nil || len(s.current.Profiles) <= 1 {
		if s.stream.Receive() {
			s.current = s.stream.Msg()
			return len(s.current.Profiles) != 0
		}
		return false
	}
	s.current.Profiles = s.current.Profiles[1:]
	return true
}

func (s *StreamProfileIterator) At() ProfileWithLabels {
	return ProfileWithLabels{
		Profile:      s.current.Profiles[0],
		Labels:       s.current.Labelsets[s.current.Profiles[0].LabelsetIndex].Labels,
		ingesterAddr: s.ingesterAddr,
	}
}

func (s *StreamProfileIterator) Err() error {
	return s.stream.Err()
}

func (s *StreamProfileIterator) Close() {
	s.stream.Close()
}

type ProfileIteratorHeap []iterator.Interface[ProfileWithLabels]

func (h ProfileIteratorHeap) Len() int { return len(h) }
func (h ProfileIteratorHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h ProfileIteratorHeap) Peek() iterator.Interface[ProfileWithLabels] { return h[0] }
func (h *ProfileIteratorHeap) Push(x interface{}) {
	*h = append(*h, x.(iterator.Interface[ProfileWithLabels]))
}

func (h *ProfileIteratorHeap) Pop() interface{} {
	n := len(*h)
	x := (*h)[n-1]
	*h = (*h)[0 : n-1]
	return x
}

func (h ProfileIteratorHeap) Less(i, j int) bool {
	p1, p2 := h[i].At(), h[j].At()
	if p1.Timestamp == p2.Timestamp {
		return firemodel.CompareLabelPairs(p1.Labels, p2.Labels) < 0
	}
	return p1.Timestamp < p2.Timestamp
}

type DedupeProfileIterator struct {
	heap *ProfileIteratorHeap
	errs []error
	curr ProfileWithLabels

	tuples []tuple
}

type tuple struct {
	ProfileWithLabels
	iterator.Interface[ProfileWithLabels]
}

// NewDedupeProfileIterator creates a new an iterator of ProfileWithLabels while
// iterating it removes duplicate Profile by ID across the set of iterators, but not within.
func NewDedupeProfileIterator(its []iterator.Interface[ProfileWithLabels]) iterator.Interface[ProfileWithLabels] {
	heap := make(ProfileIteratorHeap, 0, len(its))
	res := &DedupeProfileIterator{
		heap:   &heap,
		tuples: make([]tuple, 0, len(its)),
	}
	for _, iter := range its {
		res.requeue(iter, false)
	}
	return res
}

func (i *DedupeProfileIterator) requeue(ei iterator.Interface[ProfileWithLabels], advanced bool) {
	if advanced || ei.Next() {
		heap.Push(i.heap, ei)
		return
	}
	ei.Close()
	if err := ei.Err(); err != nil {
		i.errs = append(i.errs, err)
	}
}

func (i *DedupeProfileIterator) Next() bool {
	if i.heap.Len() == 0 {
		return false
	}
	if i.heap.Len() == 1 {
		i.curr = i.heap.Peek().At()
		if !i.heap.Peek().Next() {
			i.heap.Pop()
		}
		return true
	}

	for i.heap.Len() > 0 {
		next := i.heap.Peek()
		value := next.At()
		if len(i.tuples) > 0 && i.tuples[0].Timestamp < value.Timestamp {
			break
		}
		heap.Pop(i.heap)
		i.tuples = append(i.tuples, tuple{
			ProfileWithLabels: value,
			Interface:         next,
		})
	}
	// shortcut if we have a single tuple.
	if len(i.tuples) == 1 {
		i.curr = i.tuples[0].ProfileWithLabels
		i.requeue(i.tuples[0].Interface, false)
		i.tuples = i.tuples[:0]
		return true
	}

	// todo: we might want to pick based on ingester addr.
	t := i.tuples[0]
	i.requeue(t.Interface, false)
	i.curr = t.ProfileWithLabels

	for _, t := range i.tuples[1:] {
		if t.ProfileWithLabels.ID == i.curr.ID {
			i.requeue(t.Interface, false)
			continue
		}
		i.requeue(t.Interface, true)
	}
	i.tuples = i.tuples[:0]

	return true
}

func (i *DedupeProfileIterator) At() ProfileWithLabels {
	return i.curr
}

func (i *DedupeProfileIterator) Err() error {
	return multierror.New(i.errs...).Err()
}

func (i *DedupeProfileIterator) Close() {
	for _, s := range *i.heap {
		s.Close()
		if err := s.Err(); err != nil {
			i.errs = append(i.errs, err)
		}
	}
}
