package router

import (
	"fmt"
	"sync/atomic"
)

type RouterStat struct {
	MltiInfos    uint64
	StatResponse int32
}

// operations
/*
 * this copy content of stat to 'copy' atomicly
 * @param copy is target the content copy to
 * @return copy if it is not nil, otherwise will return a new RouterStat point
 */
func (stat *RouterStat) AtomicCopyTo(copy *RouterStat) *RouterStat {
	if copy == nil {
		copy = &RouterStat{}
	}

	if stat == copy {
		goto Exit
	}

	atomic.StoreUint64(&copy.MltiInfos, atomic.LoadUint64(&stat.MltiInfos))
	atomic.StoreInt32(&copy.StatResponse, atomic.LoadInt32(&stat.StatResponse))

Exit:
	return copy
}

/*
 * stat subtract sub and then store result in stat
 * @param sub is value will be subtracted
 * @return stat with new value
 */
func (stat *RouterStat) Sub(sub *RouterStat) *RouterStat {
	if sub == nil {
		sub = &RouterStat{}
	}

	stat.MltiInfos -= sub.MltiInfos

	return stat
}

/*
 * stat add add and then store result in stat
 * @param add is value will be added
 * @return stat with new value
 */
func (stat *RouterStat) Add(add *RouterStat) *RouterStat {
	if add == nil {
		add = &RouterStat{}
	}

	stat.MltiInfos += add.MltiInfos

	return stat
}

// format functions
const routerStatFormat = "{ \"get multiinfo\" : %v }"

func (stat *RouterStat) String() string {
	return fmt.Sprintf(routerStatFormat,
		stat.AtomicGetMltiInfos())
}

/*
 * this computes qps with stat according to interval
 * @param i is interval which unit is second
 * @return a json string that include saver QPS information
 */
func (stat *RouterStat) AtomicMakeQps(i uint64) string {
	if i == 0 {
		i = 1
	}

	atomic.StoreUint64(&stat.MltiInfos, atomic.LoadUint64(&stat.MltiInfos)/i)

	return stat.QpsString()
}

const routerStatQpsFormat = "{ \"QPS\" : { \"get multiinfo\" : %v }}"

func (stat *RouterStat) QpsString() string {
	return fmt.Sprintf(routerStatQpsFormat,
		stat.AtomicGetMltiInfos())
}

// atomic get functions

func (stat *RouterStat) AtomicGetStatResponse() bool {
	if 0 == atomic.LoadInt32(&stat.StatResponse) {
		return false
	}

	return true
}

func (stat *RouterStat) AtomicGetMltiInfos() uint64 {
	return atomic.LoadUint64(&stat.MltiInfos)
}

// atomic add functions
func (stat *RouterStat) AtomicAddMltiInfos() uint64 {
	return atomic.AddUint64(&stat.MltiInfos, 1)
}

// set functions

func (stat *RouterStat) AtomicSetStatResponse(value bool) {
	if value {
		atomic.StoreInt32(&stat.StatResponse, 1)
	} else {
		atomic.StoreInt32(&stat.StatResponse, 0)
	}
}
