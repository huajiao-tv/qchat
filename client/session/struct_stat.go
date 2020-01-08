package session

import (
	"fmt"
	"sync/atomic"
	"time"
)

type SessionStat struct {
	// user sessions
	OpenSessions      uint64
	OpenSessionFails  uint64
	CloseSessions     uint64
	CloseSessionFails uint64
	QuerySessions     uint64
	QuerySessionFails uint64
	// chatroom request count
	Joins      uint64
	JoinFails  uint64
	Quits      uint64
	QuitFails  uint64
	Querys     uint64
	QueryFails uint64
	// p2p chat response time count
	OpenRespTime         uint64
	CloseRespTime        uint64
	QuerySessionRespTime uint64
	// chatroom response time count
	JoinRespTime  uint64
	QuitRespTime  uint64
	QueryRespTime uint64
	// stat switch state
	StatResponse int32
}

// operations
/*
 * this copy content of stat to 'copy' atomicly
 * @param copy is target the content copy to
 * @return copy if it is not nil, otherwise will return a new GatewayStat point
 */
func (stat *SessionStat) AtomicCopyTo(copy *SessionStat) *SessionStat {
	if copy == nil {
		copy = &SessionStat{}
	}

	if stat == copy {
		goto Exit
	}

	atomic.StoreUint64(&copy.OpenSessions, atomic.LoadUint64(&stat.OpenSessions))
	atomic.StoreUint64(&copy.OpenSessionFails, atomic.LoadUint64(&stat.OpenSessionFails))
	atomic.StoreUint64(&copy.CloseSessions, atomic.LoadUint64(&stat.CloseSessions))
	atomic.StoreUint64(&copy.CloseSessionFails, atomic.LoadUint64(&stat.CloseSessionFails))
	atomic.StoreUint64(&copy.QuerySessions, atomic.LoadUint64(&stat.QuerySessions))
	atomic.StoreUint64(&copy.QuerySessionFails, atomic.LoadUint64(&stat.QuerySessionFails))

	atomic.StoreUint64(&copy.Joins, atomic.LoadUint64(&stat.Joins))
	atomic.StoreUint64(&copy.JoinFails, atomic.LoadUint64(&stat.JoinFails))
	atomic.StoreUint64(&copy.Quits, atomic.LoadUint64(&stat.Quits))
	atomic.StoreUint64(&copy.QuitFails, atomic.LoadUint64(&stat.QuitFails))
	atomic.StoreUint64(&copy.Querys, atomic.LoadUint64(&stat.Querys))
	atomic.StoreUint64(&copy.QueryFails, atomic.LoadUint64(&stat.QueryFails))

	atomic.StoreUint64(&copy.OpenRespTime, atomic.LoadUint64(&stat.OpenRespTime))
	atomic.StoreUint64(&copy.CloseRespTime, atomic.LoadUint64(&stat.CloseRespTime))
	atomic.StoreUint64(&copy.QuerySessionRespTime, atomic.LoadUint64(&stat.QuerySessionRespTime))
	atomic.StoreUint64(&copy.JoinRespTime, atomic.LoadUint64(&stat.JoinRespTime))
	atomic.StoreUint64(&copy.QuitRespTime, atomic.LoadUint64(&stat.QuitRespTime))
	atomic.StoreUint64(&copy.QueryRespTime, atomic.LoadUint64(&stat.QueryRespTime))
	atomic.StoreInt32(&copy.StatResponse, atomic.LoadInt32(&stat.StatResponse))

Exit:
	return copy
}

/*
 * stat subtract sub and then store result in stat
 * @param sub is value will be subtracted
 * @return stat with new value
 */
