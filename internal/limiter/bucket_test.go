package limiter

import (
	"testing"
	"time"
)

func TestBucketRefill(t *testing.T) {
	bucket := newTokenBucket(5, 1)
	for range 10 {
		if !bucket.consume() {
			t.Error("Bucker didn't refill at right time")
		}
		time.Sleep(time.Second)
	}
}

func TestTokenExhaustion(t *testing.T) {
	bucket := newTokenBucket(5, 1)
	count := 0
	for range 10 {
		if bucket.consume() {
			count += 1
		}
	}
	if count == 10 {
		t.Error("Tokens didn't exhaust")
	}
}

func TestBucketMerge(t *testing.T) {
	bucket := newTokenBucket(5, 1)
	incoming := &BucketState{
		UsableTokens: 6,
		UsedTokens:   1,
		LastRefilled: time.Now().UTC(),
	}
	bucket.merge(incoming)
	if bucket.UsableTokens != 6 || bucket.UsedTokens != 1 {
		t.Error("Bucket merging failed")
	}
}
