package util

import (
	"log"
	"sync"
	"time"
)

type (
	Item[T any] struct {
		Value      T
		Mu         sync.Mutex
		accessChan chan bool
	}

	CacheLoader[K comparable, V any] struct {
		items             map[K]*Item[V]
		ExpireAfterAccess time.Duration
		loader            func(K) (V, error)
	}
)

func NewCacheLoader[K comparable, V any](loader func(K) (V, error)) *CacheLoader[K, V] {
	return &CacheLoader[K, V]{
		items:             make(map[K]*Item[V], 10),
		ExpireAfterAccess: time.Second * 2,
		loader:            loader,
	}
}

func (c CacheLoader[K, V]) Get(key K) (v *Item[V], err error) {
	item, ok := c.items[key]
	if !ok {
		item, err = c.loadItemToCache(key)
		if err != nil {
			return nil, err
		}
	}

	item.accessChan <- true
	return item, nil
}

func (c CacheLoader[K, V]) loadItemToCache(key K) (*Item[V], error) {
	value, err := c.loader(key)
	if err != nil {
		return nil, err
	}

	item := &Item[V]{
		Value:      value,
		accessChan: make(chan bool),
	}
	c.items[key] = item

	go func() {
		timer := time.After(c.ExpireAfterAccess)
		for {
			select {
			case <-timer:
				log.Println("Hit lock")
				item.Mu.Lock()
				log.Println("Expired, Delete session")
				if c.items[key] == item {
					delete(c.items, key)
				}
				item.Mu.Unlock()
				return
			case <-item.accessChan:
				log.Println("access he session, reset timer!!")
				timer = time.After(c.ExpireAfterAccess)
			}
		}
	}()

	return item, nil
}
