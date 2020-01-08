package cpool

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	b := uint32(0)
	cp := NewConsumerPool(100, 100, func(data interface{}) {
		if i, ok := data.(uint32); ok {
			ti := atomic.AddUint32(&b, i)
			t.Log("handle:", i, ti)
		} else {
			t.Error("data is not int type")
		}
	})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				if ok := cp.Add(uint32(1)); ok {
				} else {
					t.Error("add data failed")
				}
				time.Sleep(time.Millisecond)
			}
		}()
	}
	cp.Add(uint32(1))
	time.Sleep(2 * time.Second)
	if b != 10001 {
		t.Error("result not match", b, 10001)
	}
	if cp.GetConsumerCount() != 100 || cp.GetCurrentConsumeCount() != 100 {
		t.Error(cp.GetCurrentConsumeCount(), cp.GetConsumerCount())
	}
}
