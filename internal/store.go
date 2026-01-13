package internal

import (
	"fmt"
	"os"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type BucketStore struct {
	Cache      *lru.Cache[string, *TokenBucket] `json:"cache"`
	Disk       *Disk                            `json:"disk"`
	Capacity   float64                          `json:"capacity"`
	RefillRate float64                          `json:"refill_rate"`
}

func NewBucketStore(node string, capacity float64, refillRate float64) *BucketStore {
	dir := os.Getenv("CACHE_DIRECTORY")
	disk := NewDiskDB(fmt.Sprintf("%s/%s", dir, node))
	cache, err := lru.NewWithEvict(1e6, disk.Save)
	if err != nil {
		panic(err)
	}
	return &BucketStore{
		Cache:      cache,
		Disk:       disk,
		Capacity:   capacity,
		RefillRate: refillRate,
	}
}

func (store *BucketStore) GetOrCreateBucket(userID string) *TokenBucket {
	bucket, ok := store.Cache.Get(userID)
	if ok {
		return bucket
	}
	bucket, ok = store.Disk.Get(userID)
	if !ok {
		bucket = NewTokenBucket(store.Capacity, store.RefillRate)
		go store.Disk.Save(userID, bucket)
	}
	store.Cache.Add(userID, bucket)
	return bucket
}

func (store *BucketStore) RefreshDisk() {
	for _, userId := range store.Cache.Keys() {
		bucket, _ := store.Cache.Get(userId)
		store.Disk.Save(userId, bucket)
	}
}

type BucketState struct {
	UsedTokens   float64   `json:"used_tokens"`
	UsableTokens float64   `json:"usable_tokens"`
	LastRefilled time.Time `json:"last_refilled"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewDeltaCache(size int) *lru.Cache[string, *TokenBucket] {
	cache, err := lru.New[string, *TokenBucket](size)
	if err != nil {
		panic(err)
	}
	return cache
}

func ToDeltaMessage(delta *lru.Cache[string, *TokenBucket]) map[string]*BucketState {
	message := make(map[string]*BucketState)
	currentTimestamp := time.Now()
	for _, userId := range delta.Keys() {
		bucket, _ := delta.Get(userId)
		message[userId] = &BucketState{
			UsedTokens:   bucket.UsedTokens,
			UsableTokens: bucket.UsableTokens,
			LastRefilled: bucket.LastRefilled,
			Timestamp:    currentTimestamp,
		}
	}
	return message
}
