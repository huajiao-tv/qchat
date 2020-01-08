package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type testResult struct {
	MinTime        time.Duration
	MaxTime        time.Duration
	TotalTime      time.Duration
	TotalRequests  int64
	FailedRequests int64
}

func (this *testResult) success(start time.Time) {
	cost := time.Now().Sub(start)
	this.TotalTime += cost
	this.TotalRequests++
	if cost > this.MaxTime {
		this.MaxTime = cost
	} else if cost < this.MinTime {
		this.MinTime = cost
	}
}

func (this *testResult) fail() {
	this.FailedRequests++
	this.TotalRequests++
}

func (this testResult) String() string {
	if this.TotalRequests != 0 {
		return fmt.Sprintf("Request result: min:%fs, max:%fs, average:%fs, requests:%d(failed:%d)",
			this.MinTime.Seconds(), this.MaxTime.Seconds(), this.TotalTime.Seconds()/float64(this.TotalRequests), this.TotalRequests, this.FailedRequests)
	} else {
		return "Error: not run"
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.IntVar(&num, "num", 10000, "test num every goroutinue")
	flag.IntVar(&gonum, "gonum", runtime.NumCPU(), "goroutinue number")
	flag.IntVar(&level, "level", -1, "compress level")
	flag.IntVar(&length, "length", 1024, "data length")
	flag.Parse()
}

var (
	gonum  int
	num    int
	level  int
	length int
	wg     *sync.WaitGroup
)

func main() {

	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	begin := time.Now()

	for i := 0; i < gonum; i++ {
		go func(goid, runnum int) {
			defer wg.Done()

			data := []byte(randString(length))
			res := compress(level, data)
			fmt.Println("ratio: ", float64(len(res))/float64(len(data)))

			result := &testResult{MinTime: 1e9}

			for {
				start := time.Now()
				compress(level, data)
				result.success(start)

				if runnum--; runnum == 0 {
					break
				}
			}

			fmt.Println(result.String())
		}(i, num)
	}

	wg.Wait()

	end := time.Now()

	duration := end.Sub(begin)
	sec := duration.Seconds()
	fmt.Println(float64(gonum*num)/sec, sec)
}

func randString(size int) []byte {
	base := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+-=|[]{};':,./<>?")
	result := make([]byte, size)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		result[i] = base[rand.Intn(len(base))]
	}
	return result
}

func compress(level int, data []byte) []byte {
	var buf bytes.Buffer
	w, _ := gzip.NewWriterLevel(&buf, level)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}
