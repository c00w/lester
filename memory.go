package lester

import (
	"log"

	"github.com/boltdb/bolt"
)

type BoltMemory struct {
	db *bolt.DB
}

func NewBoltMemory(path string) (*BoltMemory, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("default"))
		return err
	}); err != nil {
		return nil, err
	}
	return &BoltMemory{db}, nil

}

func (b *BoltMemory) Close() {
	b.db.Close()
}

func (b *BoltMemory) SetValue(key string, value string) {
	if err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))
		return b.Put([]byte(key), []byte(value))
	}); err != nil {
		log.Fatal(err)
	}

}
