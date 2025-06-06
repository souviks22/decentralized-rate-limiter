package utils

import (
	"testing"
)

func TestEncoding(t *testing.T) {
	bucket := make(map[string]int)
	bucket["souvik"] = 9
	bucket["bristi"] = 3
	encoded := Encode(bucket)
	decoded := Decode[map[string]int](encoded)
	if decoded["souvik"] != 9 || decoded["bristi"] != 3 {
		t.Error("Encoding failed")
	}
}
