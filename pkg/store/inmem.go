package store

import "sync"

type Bucket struct {
	Name  string
	Mutex sync.RWMutex
}

var cache = make(map[string]map[string]interface{})
var mutexes = make(map[string]*sync.RWMutex)

func GetMutex(bucket string) *sync.RWMutex {
	m := mutexes[bucket]
	if m == nil {
		mutexes[bucket] = new(sync.RWMutex)
		m = mutexes[bucket]
	}
	return m
}

func MemGet(bucket string, key string) interface{} {
	m := GetMutex(bucket)
	m.RLock()
	defer m.RUnlock()
	b := cache[bucket]
	if b == nil {
		b = make(map[string]interface{})
		cache[bucket] = b
	}
	return b[key]
}

func ClearMemBucket(bucket string) {
	m := GetMutex(bucket)
	m.Lock()
	defer m.Unlock()
	b := make(map[string]interface{})
	cache[bucket] = b
}

func MemSet(bucket string, key string, val interface{}) {
	m := GetMutex(bucket)
	m.Lock()
	defer m.Unlock()
	b := cache[bucket]
	if b == nil {
		b = make(map[string]interface{})
		cache[bucket] = b
	}
	b[key] = val
}

func MemDel(bucket string, key string) {
	m := GetMutex(bucket)
	m.Lock()
	defer m.Unlock()
	b := cache[bucket]
	if b == nil {
		b = make(map[string]interface{})
		cache[bucket] = b
	}
	delete(b, key)
}
