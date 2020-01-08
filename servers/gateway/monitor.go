package main

import (
	"fmt"
	"sync/atomic"
)

var monitor = NewMonitor()

type Monitor struct {
	readNum  uint64
	writeNum uint64
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) AddRead(i uint64) {
	atomic.AddUint64(&m.readNum, i)
}

func (m *Monitor) AddWrite(i uint64) {
	atomic.AddUint64(&m.writeNum, i)
}

func (m *Monitor) String() string {
	return fmt.Sprintf("reading = %v writing = %v", atomic.LoadUint64(&m.readNum), atomic.LoadUint64(&m.writeNum))
}

func (m *Monitor) MonitorData() (uint64, uint64) {
	return atomic.LoadUint64(&m.readNum), atomic.LoadUint64(&m.writeNum)
}
