package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/client/session"
)

var (
	requestStat        = newSessionStat()
	qpsData            = newSessionStat()
	qpsCounterInterval = uint64(1)
)

func newSessionStat() *session.SessionStat {
	stat := &session.SessionStat{}
	stat.AtomicSetStatResponse(netConf().StatResponseTime)

	return stat
}

/*
 * return QPS data
 */
func (this *GorpcService) GetSessionQps(req int, resp *session.SessionStat) error {
	qpsData.AtomicCopyTo(resp)
	return nil
}

/*
 * return all operations after start
 */
func (this *GorpcService) GetSessionTotalOps(req int, resp *session.SessionStat) error {
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

	lastReqStat := newSessionStat()

	for {
		<-ticker.C // wait a moment

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
 * this is used to count open user session response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countPeerRespTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countOpenRespTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddOpenRespTime)
}

/*
 * this is used to count close user session response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countImRespTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countCloseRespTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddCloseRespTime)
}

/*
 * this is used to count query user session response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countImRespTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countQuerySessionRespTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddQuerySessionRespTime)
}

/*
 * this is used to count join chatroom request response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countJoinRespTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countJoinRespTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddJoinRespTime)
}

/*
 * this is used to count quit chatroom request response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countQuitRespTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countQuitRespTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddQuitRespTime)
}

/*
 * this is used to count query chatroom request response time, typically caller should use defer to call this return value
 *  this should be called at first function,  which wants count response time, with defer,likes below:
 *      defer countSendRespTime(owner, traceSn, funcName, appid)()
 * @param owner is owner that used to log
 * @param traceSn is trace sequence number that used to log
 * @param funcName is caller func name that used to log
 * @param appid is application id that used to log
 * @return a function that should be called defer statement
 */
func countQueryRespTime(owner, traceSn, funcName string, appid uint16) func() {
	return countResponseTime(owner, traceSn, funcName, appid, requestStat.AtomicAddQueryRespTime)
}
