package cpool

import (
	"sync"
)

type ConsumerPool struct {
	fn           func(interface{})
	chanLen      uint64 // 一经初使化后，不可改变
	consumerChan chan interface{}

	sync.RWMutex
	consumerCount        uint          // 一共可以启动多少个goroutine数
	currentConsumerCount uint          // 当前正在运行的goroutine数
	cancel               chan struct{} // 退出时关闭此通道
}

func (cp *ConsumerPool) GetConsumerCount() uint {
	cp.RLock()
	defer cp.RUnlock()
	return cp.consumerCount
}
func (cp *ConsumerPool) GetCurrentConsumeCount() uint {
	cp.RLock()
	defer cp.RUnlock()
	return cp.currentConsumerCount
}

func (cp *ConsumerPool) Add(data interface{}) bool {
	select {
	case cp.consumerChan <- data:
		return true
	default:
		return false
	}
}

func (cp *ConsumerPool) Cancel() {
	select {
	case <-cp.cancel:
	default:
		close(cp.cancel)
	}
}

func NewConsumerPool(consumerCount uint, chanLen uint64, callback func(data interface{})) *ConsumerPool {
	if consumerCount <= 0 || chanLen <= 0 {
		panic("consumerCount and chanLen must > 0")
	}
	if callback == nil {
		panic("callback can't not be nil")
	}

	cp := &ConsumerPool{
		consumerCount:        consumerCount,
		chanLen:              chanLen,
		consumerChan:         make(chan interface{}, chanLen),
		currentConsumerCount: consumerCount,
		fn:                   callback,
		cancel:               make(chan struct{}),
	}

	for i := uint(0); i <= consumerCount; i++ {
		go cp.startConsumer(callback)
	}
	return cp
}

func (cp *ConsumerPool) startConsumer(callback func(data interface{})) {
	for {
		select {
		case d := <-cp.consumerChan:
			callback(d)
		case <-cp.cancel:
			return
		}
	}
}