func (stat *SessionStat) Sub(sub *SessionStat) *SessionStat {
	if sub == nil {
		sub = &SessionStat{}
	}

	stat.OpenSessions -= sub.OpenSessions
	stat.OpenSessionFails -= sub.OpenSessionFails
	stat.CloseSessions -= sub.CloseSessions
	stat.CloseSessionFails -= sub.CloseSessionFails
	stat.QuerySessions -= sub.QuerySessions
	stat.QuerySessionFails -= sub.QuerySessionFails

	stat.Joins -= sub.Joins
	stat.JoinFails -= sub.JoinFails
	stat.Quits -= sub.Quits
	stat.QuitFails -= sub.QuitFails
	stat.Querys -= sub.Querys
	stat.QueryFails -= sub.QueryFails

	stat.OpenRespTime -= sub.OpenRespTime
	stat.CloseRespTime -= sub.CloseRespTime
	stat.QuerySessionRespTime -= sub.QuerySessionRespTime
	stat.JoinRespTime -= sub.JoinRespTime
	stat.QuitRespTime -= sub.QuitRespTime
	stat.QueryRespTime -= sub.QueryRespTime

	return stat
}

/*
 * stat add add and then store result in stat
 * @param add is value will be added
 * @return stat with new value
 */
func (stat *SessionStat) Add(add *SessionStat) *SessionStat {
	if add == nil {
		add = &SessionStat{}
	}

	stat.OpenSessions += add.OpenSessions
	stat.OpenSessionFails += add.OpenSessionFails
	stat.CloseSessions += add.CloseSessions
	stat.CloseSessionFails += add.CloseSessionFails
	stat.QuerySessions += add.QuerySessions
	stat.QuerySessionFails += add.QuerySessionFails
	stat.QuerySessionRespTime += add.QuerySessionRespTime

	stat.Joins += add.Joins
	stat.JoinFails += add.JoinFails
	stat.Quits += add.Quits
	stat.QuitFails += add.QuitFails
	stat.Querys += add.Querys
	stat.QueryFails += add.QueryFails

	stat.OpenRespTime += add.OpenRespTime
	stat.CloseRespTime += add.CloseRespTime
	stat.JoinRespTime += add.JoinRespTime
	stat.QuitRespTime += add.QuitRespTime
	stat.QueryRespTime += add.QueryRespTime

	return stat
}

// format functions
const sessionStatFormat = "{ " +
	"\"open session\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"close session\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"query session\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"join chatroom\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"quit chatroom\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"query chatroom\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"response time stat switch\" : %v" + " }"

func (stat *SessionStat) String() string {
	return fmt.Sprintf(sessionStatFormat,
		stat.AtomicGetOpenSessions(), stat.AtomicGetOpenSessionFails(), stat.AtomicGetOpenRespTime(),
		stat.AtomicGetCloseSessions(), stat.AtomicGetCloseSessionFails(), stat.AtomicGetCloseRespTime(),
		stat.AtomicGetQuerySessions(), stat.AtomicGetQuerySessionFails(), stat.AtomicGetQuerySessionRespTime(),
		stat.AtomicGetJoins(), stat.AtomicGetJoinFails(), stat.AtomicGetJoinRespTime(),
		stat.AtomicGetQuits(), stat.AtomicGetQuitFails(), stat.AtomicGetQuitRespTime(),
		stat.AtomicGetQuerys(), stat.AtomicGetQueryFails(), stat.AtomicGetQueryRespTime(),
		stat.AtomicGetStatResponse())
}

/*
 * this computes qps with stat according to interval
 * @param i is interval which unit is second
 * @return a json string that include saver QPS information
 */
