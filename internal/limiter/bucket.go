package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity		float64
	usedTokens		float64
	usableTokens	float64
	refillRate		float64
	lastRefilled	time.Time
	mutex			sync.Mutex
}

func newTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		usedTokens: 0,
		usableTokens: capacity,
		refillRate: refillRate,
		lastRefilled: time.Now().UTC(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now().UTC()
	elapsed := now.Sub(tb.lastRefilled).Seconds()
	tb.usableTokens += elapsed * tb.refillRate
	tb.usableTokens = min(tb.usedTokens + tb.capacity, tb.usableTokens)
	tb.lastRefilled = now
}

func (tb *TokenBucket) consume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.refill()
	if tb.usableTokens - tb.usedTokens >= 1 {
		tb.usedTokens++
		return true
	}
	return false
}

func (tb *TokenBucket) merge(incoming *TokenBucket) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.usedTokens = max(incoming.usedTokens, tb.usedTokens)
	tb.usableTokens = max(incoming.usableTokens, tb.usableTokens)
	if incoming.lastRefilled.After(tb.lastRefilled) {
		tb.lastRefilled = incoming.lastRefilled
	}
}