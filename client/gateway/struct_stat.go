package gateway

import (
	"fmt"
	"sync/atomic"
	"time"
)

type GatewayStat struct {
	Reads                 uint64
	Writes                uint64
	OpenConns             uint64
	OpenConnFails         uint64
	CloseConns            uint64
	CloseConnFails        uint64
	OpenConnResponseTime  uint64
	CloseConnResponseTime uint64
	StatResponse          int32
	LastSecondFlow        uint64
	ThisSecondFlow        uint64
}

// operations
/*
 * this copy content of stat to 'copy' atomicly
 * @param copy is target the content copy to
 * @return copy if it is not nil, otherwise will return a new GatewayStat point
 */
func (stat *GatewayStat) AtomicCopyTo(copy *GatewayStat) *GatewayStat {
	if copy == nil {
		copy = &GatewayStat{}
	}

	if stat == copy {
		goto Exit
	}

	atomic.StoreUint64(&copy.Reads, atomic.LoadUint64(&stat.Reads))
	atomic.StoreUint64(&copy.Writes, atomic.LoadUint64(&stat.Writes))

	atomic.StoreUint64(&copy.OpenConns, atomic.LoadUint64(&stat.OpenConns))
	atomic.StoreUint64(&copy.OpenConnFails, atomic.LoadUint64(&stat.OpenConnFails))
	atomic.StoreUint64(&copy.CloseConns, atomic.LoadUint64(&stat.CloseConns))
	atomic.StoreUint64(&copy.CloseConnFails, atomic.LoadUint64(&stat.CloseConnFails))
	atomic.StoreUint64(&copy.OpenConnResponseTime, atomic.LoadUint64(&stat.OpenConnResponseTime))
	atomic.StoreUint64(&copy.CloseConnResponseTime, atomic.LoadUint64(&stat.CloseConnResponseTime))
	atomic.StoreInt32(&copy.StatResponse, atomic.LoadInt32(&stat.StatResponse))

Exit:
	return copy
}

/*
 * stat subtract sub and then store result in stat
 * @param sub is value will be subtracted
 * @return stat with new value
 */
func (stat *GatewayStat) Sub(sub *GatewayStat) *GatewayStat {
	if sub == nil {
		sub = &GatewayStat{}
	}

	stat.Reads -= sub.Reads
	stat.Writes -= sub.Writes

	stat.OpenConns -= sub.OpenConns
	stat.OpenConnFails -= sub.OpenConnFails
	stat.CloseConns -= sub.CloseConns
	stat.CloseConnFails -= sub.CloseConnFails
	stat.OpenConnResponseTime -= sub.OpenConnResponseTime
	stat.CloseConnResponseTime -= sub.CloseConnResponseTime

	return stat
}

/*
 * stat add add and then store result in stat
 * @param add is value will be added
 * @return stat with new value
 */
func (stat *GatewayStat) Add(add *GatewayStat) *GatewayStat {
	if add == nil {
		add = &GatewayStat{}
	}

	stat.Reads += add.Reads
	stat.Writes += add.Writes

	stat.OpenConns += add.OpenConns
	stat.OpenConnFails += add.OpenConnFails
	stat.CloseConns += add.CloseConns
	stat.CloseConnFails += add.CloseConnFails
	stat.OpenConnResponseTime += add.OpenConnResponseTime
	stat.CloseConnResponseTime += add.CloseConnResponseTime

	return stat
}

// format functions
const gatewayStatFormat = "{ \"total open/close connections\" : %v, \"successfully open\" : %v, " +
	"\"failed open\" : %v, " + "\"successfully close\" : %v, \"failed close\" : %v," +
	"\"total open response time\" : %v, \"total close response time\" : %v, \"read number\" : %v, \"write number\" : %v }"

func (stat *GatewayStat) String() string {
	return fmt.Sprintf(gatewayStatFormat,
		stat.AtomicGetTotalRequests(), stat.AtomicGetOpenConns(), stat.AtomicGetOpenConnFails(),
		stat.AtomicGetCloseConns(), stat.AtomicGetCloseConnFails(), stat.AtomicGetOpenConnResponseTime(),
		stat.AtomicGetCloseConnResponseTime(), stat.AtomicGetReads(), stat.AtomicGetWrites())
}

/*
 * this computes qps with stat according to interval
 * @param i is interval which unit is second
 * @return a json string that include saver QPS information
 */
func (stat *GatewayStat) AtomicMakeQps(i uint64) string {
	if i == 0 {
		i = 1
	}

	atomic.StoreUint64(&stat.OpenConns, atomic.LoadUint64(&stat.OpenConns)/i)
	atomic.StoreUint64(&stat.OpenConnFails, atomic.LoadUint64(&stat.OpenConnFails)/i)
	atomic.StoreUint64(&stat.CloseConns, atomic.LoadUint64(&stat.CloseConns)/i)
	atomic.StoreUint64(&stat.CloseConnFails, atomic.LoadUint64(&stat.CloseConnFails)/i)
	atomic.StoreUint64(&stat.OpenConnResponseTime, atomic.LoadUint64(&stat.OpenConnResponseTime)/i)
	atomic.StoreUint64(&stat.CloseConnResponseTime, atomic.LoadUint64(&stat.CloseConnResponseTime)/i)

	return stat.QpsString()
}