func (stat *SessionStat) AtomicMakeQps(i uint64) string {
	if i == 0 {
		i = 1
	}

	atomic.StoreUint64(&stat.OpenSessions, atomic.LoadUint64(&stat.OpenSessions)/i)
	atomic.StoreUint64(&stat.OpenSessionFails, atomic.LoadUint64(&stat.OpenSessionFails)/i)
	atomic.StoreUint64(&stat.CloseSessions, atomic.LoadUint64(&stat.CloseSessions)/i)
	atomic.StoreUint64(&stat.CloseSessionFails, atomic.LoadUint64(&stat.CloseSessionFails)/i)
	atomic.StoreUint64(&stat.QuerySessions, atomic.LoadUint64(&stat.QuerySessions)/i)
	atomic.StoreUint64(&stat.QuerySessionFails, atomic.LoadUint64(&stat.QuerySessionFails)/i)

	atomic.StoreUint64(&stat.Joins, atomic.LoadUint64(&stat.Joins)/i)
	atomic.StoreUint64(&stat.JoinFails, atomic.LoadUint64(&stat.JoinFails)/i)
	atomic.StoreUint64(&stat.Quits, atomic.LoadUint64(&stat.Quits)/i)
	atomic.StoreUint64(&stat.QuitFails, atomic.LoadUint64(&stat.QuitFails)/i)
	atomic.StoreUint64(&stat.Querys, atomic.LoadUint64(&stat.Querys)/i)
	atomic.StoreUint64(&stat.QueryFails, atomic.LoadUint64(&stat.QueryFails)/i)

	atomic.StoreUint64(&stat.OpenRespTime, atomic.LoadUint64(&stat.OpenRespTime)/i)
	atomic.StoreUint64(&stat.CloseRespTime, atomic.LoadUint64(&stat.CloseRespTime)/i)
	atomic.StoreUint64(&stat.QuerySessionRespTime, atomic.LoadUint64(&stat.QuerySessionRespTime)/i)
	atomic.StoreUint64(&stat.JoinRespTime, atomic.LoadUint64(&stat.JoinRespTime)/i)
	atomic.StoreUint64(&stat.QuitRespTime, atomic.LoadUint64(&stat.QuitRespTime)/i)
	atomic.StoreUint64(&stat.QueryRespTime, atomic.LoadUint64(&stat.QueryRespTime)/i)

	return stat.QpsString()
}

const sessionStatQpsFormat = "{ \"QPS\" : { " +
	"\"user session\" : { \"total\" : %v, \"open\" : %v, \"close\" : %v, \"query\" : %v }, " +
	"\"chatroom\" : { \"total\" : %v, \"join\" : %v, \"quit\" : %v, \"query\" : %v } }, " +
	"\"average response time(ms)\" : { \"stat\" : %v, " +
	"\"user session\" : { \"total\" : %.3f, \"open\" : %.3f, \"close\" : %.3f, \"query\" : %.3f }, " +
	"\"chatroom\" : { \"total\" : %.3f, \"join\" : %.3f, \"quit\" : %.3f, \"query\" : %.3f } } }"

