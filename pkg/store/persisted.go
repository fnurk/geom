package store

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/fnurk/geom/pkg/model"
	"github.com/fnurk/geom/pkg/pubsub"
	bolt "go.etcd.io/bbolt"
)

type DbInitHook func(*bolt.DB) error

var store *bolt.DB
var Changes *pubsub.Pubsub

var initHooks = make([]DbInitHook, 0)

func AddInitHook(hook DbInitHook) {
	initHooks = append(initHooks, hook)
}

func Init() error {
	db, err := bolt.Open("test.db", 0666, nil)

	if err != nil {
		return err
	}

	Changes = pubsub.New(1000)

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

	//DUMP DATABASE TO OTHER FILE FOR VIEWING
	go func(*bolt.DB) {
		for {
			c := time.Tick(2 * time.Second)
			for range c {
				db.View(func(tx *bolt.Tx) error {
					tx.CopyFile("copy.db", 0666)
					return nil
				})
			}
		}
	}(store)

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

func Get(t string, id string) ([]byte, error) {
	var vCopy []byte

	store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t))
		v := b.Get(strtob(id))

		if v != nil {
			vCopy = make([]byte, len(v))
			copy(vCopy, v)
		}

		return nil
	})

	if vCopy == nil {
		return nil, nil
	}

	return vCopy, nil
}

func Put(t string, id uint64, d []byte) (uint64, error) {
	retId := id

	// indexes := GetIndexes(t)

	store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t))

		var bid []byte

		if id == 0 { //id empty? autoincrement
			i, _ := b.NextSequence()
			retId = i
			bid = itob(i)
		} else {
			retId = id
			bid = itob(id)
		}

		err := b.Put(bid, []byte(d))
		if err != nil {
			return err
		}

		// for _, ix := range indexes {
		// 	ixb := tx.Bucket([]byte("index"))
		//
		// 	val := d.(map[string][]byte)[ix.FieldName]
		//
		// 	var ixkey string
		// 	switch val.(type) {
		// 	case string:
		// 		ixkey = ix.IndexName + "." + val.(string)
		// 	case float64:
		// 		ixkey = ix.IndexName + "." + fmt.Sprintf("%f", val.(float64))
		// 	case int:
		// 		ixkey = ix.IndexName + "." + fmt.Sprintf("%d", val.(int))
		// 	case uint64:
		// 		ixkey = ix.IndexName + "." + fmt.Sprintf("%d", val.(uint64))
		// 	}
		//
		// 	bixkey := []byte(ixkey)
		// 	err := ixb.Put(bixkey, bid)
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		return nil
	})

	Changes.Publish(pubsub.Message{
		Topic: fmt.Sprintf("%s.%d", t, retId),
		Body:  string(d),
	})

	return retId, nil
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

func strtob(v string) []byte {
	i, _ := strconv.Atoi(v)
	return itob(uint64(i))
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
