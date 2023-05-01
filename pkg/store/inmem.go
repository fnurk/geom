package store

import "sync"

type InMemKV struct {
	buckets map[string]map[string][]byte
	mutexes map[string]*sync.RWMutex
}

type Bucket struct {
	Name  string
	Mutex sync.RWMutex
}

func NewInMemKV() *InMemKV {
	return &InMemKV{
		buckets: map[string]map[string][]byte{},
		mutexes: map[string]*sync.RWMutex{},
	}
}

func (kv *InMemKV) getMutex(bucket string) *sync.RWMutex {
	m := kv.mutexes[bucket]
	if m == nil {
		kv.mutexes[bucket] = new(sync.RWMutex)
		m = kv.mutexes[bucket]
	}
	return m
}

func (kv *InMemKV) Get(bucket string, key string) []byte {
	m := kv.getMutex(bucket)
	m.RLock()
	defer m.RUnlock()
	b := kv.buckets[bucket]
	if b == nil {
		b = make(map[string][]byte)
		kv.buckets[bucket] = b
	}
	return b[key]
}

func (kv *InMemKV) ClearBucket(bucket string) {
	m := kv.getMutex(bucket)
	m.Lock()
	defer m.Unlock()
	b := make(map[string][]byte)
	kv.buckets[bucket] = b
}

func (kv *InMemKV) Set(bucket string, key string, val []byte) {
	m := kv.getMutex(bucket)
	m.Lock()
	defer m.Unlock()
	b := kv.buckets[bucket]
	if b == nil {
		b = make(map[string][]byte)
		kv.buckets[bucket] = b
	}
	b[key] = val
}

func (kv *InMemKV) Del(bucket string, key string) {
	m := kv.getMutex(bucket)
	m.Lock()
	defer m.Unlock()
	b := kv.buckets[bucket]
	if b == nil {
		b = make(map[string][]byte)
		kv.buckets[bucket] = b
	}
	delete(b, key)
}