const gatewayStatQpsFormat = "{ \"QPS\" : { \"request\" : %v, \"open\" : %v, \"close\" : %v, \"read\" : %v, \"write\" : %v }, " +
	"\"average response time(ms)\" : { \"stat\" : %v, \"open\" : %.3f, \"close\" : %.3f } }"

func (stat *GatewayStat) QpsString() string {
	// compute open connection average response time
	openAverage := float64(stat.AtomicGetOpenConnResponseTime())
	openReqs := stat.AtomicGetOpenConns() + stat.AtomicGetOpenConnFails()
	if openReqs == 0 {
		openAverage = 0
	} else {
		openAverage = openAverage / float64(openReqs) / float64(time.Millisecond)
	}

	// compute close connection request average response time
	closeAverage := float64(stat.AtomicGetCloseConnResponseTime())
	closeReqs := stat.AtomicGetCloseConns() + stat.AtomicGetCloseConnFails()
	if closeReqs == 0 {
		closeAverage = 0
	} else {
		closeAverage = closeAverage / float64(closeReqs) / float64(time.Millisecond)
	}

	return fmt.Sprintf(gatewayStatQpsFormat,
		stat.AtomicGetTotalRequests(), (stat.AtomicGetOpenConns() + stat.AtomicGetOpenConnFails()),
		(stat.AtomicGetCloseConns() + stat.AtomicGetCloseConnFails()), stat.AtomicGetReads(),
		stat.AtomicGetWrites(), stat.AtomicGetStatResponse(), openAverage, closeAverage)
}

// atomic get functions

func (stat *GatewayStat) AtomicGetStatResponse() bool {
	if 0 == atomic.LoadInt32(&stat.StatResponse) {
		return false
	}

	return true
}

func (stat *GatewayStat) AtomicGetReads() uint64 {
	return atomic.LoadUint64(&stat.Reads)
}

func (stat *GatewayStat) AtomicGetWrites() uint64 {
	return atomic.LoadUint64(&stat.Writes)
}

func (stat *GatewayStat) AtomicGetTotalRequests() uint64 {
	return atomic.LoadUint64(&stat.OpenConns) + atomic.LoadUint64(&stat.OpenConnFails) +
		atomic.LoadUint64(&stat.CloseConns) + atomic.LoadUint64(&stat.CloseConnFails)
}

func (stat *GatewayStat) AtomicGetOpenConns() uint64 {
	return atomic.LoadUint64(&stat.OpenConns)
}

func (stat *GatewayStat) AtomicGetOpenConnFails() uint64 {
	return atomic.LoadUint64(&stat.OpenConnFails)
}

func (stat *GatewayStat) AtomicGetCloseConns() uint64 {
	return atomic.LoadUint64(&stat.CloseConns)
}

func (stat *GatewayStat) AtomicGetCloseConnFails() uint64 {
	return atomic.LoadUint64(&stat.CloseConnFails)
}

func (stat *GatewayStat) AtomicGetOpenConnResponseTime() uint64 {
	return atomic.LoadUint64(&stat.OpenConnResponseTime)
}

func (stat *GatewayStat) AtomicGetCloseConnResponseTime() uint64 {
	return atomic.LoadUint64(&stat.CloseConnResponseTime)
}

func (stat *GatewayStat) AtomicGetLastSecondFlow() uint64 {
	return atomic.LoadUint64(&stat.LastSecondFlow)
}

func (stat *GatewayStat) AtomicGetThisSecondFlow() uint64 {
	return atomic.LoadUint64(&stat.ThisSecondFlow)
}

// atomic add functions
func (stat *GatewayStat) AtomicAddReads(reads uint64) uint64 {
	return atomic.AddUint64(&stat.Reads, reads)
}

func (stat *GatewayStat) AtomicAddWrites(writes uint64) uint64 {
	return atomic.AddUint64(&stat.Writes, writes)
}

func (stat *GatewayStat) AtomicAddOpenConns(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenConns, i)
}

func (stat *GatewayStat) AtomicAddOpenConnFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenConnFails, i)
}

func (stat *GatewayStat) AtomicAddCloseConns(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseConns, i)
}

func (stat *GatewayStat) AtomicAddCloseConnFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseConnFails, i)
}

func (stat *GatewayStat) AtomicAddOpenConnResponseTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenConnResponseTime, i)
}

func (stat *GatewayStat) AtomicAddCloseConnResponseTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseConnResponseTime, i)
}

func (stat *GatewayStat) AtomicAddThisSecondFlow(i uint64) uint64 {
	return atomic.AddUint64(&stat.ThisSecondFlow, i)
}

// set functions

func (stat *GatewayStat) AtomicSetStatResponse(value bool) {
	if value {
		atomic.StoreInt32(&stat.StatResponse, 1)
	} else {
		atomic.StoreInt32(&stat.StatResponse, 0)
	}
}

func (stat *GatewayStat) AtomicSetReadWrites(reads, writes uint64) {
	atomic.StoreUint64(&stat.Reads, reads)
	atomic.StoreUint64(&stat.Writes, writes)
}

func (stat *GatewayStat) AtomicSetThisSecondFlow(flow uint64) {
	atomic.StoreUint64(&stat.ThisSecondFlow, flow)
}

func (stat *GatewayStat) AtomicSetLastSecondFlow(flow uint64) {
	atomic.StoreUint64(&stat.LastSecondFlow, flow)
}
