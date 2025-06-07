package limiter

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/souviks22/decentralized-rate-limiter/internal/p2p"
	"github.com/souviks22/decentralized-rate-limiter/internal/utils"
)

type CRDT struct {
	node    *p2p.Node
	buckets *BucketCache
	deltas  *lru.Cache[string, *TokenBucket]
	mutex   sync.Mutex
}

const (
	MaxBatchSize      = 100
	BatchInterval     = 100 * time.Millisecond
	DiskWriteInterval = time.Hour
)

func New(capacity float64, refillRate float64) *CRDT {
	p2pNode := p2p.NewNode(context.Background(), "crdt-buckets")
	bucketCache := newBucketCache(p2pNode.Host.String(), capacity, refillRate)
	crdt := CRDT{
		node:    p2pNode,
		buckets: bucketCache,
		deltas:  newCache(MaxBatchSize),
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
		crdt.buckets.getOrCreateBucket(userId).merge(message[userId])
	}
}

func (crdt *CRDT) broadcast() {
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	crdt.node.Broadcast(utils.Encode(toMessage(crdt.deltas)))
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
}
