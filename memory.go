package lester

import (
	"log"
	"strings"

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

type BoltPair struct {
	Key   string
	Value string
}

func (b *BoltMemory) GetPrefix(prefix string) []BoltPair {
	out := make([]BoltPair, 0)
	if err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))
		return b.ForEach(func(k, v []byte) error {
			if strings.HasPrefix(string(k), prefix) {
				out = append(out, BoltPair{string(k), string(v)})
			}
			return nil
		})
	}); err != nil {
		log.Fatal(err)
	}
	return out
}
