package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	Capacity     float64
	RefillRate   float64
	UsedTokens   float64
	UsableTokens float64
	LastRefilled time.Time
	mutex        sync.RWMutex
}

func newTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		Capacity:     capacity,
		RefillRate:   refillRate,
		UsedTokens:   0,
		UsableTokens: capacity,
		LastRefilled: time.Now().UTC(),
	}
}

func (tb *TokenBucket) refill() {
	elapsed := time.Since(tb.LastRefilled).Seconds()
	tb.LastRefilled = time.Now().UTC()
	tb.UsableTokens += elapsed * tb.RefillRate
	tb.UsableTokens = min(tb.UsedTokens+tb.Capacity, tb.UsableTokens)
}

func (tb *TokenBucket) consume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.refill()
	if tb.UsableTokens-tb.UsedTokens > 0 {
		tb.UsedTokens += 1
		return true
	}
	return false
}

func (tb *TokenBucket) merge(incoming *BucketState) time.Duration {
	if incoming == nil {
		return 0
	}
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.UsedTokens = max(incoming.UsedTokens, tb.UsedTokens)
	tb.UsableTokens = max(incoming.UsableTokens, tb.UsableTokens)
	tb.LastRefilled = time.Unix(max(incoming.LastRefilled.Unix(), tb.LastRefilled.Unix()), 0)
	return time.Since(incoming.Timestamp)
}
