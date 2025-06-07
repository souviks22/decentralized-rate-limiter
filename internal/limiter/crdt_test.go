package limiter

import "testing"

func TestRateLimiter(t *testing.T) {
	limiter := New(5, 1)
	for range 100 {
		if !limiter.AllowRequest("souvik") {
			return
		}
	}
	t.Error("Request wasn't rate limited")
}
