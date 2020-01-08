package saver

import (
	"fmt"
	"sync/atomic"
	"time"
)

type SaverStat struct {
	OpenSessions               uint64
	OpenSessionFails           uint64
	CloseSessions              uint64
	CloseSessionFails          uint64
	QuerySessions              uint64
	QuerySessionSuccess        uint64
	QuerySessionFails          uint64
	TotalSessionResponseTime   uint64
	StoreIMs                   uint64
	StoreIMFails               uint64
	RetrieveIms                uint64
	RetrieveImFails            uint64
	StorePeers                 uint64
	StorePeerFails             uint64
	RetrievePeers              uint64
	RetrievePeerFails          uint64
	TotalP2pResponseTime       uint64
	RetrievePublics            uint64
	RetrievePublicResponseTime uint64
	StatResponse               int32
}

/*
 * this copy content of stat to 'copy' atomicly
 * @param copy is target the content copy to
 * @return copy if it is not nil, otherwise will return a new SaverStat point
 */
func (stat *SaverStat) AtomicCopyTo(copy *SaverStat) *SaverStat {
	if copy == nil {
		copy = &SaverStat{}
	}

	if stat == copy {
		goto Exit
	}

	atomic.StoreUint64(&copy.OpenSessions, atomic.LoadUint64(&stat.OpenSessions))
	atomic.StoreUint64(&copy.OpenSessionFails, atomic.LoadUint64(&stat.OpenSessionFails))
	atomic.StoreUint64(&copy.CloseSessions, atomic.LoadUint64(&stat.CloseSessions))
	atomic.StoreUint64(&copy.CloseSessionFails, atomic.LoadUint64(&stat.CloseSessionFails))
	atomic.StoreUint64(&copy.QuerySessions, atomic.LoadUint64(&stat.QuerySessions))
	atomic.StoreUint64(&copy.QuerySessionSuccess, atomic.LoadUint64(&stat.QuerySessionSuccess))
	atomic.StoreUint64(&copy.QuerySessionFails, atomic.LoadUint64(&stat.QuerySessionFails))
	atomic.StoreUint64(&copy.StoreIMs, atomic.LoadUint64(&stat.StoreIMs))
	atomic.StoreUint64(&copy.StoreIMFails, atomic.LoadUint64(&stat.StoreIMFails))
	atomic.StoreUint64(&copy.RetrieveIms, atomic.LoadUint64(&stat.RetrieveIms))
	atomic.StoreUint64(&copy.RetrieveImFails, atomic.LoadUint64(&stat.RetrieveImFails))
	atomic.StoreUint64(&copy.StorePeers, atomic.LoadUint64(&stat.StorePeers))
	atomic.StoreUint64(&copy.StorePeerFails, atomic.LoadUint64(&stat.StorePeerFails))
	atomic.StoreUint64(&copy.RetrievePeers, atomic.LoadUint64(&stat.RetrievePeers))
	atomic.StoreUint64(&copy.RetrievePeerFails, atomic.LoadUint64(&stat.RetrievePeerFails))
	atomic.StoreUint64(&copy.RetrievePublics, atomic.LoadUint64(&stat.RetrievePublics))
	atomic.StoreUint64(&copy.RetrievePublicResponseTime, atomic.LoadUint64(&stat.RetrievePublicResponseTime))
	atomic.StoreUint64(&copy.TotalP2pResponseTime, atomic.LoadUint64(&stat.TotalP2pResponseTime))
	atomic.StoreInt32(&copy.StatResponse, atomic.LoadInt32(&stat.StatResponse))

Exit:
	return copy
}

/*
 * stat subtract sub and then store result in stat
 * @param sub is value will be subtracted
 * @return stat with new value
 */
