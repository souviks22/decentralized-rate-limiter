package drl

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	internal "github.com/souviks22/decentralized-rate-limiter/internal"
)

type CRDT struct {
	Node        *internal.P2PNode                         `json:"node"`
	Buckets     *internal.BucketStore                     `json:"buckets"`
	Delta       *lru.Cache[string, *internal.TokenBucket] `json:"deltas"`
	SyncLatency *Recorder                                 `json:"sync_latency"`
	MessageSize *Recorder                                 `json:"message_size"`
	mutex       sync.Mutex
}

const (
	MaxBatchSize      = 100
	BatchInterval     = 100 * time.Millisecond
	DiskWriteInterval = time.Hour
)

func millisecondScale(v any) float64 {
	return float64(v.(time.Duration).Milliseconds())
}

func byteScale(v any) float64 {
	return float64(v.(int))
}

func NewRateLimiter(capacity float64, refillRate float64) *CRDT {
	node := internal.NewP2PNode(context.Background(), "crdt-buckets")
	buckets := internal.NewBucketStore(node.Host.String(), capacity, refillRate)
	deltas := internal.NewDeltaCache(MaxBatchSize)
	syncLatency := NewRecorder("SYNC-LATENCY", "ms", millisecondScale)
	messageSize := NewRecorder("MESSAGE-SIZE", "B", byteScale)
	crdt := &CRDT{
		Node:        node,
		Buckets:     buckets,
		Delta:       deltas,
		SyncLatency: syncLatency,
		MessageSize: messageSize,
	}
	crdt.start()
	return crdt
}

func (crdt *CRDT) start() {
	go crdt.Node.ReadLoop(crdt.merge)
	go crdt.SendDeltaEvery(BatchInterval)
	go crdt.BackupToDiskEvery(DiskWriteInterval)
	go crdt.LogMetricsEvery(10 * time.Second)
}

func (crdt *CRDT) SendDeltaEvery(batchInterval time.Duration) {
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()
	for range ticker.C {
		if crdt.Delta.Len() > 0 {
			crdt.broadcast()
		}
	}
}

func (crdt *CRDT) BackupToDiskEvery(diskWriteInterval time.Duration) {
	ticker := time.NewTicker(diskWriteInterval)
	defer ticker.Stop()
	for range ticker.C {
		crdt.Buckets.RefreshDisk()
	}
}

func (crdt *CRDT) LogMetricsEvery(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		crdt.SyncLatency.LogSnapshot()
		crdt.MessageSize.LogSnapshot()
	}
}

func (crdt *CRDT) AllowRequest(userId string) bool {
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	bucket := crdt.Buckets.GetOrCreateBucket(userId)
	crdt.Delta.ContainsOrAdd(userId, bucket)
	if crdt.Delta.Len() == MaxBatchSize {
		go crdt.broadcast()
	}
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
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	delta := internal.ToDeltaMessage(crdt.Delta)
	crdt.Delta.Purge()
	data := internal.Encode(delta)
	crdt.MessageSize.Record(len(data))
	crdt.Node.Broadcast(data)
}
