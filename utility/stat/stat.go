package stat

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	DefaultStatCap = 100
)

type Stat struct {
	acc uint32

	last  time.Time
	cache map[string]uint64

	l    sync.RWMutex
	reqs map[string]uint64

	qps unsafe.Pointer
}

func NewStat(interval int) *Stat {
	s := &Stat{
		cache: make(map[string]uint64, DefaultStatCap),
		reqs:  make(map[string]uint64, DefaultStatCap),
		qps:   unsafe.Pointer(&map[string]float64{}),
		acc:   uint32(interval),
	}
	go s.CalFn(uint32(interval))
	return s
}

func (s *Stat) Incr(field string) {
	s.l.Lock()
	s.reqs[field] = s.reqs[field] + 1
	s.l.Unlock()
}

func (s *Stat) SetInterval(interval int) {
	atomic.StoreUint32(&s.acc, uint32(interval))
}

func (s *Stat) GetQps() map[string]float64 {
	qps := atomic.LoadPointer(&s.qps)
	return *(*map[string]float64)(qps)
}

func (s *Stat) CalFn(interval uint32) {
	s.last = time.Now()
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		<-ticker.C

		s.l.Lock()
		now := time.Now()
		d := now.Sub(s.last).Seconds()
		qps := make(map[string]float64, len(s.reqs))

		for item, ops := range s.reqs {
			qps[item] = float64(ops-s.cache[item]) / d
			s.cache[item] = ops
		}
		s.l.Unlock()
		s.last = now
		atomic.StorePointer(&s.qps, unsafe.Pointer(&qps))

		if new := atomic.LoadUint32(&s.acc); new != interval {
			interval = new
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(interval) * time.Second)
		}
	}
}