func (stat *SaverStat) Sub(sub *SaverStat) *SaverStat {
	if sub == nil {
		sub = &SaverStat{}
	}

	stat.OpenSessions -= sub.OpenSessions
	stat.OpenSessionFails -= sub.OpenSessionFails
	stat.CloseSessions -= sub.CloseSessions
	stat.CloseSessionFails -= sub.CloseSessionFails
	stat.QuerySessions -= sub.QuerySessions
	stat.QuerySessionSuccess -= sub.QuerySessionSuccess
	stat.QuerySessionFails -= sub.QuerySessionFails
	stat.TotalSessionResponseTime -= sub.TotalSessionResponseTime
	stat.StoreIMs -= sub.StoreIMs
	stat.StoreIMFails -= sub.StoreIMFails
	stat.RetrieveIms -= sub.RetrieveIms
	stat.RetrieveImFails -= sub.RetrieveImFails
	stat.StorePeers -= sub.StorePeers
	stat.StorePeerFails -= sub.StorePeerFails
	stat.RetrievePeers -= sub.RetrievePeers
	stat.RetrievePeerFails -= sub.RetrievePeerFails
	stat.RetrievePublics -= sub.RetrievePublics
	stat.RetrievePublicResponseTime -= sub.RetrievePublicResponseTime
	stat.TotalP2pResponseTime -= sub.TotalP2pResponseTime

	return stat
}

/*
 * stat add add and then store result in stat
 * @param add is value will be added
 * @return stat with new value
 */
func (stat *SaverStat) Add(add *SaverStat) *SaverStat {
	if add == nil {
		add = &SaverStat{}
	}

	stat.OpenSessions += add.OpenSessions
	stat.OpenSessionFails += add.OpenSessionFails
	stat.CloseSessions += add.CloseSessions
	stat.CloseSessionFails += add.CloseSessionFails
	stat.QuerySessions += add.QuerySessions
	stat.QuerySessionSuccess += add.QuerySessionSuccess
	stat.QuerySessionFails += add.QuerySessionFails
	stat.TotalSessionResponseTime += add.TotalSessionResponseTime
	stat.StoreIMs += add.StoreIMs
	stat.StoreIMFails += add.StoreIMFails
	stat.RetrieveIms += add.RetrieveIms
	stat.RetrieveImFails += add.RetrieveImFails
	stat.StorePeers += add.StorePeers
	stat.StorePeerFails += add.StorePeerFails
	stat.RetrievePeers += add.RetrievePeers
	stat.RetrievePeerFails += add.RetrievePeerFails
	stat.RetrievePublics += add.RetrievePublics
	stat.RetrievePublicResponseTime += add.RetrievePublicResponseTime
	stat.TotalP2pResponseTime += add.TotalP2pResponseTime

	return stat
}

const saverStatFormat = "{ \"session requests\" : %v, \"successfully open sessions\" : %v, \"failed open sessions\" : %v, " +
	"\"successfully close sessions\" : %v, \"failed close sessions\" : %v, \"query session requests\" : %v, " +
	"\"successfully query sessions\" : %v, \"failed query sessions\" : %v, \"p2p message requests\" : %v, " +
	"\"successfully store peers\" : %v, \"failed store peers\" : %v, \"successfully store ims\" : %v, " +
	"\"failed store ims\" : %v, \"successfully retrieve peers\" : %v, \"failed retrieve peers\" : %v, " +
	"\"successfully retrieve ims\" : %v, \"failed retrieve ims\" : %v, \"retrieve publics\" : %v }"

func (stat *SaverStat) String() string {
	return fmt.Sprintf(saverStatFormat,
		stat.AtomicTotalSessionRequests(), stat.AtomicGetOpenSessions(), stat.AtomicGetOpenSessionFails(), stat.AtomicGetCloseSessions(),
		stat.AtomicGetCloseSessionFails(), stat.AtomicGetQuerySessions(), stat.AtomicGetQuerySessionSuccess(),
		stat.AtomicGetQuerySessionFails(), stat.AtomicTotalP2PMsgRequests(), stat.AtomicGetStorePeers(), stat.AtomicGetStorePeerFails(),
		stat.AtomicGetStoreIMs(), stat.AtomicGetStoreIMFails(), stat.AtomicGetRetrievePeers(), stat.AtomicGetRetrievePeerFails(),
		stat.AtomicGetRetrieveIms(), stat.AtomicGetRetrieveImFails(), stat.AtomicGetRetrievePublics())
}

