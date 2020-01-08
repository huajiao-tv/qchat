package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/logic"
)

var (
	requestStat        = newGatewayStat()
	qpsData            = newGatewayStat()
	qpsCounterInterval = uint64(1)
)

func newGatewayStat() *gateway.GatewayStat {
	stat := &gateway.GatewayStat{}
	stat.AtomicSetStatResponse(netConf().StatResponseTime)

	return stat
}

/*
 * return QPS data
 */
func (this *GorpcService) GetGatewayQps(req int, resp *gateway.GatewayStat) error {
	qpsData.AtomicCopyTo(resp)
	return nil
}

/*
 * return all operations after start
 */
func (this *GorpcService) GetGatewayTotalOps(req int, resp *gateway.GatewayStat) error {
	if netConf().StatResponseTime != requestStat.AtomicGetStatResponse() {
		requestStat.AtomicSetStatResponse(netConf().StatResponseTime)
	}
	requestStat.AtomicCopyTo(resp)
	return nil
}

/*
 * this will stat saver QPS according to stat interval config
 */
func StatQps() {
	if netConf().QpsCountInterval > 0 {
		atomic.StoreUint64(&qpsCounterInterval, uint64(netConf().QpsCountInterval))
	}

	ticker := time.NewTicker(time.Second * time.Duration(qpsCounterInterval))
	defer func() {
		ticker.Stop()
	}()

	lastReqStat := newGatewayStat()

	for {
		<-ticker.C // wait a moment

		// update read/write data
		requestStat.AtomicSetReadWrites(monitor.MonitorData())

		// get current stat data atomicly
		currentReqStat := requestStat.AtomicCopyTo(nil)
		// update stat switch state
		if netConf().StatResponseTime != currentReqStat.AtomicGetStatResponse() {
			currentReqStat.AtomicSetStatResponse(netConf().StatResponseTime)
		}
		// record last stat data before update it
		sub := lastReqStat.AtomicCopyTo(nil)
		// update last stat data atomicly
		currentReqStat.AtomicCopyTo(lastReqStat)
		// count current stat, not atomic for currentReqStat and sub both are local variables
		currentReqStat.Sub(sub)
		// make qps and update result atomicly
		currentReqStat.AtomicMakeQps(qpsCounterInterval)
		currentReqStat.AtomicCopyTo(qpsData) // note that qps data is computed yet
		// log QPS data to trace log
		Logger.Trace("", "", "", "StatQps", qpsData.QpsString())

		// update tick timer if config is changed
		if uint64(netConf().QpsCountInterval) != qpsCounterInterval && netConf().QpsCountInterval > 0 {
			ticker.Stop() // should stop old ticker explicitly
			Logger.Debug("", "", "", "StatQps",
				fmt.Sprintf("QPS counter interval is changed from %v to %v.",
					qpsCounterInterval, netConf().QpsCountInterval))

			// update QPS counter interval to new value
			atomic.StoreUint64(&qpsCounterInterval, uint64(netConf().QpsCountInterval))

			// make new time ticker
			ticker = time.NewTicker(time.Second * time.Duration(qpsCounterInterval))
		}
	}
}

// 流量统计
func StartCountFlow() {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
	}()

	for {
		// get this second flow
		thisSecondFlow := requestStat.AtomicGetThisSecondFlow()

		// reset this second flow to 0
		requestStat.AtomicSetThisSecondFlow(0)

		// set last second flow
		requestStat.AtomicSetLastSecondFlow(thisSecondFlow)

		<-ticker.C
	}
}

/*
 * this is used to count open connection response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countOpenResponseTime(funcName)()
 * @param funcName is caller func name that used to log
 * @return a function that should be called defer statement
 */
func countOpenResponseTime(funcName string) func() {
	return countResponseTime("", "", funcName, logic.DEFAULT_APPID, requestStat.AtomicAddOpenConnResponseTime)
}

/*
 * this is used to count close connection response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countCloseResponseTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countCloseResponseTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddCloseConnResponseTime)
}

/*
 * this is used to count request response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countResponseTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @param countFunc is correspond count response time function
 * @return a function that should be called defer statement
 */
func countResponseTime(owner, traceSn, funcName string, appid uint16, countFunc func(uint64) uint64) func() {
	start := time.Now()
	return func() {
		elapsed := int64(time.Since(start))
		if elapsed > netConf().ResponseSlowThreshold {
			Logger.Warn(owner, appid, traceSn, funcName, "Slow hanlding!",
				fmt.Sprintf("used %.3f ms for reuqtest", float64(elapsed)/float64(time.Millisecond)))
		}
		countFunc(uint64(elapsed))
	}
}

/*
 * this is used to count operation latency time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countGatewayOperation(timestamp)()
 * @param timestamp is timestamp of chat room message if not zero
 * @param owner is owner information
 * @param appid is application id
 * @param traceId is trace id
 * @param funcName is function name who called this
 * @return a function that should be called defer statement
 */
func countGatewayOperation(timestamp int64, owner, appid, traceId, funcName string) func() {
	start := time.Now()

	return func() {
		elapsed := int64(time.Since(start)) / 1e6
		if elapsed > netConf().OperationSlowThreshold {
			Logger.Warn(owner, appid, traceId, funcName, "Slow hanlding!", fmt.Sprintf("used %v ms for operation.", elapsed))
		}

		if timestamp > 0 {
			gone := time.Now().UnixNano()/1e6 - timestamp
			if gone >= netConf().ChatroomMsgNormalThreshold && gone < netConf().ChatroomMsgSlowThreshold {
				Logger.Error(owner, appid, traceId, funcName, "Slow chatroom message hanlding!", fmt.Sprintf("used %v ms for chatroom sending.", gone))
			} else if gone >= netConf().ChatroomMsgSlowThreshold {
				Logger.Error(owner, appid, traceId, funcName, "Very slow chatroom message hanlding!", fmt.Sprintf("used %v ms for chatroom sending.", gone))
			}
		}
	}
}
