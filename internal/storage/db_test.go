package storage

import (
	"log"
	"testing"

	"github.com/souviks22/decentralized-rate-limiter/internal/utils"
)

func TestStorage(t *testing.T){
	db := NewDiskDB("../../data")
	bucket := make(map[string]int)
	bucket["souvik"] = 9
	bucket["bristi"] = 3
	encoded := utils.Encode(bucket)
	db.Save("never", encoded)
	data, ok := db.Get("never")
	if !ok {
		t.Error("Storage could not save data")
	}
	decoded := utils.Decode[map[string]int](data)
	log.Println(decoded)
}