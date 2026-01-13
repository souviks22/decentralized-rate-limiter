package internal

import "github.com/dgraph-io/badger/v4"

type Disk struct {
	DB *badger.DB `json:"db"`
}

func NewDiskDB(path string) *Disk {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &Disk{
		DB: db,
	}
}

func (disk *Disk) Save(userID string, bucket *TokenBucket) {
	encodedUserID := []byte(userID)
	encodedBucket := Encode(bucket)
	insertBucket := func(txn *badger.Txn) error {
		return txn.Set(encodedUserID, encodedBucket)
	}
	disk.DB.Update(insertBucket)
}

func (disk *Disk) Get(userID string) (*TokenBucket, bool) {
	encodedUserID := []byte(userID)
	var encodedBucket []byte
	retrieveBucket := func(txn *badger.Txn) error {
		item, err := txn.Get(encodedUserID)
		if err != nil {
			return err
		}
		extractBucket := func(val []byte) error {
			encodedBucket = val
			return err
		}
		return item.Value(extractBucket)
	}
	err := disk.DB.View(retrieveBucket)
	if err != nil {
		return nil, false
	}
	decodedBucket := Decode[*TokenBucket](encodedBucket)
	return decodedBucket, true
}
