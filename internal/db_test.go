package internal

import "testing"

func TestDiskStorage(t *testing.T) {
	path, userId := "../demo/data", "souvik"
	db := NewDiskDB(path)
	buckets := make(map[string]*TokenBucket)
	buckets[userId] = NewTokenBucket(10, 1)
	insertedBucket := buckets[userId]
	db.Save(userId, insertedBucket)
	retrievedBucket, ok := db.Get(userId)
	if !ok || !equal(insertedBucket, retrievedBucket) {
		t.Error("Storage could not save data")
	}
}

func equal(tb1 *TokenBucket, tb2 *TokenBucket) bool {
	a, b := Encode(tb1), Encode(tb2)
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
