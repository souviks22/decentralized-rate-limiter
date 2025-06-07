package limiter

import "testing"

func TestBucketCache(t *testing.T) {
	cache := newBucketCache("random", 5, 1)
	bucket := cache.getOrCreateBucket("souvik")
	bucket.consume()
	cache.refreshDisk()
	cache.buckets.Remove("souvik")
	bucket = cache.getOrCreateBucket("souvik")
	if bucket.UsedTokens != 1 {
		t.Error("Cache eviction failed")
	}
}

func TestCacheMessaging(t *testing.T) {
	cache := newCache(10)
	cache.Add("souvik", newTokenBucket(5, 1))
	cache.Add("bristi", newTokenBucket(5, 1))
	m := toMessage(cache)
	_, ok1 := m["souvik"]
	_, ok2 := m["bristi"]
	if cache.Len() > 0 || !ok1 || !ok2 {
		t.Error("Bucket state wasn't generated")
	}
}
