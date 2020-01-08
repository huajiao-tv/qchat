package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/client/router"
)

var (
	requestStat        = newRouterStat()
	qpsData            = newRouterStat()
	qpsCounterInterval = uint64(1)
)

func newRouterStat() *router.RouterStat {
	stat := &router.RouterStat{}
	stat.AtomicSetStatResponse(netConf().StatResponseTime)

	return stat
}

/*
 * return QPS data
 */
func (this *GorpcService) GetRouterQps(req int, resp *router.RouterStat) error {
	qpsData.AtomicCopyTo(resp)
	return nil
}

/*
 * return all operations after start
 */
func (this *GorpcService) GetRouterTotalOps(req int, resp *router.RouterStat) error {
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

	lastReqStat := newRouterStat()

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