func (stat *SessionStat) QpsString() string {
	// compute open user session  average response time
	openAverage := float64(stat.AtomicGetOpenRespTime())
	openReqs := stat.AtomicGetOpenSessions() + stat.AtomicGetOpenSessionFails()
	if openReqs == 0 {
		openAverage = 0
	} else {
		openAverage = openAverage / float64(openReqs) / float64(time.Millisecond)
	}

	// compute close user session average response time
	closeAverage := float64(stat.AtomicGetCloseRespTime())
	closeReqs := stat.AtomicGetCloseSessions() + stat.AtomicGetCloseSessionFails()
	if closeReqs == 0 {
		closeAverage = 0
	} else {
		closeAverage = closeAverage / float64(closeReqs) / float64(time.Millisecond)
	}

	// compute query user session average response time
	querySessionAverage := float64(stat.AtomicGetQuerySessionRespTime())
	querySessionReqs := stat.AtomicGetQuerySessions() + stat.AtomicGetQuerySessionFails()
	if querySessionReqs == 0 {
		querySessionAverage = 0
	} else {
		querySessionAverage = querySessionAverage / float64(querySessionReqs) / float64(time.Millisecond)
	}

	// coumpte total session request average response time
	totalUserSessionAverage := float64(stat.AtomicGetOpenRespTime()) + float64(stat.AtomicGetCloseRespTime()) +
		float64(stat.AtomicGetQuerySessionRespTime())
	totalUserSessions := openReqs + closeReqs + querySessionReqs
	if totalUserSessions == 0 {
		totalUserSessionAverage = 0
	} else {
		totalUserSessionAverage = totalUserSessionAverage / float64(totalUserSessions) / float64(time.Millisecond)
	}

	// compute join chatroom request average response time
	joinAverage := float64(stat.AtomicGetJoinRespTime())
	joinReqs := stat.AtomicGetJoins() + stat.AtomicGetJoinFails()
	if joinReqs == 0 {
		joinAverage = 0
	} else {
		joinAverage = joinAverage / float64(joinReqs) / float64(time.Millisecond)
	}

	// compute quit chatroom request average response time
	quitAverage := float64(stat.AtomicGetQuitRespTime())
	quitReqs := stat.AtomicGetQuits() + stat.AtomicGetQuitFails()
	if quitReqs == 0 {
		quitAverage = 0
	} else {
		quitAverage = quitAverage / float64(quitReqs) / float64(time.Millisecond)
	}

	// compute query chatroom average response time
	queryAverage := float64(stat.AtomicGetQueryRespTime())
	queryReqs := stat.AtomicGetQuerys() + stat.AtomicGetQueryFails()
	if queryReqs == 0 {
		queryAverage = 0
	} else {
		queryAverage = queryAverage / float64(queryReqs) / float64(time.Millisecond)
	}

	// coumpte total chatroom request average response time
	totalChatroomAverage := float64(stat.AtomicGetOpenRespTime()) + float64(stat.AtomicGetCloseRespTime()) +
		float64(stat.AtomicGetQuerySessionRespTime())
	totalChatrooms := joinReqs + quitReqs + queryReqs
	if totalChatrooms == 0 {
		totalChatroomAverage = 0
	} else {
		totalChatroomAverage = totalChatroomAverage / float64(totalChatrooms) / float64(time.Millisecond)
	}

	return fmt.Sprintf(sessionStatQpsFormat,
		totalUserSessions, openReqs, closeReqs, querySessionReqs,
		totalChatrooms, joinReqs, quitReqs, queryReqs,
		stat.AtomicGetStatResponse(),
		totalUserSessionAverage, openAverage, closeAverage, querySessionAverage,
		totalChatroomAverage, joinAverage, quitAverage, queryAverage)
}

// atomic get functions
// ---------------------------------------------------------------------------------------------------------

// response time stat related

func (stat *SessionStat) AtomicGetStatResponse() bool {
	if 0 == atomic.LoadInt32(&stat.StatResponse) {
		return false
	}

	return true
}

func (stat *SessionStat) AtomicSetStatResponse(value bool) {
	if value {
		atomic.StoreInt32(&stat.StatResponse, 1)
	} else {
		atomic.StoreInt32(&stat.StatResponse, 0)
	}
}

func (stat *SessionStat) AtomicGetOpenRespTime() uint64 {
	return atomic.LoadUint64(&stat.OpenRespTime)
}

func (stat *SessionStat) AtomicGetCloseRespTime() uint64 {
	return atomic.LoadUint64(&stat.CloseRespTime)
}

func (stat *SessionStat) AtomicGetQuerySessionRespTime() uint64 {
	return atomic.LoadUint64(&stat.QuerySessionRespTime)
}

func (stat *SessionStat) AtomicGetJoinRespTime() uint64 {
	return atomic.LoadUint64(&stat.JoinRespTime)
}

func (stat *SessionStat) AtomicGetQuitRespTime() uint64 {
	return atomic.LoadUint64(&stat.QuitRespTime)
}

func (stat *SessionStat) AtomicGetQueryRespTime() uint64 {
	return atomic.LoadUint64(&stat.QueryRespTime)
}

// user session related

