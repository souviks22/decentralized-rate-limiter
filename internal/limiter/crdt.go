package limiter

type CRDT struct {
	buckets		map[string]*TokenBucket
	deltas		map[string]*TokenBucket
	capacity	float64
	refillRate	float64
}

func newCRDT(capacity float64, refillRate float64) *CRDT {
	return &CRDT{
		buckets: make(map[string]*TokenBucket),
		deltas: make(map[string]*TokenBucket),
		capacity: capacity,
		refillRate: refillRate,
	}
}

func (crdt *CRDT) getBucket(userId string) *TokenBucket {
	if crdt.buckets[userId] == nil {
		crdt.buckets[userId] = newTokenBucket(crdt.capacity, crdt.refillRate)
	}
	return crdt.buckets[userId]
}

func (crdt *CRDT) broadcast(userId string) {
	crdt.deltas[userId] = crdt.buckets[userId]
}
