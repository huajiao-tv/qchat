package main

import (
	"sync"

	"github.com/huajiao-tv/qchat/utility/cpool"
)

type MultiConsumerPools struct {
	ConsumerCount uint
	ChanLen       uint64
	data          map[string]*cpool.ConsumerPool
	mutex         *sync.Mutex
	callback      func(interface{})
}

func NewMultiConsumerPools(consumerCount uint, chanLen uint64, callback func(interface{})) *MultiConsumerPools {
	return &MultiConsumerPools{
		ConsumerCount: consumerCount,
		ChanLen:       chanLen,
		data:          make(map[string]*cpool.ConsumerPool),
		mutex:         &sync.Mutex{},
		callback:      callback,
	}
}

func (pools *MultiConsumerPools) NewPool(index string) *cpool.ConsumerPool {
	pool := cpool.NewConsumerPool(pools.ConsumerCount, pools.ChanLen, pools.callback)
	pools.data[index] = pool
	return pool
}

func (pools *MultiConsumerPools) GetPool(index string) *cpool.ConsumerPool {
	pools.mutex.Lock()

	var pool *cpool.ConsumerPool

	if p, ok := pools.data[index]; !ok {
		pool = pools.NewPool(index)
	} else {
		pool = p
	}

	pools.mutex.Unlock()
	return pool
}
