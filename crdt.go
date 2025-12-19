package drl

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	internal "github.com/souviks22/decentralized-rate-limiter/internal"
)

type CRDT struct {
	Node        *internal.P2PNode
	Buckets     *internal.BucketCache
	Deltas      *lru.Cache[string, *internal.TokenBucket]
	Mutex       sync.RWMutex
	SyncLatency *Recorder
	MessageSize *Recorder
}

const (
	MaxBatchSize      = 100
	BatchInterval     = 100 * time.Millisecond
	DiskWriteInterval = time.Hour
)

func NewRateLimiter(capacity float64, refillRate float64) *CRDT {
	p2pNode := internal.NewP2PNode(context.Background(), "crdt-buckets")
	bucketCache := internal.NewBucketCache(p2pNode.Host.String(), capacity, refillRate)
	syncLatency := NewRecorder("SYNC-LATENCY", "ms", func(v any) float64 {
		return float64(v.(time.Duration).Milliseconds())
	})
	messageSize := NewRecorder("MESSAGE-SIZE", "B", func(v any) float64 {
		return float64(v.(int))
	})
	crdt := CRDT{
		Node:        p2pNode,
		Buckets:     bucketCache,
		Deltas:      internal.NewDeltaCache(MaxBatchSize),
		SyncLatency: syncLatency,
		MessageSize: messageSize,
	}
	crdt.start()
	return &crdt
}

func (crdt *CRDT) AllowRequest(userId string) bool {
	crdt.Mutex.Lock()
	defer crdt.Mutex.Unlock()
	bucket := crdt.Buckets.GetOrCreateBucket(userId)
	crdt.Deltas.ContainsOrAdd(userId, bucket)
	go func() {
		if crdt.Deltas.Len() == MaxBatchSize {
			crdt.broadcast()
		}
	}()
	return bucket.Consume()
}

func (crdt *CRDT) merge(data []byte) {
	message := internal.Decode[map[string]*internal.BucketState](data)
	for userId := range message {
		delay := crdt.Buckets.GetOrCreateBucket(userId).Merge(message[userId])
		crdt.SyncLatency.Record(delay)
	}
}

func (crdt *CRDT) broadcast() {
	crdt.Mutex.Lock()
	defer crdt.Mutex.Unlock()
	data := internal.Encode(internal.ToDeltaMessage(crdt.Deltas))
	crdt.MessageSize.Record(len(data))
	crdt.Node.Broadcast(data)
}

func (crdt *CRDT) start() {
	crdt.Node.ReadLoop(crdt.merge)
	go func() {
		ticker := time.NewTicker(BatchInterval)
		defer ticker.Stop()
		for range ticker.C {
			if crdt.Deltas.Len() > 0 {
				crdt.broadcast()
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(DiskWriteInterval)
		defer ticker.Stop()
		for range ticker.C {
			crdt.Buckets.RefreshDisk()
		}
	}()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			crdt.SyncLatency.LogSnapshot()
			crdt.MessageSize.LogSnapshot()
		}
	}()
}
