package limiter

import (
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/souviks22/decentralized-rate-limiter/internal/storage"
	"github.com/souviks22/decentralized-rate-limiter/internal/utils"
)

type BucketCache struct {
	buckets    *lru.Cache[string, *TokenBucket]
	disk       *storage.Disk
	capacity   float64
	refillRate float64
}

func newBucketCache(node string, capacity float64, refillRate float64) *BucketCache {
	disk := storage.NewDiskDB(fmt.Sprintf("/data/%s", node))
	buckets, err := lru.NewWithEvict(1e6, func(userId string, tb *TokenBucket) {
		disk.Save(userId, utils.Encode(tb))
	})
	if err != nil {
		panic(err)
	}
	return &BucketCache{
		buckets:    buckets,
		disk:       disk,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (cache *BucketCache) getOrCreateBucket(userID string) *TokenBucket {
	bucket, ok := cache.buckets.Get(userID)
	if ok {
		return bucket
	}
	data, ok := cache.disk.Get(userID)
	if ok {
		bucket = utils.Decode[*TokenBucket](data)
	} else {
		bucket = newTokenBucket(cache.capacity, cache.refillRate)
		cache.disk.Save(userID, utils.Encode(bucket))
	}
	cache.buckets.Add(userID, bucket)
	return bucket
}

type BucketState struct {
	UsedTokens   float64
	UsableTokens float64
	LastRefilled time.Time
}

func newCache(size int) *lru.Cache[string, *TokenBucket] {
	cache, err := lru.New[string, *TokenBucket](size)
	if err != nil {
		panic(err)
	}
	return cache
}

func toMessage(cache *lru.Cache[string, *TokenBucket]) map[string]*BucketState {
	message := make(map[string]*BucketState)
	for _, userId := range cache.Keys() {
		bucket, _ := cache.Get(userId)
		message[userId] = &BucketState{
			UsedTokens:   bucket.UsedTokens,
			UsableTokens: bucket.UsableTokens,
			LastRefilled: bucket.LastRefilled,
		}
	}
	cache.Purge()
	return message
}
