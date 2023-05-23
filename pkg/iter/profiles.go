package iter

import (
	"github.com/grafana/dskit/multierror"
	"github.com/prometheus/common/model"

	phlaremodel "github.com/grafana/phlare/pkg/model"
	"github.com/grafana/phlare/pkg/util/loser"
)

type Timestamp interface {
	Timestamp() model.Time
}

type Profile interface {
	Labels() phlaremodel.Labels
	Timestamp
}

func lessProfile(p1, p2 Profile) bool {
	if p1.Timestamp() == p2.Timestamp() {
		// todo we could compare SeriesRef here
		return phlaremodel.CompareLabelPairs(p1.Labels(), p2.Labels()) < 0
	}
	return p1.Timestamp() < p2.Timestamp()
}

type MergeIterator[P Profile] struct {
	tree        *loser.Tree[P, Iterator[P]]
	errs        multierror.MultiError
	current     P
	deduplicate bool
}

// NewMergeIterator returns an iterator that k-way merges the given iterators.
// The given iterators must be sorted by timestamp and labels themselves.
// Optionally, the iterator can deduplicate profiles with the same timestamp and labels.
func NewMergeIterator[P Profile](maxVal P, deduplicate bool, iters ...Iterator[P]) Iterator[P] {
	iter := &MergeIterator[P]{
		deduplicate: deduplicate,
	}
	iter.tree = loser.New(
		iters,
		maxVal,
		func(s Iterator[P]) P {
			return s.At()
		},
		func(p1, p2 P) bool {
			return lessProfile(p1, p2)
		},
		func(s Iterator[P]) {
			if err := s.Close(); err != nil {
				iter.errs.Add(err)
			}
		})
	iter.current = maxVal
	return iter
}

func (i *MergeIterator[P]) Next() bool {
	for i.tree.Next() {
		next := i.tree.Winner()
		if !i.deduplicate || (next.At().Timestamp() != i.current.Timestamp() || phlaremodel.CompareLabelPairs(next.At().Labels(), i.current.Labels()) != 0) {
			i.current = next.At()
			return true
		}
	}
	return false
}

func (i *MergeIterator[P]) At() P {
	return i.current
}

func (i *MergeIterator[P]) Err() error {
	return i.errs.Err()
}

func (i *MergeIterator[P]) Close() error {
	i.tree.Close()
	return i.Err()
}

type TimeRangedIterator[T Timestamp] struct {
	Iterator[T]
	min, max model.Time
}

func NewTimeRangedIterator[T Timestamp](it Iterator[T], min, max model.Time) Iterator[T] {
	return &TimeRangedIterator[T]{
		Iterator: it,
		min:      min,
		max:      max,
	}
}

func (it *TimeRangedIterator[T]) Next() bool {
	for it.Iterator.Next() {
		if it.At().Timestamp() < it.min {
			continue
		}
		if it.At().Timestamp() > it.max {
			return false
		}
		return true
	}
	return false
}
