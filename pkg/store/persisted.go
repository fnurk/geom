package store

type DbInitHook func(*Datastore) error
type DbPutHook func(t string, id string, value []byte)

type Database interface {
	Init() error
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

type Datastore struct {
	db        Database
	cache     Cache
	initHooks []DbInitHook
	putHooks  []DbPutHook
}

func NewDatastore(db Database, cache Cache) *Datastore {
	return &Datastore{
		db:        db,
		cache:     cache,
		initHooks: []DbInitHook{},
		putHooks:  []DbPutHook{},
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

	for _, ih := range ds.initHooks {
		err := ih(ds)
		if err != nil {
			return err
		}
	}

	return err
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
