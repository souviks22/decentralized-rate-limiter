package limiter

import (
	"fmt"
	"log"
	"math"
	"testing"
	"time"
)

type TokenBucket struct {
	capacity   float64
	refillRate float64
	tokens     float64
	timestamp  time.Time
}

func newTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
		timestamp:  time.Now(),
	}
}

func (tb *TokenBucket) allowRequest() bool {
	elapsed := time.Since(tb.timestamp).Seconds()
	tb.timestamp = time.Now()
	tb.tokens = math.Min(tb.tokens+elapsed*tb.refillRate, tb.capacity)
	if tb.tokens > 0 {
		tb.tokens -= 1
		return true
	}
	return false
}

func TestCorrectness(t *testing.T) {
	capacity, refillRate := 10.0, 1.0
	buckets := make(map[string]*TokenBucket)
	for i := range 1000 {
		buckets[fmt.Sprintf("user_%d", i)] = newTokenBucket(capacity, refillRate)
	}
	limiter := NewRateLimiter(int(capacity), refillRate)
	correct := 0.0
	for range 10000 {
		for i := range 1000 {
			userId := fmt.Sprintf("user_%d", i)
			if limiter.AllowRequest(userId) == buckets[userId].allowRequest() {
				correct += 1
			}
		}
	}
	accuracy := correct / 10000000
	log.Println("Correctness Accuracy", accuracy)
	if accuracy < 0.9 {
		t.Error("Correctness is too low")
	}
}