/*
 * this computes qps with stat according to interval
 * @param i is interval which unit is second
 * @return a json string that include saver QPS information
 */
func (stat *SaverStat) AtomicMakeQps(i uint64) string {
	if i == 0 {
		i = 1
	}

	atomic.StoreUint64(&stat.OpenSessions, atomic.LoadUint64(&stat.OpenSessions)/i)
	atomic.StoreUint64(&stat.OpenSessionFails, atomic.LoadUint64(&stat.OpenSessionFails)/i)
	atomic.StoreUint64(&stat.CloseSessions, atomic.LoadUint64(&stat.CloseSessions)/i)
	atomic.StoreUint64(&stat.CloseSessionFails, atomic.LoadUint64(&stat.CloseSessionFails)/i)
	atomic.StoreUint64(&stat.QuerySessions, atomic.LoadUint64(&stat.QuerySessions)/i)
	atomic.StoreUint64(&stat.QuerySessionSuccess, atomic.LoadUint64(&stat.QuerySessionSuccess)/i)
	atomic.StoreUint64(&stat.QuerySessionFails, atomic.LoadUint64(&stat.QuerySessionFails)/i)
	atomic.StoreUint64(&stat.StoreIMs, atomic.LoadUint64(&stat.StoreIMs)/i)
	atomic.StoreUint64(&stat.StoreIMFails, atomic.LoadUint64(&stat.StoreIMFails)/i)
	atomic.StoreUint64(&stat.RetrieveIms, atomic.LoadUint64(&stat.RetrieveIms)/i)
	atomic.StoreUint64(&stat.RetrieveImFails, atomic.LoadUint64(&stat.RetrieveImFails)/i)
	atomic.StoreUint64(&stat.StorePeers, atomic.LoadUint64(&stat.StorePeers)/i)
	atomic.StoreUint64(&stat.StorePeerFails, atomic.LoadUint64(&stat.StorePeerFails)/i)
	atomic.StoreUint64(&stat.RetrievePeers, atomic.LoadUint64(&stat.RetrievePeers)/i)
	atomic.StoreUint64(&stat.RetrievePeerFails, atomic.LoadUint64(&stat.RetrievePeerFails)/i)
	atomic.StoreUint64(&stat.RetrievePublics, atomic.LoadUint64(&stat.RetrievePublics)/i)
	atomic.StoreUint64(&stat.RetrievePublicResponseTime, atomic.LoadUint64(&stat.RetrievePublicResponseTime)/i)
	atomic.StoreUint64(&stat.TotalP2pResponseTime, atomic.LoadUint64(&stat.TotalP2pResponseTime)/i)

	return stat.QpsString()
}

const saverStatQpsFormat = "{ \"QPS\" : { \"session request\" : %v, \"open session\" : %v, \"close session\" : %v, " +
	"\"query session\" : %v, \"p2p message\" : %v, \"peers\" : %v, \"ims\" : %v, \"publics\" : %v }, " +
	"\"average response time(ms)\" : { \"stat\" : %v, \"session\" : %.3f, \"p2p message\" : %.3f , \"public message\" : %.3f } }"

func (stat *SaverStat) QpsString() string {
	// compute session average response time
	sessionAverage := float64(stat.AtomicGetTotalSessionResponseTime())
	sessionReqs := stat.AtomicTotalSessionRequests()
	if sessionReqs == 0 {
		sessionAverage = 0
	} else {
		sessionAverage = sessionAverage / float64(sessionReqs) / float64(time.Millisecond)
	}

	// compute p2p message request average response time
	p2pAverage := float64(stat.AtomicGetTotalP2pResponseTime())
	p2pReqs := stat.AtomicTotalP2PMsgRequests()
	if p2pReqs == 0 {
		p2pAverage = 0
	} else {
		p2pAverage = p2pAverage / float64(p2pReqs) / float64(time.Millisecond)
	}

	// compute public message average response time
	publicAverage := float64(stat.AtomicGetRetrievePublicResponseTime())
	retrievePublics := stat.AtomicGetRetrievePublics()
	if retrievePublics == 0 {
		publicAverage = 0
	} else {
		publicAverage = publicAverage / float64(retrievePublics) / float64(time.Millisecond)
	}

	return fmt.Sprintf(saverStatQpsFormat,
		stat.AtomicTotalSessionRequests(), (stat.AtomicGetOpenSessions() + stat.AtomicGetOpenSessionFails()),
		(stat.AtomicGetCloseSessions() + stat.AtomicGetCloseSessionFails()), stat.AtomicGetQuerySessions(),
		stat.AtomicTotalP2PMsgRequests(), stat.AtomicTotalPeerRequests(), stat.AtomicTotalIMRequests(), retrievePublics,
		stat.AtomicGetStatResponse(), sessionAverage, p2pAverage, publicAverage)
}

