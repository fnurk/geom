package store

import (
	"encoding/binary"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
)

type BoltDatabase struct {
	DB           *bolt.DB
	PeriodicDump bool
}

func NewBoltDb(filename string) (*BoltDatabase, error) {
	boltdb, err := bolt.Open(filename, 0666, nil)

	if err != nil {
		return nil, err
	}

	return &BoltDatabase{
		DB:           boltdb,
		PeriodicDump: true,
	}, nil

}

func (db BoltDatabase) Init() error {

	//DUMP DATABASE TO OTHER FILE FOR VIEWING
	if db.PeriodicDump {
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
	}

	return nil
}

func (db BoltDatabase) CreateBucketIfNotExists(bucketName string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})
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

		return nil
	})

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
