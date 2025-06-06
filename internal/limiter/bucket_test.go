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