/*
 * get total session request count atomicly
 */
func (stat *SaverStat) AtomicTotalSessionRequests() uint64 {
	return atomic.LoadUint64(&stat.OpenSessions) + atomic.LoadUint64(&stat.OpenSessionFails) +
		atomic.LoadUint64(&stat.CloseSessions) + atomic.LoadUint64(&stat.CloseSessionFails) +
		atomic.LoadUint64(&stat.QuerySessions)
}

/*
 * get total open session request count atomicly
 */
func (stat *SaverStat) AtomicGetOpenSessions() uint64 {
	return atomic.LoadUint64(&stat.OpenSessions)
}

func (stat *SaverStat) AtomicAddOpenSessions(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenSessions, i)
}

func (stat *SaverStat) AtomicGetOpenSessionFails() uint64 {
	return atomic.LoadUint64(&stat.OpenSessionFails)
}

func (stat *SaverStat) AtomicAddOpenSessionFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenSessionFails, i)
}

func (stat *SaverStat) AtomicGetCloseSessions() uint64 {
	return atomic.LoadUint64(&stat.CloseSessions)
}

func (stat *SaverStat) AtomicAddCloseSessions(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseSessions, i)
}

func (stat *SaverStat) AtomicGetCloseSessionFails() uint64 {
	return atomic.LoadUint64(&stat.CloseSessionFails)
}

func (stat *SaverStat) AtomicAddCloseSessionFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseSessionFails, i)
}

func (stat *SaverStat) AtomicGetQuerySessions() uint64 {
	return atomic.LoadUint64(&stat.QuerySessions)
}

func (stat *SaverStat) AtomicAddQuerySessions(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessions, i)
}

func (stat *SaverStat) AtomicGetQuerySessionSuccess() uint64 {
	return atomic.LoadUint64(&stat.QuerySessionSuccess)
}

func (stat *SaverStat) AtomicAddQuerySessionSuccess(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessionSuccess, i)
}

func (stat *SaverStat) AtomicGetQuerySessionFails() uint64 {
	return atomic.LoadUint64(&stat.QuerySessionFails)
}

func (stat *SaverStat) AtomicAddQuerySessionFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessionFails, i)
}

func (stat *SaverStat) AtomicGetTotalSessionResponseTime() uint64 {
	return atomic.LoadUint64(&stat.TotalSessionResponseTime)
}

/*
 * add a session request response time to total atomicly
 * @param i is session request response time with unit nanosecond
 * @return new total session request response time
 */
func (stat *SaverStat) AtomicAddSessionResponseTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.TotalSessionResponseTime, i)
}

/*
 * get total p2p message request count atomicly
 */
func (stat *SaverStat) AtomicTotalP2PMsgRequests() uint64 {
	return atomic.LoadUint64(&stat.StoreIMs) + atomic.LoadUint64(&stat.StoreIMFails) +
		atomic.LoadUint64(&stat.RetrieveIms) + atomic.LoadUint64(&stat.RetrieveImFails) +
		atomic.LoadUint64(&stat.StorePeers) + atomic.LoadUint64(&stat.StorePeerFails) +
		atomic.LoadUint64(&stat.RetrievePeers) + atomic.LoadUint64(&stat.RetrievePeerFails)
}

/*
 * get total IM request count atomicly
 */
