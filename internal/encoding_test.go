package internal

import "testing"

func TestEncoding(t *testing.T) {
	key1, key2 := "souvik", "bristi"
	cache := make(map[string]int)
	cache[key1] = 9
	cache[key2] = 3
	encoded := Encode(cache)
	decoded := Decode[map[string]int](encoded)
	if decoded[key1] != 9 || decoded[key2] != 3 {
		t.Error("Encoding failed")
	}
}
