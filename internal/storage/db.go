package storage

import "github.com/dgraph-io/badger/v4"

type Disk struct {
	db *badger.DB
}

func NewDiskDB(path string) *Disk {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &Disk{
		db: db,
	}
}

func (disk *Disk) Save(userID string, bucket []byte) error {
	return disk.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(userID), bucket)
	})
}

func (disk *Disk) Get(userID string) ([]byte, bool) {
	var data []byte
	err := disk.db.View(func(txn *badger.Txn) error {
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
