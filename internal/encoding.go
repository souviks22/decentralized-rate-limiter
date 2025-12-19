package internal

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/golang/snappy"
	lru "github.com/hashicorp/golang-lru/v2"
)

type BucketState struct {
	UsedTokens   float64
	UsableTokens float64
	LastRefilled time.Time
	Timestamp    time.Time
}

func ToDeltaMessage(cache *lru.Cache[string, *TokenBucket]) map[string]*BucketState {
	message := make(map[string]*BucketState)
	for _, userId := range cache.Keys() {
		bucket, _ := cache.Get(userId)
		message[userId] = &BucketState{
			UsedTokens:   bucket.UsedTokens,
			UsableTokens: bucket.UsableTokens,
			LastRefilled: bucket.LastRefilled,
			Timestamp:    time.Now(),
		}
	}
	cache.Purge()
	return message
}

func Encode(value any) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return nil
	}
	compressed := snappy.Encode(nil, buf.Bytes())
	return compressed
}

func Decode[V any](data []byte) V {
	decompressed, _ := snappy.Decode(nil, data)
	var value V
	buf := bytes.NewBuffer(decompressed)
	dec := gob.NewDecoder(buf)
	dec.Decode(&value)
	return value
}
