package internal

import "github.com/dgraph-io/badger/v4"

type Disk struct {
	DB *badger.DB
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

func (disk *Disk) Save(userID string, bucket []byte) error {
	return disk.DB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(userID), bucket)
	})
}

func (disk *Disk) Get(userID string) ([]byte, bool) {
	var data []byte
	err := disk.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			data = val
			return err
		})
	})
	return data, err == nil
}
