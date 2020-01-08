package arrays

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

const (
	ASC = iota
	DESC
)

// Sort 将data表示的slice 排序。sortFlag表示升序还是降序
func Sort(data interface{}, sortFlag int) {
	switch tmpSlice := data.(type) {
	case []int:
		if sortFlag == ASC {
			sort.IntSlice(tmpSlice).Sort()
		} else {
			slice := intSlice{sort.IntSlice(tmpSlice), true}
			sort.Sort(slice)
		}
	case []float64:
		if sortFlag == ASC {
			sort.Float64Slice(tmpSlice).Sort()
		} else {
			slice := float64Slice{slice: sort.Float64Slice(tmpSlice)}
			sort.Sort(slice)
		}
	case []string:
		if sortFlag == ASC {
			sort.StringSlice(tmpSlice).Sort()
		} else {
			slice := stringSlice{slice: sort.StringSlice(tmpSlice)}
			sort.Sort(slice)
		}
	}
}

var initRand bool

// Shuffle 将data表示的slice中的元素打乱
func Shuffle(data interface{}) {
	if initRand == false {
		rand.Seed(time.Now().UnixNano())
		initRand = true
	}
	switch tmpSlice := data.(type) {
	case []int:
		slice := intSlice{sort.IntSlice(tmpSlice), false}
		sort.Sort(slice)
	case []float64:
		slice := float64Slice{sort.Float64Slice(tmpSlice), false}
		sort.Sort(slice)
	case []string:
		slice := stringSlice{sort.StringSlice(tmpSlice), false}
		sort.Sort(slice)
	}
}

type intSlice struct {
	slice  sort.IntSlice
	isSort bool
}

func (self intSlice) Len() int {
	return len(self.slice)
}

func (self intSlice) Less(i, j int) bool {
	if self.isSort {
		return self.slice[i] > self.slice[j]
	}
	if rand.Intn(2) == 1 {
		return true
	}
	return false
}

func (self intSlice) Swap(i, j int) {
	self.slice[i], self.slice[j] = self.slice[j], self.slice[i]
}

type float64Slice struct {
	slice  sort.Float64Slice
	isSort bool
}

func (self float64Slice) Len() int {
	return len(self.slice)
}

func (self float64Slice) Less(i, j int) bool {
	if self.isSort {
		return self.slice[i] > self.slice[j] || !math.IsNaN(self.slice[i]) && math.IsNaN(self.slice[j])
	}

	if rand.Intn(2) == 1 {
		return true
	}
	return false
}

func (self float64Slice) Swap(i, j int) {
	self.slice[i], self.slice[j] = self.slice[j], self.slice[i]
}

type stringSlice struct {
	slice  sort.StringSlice
	isSort bool
}

func (self stringSlice) Len() int {
	return len(self.slice)
}

func (self stringSlice) Less(i, j int) bool {
	if self.isSort {
		return self.slice[i] > self.slice[j]
	}
	if rand.Intn(2) == 1 {
		return true
	}
	return false
}

func (self stringSlice) Swap(i, j int) {
	self.slice[i], self.slice[j] = self.slice[j], self.slice[i]
}
