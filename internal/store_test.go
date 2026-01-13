package internal

import (
	"os"
	"testing"
)

func TestBucketStore(t *testing.T) {
	dir, node, userID := "./data", "random", "souvik"
	os.Mkdir(dir, 0755)
	os.Setenv("CACHE_DIRECTORY", dir)
	store := NewBucketStore(node, 10, 1)
	bucket := store.GetOrCreateBucket(userID)
	bucket.Consume()
	store.RefreshDisk()
	store.Cache.Remove(userID)
	bucket = store.GetOrCreateBucket(userID)
	os.Unsetenv("CACHE_DIRECTORY")
	os.RemoveAll(dir)
	if bucket.UsedTokens != 1 {
		t.Error("Cache eviction failed")
	}
}

func TestDeltaCacheConversionToMessage(t *testing.T) {
	delta := NewDeltaCache(10)
	delta.Add("souvik", NewTokenBucket(10, 1))
	delta.Add("bristi", NewTokenBucket(10, 1))
	message := ToDeltaMessage(delta)
	delta.Purge()
	_, ok1 := message["souvik"]
	_, ok2 := message["bristi"]
	if delta.Len() > 0 || !ok1 || !ok2 {
		t.Error("Bucket state wasn't generated")
	}
}
