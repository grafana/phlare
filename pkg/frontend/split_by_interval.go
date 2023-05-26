package frontend

import "time"

// TimeIntervalIterator splits a time range into non-overlapping sub-ranges
// are aligned to the interval, where the boundary adjoining on the left is
// not included, e.g: [t1, t2), [t3, t4), ..., [tn-1, tn].
type TimeIntervalIterator struct {
	startTime int64
	endTime   int64
	interval  int64
}

type TimeInterval struct{ Start, End time.Time }

// NewTimeIntervalIterator returns a new interval iterator.
// If the interval is zero, the entire time span is taken as a single interval.
func NewTimeIntervalIterator(startTime, endTime time.Time, interval time.Duration) *TimeIntervalIterator {
	i := TimeIntervalIterator{
		startTime: startTime.UnixNano(),
		endTime:   endTime.UnixNano(),
		interval:  interval.Nanoseconds(),
	}
	if interval == 0 {
		i.interval = 2 * endTime.Sub(startTime).Nanoseconds()
	}
	return &i
}

func (i *TimeIntervalIterator) Next() bool { return i.startTime < i.endTime }

func (i *TimeIntervalIterator) At() TimeInterval {
	t := TimeInterval{Start: time.Unix(0, i.startTime)}
	if i.startTime += i.interval - i.startTime%i.interval; i.endTime > i.startTime {
		t.End = time.Unix(0, i.startTime-1)
	} else {
		t.End = time.Unix(0, i.endTime)
	}
	return t
}

func (*TimeIntervalIterator) Err() error { return nil }

func (*TimeIntervalIterator) Close() error { return nil }
