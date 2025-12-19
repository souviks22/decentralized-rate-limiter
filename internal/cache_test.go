package internal

import "testing"

func TestBucketCache(t *testing.T) {
	cache := NewBucketCache("random", 5, 1)
	bucket := cache.GetOrCreateBucket("souvik")
	bucket.Consume()
	cache.RefreshDisk()
	cache.Buckets.Remove("souvik")
	bucket = cache.GetOrCreateBucket("souvik")
	if bucket.UsedTokens != 1 {
		t.Error("Cache eviction failed")
	}
}

func TestCacheMessaging(t *testing.T) {
	cache := NewDeltaCache(10)
	cache.Add("souvik", NewTokenBucket(5, 1))
	cache.Add("bristi", NewTokenBucket(5, 1))
	m := ToDeltaMessage(cache)
	_, ok1 := m["souvik"]
	_, ok2 := m["bristi"]
	if cache.Len() > 0 || !ok1 || !ok2 {
		t.Error("Bucket state wasn't generated")
	}
}
