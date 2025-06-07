package storage

import (
	"testing"

	"github.com/souviks22/decentralized-rate-limiter/internal/utils"
)

func TestStorage(t *testing.T) {
	db := NewDiskDB("../../data")
	bucket := make(map[string]int)
	bucket["souvik"] = 9
	bucket["bristi"] = 3
	encoded := utils.Encode(bucket)
	db.Save("never", encoded)
	data, ok := db.Get("never")
	if !ok || !equal(data, encoded) {
		t.Error("Storage could not save data")
	}
}

func equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
