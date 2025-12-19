package internal

import (
	"fmt"
	"os"

	lru "github.com/hashicorp/golang-lru/v2"
)

var dir = os.Getenv("CACHE_DIRECTORY")

type BucketCache struct {
	Buckets    *lru.Cache[string, *TokenBucket]
	Disk       *Disk
	Capacity   float64
	RefillRate float64
}

func NewBucketCache(node string, capacity float64, refillRate float64) *BucketCache {
	disk := NewDiskDB(fmt.Sprintf("%s/%s", dir, node))
	buckets, err := lru.NewWithEvict(1e6, func(userId string, tb *TokenBucket) {
		disk.Save(userId, Encode(tb))
	})
	if err != nil {
		panic(err)
	}
	return &BucketCache{
		Buckets:    buckets,
		Disk:       disk,
		Capacity:   capacity,
		RefillRate: refillRate,
	}
}

func (cache *BucketCache) GetOrCreateBucket(userID string) *TokenBucket {
	bucket, ok := cache.Buckets.Get(userID)
	if ok {
		return bucket
	}
	data, ok := cache.Disk.Get(userID)
	if ok {
		bucket = Decode[*TokenBucket](data)
	} else {
		bucket = NewTokenBucket(cache.Capacity, cache.RefillRate)
		cache.Disk.Save(userID, Encode(bucket))
	}
	cache.Buckets.Add(userID, bucket)
	return bucket
}

func (cache *BucketCache) RefreshDisk() {
	for _, userId := range cache.Buckets.Keys() {
		bucket, _ := cache.Buckets.Get(userId)
		cache.Disk.Save(userId, Encode(bucket))
	}
}

func NewDeltaCache(size int) *lru.Cache[string, *TokenBucket] {
	cache, err := lru.New[string, *TokenBucket](size)
	if err != nil {
		panic(err)
	}
	return cache
}
