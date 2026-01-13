package internal

import (
	"testing"
	"time"
)

func TestBucketRefill(t *testing.T) {
	capacity, refillRate := 10.0, 1.0
	bucket := NewTokenBucket(capacity, refillRate)
	for range 20 {
		if !bucket.Consume() {
			t.Error("Bucker didn't refill at right time")
		}
		time.Sleep(time.Second)
	}
}

func TestTokenExhaustion(t *testing.T) {
	capacity, refillRate := 10.0, 1.0
	bucket := NewTokenBucket(capacity, refillRate)
	successfulConsumptions := 0
	for range 20 {
		if bucket.Consume() {
			successfulConsumptions += 1
		}
	}
	if successfulConsumptions == 20 {
		t.Error("Tokens didn't exhaust")
	}
}

func TestBucketMerge(t *testing.T) {
	capacity, refillRate := 10.0, 1.0
	bucket := NewTokenBucket(capacity, refillRate)
	incoming := &BucketState{
		UsableTokens: 11,
		UsedTokens:   1,
		LastRefilled: time.Now().UTC(),
	}
	bucket.Merge(incoming)
	if bucket.UsableTokens != 11 || bucket.UsedTokens != 1 {
		t.Error("Bucket merging failed")
	}
}
