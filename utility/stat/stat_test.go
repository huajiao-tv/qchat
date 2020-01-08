package stat

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestInc(t *testing.T) {
	s := NewStat(1)
	for i := 0; i < 10; i++ {
		s.Incr("test")
	}
	time.Sleep(time.Duration(2 * time.Second))
	s.RLock()
	fmt.Println("reqs", s.GetReqs(), "qps", s.GetQps())
	s.RUnlock()
}

func TestRadomInc(t *testing.T) {
	s := NewStat(1)
	close := make(chan bool)
	go func(close chan bool) {
		for i := 0; i < 1000; i++ {
			s.Incr("test")
			time.Sleep(time.Duration(time.Duration(rand.Intn(100)) * time.Millisecond))
		}
		close <- true
	}(close)
	ticker := time.NewTicker(time.Duration(1) * time.Second)

Loop:
	for {
		select {
		case <-close:
			ticker.Stop()
			break Loop
		case <-ticker.C:
			start := time.Now()
			s.RLock()
			fmt.Println("reqs", s.GetReqs(), "qps", s.GetQps())
			s.RUnlock()
			fmt.Println("lock time", time.Now().Sub(start))
		}
	}
}
