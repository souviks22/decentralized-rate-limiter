package limiter

import (
	"sync"
	"time"
)

type BucketSketch struct {
	capacity     float64
	refillRate   float64
	usedTokens   *CountMinSketch
	usableTokens *CountMinSketch
	lastRefilled *CountMinSketch
	mutex        sync.RWMutex
}

func newBucketSketch(capacity float64, refillRate float64) *BucketSketch {
	return &BucketSketch{
		capacity:     capacity,
		refillRate:   refillRate,
		usedTokens:   newCountMinSketch(0),
		usableTokens: newCountMinSketch(capacity),
		lastRefilled: newCountMinSketch(float64(time.Now().Unix())),
	}
}

func (bs *BucketSketch) refill(userId string) {
	elapsed := float64(time.Now().Unix()) - bs.lastRefilled.estimate(userId)
	usableTokens := bs.usableTokens.estimate(userId)
	possibleUsableTokens := usableTokens + elapsed*bs.refillRate
	possibleUsableTokens = min(bs.usedTokens.estimate(userId)+bs.capacity, possibleUsableTokens)
	bs.usableTokens.add(userId, possibleUsableTokens-usableTokens)
	bs.lastRefilled.add(userId, elapsed)
}

func (bs *BucketSketch) consume(userId string) bool {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.refill(userId)
	if bs.usableTokens.estimate(userId)-bs.usedTokens.estimate(userId) >= 1 {
		bs.usedTokens.add(userId, 1)
		return true
	}
	return false
}

func (bs *BucketSketch) merge(ds *EncodedDeltaSketch) {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	bs.usedTokens.merge(ds.UsedTokens)
	bs.usableTokens.merge(ds.UsableTokens)
	bs.lastRefilled.merge(ds.LastRefilled)
}

func (bs *BucketSketch) getDeltaSketch() *EncodedDeltaSketch {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	return &EncodedDeltaSketch{
		UsedTokens:   bs.usedTokens.encodeDelta(),
		UsableTokens: bs.usableTokens.encodeDelta(),
		LastRefilled: bs.lastRefilled.encodeDelta(),
	}
}

func (bs *BucketSketch) hasDeltaSketch() bool {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	return len(bs.usedTokens.delta) > 0 &&
		len(bs.usableTokens.delta) > 0 &&
		len(bs.lastRefilled.delta) > 0
}
