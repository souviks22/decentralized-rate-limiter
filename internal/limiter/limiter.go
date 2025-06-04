package limiter

type DecentralizedRateLimiter struct {
	crdt *CRDT
}

func NewRateLimiter(capacity int, refillRate float64) *DecentralizedRateLimiter {
	crdt := newCRDT(float64(capacity), refillRate)
	return &DecentralizedRateLimiter{
		crdt: crdt,
	}
}

func (drt *DecentralizedRateLimiter) AllowRequest(userId string) bool {
	return drt.crdt.consume(userId)
}
