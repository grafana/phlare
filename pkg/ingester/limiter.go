package ingester

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/common/model"

	phlaremodel "github.com/grafana/phlare/pkg/model"
	"github.com/grafana/phlare/pkg/validation"
)

var (
	activeSeriesTimeout = 10 * time.Minute
	activeSeriesCleanup = time.Minute
)

type RingCount interface {
	HealthyInstancesCount() int
}

type Limits interface {
	MaxLocalSeriesPerUser(userID string) int
	MaxGlobalSeriesPerUser(userID string) int
}

type Limiter interface {
	// AllowProfile returns an error if the profile is not allowed to be ingested.
	// The error is a validation error and can be out of order or max series limit reached.
	AllowProfile(fp model.Fingerprint, lbs phlaremodel.Labels, tsNano int64) error
	Stop()
}

type limiter struct {
	limits            Limits
	ring              RingCount
	replicationFactor int
	tenantID          string

	activeSeries  map[model.Fingerprint]int64
	lastTimestamp map[model.Fingerprint]int64

	mtx sync.Mutex // todo: may be shard the lock to avoid latency spikes.

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewLimiter(tenantID string, limits Limits, ring RingCount, replicationFactor int) Limiter {
	ctx, cancel := context.WithCancel(context.Background())

	l := &limiter{
		tenantID:          tenantID,
		limits:            limits,
		ring:              ring,
		replicationFactor: replicationFactor,
		activeSeries:      map[model.Fingerprint]int64{},
		lastTimestamp:     map[model.Fingerprint]int64{},
		cancel:            cancel,
		ctx:               ctx,
	}

	l.wg.Add(1)
	go l.loop()

	return l
}

func (l *limiter) Stop() {
	l.cancel()
	l.wg.Wait()
}

func (l *limiter) loop() {
	defer l.wg.Done()

	ticker := time.NewTicker(activeSeriesCleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.ctx.Done():
			return
		}
	}
}

// cleanup removes the series that have not been used for a while.
func (l *limiter) cleanup() {
	now := time.Now().UnixNano()
	l.mtx.Lock()
	defer l.mtx.Unlock()

	for fp, lastUsed := range l.activeSeries {
		if now-lastUsed > int64(activeSeriesTimeout) {
			delete(l.activeSeries, fp)
		}
	}
}

func (l *limiter) AllowProfile(fp model.Fingerprint, lbs phlaremodel.Labels, tsNano int64) error {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	if err := l.allowNewProfile(fp, lbs, tsNano); err != nil {
		return err
	}
	return l.allowNewSeries(fp)
}

func (l *limiter) allowNewProfile(fp model.Fingerprint, lbs phlaremodel.Labels, tsNano int64) error {
	max, ok := l.lastTimestamp[fp]
	if ok {
		// profile is before the last timestamp
		if tsNano < max {
			return validation.NewErrorf(validation.OutOfOrder, "profile for series %s out of order (received %s last %s)", phlaremodel.LabelPairsString(lbs), time.Unix(0, tsNano), time.Unix(0, max))
		}
	}

	// set the last timestamp
	l.lastTimestamp[fp] = tsNano
	return nil
}

func (l *limiter) allowNewSeries(fp model.Fingerprint) error {
	_, ok := l.activeSeries[fp]
	series := len(l.activeSeries)
	if !ok {
		// can this series be added?
		if err := l.assertMaxSeriesPerUser(l.tenantID, series); err != nil {
			return err
		}
	}

	// update time or add it
	l.activeSeries[fp] = time.Now().UnixNano()
	return nil
}

func (l *limiter) assertMaxSeriesPerUser(tenantID string, series int) error {
	// Start by setting the local limit either from override or default
	localLimit := l.limits.MaxLocalSeriesPerUser(tenantID)

	// We can assume that series are evenly distributed across ingesters
	// so we do convert the global limit into a local limit
	globalLimit := l.limits.MaxGlobalSeriesPerUser(tenantID)
	adjustedGlobalLimit := convertGlobalToLocalLimit(globalLimit, l.ring, l.replicationFactor)

	// Set the calculated limit to the lesser of the local limit or the new calculated global limit
	calculatedLimit := minNonZero(localLimit, adjustedGlobalLimit)

	// If both the local and global limits are disabled, we just
	// use the largest int value
	if calculatedLimit == 0 {
		return nil
	}

	if series < calculatedLimit {
		return nil
	}
	return validation.NewErrorf(validation.StreamLimit, validation.StreamLimitErrorMsg, series, calculatedLimit)
}

func convertGlobalToLocalLimit(globalLimit int, ringCount RingCount, replicationFactor int) int {
	if globalLimit == 0 {
		return 0
	}

	// Given we don't need a super accurate count (ie. when the ingesters
	// topology changes) and we prefer to always be in favor of the tenant,
	// we can use a per-ingester limit equal to:
	// (global limit / number of ingesters) * replication factor
	numIngesters := ringCount.HealthyInstancesCount()

	// May happen because the number of ingesters is asynchronously updated.
	// If happens, we just temporarily ignore the global limit.
	if numIngesters > 0 {
		return int((float64(globalLimit) / float64(numIngesters)) * float64(replicationFactor))
	}

	return 0
}

func minNonZero(first, second int) int {
	if first == 0 || (second != 0 && first > second) {
		return second
	}

	return first
}
