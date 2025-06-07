package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity     float64
	refillRate   float64
	usedTokens   float64
	usableTokens float64
	lastRefilled time.Time
	mutex        sync.RWMutex
}

func newTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:     capacity,
		refillRate:   refillRate,
		usedTokens:   0,
		usableTokens: capacity,
		lastRefilled: time.Now().UTC(),
	}
}

func (tb *TokenBucket) refill() {
	elapsed := time.Since(tb.lastRefilled).Seconds()
	tb.lastRefilled = time.Now().UTC()
	tb.usableTokens += elapsed * tb.refillRate
	tb.usableTokens = min(tb.usedTokens+tb.capacity, tb.usableTokens)
}

func (tb *TokenBucket) consume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.refill()
	if tb.usableTokens-tb.usedTokens > 0 {
		tb.usedTokens += 1
		return true
	}
	return false
}

func (tb *TokenBucket) merge(incoming *BucketState) {
	if incoming == nil {
		return
	}
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.usedTokens = max(incoming.UsedTokens, tb.usedTokens)
	tb.usableTokens = max(incoming.UsableTokens, tb.usableTokens)
	tb.lastRefilled = time.Unix(max(incoming.LastRefilled.Unix(), tb.lastRefilled.Unix()), 0)
}
