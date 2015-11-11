package metrik

import (
	"sync"
)

//type Store represents a persistent storage mechanism for the metric data. A simple key-value store would do fine.
//Persistence is not that important, since this data is real-time so it can be recreated easily. But to allow the
//server to scale out easily, we need to be able to safely call the below methods from multiple threads/processes/machines.
type Store interface {
	Initialize(...interface{}) error  //Initialize the store
	Update(Metric, MetricValue) error //Update all points for that metric
	Get(Metric) (MetricValue, error)  //Get all values for that metric
	//GetByTag(Metric, Tags) (MetricValue, error) // Get values with given tags (not )
}

type InMemoryStore struct {
	sync.RWMutex
	store map[string]MetricValue
}

func (i InMemoryStore) Initialize(...interface{}) error {
	return nil
}

func (i InMemoryStore) Update(m Metric, mv MetricValue) error {
	i.Lock()
	i.store[m.Name()] = mv
	i.Unlock()
	return nil
}

func (i InMemoryStore) Get(m Metric) (MetricValue, error) {
	i.RLock()
	defer i.RUnlock()
	return i.store[m.Name()], nil
}

//func (i InMemoryStore) GetByTag(m Metric, t Tags) (MetricValue, error) {
//	return nil, nil
//}
