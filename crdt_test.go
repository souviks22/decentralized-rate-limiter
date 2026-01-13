package drl

import (
	"os"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	dir := "./data"
	os.Mkdir(dir, 0755)
	os.Setenv("CACHE_DIRECTORY", dir)
	limiter := NewRateLimiter(10, 1)
	err := true
	for range 100 {
		if !limiter.AllowRequest("souvik") {
			err = false
			break
		}
	}
	os.Unsetenv("CACHE_DIRECTORY")
	os.RemoveAll(dir)
	if err {
		t.Error("Request wasn't rate limited")
	}
}
