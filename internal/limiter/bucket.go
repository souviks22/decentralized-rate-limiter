package limiter

import (
	"sync"
	"time"
	"fmt"
)

type TokenBucket struct {
	Capacity		float64
	UsedTokens		float64
	UsableTokens	float64
	RefillRate		float64
	LastRefilled	time.Time
	Mutex			sync.Mutex
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		Capacity: capacity,
		UsedTokens: 0,
		UsableTokens: capacity,
		RefillRate: refillRate,
		LastRefilled: time.Now().UTC(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now().UTC()
	elapsed := now.Sub(tb.LastRefilled).Seconds()
	tb.UsableTokens += elapsed * tb.RefillRate
	tb.UsableTokens = min(tb.UsedTokens + tb.Capacity, tb.UsableTokens)
	tb.LastRefilled = now
}

func (tb *TokenBucket) consume() bool {
	tb.Mutex.Lock()
	defer tb.Mutex.Unlock()
	tb.refill()
	if tb.UsableTokens - tb.UsedTokens >= 1 {
		tb.UsedTokens++
		return true
	}
	return false
}

func (tb *TokenBucket) merge(incoming *TokenBucket) {
	tb.Mutex.Lock()
	defer tb.Mutex.Unlock()
	tb.UsedTokens = max(incoming.UsedTokens, tb.UsedTokens)
	tb.UsableTokens = max(incoming.UsableTokens, tb.UsableTokens)
	if incoming.LastRefilled.After(tb.LastRefilled) {
		tb.LastRefilled = incoming.LastRefilled
	}
}

func (tb *TokenBucket) String() string {
	return fmt.Sprintf("%f", tb.UsedTokens)
}