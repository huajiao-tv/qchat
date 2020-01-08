package main

import (
	"sync"
	"time"
)

type Task struct {
	Begin    time.Time
	Freq     map[string]int
	Id       int64
	close    chan bool
	Center   string
	Receiver string
	Roomid   string
	Last     int //持续多少秒
}

var tasks map[int64]*Task = map[int64]*Task{}
var lock sync.RWMutex

func addTask(freq map[string]int, center, roomid, receiver string, last int) *Task {
	t := &Task{
		time.Now(),
		freq,
		time.Now().UnixNano(),
		make(chan bool),
		center,
		receiver,
		roomid,
		last,
	}

	lock.Lock()
	tasks[t.Id] = t
	go func() {
		select {
		case <-t.close:
			return
		default:
		}
		var wg sync.WaitGroup
		wg.Add(len(freq))
		for k, v := range freq {
			go func(k string, v int) {
				ExecNTimes(v, int64(v*last), genSendType(k, center, roomid, receiver))
				wg.Done()
			}(k, v)
		}
		wg.Wait()
		lock.Lock()
		delete(tasks, t.Id)
		lock.Unlock()

	}()
	lock.Unlock()
	return t
}

func getTasks() []*Task {
	t := []*Task{}
	lock.RLock()
	for _, v := range tasks {
		t = append(t, v)
	}
	lock.RUnlock()
	return t

}
