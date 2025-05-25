package limiter

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/souviks22/decentralized-rate-limiter/internal/p2p"
)

type CRDT struct {
	buckets		map[string]*TokenBucket
	deltas		map[string]*TokenBucket
	capacity	float64
	refillRate	float64
	node		*p2p.Node
	mutex		sync.Mutex
}

func newCRDT(capacity float64, refillRate float64) *CRDT {
	p2pNode := p2p.NewNode(context.Background(), "crdt-buckets")
	crdt := CRDT{
		buckets: make(map[string]*TokenBucket),
		deltas: make(map[string]*TokenBucket),
		capacity: capacity,
		refillRate: refillRate,
		node: p2pNode,
	}
	crdt.start()
	return &crdt
}

func (crdt *CRDT) getBucket(userId string) *TokenBucket {
	if crdt.buckets[userId] == nil {
		crdt.buckets[userId] = NewTokenBucket(crdt.capacity, crdt.refillRate)
	}
	return crdt.buckets[userId]
}

func (crdt *CRDT) broadcast(userId string) {
	crdt.deltas[userId] = crdt.buckets[userId]
}

func (crdt *CRDT) serialize() []byte {
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	payload, _ := json.Marshal(crdt.deltas)
	crdt.deltas = make(map[string]*TokenBucket)
	return payload
}

func (crdt *CRDT) deserializeAndMerge(payload []byte) {
	var deltas map[string]*TokenBucket
	err := json.Unmarshal(payload, &deltas)
	if err != nil { 
		return 
	}
	crdt.mutex.Lock()
	defer crdt.mutex.Unlock()
	for userId := range deltas {
		crdt.getBucket(userId).merge(deltas[userId])
	}
}

func (crdt *CRDT) start() {
	crdt.node.ReadLoop(crdt.deserializeAndMerge)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if len(crdt.deltas) == 0 { continue }
			crdt.node.Broadcast(crdt.serialize())
		}
	}()
}