package limiter

import (
	"context"
	"time"

	"github.com/souviks22/decentralized-rate-limiter/internal/p2p"
)

type CRDT struct {
	buckets    *BucketSketch
	capacity   float64
	refillRate float64
	node       *p2p.Node
}

func newCRDT(capacity float64, refillRate float64) *CRDT {
	p2pNode := p2p.NewNode(context.Background(), "crdt-buckets")
	crdt := CRDT{
		buckets:    newBucketSketch(capacity, refillRate),
		capacity:   capacity,
		refillRate: refillRate,
		node:       p2pNode,
	}
	crdt.start()
	return &crdt
}

func (crdt *CRDT) consume(userId string) bool {
	return crdt.buckets.consume(userId)
}

func (crdt *CRDT) serialize() []byte {
	ds := crdt.buckets.getDeltaSketch()
	payload := ds.encode()
	return payload
}

func (crdt *CRDT) deserialize(payload []byte) {
	ds := decode(payload)
	crdt.buckets.merge(ds)
}

func (crdt *CRDT) start() {
	crdt.node.ReadLoop(crdt.deserialize)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if crdt.buckets.hasDeltaSketch() {
				crdt.node.Broadcast(crdt.serialize())
			}
		}
	}()
}
