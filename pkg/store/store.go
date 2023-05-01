package store

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fnurk/geom/pkg/model"
)

type DbInitHook func(*Datastore) error
type DbPutHook func(t string, id string, value []byte)

type Database interface {
	Init() error
	CreateBucketIfNotExists(bucketName string) error
	Get(bucket string, id string) ([]byte, error)
	//empty id -> autoincrement the ID aka "create new"
	Put(bucket string, id string, data []byte) (string, error)
	Delete(bucket string, id string) error
	Close()
}

type Cache interface {
	Get(bucket string, key string) []byte
	Set(bucket string, key string, val []byte)
	Del(bucket string, key string)
	ClearBucket(bucket string)
}

type IndexType string

const (
	INMEM   = "inmem"
	PERSIST = "persist"
)

type Index struct {
	indexType IndexType
	fieldName string
}

type Datastore struct {
	db        Database
	cache     Cache
	initHooks []DbInitHook
	putHooks  []DbPutHook
	indexMap  map[string][]Index
}

func NewDatastore(db Database, cache Cache) *Datastore {
	return &Datastore{
		db:        db,
		cache:     cache,
		initHooks: []DbInitHook{},
		putHooks:  []DbPutHook{},
		indexMap:  map[string][]Index{},
	}
}

func (ds *Datastore) AddInitHook(hook DbInitHook) {
	ds.initHooks = append(ds.initHooks, hook)
}

func (ds *Datastore) AddPutHook(hook DbPutHook) {
	ds.putHooks = append(ds.putHooks, hook)
}

func (ds *Datastore) Init() error {
	err := ds.db.Init()

	ds.db.CreateBucketIfNotExists("index")

	for k := range model.Types {
		err := ds.db.CreateBucketIfNotExists(k)
		if err != nil {
			return err
		}
	}

	ds.populateIndexTypes()
	ds.populateIndexes()

	for _, ih := range ds.initHooks {
		err := ih(ds)
		if err != nil {
			return err
		}
	}

	return err
}

func (ds *Datastore) CreateBucketIfNotExists(bucketName string) error {
	return ds.db.CreateBucketIfNotExists(bucketName)
}

func (ds *Datastore) Get(bucket string, id string) ([]byte, error) {
	return ds.db.Get(bucket, id)
}

func (ds *Datastore) Put(bucket string, id string, data []byte) (string, error) {
	id, err := ds.db.Put(bucket, id, data)
	if err != nil {
		return "", err
	}

	for _, ph := range ds.putHooks {
		ph(bucket, id, data)
	}

	return id, nil
}

func (ds *Datastore) Delete(bucket string, id string) error {
	return ds.db.Delete(bucket, id)
}

func (ds *Datastore) Close() {
	ds.db.Close()
}

func (ds *Datastore) populateIndexes() {
	for typeName, idxs := range ds.indexMap {
		for _, idx := range idxs {
			if idx.indexType == INMEM {
				prefix := fmt.Sprintf("%s_%s:", typeName, idx.fieldName)
				fmt.Printf("Prefix: %s\n", prefix)
			}
		}

	}

}

func (ds *Datastore) populateIndexTypes() {
	for k, t := range model.Types {
		val := reflect.ValueOf(t)
		for i := 0; i < val.Type().NumField(); i++ {
			t := val.Type().Field(i)

			indexTag := t.Tag.Get("index")
			jsonTag := t.Tag.Get("json")
			if indexTag != "" {
				parts := strings.Split(indexTag, ",")
				fieldName := t.Name
				if jsonTag != "" {
					jsonParts := strings.Split(jsonTag, ",")
					fieldName = jsonParts[0]
				}
				for _, part := range parts {
					if ds.indexMap[k] == nil {
						ds.indexMap[k] = make([]Index, 0)
					}
					idx := Index{
						fieldName: fieldName,
					}
					switch part {
					case INMEM:
						idx.indexType = INMEM
						ds.indexMap[k] = append(ds.indexMap[k], idx)
					case PERSIST:
						idx.indexType = PERSIST
						ds.indexMap[k] = append(ds.indexMap[k], idx)
					}
				}
			}
		}
	}
	fmt.Printf("INDEXES: %+v\n", ds.indexMap)
}
