package store

import (
	"encoding/binary"
	"strconv"
	"time"

	"github.com/fnurk/geom/pkg/model"
	bolt "go.etcd.io/bbolt"
)

type BoltDatabase struct {
	DB        *bolt.DB
	InitHooks []DbInitHook
	PutHooks  []DbPutHook
}

func NewBoltDb(filename string) (*BoltDatabase, error) {
	boltdb, err := bolt.Open(filename, 0666, nil)

	if err != nil {
		return nil, err
	}

	return &BoltDatabase{
		DB:        boltdb,
		InitHooks: []DbInitHook{},
		PutHooks:  []DbPutHook{},
	}, nil

}

func (db BoltDatabase) AddInitHook(hook DbInitHook) {
	db.InitHooks = append(db.InitHooks, hook)
}

func (db BoltDatabase) AddPutHook(hook DbPutHook) {
	db.PutHooks = append(db.PutHooks, hook)
}

func (db BoltDatabase) Init() error {

	db.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("index"))
		if err != nil {
			return err
		}
		return nil
	})

	db.DB.Update(func(tx *bolt.Tx) error {
		for k := range model.Types {
			_, err := tx.CreateBucketIfNotExists([]byte(k))
			if err != nil {
				return err
			}
		}
		return nil
	})

	for _, ih := range db.InitHooks {
		err := ih(db)
		if err != nil {
			return err
		}
	}

	//DUMP DATABASE TO OTHER FILE FOR VIEWING
	go func(db *bolt.DB) {
		for {
			c := time.Tick(2 * time.Second)
			for range c {
				db.View(func(tx *bolt.Tx) error {
					tx.CopyFile("copy.db", 0666)
					return nil
				})
			}
		}
	}(db.DB)

	return nil
}

func (db BoltDatabase) Close() {
	db.DB.Close()
}

func (db BoltDatabase) Get(t string, id string) ([]byte, error) {
	var vCopy []byte

	db.DB.View(func(tx *bolt.Tx) error {
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

func (db BoltDatabase) Put(t string, id string, data []byte) (string, error) {
	retId := id

	// indexes := GetIndexes(t)

	db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t))

		var bid []byte

		if id == "" { //id empty? autoincrement
			i, _ := b.NextSequence()
			retId = id
			bid = itob(i)
		} else {
			retId = id
			bid = strtob(id)
		}

		err := b.Put(bid, []byte(data))
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

	for _, ph := range db.PutHooks {
		ph(t, retId, data)
	}

	return retId, nil
}

func (db BoltDatabase) Delete(t string, id string) error {
	db.DB.Update(func(tx *bolt.Tx) error {
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
