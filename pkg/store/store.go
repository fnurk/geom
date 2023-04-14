package store

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Document struct {
	ContentType string
	Content     []byte
}

type DocStore interface {
	Get(id int) (*Document, error)
	Create(d Document) (int, error)
	Put(id int, d Document) error
	Delete(id int) error
}

type Store struct {
	db *bolt.DB
}

var store Store

func Init() (DocStore, error) {
	boltdb, err := bolt.Open("test.db", 0666, nil)

	if err != nil {
		return nil, err
	}

	boltdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("docs"))
		if err != nil {
			return err
		}
		return nil
	})

	store := Store{
		db: boltdb,
	}

	return store, nil
}

func Close() {
	store.db.Close()
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func (s Store) Get(id int) (*Document, error) {
	var vCopy []byte
	var result Document

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("docs"))
		v := b.Get(itob(id))

		if v != nil {
			vCopy = make([]byte, len(v))
			copy(vCopy, v)
		}

		return nil
	})

	if vCopy == nil {
		return nil, nil
	}

	err := gob.NewDecoder(bytes.NewReader(vCopy)).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s Store) Create(d Document) (int, error) {
	var idInt int
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(d)

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("docs"))
		if b == nil {
			return fmt.Errorf("No bucket found")
		}

		id, _ := b.NextSequence()
		idInt = int(id)
		fmt.Printf("Buffer: %v\n", buffer.Bytes())
		fmt.Printf("id: %v\n", idInt)

		return b.Put(itob(idInt), buffer.Bytes())
	})
	if err != nil {
		return 0, err
	}
	return idInt, nil
}

func (s Store) Put(id int, d Document) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(d)

	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("docs"))
		err := b.Put(itob(id), buffer.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func (s Store) Delete(id int) error {
	s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("docs"))
		err := b.Delete(itob(id))
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}
