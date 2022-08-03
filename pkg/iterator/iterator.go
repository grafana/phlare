package iterator

type Interface[T any] interface {
	Next() bool
	At() T
	Err() error
	Close()
}

type errIterator[T any] struct {
	err error
}

func (i *errIterator[T]) Err() error {
	return i.err
}

func (*errIterator[T]) Next() bool {
	return false
}

func (*errIterator[T]) At() (t T) {
	return
}

func (*errIterator[T]) Close() {
}

func NewErrIterator[T any](err error) Interface[T] {
	return &errIterator[T]{
		err: err,
	}
}

type sliceIterator[T any] struct {
	s   []T
	pos int
}

func (i *sliceIterator[T]) Err() error {
	return nil
}

func (i *sliceIterator[T]) Next() bool {
	if i.pos >= len(i.s) {
		return false
	}
	i.pos++
	return true
}

func (i *sliceIterator[T]) At() (t T) {
	return i.s[i.pos-1]
}

func (i *sliceIterator[T]) Close() {
}

func NewSliceIterator[T any](s []T) Interface[T] {
	return &sliceIterator[T]{
		s: s,
	}
}

func Slice[T any](it Interface[T]) ([]T, error) {
	var res []T
	for it.Next() {
		res = append(res, it.At())
	}
	return res, it.Err()
}
