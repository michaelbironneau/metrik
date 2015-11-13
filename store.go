package metrik

import (
	"sync"
)

//type Store represents a storage mechanism for the metric data. Any key-value store will do.
type Store interface {
	Initialize(...interface{}) error //Initialize the store
	Put([]byte, []byte) error        //Put value
	Get([]byte) ([]byte, error)      //Retrieve value
}

type inMemoryStore struct {
	sync.RWMutex
	store map[string][]byte
}

func (i inMemoryStore) Initialize(...interface{}) error {
	return nil
}

func (i inMemoryStore) Put(key []byte, val []byte) error {
	i.Lock()
	i.store[string(key)] = val
	i.Unlock()
	return nil
}

func (i inMemoryStore) Get(key []byte) ([]byte, error) {
	i.RLock()
	defer i.RUnlock()
	return i.store[string(key)], nil
}
