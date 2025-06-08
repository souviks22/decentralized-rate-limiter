package limiter

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/souviks22/decentralized-rate-limiter/internal/metrics"
	"github.com/souviks22/decentralized-rate-limiter/internal/p2p"
	"github.com/souviks22/decentralized-rate-limiter/internal/utils"
)

type CRDT struct {
	node        *p2p.Node
	buckets     *BucketCache
	deltas      *lru.Cache[string, *TokenBucket]
	mutex       sync.Mutex
	syncLatency *metrics.Recorder
	messageSize *metrics.Recorder
}

const (
	MaxBatchSize      = 100
	BatchInterval     = 100 * time.Millisecond
	DiskWriteInterval = time.Hour
)

func New(capacity float64, refillRate float64) *CRDT {
	p2pNode := p2p.NewNode(context.Background(), "crdt-buckets")
	bucketCache := newBucketCache(p2pNode.Host.String(), capacity, refillRate)
	syncLatency := metrics.NewRecorder("SYNC-LATENCY", "ms", func(v any) float64 {
		return float64(v.(time.Duration).Milliseconds())
	})
	messageSize := metrics.NewRecorder("MESSAGE-SIZE", "B", func(v any) float64 {
		return float64(v.(int))
	})
	crdt := CRDT{
		node:        p2pNode,
		buckets:     bucketCache,
		deltas:      newCache(MaxBatchSize),
		syncLatency: syncLatency,
		messageSize: messageSize,
	}
	crdt.start()
	return &crdt
}

func (crdt *CRDT) AllowRequest(userId string) bool {
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	bucket := crdt.buckets.getOrCreateBucket(userId)
	crdt.deltas.ContainsOrAdd(userId, bucket)
	go func() {
		if crdt.deltas.Len() == MaxBatchSize {
			crdt.broadcast()
		}
	}()
	return bucket.consume()
}

func (crdt *CRDT) merge(data []byte) {
	message := utils.Decode[map[string]*BucketState](data)
	for userId := range message {
		delay := crdt.buckets.getOrCreateBucket(userId).merge(message[userId])
		crdt.syncLatency.Record(delay)
	}
}

func (crdt *CRDT) broadcast() {
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	data := utils.Encode(toMessage(crdt.deltas))
	crdt.messageSize.Record(len(data))
	crdt.node.Broadcast(data)
}

func (crdt *CRDT) start() {
	crdt.node.ReadLoop(crdt.merge)
	go func() {
		ticker := time.NewTicker(BatchInterval)
		defer ticker.Stop()
		for range ticker.C {
			if crdt.deltas.Len() > 0 {
				crdt.broadcast()
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(DiskWriteInterval)
		defer ticker.Stop()
		for range ticker.C {
			crdt.buckets.refreshDisk()
		}
	}()
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			crdt.syncLatency.LogSnapshot()
			crdt.messageSize.LogSnapshot()
		}
	}()
}
