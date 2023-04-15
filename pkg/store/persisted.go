package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/fnurk/geom/pkg/model"
	bolt "go.etcd.io/bbolt"
)

type DbInitHook func(*bolt.DB) error

var store *bolt.DB

var initHooks = make([]DbInitHook, 0)

func AddInitHook(hook DbInitHook) {
	initHooks = append(initHooks, hook)
}

func Init() error {
	db, err := bolt.Open("test.db", 0666, nil)

	if err != nil {
		return err
	}

	store = db

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("index"))
		if err != nil {
			return err
		}
		return nil
	})

	db.Update(func(tx *bolt.Tx) error {
		for k := range model.Types {
			_, err := tx.CreateBucketIfNotExists([]byte(k))
			if err != nil {
				return err
			}
		}
		return nil
	})

	for _, ih := range initHooks {
		err = ih(store)
		if err != nil {
			return err
		}
	}

	return nil
}

func Close() {
	store.Close()
}

func Get(t string, id string) (interface{}, error) {
	var vCopy []byte

	store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t))
		v := b.Get([]byte(id))

		if v != nil {
			vCopy = make([]byte, len(v))
			copy(vCopy, v)
		}

		return nil
	})

	if vCopy == nil {
		return nil, nil
	}

	result, err := model.Decode(t, vCopy)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Put(t string, id string, d interface{}) (*string, error) {
	bytes, err := json.Marshal(d)

	if err != nil {
		return nil, err
	}

	retId := id

	indexes := GetIndexes(t)

	store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t))

		var bid []byte

		if len(id) == 0 { //id empty? autoincrement
			i, _ := b.NextSequence()
			retId = fmt.Sprint(i)
			bid = itob(i)
		} else {
			retId = id
			bid = []byte(id)
		}

		err := b.Put(bid, bytes)
		if err != nil {
			return err
		}

		for _, ix := range indexes {
			ixb := tx.Bucket([]byte("index"))

			val := d.(map[string]interface{})[ix.FieldName]
			ixkey := ix.IndexName + "." + val.(string)
			bixkey := []byte(ixkey)
			err := ixb.Put(bixkey, bid)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return &retId, nil
}

func Delete(t string, id string) error {
	store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t))
		err := b.Delete([]byte(id))
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