func (stat *SaverStat) AtomicTotalIMRequests() uint64 {
	return atomic.LoadUint64(&stat.StoreIMs) + atomic.LoadUint64(&stat.StoreIMFails) +
		atomic.LoadUint64(&stat.RetrieveIms) + atomic.LoadUint64(&stat.RetrieveImFails)
}

func (stat *SaverStat) AtomicGetStoreIMs() uint64 {
	return atomic.LoadUint64(&stat.StoreIMs)
}

func (stat *SaverStat) AtomicAddStoreIMs(i uint64) uint64 {
	return atomic.AddUint64(&stat.StoreIMs, i)
}

func (stat *SaverStat) AtomicGetStoreIMFails() uint64 {
	return atomic.LoadUint64(&stat.StoreIMFails)
}

func (stat *SaverStat) AtomicAddStoreIMFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.StoreIMFails, i)
}

func (stat *SaverStat) AtomicGetRetrieveIms() uint64 {
	return atomic.LoadUint64(&stat.RetrieveIms)
}

func (stat *SaverStat) AtomicAddRetrieveIms(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrieveIms, i)
}

func (stat *SaverStat) AtomicGetRetrieveImFails() uint64 {
	return atomic.LoadUint64(&stat.RetrieveImFails)
}

func (stat *SaverStat) AtomicAddRetrieveImFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrieveImFails, i)
}

/*
 * get total peer request count atomicly
 */
func (stat *SaverStat) AtomicTotalPeerRequests() uint64 {
	return atomic.LoadUint64(&stat.StorePeers) + atomic.LoadUint64(&stat.StorePeerFails) +
		atomic.LoadUint64(&stat.RetrievePeers) + atomic.LoadUint64(&stat.RetrievePeerFails)
}

func (stat *SaverStat) AtomicGetStorePeers() uint64 {
	return atomic.LoadUint64(&stat.StorePeers)
}

func (stat *SaverStat) AtomicAddStorePeers(i uint64) uint64 {
	return atomic.AddUint64(&stat.StorePeers, i)
}

func (stat *SaverStat) AtomicGetStorePeerFails() uint64 {
	return atomic.LoadUint64(&stat.StorePeerFails)
}

func (stat *SaverStat) AtomicAddStorePeerFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.StorePeerFails, i)
}

func (stat *SaverStat) AtomicGetRetrievePeers() uint64 {
	return atomic.LoadUint64(&stat.RetrievePeers)
}

func (stat *SaverStat) AtomicAddRetrievePeers(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrievePeers, i)
}

func (stat *SaverStat) AtomicGetRetrievePeerFails() uint64 {
	return atomic.LoadUint64(&stat.RetrievePeerFails)
}

func (stat *SaverStat) AtomicAddRetrievePeerFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrievePeerFails, i)
}

func (stat *SaverStat) AtomicGetRetrievePublics() uint64 {
	return atomic.LoadUint64(&stat.RetrievePublics)
}

func (stat *SaverStat) AtomicAddRetrievePublics(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrievePublics, i)
}

func (stat *SaverStat) AtomicGetRetrievePublicResponseTime() uint64 {
	return atomic.LoadUint64(&stat.RetrievePublicResponseTime)
}

func (stat *SaverStat) AtomicAddRetrievePublicResponseTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrievePublicResponseTime, i)
}

func (stat *SaverStat) AtomicGetTotalP2pResponseTime() uint64 {
	return atomic.LoadUint64(&stat.TotalP2pResponseTime)
}

/*
 * add a p2p message request response time to total atomicly
 * @param i is p2p message request response time with unit nanosecond
 * @return new total p2p message request response time
 */
func (stat *SaverStat) AtomicAddP2pResponseTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.TotalP2pResponseTime, i)
}

func (stat *SaverStat) AtomicGetStatResponse() bool {
	if 0 == atomic.LoadInt32(&stat.StatResponse) {
		return false
	}

	return true
}

func (stat *SaverStat) AtomicSetStatResponse(value bool) {
	if value {
		atomic.StoreInt32(&stat.StatResponse, 1)
	} else {
		atomic.StoreInt32(&stat.StatResponse, 0)
	}
}