func (stat *SessionStat) AtomicGetUserSessionRequests() uint64 {
	return atomic.LoadUint64(&stat.OpenSessions) + atomic.LoadUint64(&stat.OpenSessionFails) +
		atomic.LoadUint64(&stat.CloseSessions) + atomic.LoadUint64(&stat.CloseSessionFails) +
		atomic.LoadUint64(&stat.QuerySessions) + atomic.LoadUint64(&stat.QuerySessionFails)
}

func (stat *SessionStat) AtomicGetOpenSessions() uint64 {
	return atomic.LoadUint64(&stat.OpenSessions)
}

func (stat *SessionStat) AtomicGetOpenSessionFails() uint64 {
	return atomic.LoadUint64(&stat.OpenSessionFails)
}

func (stat *SessionStat) AtomicGetCloseSessions() uint64 {
	return atomic.LoadUint64(&stat.CloseSessions)
}

func (stat *SessionStat) AtomicGetCloseSessionFails() uint64 {
	return atomic.LoadUint64(&stat.CloseSessionFails)
}

func (stat *SessionStat) AtomicGetQuerySessions() uint64 {
	return atomic.LoadUint64(&stat.QuerySessions)
}

func (stat *SessionStat) AtomicGetQuerySessionFails() uint64 {
	return atomic.LoadUint64(&stat.QuerySessionFails)
}

// chatroom related

func (stat *SessionStat) AtomicGetJoins() uint64 {
	return atomic.LoadUint64(&stat.Joins)
}

func (stat *SessionStat) AtomicGetJoinFails() uint64 {
	return atomic.LoadUint64(&stat.JoinFails)
}

func (stat *SessionStat) AtomicGetQuits() uint64 {
	return atomic.LoadUint64(&stat.Quits)
}

func (stat *SessionStat) AtomicGetQuitFails() uint64 {
	return atomic.LoadUint64(&stat.QuitFails)
}

func (stat *SessionStat) AtomicGetQuerys() uint64 {
	return atomic.LoadUint64(&stat.Querys)
}

func (stat *SessionStat) AtomicGetQueryFails() uint64 {
	return atomic.LoadUint64(&stat.QueryFails)
}

// atomic add functions
// ---------------------------------------------------------------------------------------------------------

// response time stat related

func (stat *SessionStat) AtomicAddOpenRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenRespTime, i)
}

func (stat *SessionStat) AtomicAddCloseRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseRespTime, i)
}

func (stat *SessionStat) AtomicAddQuerySessionRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessionRespTime, i)
}

func (stat *SessionStat) AtomicAddJoinRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.JoinRespTime, i)
}

func (stat *SessionStat) AtomicAddQuitRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuitRespTime, i)
}

func (stat *SessionStat) AtomicAddQueryRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.QueryRespTime, i)
}

// user session related

func (stat *SessionStat) AtomicAddOpenSessions(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenSessions, i)
}

func (stat *SessionStat) AtomicAddOpenSessionFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.OpenSessionFails, i)
}

func (stat *SessionStat) AtomicAddCloseSessions(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseSessions, i)
}

func (stat *SessionStat) AtomicAddCloseSessionFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.CloseSessionFails, i)
}

func (stat *SessionStat) AtomicAddQuerySessions(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessions, i)
}

func (stat *SessionStat) AtomicAddQuerySessionFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessionFails, i)
}

// chatroom related

func (stat *SessionStat) AtomicAddJoins(i uint64) uint64 {
	return atomic.AddUint64(&stat.Joins, i)
}

func (stat *SessionStat) AtomicAddJoinFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.JoinFails, i)
}

func (stat *SessionStat) AtomicAddQuits(i uint64) uint64 {
	return atomic.AddUint64(&stat.Quits, i)
}

func (stat *SessionStat) AtomicAddQuitFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuitFails, i)
}

func (stat *SessionStat) AtomicAddQuerys(i uint64) uint64 {
	return atomic.AddUint64(&stat.Querys, i)
}

func (stat *SessionStat) AtomicAddQueryFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.QueryFails, i)
}
