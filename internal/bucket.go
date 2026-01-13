package internal

import (
	"sync"
	"time"
)

type TokenBucket struct {
	Capacity     float64   `json:"capacity"`
	RefillRate   float64   `json:"refill_rate"`
	UsedTokens   float64   `json:"used_tokens"`
	UsableTokens float64   `json:"usable_tokens"`
	LastRefilled time.Time `json:"last_refilled"`
	mutex        sync.Mutex
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	initiallyUsedTokens := 0.0
	currentTimestamp := time.Now().UTC()
	return &TokenBucket{
		Capacity:     capacity,
		RefillRate:   refillRate,
		UsedTokens:   initiallyUsedTokens,
		UsableTokens: capacity,
		LastRefilled: currentTimestamp,
	}
}

func (tb *TokenBucket) Refill() {
	currentTimestamp := time.Now().UTC()
	elapsedTime := time.Since(tb.LastRefilled).Seconds()
	maximumGrantableTokens := tb.UsableTokens + elapsedTime*tb.RefillRate
	maximumAllowedTokens := tb.UsedTokens + tb.Capacity
	tb.UsableTokens = min(maximumGrantableTokens, maximumAllowedTokens)
	tb.LastRefilled = currentTimestamp
}

func (tb *TokenBucket) Consume() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.Refill()
	spendableTokens := tb.UsableTokens - tb.UsedTokens
	if spendableTokens > 0 {
		tb.UsedTokens += 1
		return true
	}
	return false
}

func (tb *TokenBucket) Merge(incoming *BucketState) time.Duration {
	if incoming == nil {
		return 0
	}
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.UsedTokens = max(incoming.UsedTokens, tb.UsedTokens)
	tb.UsableTokens = max(incoming.UsableTokens, tb.UsableTokens)
	latestTimestamp := max(incoming.LastRefilled.Unix(), tb.LastRefilled.Unix())
	tb.LastRefilled = time.Unix(latestTimestamp, 0)
	return time.Since(incoming.Timestamp)
}
