package stat

import (
	"sync/atomic"
	"time"
)

type Threshold struct {
	ops uint64
	max uint64
}

func NewThreshold(limit int) *Threshold {
	t := &Threshold{
		max: uint64(limit),
	}
	go t.TimerFunc()
	return t
}

func (t *Threshold) Incr() bool {
	return atomic.AddUint64(&t.ops, 1) < atomic.LoadUint64(&t.max)
}

func (t *Threshold) SetLimit(limit int) {
	atomic.StoreUint64(&t.max, uint64(limit))
}

func (t *Threshold) TimerFunc() {
	ticker := time.NewTicker(time.Second)

	for {
		<- ticker.C

		atomic.StoreUint64(&t.ops, 0)
	}
}
