package center

import (
	"fmt"
	"sync/atomic"
	"time"
)

type CenterStat struct {
	// p2p chat request count
	PublicHots     uint64
	PublicHotFails uint64
	Peers          uint64
	PeerFails      uint64
	IMs            uint64
	IMFails        uint64
	Retrieves      uint64
	RetrieveFails  uint64
	Recalls        uint64
	RecallFails    uint64
	// chatroom request count
	Joins                uint64
	JoinFails            uint64
	Quits                uint64
	QuitFails            uint64
	Sends                uint64
	SendFails            uint64
	Broadcast            uint64
	BroadcastFails       uint64
	OnlineBroadcast      uint64
	OnlineBroadcastFails uint64
	QueryMemberCounts    uint64
	QueryMemberDetails   uint64
	QueryUserInRoom      uint64
	QuerySessionInRoom   uint64
	// p2p chat response time count
	PeerRespTime uint64
	ImRespTime   uint64
	// chatroom response time count
	JoinRespTime uint64
	QuitRespTime uint64
	SendRespTime uint64
	// stat switch state
	StatResponse int32
	// chat group count
	GroupJoins            uint64
	GroupJoinFails        uint64
	GroupCreates          uint64
	GroupCreateFails      uint64
	GroupQuits            uint64
	GroupQuitFails        uint64
	GroupDismiss          uint64
	GroupDismissFails     uint64
	GroupSends            uint64
	GroupSendFails        uint64
	GroupUserLists        uint64
	GroupUserListFails    uint64
	GroupIsMemebers       uint64
	GroupIsMemebersFails  uint64
	GroupInGroups         uint64
	GroupInGroupsFails    uint64
	GroupInfos            uint64
	GroupInfoFails        uint64
	GroupCreatedList      uint64
	GroupCreatedListFails uint64
	GroupMessageLists     uint64
	GroupMessageListFails uint64
	GroupJoinCounts       uint64
	GroupJoinCountFails   uint64
}

// operations
/*
 * this copy content of stat to 'copy' atomicly
 * @param copy is target the content copy to
 * @return copy if it is not nil, otherwise will return a new GatewayStat point
 */
func (stat *CenterStat) AtomicCopyTo(copy *CenterStat) *CenterStat {
	if copy == nil {
		copy = &CenterStat{}
	}

	if stat == copy {
		goto Exit
	}

	atomic.StoreUint64(&copy.PublicHots, atomic.LoadUint64(&stat.PublicHots))
	atomic.StoreUint64(&copy.PublicHotFails, atomic.LoadUint64(&stat.PublicHotFails))
	atomic.StoreUint64(&copy.Peers, atomic.LoadUint64(&stat.Peers))
	atomic.StoreUint64(&copy.PeerFails, atomic.LoadUint64(&stat.PeerFails))
	atomic.StoreUint64(&copy.IMs, atomic.LoadUint64(&stat.IMs))
	atomic.StoreUint64(&copy.IMFails, atomic.LoadUint64(&stat.IMFails))
	atomic.StoreUint64(&copy.Retrieves, atomic.LoadUint64(&stat.Retrieves))
	atomic.StoreUint64(&copy.RetrieveFails, atomic.LoadUint64(&stat.RetrieveFails))
	atomic.StoreUint64(&copy.Recalls, atomic.LoadUint64(&stat.Recalls))
	atomic.StoreUint64(&copy.RecallFails, atomic.LoadUint64(&stat.RecallFails))

	atomic.StoreUint64(&copy.Joins, atomic.LoadUint64(&stat.Joins))
	atomic.StoreUint64(&copy.JoinFails, atomic.LoadUint64(&stat.JoinFails))
	atomic.StoreUint64(&copy.Quits, atomic.LoadUint64(&stat.Quits))
	atomic.StoreUint64(&copy.QuitFails, atomic.LoadUint64(&stat.QuitFails))
	atomic.StoreUint64(&copy.Sends, atomic.LoadUint64(&stat.Sends))
	atomic.StoreUint64(&copy.SendFails, atomic.LoadUint64(&stat.SendFails))
	atomic.StoreUint64(&copy.Broadcast, atomic.LoadUint64(&stat.Broadcast))
	atomic.StoreUint64(&copy.BroadcastFails, atomic.LoadUint64(&stat.BroadcastFails))
	atomic.StoreUint64(&copy.OnlineBroadcast, atomic.LoadUint64(&stat.OnlineBroadcast))
	atomic.StoreUint64(&copy.OnlineBroadcastFails, atomic.LoadUint64(&stat.OnlineBroadcastFails))
	atomic.StoreUint64(&copy.QueryMemberCounts, atomic.LoadUint64(&stat.QueryMemberCounts))
	atomic.StoreUint64(&copy.QueryMemberDetails, atomic.LoadUint64(&stat.QueryMemberDetails))
	atomic.StoreUint64(&copy.QueryUserInRoom, atomic.LoadUint64(&stat.QueryUserInRoom))
	atomic.StoreUint64(&copy.QuerySessionInRoom, atomic.LoadUint64(&stat.QuerySessionInRoom))

	atomic.StoreUint64(&copy.PeerRespTime, atomic.LoadUint64(&stat.PeerRespTime))
	atomic.StoreUint64(&copy.ImRespTime, atomic.LoadUint64(&stat.ImRespTime))
	atomic.StoreUint64(&copy.JoinRespTime, atomic.LoadUint64(&stat.JoinRespTime))
	atomic.StoreUint64(&copy.QuitRespTime, atomic.LoadUint64(&stat.QuitRespTime))
	atomic.StoreUint64(&copy.SendRespTime, atomic.LoadUint64(&stat.SendRespTime))
	atomic.StoreInt32(&copy.StatResponse, atomic.LoadInt32(&stat.StatResponse))

	//chat group
	atomic.StoreUint64(&copy.GroupJoins, atomic.LoadUint64(&stat.GroupJoins))
	atomic.StoreUint64(&copy.GroupJoinFails, atomic.LoadUint64(&stat.GroupJoinFails))
	atomic.StoreUint64(&copy.GroupCreates, atomic.LoadUint64(&stat.GroupCreates))
	atomic.StoreUint64(&copy.GroupCreateFails, atomic.LoadUint64(&stat.GroupCreateFails))
	atomic.StoreUint64(&copy.GroupQuits, atomic.LoadUint64(&stat.GroupQuits))
	atomic.StoreUint64(&copy.GroupQuitFails, atomic.LoadUint64(&stat.GroupQuitFails))
	atomic.StoreUint64(&copy.GroupDismiss, atomic.LoadUint64(&stat.GroupDismiss))
	atomic.StoreUint64(&copy.GroupDismissFails, atomic.LoadUint64(&stat.GroupDismissFails))
	atomic.StoreUint64(&copy.GroupSends, atomic.LoadUint64(&stat.GroupSends))
	atomic.StoreUint64(&copy.GroupSendFails, atomic.LoadUint64(&stat.GroupSendFails))
	atomic.StoreUint64(&copy.GroupUserLists, atomic.LoadUint64(&stat.GroupUserLists))
	atomic.StoreUint64(&copy.GroupUserListFails, atomic.LoadUint64(&stat.GroupUserListFails))
	atomic.StoreUint64(&copy.GroupInfos, atomic.LoadUint64(&stat.GroupInfos))
	atomic.StoreUint64(&copy.GroupInfoFails, atomic.LoadUint64(&stat.GroupInfoFails))
	atomic.StoreUint64(&copy.GroupIsMemebers, atomic.LoadUint64(&stat.GroupIsMemebers))
	atomic.StoreUint64(&copy.GroupIsMemebersFails, atomic.LoadUint64(&stat.GroupIsMemebersFails))
	atomic.StoreUint64(&copy.GroupInGroups, atomic.LoadUint64(&stat.GroupInGroups))
	atomic.StoreUint64(&copy.GroupInGroupsFails, atomic.LoadUint64(&stat.GroupInGroupsFails))
	atomic.StoreUint64(&copy.GroupJoinCounts, atomic.LoadUint64(&stat.GroupJoinCounts))
	atomic.StoreUint64(&copy.GroupJoinCountFails, atomic.LoadUint64(&stat.GroupJoinCountFails))
	atomic.StoreUint64(&copy.GroupMessageLists, atomic.LoadUint64(&stat.GroupMessageLists))
	atomic.StoreUint64(&copy.GroupMessageListFails, atomic.LoadUint64(&stat.GroupMessageListFails))
	atomic.StoreUint64(&copy.GroupCreatedList, atomic.LoadUint64(&stat.GroupCreatedList))
	atomic.StoreUint64(&copy.GroupCreatedListFails, atomic.LoadUint64(&stat.GroupCreatedListFails))

Exit:
	return copy
}

/*
 * stat subtract sub and then store result in stat
 * @param sub is value will be subtracted
 * @return stat with new value
 */
func (stat *CenterStat) Sub(sub *CenterStat) *CenterStat {
	if sub == nil {
		sub = &CenterStat{}
	}

	stat.PublicHots -= sub.PublicHots
	stat.PublicHotFails -= sub.PublicHotFails
	stat.Peers -= sub.Peers
	stat.PeerFails -= sub.PeerFails
	stat.IMs -= sub.IMs
	stat.IMFails -= sub.IMFails
	stat.Retrieves -= sub.Retrieves
	stat.RetrieveFails -= sub.RetrieveFails
	stat.Recalls -= sub.Recalls
	stat.RecallFails -= sub.RecallFails

	stat.Joins -= sub.Joins
	stat.JoinFails -= sub.JoinFails
	stat.Quits -= sub.Quits
	stat.QuitFails -= sub.QuitFails
	stat.Sends -= sub.Sends
	stat.SendFails -= sub.SendFails
	stat.Broadcast -= sub.Broadcast
	stat.BroadcastFails -= sub.BroadcastFails
	stat.OnlineBroadcast -= sub.OnlineBroadcast
	stat.OnlineBroadcastFails -= sub.OnlineBroadcastFails
	stat.QueryMemberCounts -= sub.QueryMemberCounts
	stat.QueryMemberDetails -= sub.QueryMemberDetails
	stat.QueryUserInRoom -= sub.QueryUserInRoom
	stat.QuerySessionInRoom -= sub.QuerySessionInRoom

	stat.PeerRespTime -= sub.PeerRespTime
	stat.ImRespTime -= sub.ImRespTime
	stat.JoinRespTime -= sub.JoinRespTime
	stat.QuitRespTime -= sub.QuitRespTime
	stat.SendRespTime -= sub.SendRespTime

	// chat group
	stat.GroupJoins -= sub.GroupJoins
	stat.GroupJoinFails -= sub.GroupJoinFails
	stat.GroupQuits -= sub.GroupQuits
	stat.GroupQuitFails -= sub.GroupQuitFails
	stat.GroupSends -= sub.GroupSends
	stat.GroupSendFails -= sub.GroupSendFails
	stat.GroupCreates -= sub.GroupCreates
	stat.GroupCreateFails -= sub.GroupCreateFails
	stat.GroupDismiss -= sub.GroupDismiss
	stat.GroupDismissFails -= sub.GroupDismissFails
	stat.GroupInfos -= sub.GroupInfos
	stat.GroupInfoFails -= sub.GroupInfoFails
	stat.GroupCreatedList -= sub.GroupCreatedList
	stat.GroupCreatedListFails -= sub.GroupCreatedListFails
	stat.GroupUserLists -= sub.GroupUserLists
	stat.GroupUserListFails -= sub.GroupUserListFails
	stat.GroupIsMemebers -= sub.GroupIsMemebers
	stat.GroupIsMemebersFails -= sub.GroupIsMemebersFails
	stat.GroupInGroups -= sub.GroupInGroups
	stat.GroupInGroupsFails -= sub.GroupInGroupsFails
	stat.GroupJoinCounts -= sub.GroupJoinCounts
	stat.GroupJoinCountFails -= sub.GroupJoinCountFails
	stat.GroupMessageLists -= sub.GroupMessageLists
	stat.GroupMessageListFails -= sub.GroupMessageListFails

	return stat
}

/*
 * stat add add and then store result in stat
 * @param add is value will be added
 * @return stat with new value
 */
func (stat *CenterStat) Add(add *CenterStat) *CenterStat {
	if add == nil {
		add = &CenterStat{}
	}

	stat.PublicHots += add.PublicHots
	stat.PublicHotFails += add.PublicHotFails
	stat.Peers += add.Peers
	stat.PeerFails += add.PeerFails
	stat.IMs += add.IMs
	stat.IMFails += add.IMFails
	stat.Retrieves += add.Retrieves
	stat.RetrieveFails += add.RetrieveFails
	stat.Recalls += add.Recalls
	stat.RecallFails += add.RecallFails

	stat.Joins += add.Joins
	stat.JoinFails += add.JoinFails
	stat.Quits += add.Quits
	stat.QuitFails += add.QuitFails
	stat.Sends += add.Sends
	stat.SendFails += add.SendFails
	stat.Broadcast += add.Broadcast
	stat.BroadcastFails += add.BroadcastFails
	stat.OnlineBroadcast += add.OnlineBroadcast
	stat.OnlineBroadcastFails += stat.OnlineBroadcastFails
	stat.QueryMemberCounts += add.QueryMemberCounts
	stat.QueryMemberDetails += add.QueryMemberDetails
	stat.QueryUserInRoom += add.QueryUserInRoom
	stat.QuerySessionInRoom += add.QuerySessionInRoom

	stat.PeerRespTime += add.PeerRespTime
	stat.ImRespTime += add.ImRespTime
	stat.JoinRespTime += add.JoinRespTime
	stat.QuitRespTime += add.QuitRespTime
	stat.SendRespTime += add.SendRespTime

	//chat group
	stat.GroupJoins += add.GroupJoins
	stat.GroupJoinFails += add.GroupJoinFails
	stat.GroupQuits += add.GroupQuits
	stat.GroupQuitFails += add.GroupQuitFails
	stat.GroupSends += add.GroupSends
	stat.GroupSendFails += add.GroupSendFails
	stat.GroupCreates += add.GroupCreates
	stat.GroupCreateFails += add.GroupCreateFails
	stat.GroupDismiss += add.GroupDismiss
	stat.GroupDismissFails += add.GroupDismissFails
	stat.GroupInfos += add.GroupInfos
	stat.GroupInfoFails += add.GroupInfoFails
	stat.GroupCreatedList += add.GroupCreatedList
	stat.GroupCreatedListFails += add.GroupCreatedListFails
	stat.GroupUserLists += add.GroupUserLists
	stat.GroupUserListFails += add.GroupUserListFails
	stat.GroupIsMemebers += add.GroupIsMemebers
	stat.GroupIsMemebersFails += add.GroupIsMemebersFails
	stat.GroupInGroups += add.GroupInGroups
	stat.GroupInGroupsFails += add.GroupInGroupsFails
	stat.GroupJoinCounts += add.GroupJoinCounts
	stat.GroupJoinCountFails += add.GroupJoinCountFails
	stat.GroupMessageLists += add.GroupMessageLists
	stat.GroupMessageListFails += add.GroupMessageListFails

	return stat
}

// format functions
const centerStatFormat = "{ " +
	"\"im\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"peer\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"public/hots\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"retrieve\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"recall\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"robot join chatroom\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"robot quit chatroom\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"send chatroom message\" : { \"success\" : %v, \"failed\" : %v, \"total response time\" : %v }, " +
	"\"chatroom broadcast\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"online broadcast\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"query chatroom member count\" : %v, \"query chatroom member detail\" : %v, " +
	"\"query user in room\" : %v, \"query user session in room\" : %v, " +
	"\"response time stat switch\" : %v," +
	"\"GroupJoins\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupCreates\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupQuits\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupDismiss\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupUserLists\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupIsMemebers\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupInGroups\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupInfos\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupCreatedList\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupJoinCounts\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupSends\" : { \"success\" : %v, \"failed\" : %v }, " +
	"\"GroupMessageLists\" : { \"success\" : %v, \"failed\" : %v }" +
	" }"

func (stat *CenterStat) String() string {
	return fmt.Sprintf(centerStatFormat,
		stat.AtomicGetIms(), stat.AtomicGetImFails(), stat.AtomicGetImRespTime(),
		stat.AtomicGetPeers(), stat.AtomicGetPeerFails(), stat.AtomicGetPeerRespTime(),
		stat.AtomicGetPublicHots(), stat.AtomicGetPublicHotFails(),
		stat.AtomicGetRetrieves(), stat.AtomicGetRetrieveFails(),
		stat.AtomicGetRecalls(), stat.AtomicGetRecallFails(),
		stat.AtomicGetJoins(), stat.AtomicGetJoinFails(), stat.AtomicGetJoinRespTime(),
		stat.AtomicGetQuits(), stat.AtomicGetQuitFails(), stat.AtomicGetQuitRespTime(),
		stat.AtomicGetSends(), stat.AtomicGetSendFails(), stat.AtomicGetSendRespTime(),
		stat.AtomicGetBroadcasts(), stat.AtomicGetBroadcastFails(),
		stat.AtomicGetOnlineBroadcasts(), stat.AtomicGetBroadcastFails(),
		stat.AtomicGetQueryMemberCounts(), stat.AtomicGetQueryMemberDetails(),
		stat.AtomicGetQueryUserInRoom(), stat.AtomicGetQuerySessionInRoom(),
		stat.AtomicGetStatResponse(),
		stat.AtomicGetGroupJoins(), stat.AtomicGetGroupJoinsFails(),
		stat.AtomicGetGroupCreates(), stat.AtomicGetGroupCreateFails(),
		stat.AtomicGetGroupQuits(), stat.AtomicGetGroupQuitFails(),
		stat.AtomicGetGroupDismiss(), stat.AtomicGetGroupDismissFails(),
		stat.AtomicGetGroupUserLists(), stat.AtomicGetGroupUserListFails(),
		stat.AtomicGetGroupIsMemebers(), stat.AtomicGetGroupIsMemebersFails(),
		stat.AtomicGetGroupInGroups(), stat.AtomicGetGroupInGroupsFails(),
		stat.AtomicGetGroupInfos(), stat.AtomicGetGroupInfoFails(),
		stat.AtomicGetGroupCreatedList(), stat.AtomicGetGroupCreatedListFails(),
		stat.AtomicGetGroupJoinCounts(), stat.AtomicGetGroupJoinCountFails(),
		stat.AtomicGetGroupSends(), stat.AtomicGetGroupSendFails(),
		stat.AtomicGetGroupMessageLists(), stat.AtomicGetGroupMessageListFails())
}

/*
 * this computes qps with stat according to interval
 * @param i is interval which unit is second
 * @return a json string that include saver QPS information
 */
func (stat *CenterStat) AtomicMakeQps(i uint64) string {
	if i == 0 {
		i = 1
	}

	atomic.StoreUint64(&stat.PublicHots, atomic.LoadUint64(&stat.PublicHots)/i)
	atomic.StoreUint64(&stat.PublicHotFails, atomic.LoadUint64(&stat.PublicHotFails)/i)
	atomic.StoreUint64(&stat.Peers, atomic.LoadUint64(&stat.Peers)/i)
	atomic.StoreUint64(&stat.PeerFails, atomic.LoadUint64(&stat.PeerFails)/i)
	atomic.StoreUint64(&stat.IMs, atomic.LoadUint64(&stat.IMs)/i)
	atomic.StoreUint64(&stat.IMFails, atomic.LoadUint64(&stat.IMFails)/i)
	atomic.StoreUint64(&stat.Retrieves, atomic.LoadUint64(&stat.Retrieves)/i)
	atomic.StoreUint64(&stat.RetrieveFails, atomic.LoadUint64(&stat.RetrieveFails)/i)
	atomic.StoreUint64(&stat.Recalls, atomic.LoadUint64(&stat.Recalls)/i)
	atomic.StoreUint64(&stat.RecallFails, atomic.LoadUint64(&stat.RecallFails)/i)

	atomic.StoreUint64(&stat.Joins, atomic.LoadUint64(&stat.Joins)/i)
	atomic.StoreUint64(&stat.JoinFails, atomic.LoadUint64(&stat.JoinFails)/i)
	atomic.StoreUint64(&stat.Quits, atomic.LoadUint64(&stat.Quits)/i)
	atomic.StoreUint64(&stat.QuitFails, atomic.LoadUint64(&stat.QuitFails)/i)
	atomic.StoreUint64(&stat.Sends, atomic.LoadUint64(&stat.Sends)/i)
	atomic.StoreUint64(&stat.SendFails, atomic.LoadUint64(&stat.SendFails)/i)
	atomic.StoreUint64(&stat.Broadcast, atomic.LoadUint64(&stat.Broadcast)/i)
	atomic.StoreUint64(&stat.BroadcastFails, atomic.LoadUint64(&stat.BroadcastFails)/i)
	atomic.StoreUint64(&stat.OnlineBroadcast, atomic.LoadUint64(&stat.OnlineBroadcast)/i)
	atomic.StoreUint64(&stat.OnlineBroadcastFails, atomic.LoadUint64(&stat.OnlineBroadcastFails)/i)
	atomic.StoreUint64(&stat.QueryMemberCounts, atomic.LoadUint64(&stat.QueryMemberCounts)/i)
	atomic.StoreUint64(&stat.QueryMemberDetails, atomic.LoadUint64(&stat.QueryMemberDetails)/i)
	atomic.StoreUint64(&stat.QueryUserInRoom, atomic.LoadUint64(&stat.QueryUserInRoom)/i)
	atomic.StoreUint64(&stat.QuerySessionInRoom, atomic.LoadUint64(&stat.QuerySessionInRoom)/i)

	atomic.StoreUint64(&stat.PeerRespTime, atomic.LoadUint64(&stat.PeerRespTime)/i)
	atomic.StoreUint64(&stat.ImRespTime, atomic.LoadUint64(&stat.ImRespTime)/i)
	atomic.StoreUint64(&stat.JoinRespTime, atomic.LoadUint64(&stat.JoinRespTime)/i)
	atomic.StoreUint64(&stat.QuitRespTime, atomic.LoadUint64(&stat.QuitRespTime)/i)
	atomic.StoreUint64(&stat.SendRespTime, atomic.LoadUint64(&stat.SendRespTime)/i)

	//Group relate
	atomic.StoreUint64(&stat.GroupJoins, atomic.LoadUint64(&stat.GroupJoins)/i)
	atomic.StoreUint64(&stat.GroupJoinFails, atomic.LoadUint64(&stat.GroupJoinFails)/i)
	atomic.StoreUint64(&stat.GroupCreates, atomic.LoadUint64(&stat.GroupCreates)/i)
	atomic.StoreUint64(&stat.GroupCreateFails, atomic.LoadUint64(&stat.GroupCreateFails)/i)
	atomic.StoreUint64(&stat.GroupQuits, atomic.LoadUint64(&stat.GroupQuits)/i)
	atomic.StoreUint64(&stat.GroupQuitFails, atomic.LoadUint64(&stat.GroupQuitFails)/i)
	atomic.StoreUint64(&stat.GroupDismiss, atomic.LoadUint64(&stat.GroupDismiss)/i)
	atomic.StoreUint64(&stat.GroupDismissFails, atomic.LoadUint64(&stat.GroupDismissFails)/i)
	atomic.StoreUint64(&stat.GroupSends, atomic.LoadUint64(&stat.GroupSends)/i)
	atomic.StoreUint64(&stat.GroupSendFails, atomic.LoadUint64(&stat.GroupSendFails)/i)
	atomic.StoreUint64(&stat.GroupUserLists, atomic.LoadUint64(&stat.GroupUserLists)/i)
	atomic.StoreUint64(&stat.GroupUserListFails, atomic.LoadUint64(&stat.GroupUserListFails)/i)
	atomic.StoreUint64(&stat.GroupIsMemebers, atomic.LoadUint64(&stat.GroupIsMemebers)/i)
	atomic.StoreUint64(&stat.GroupIsMemebersFails, atomic.LoadUint64(&stat.GroupIsMemebersFails)/i)
	atomic.StoreUint64(&stat.GroupInGroups, atomic.LoadUint64(&stat.GroupInGroups)/i)
	atomic.StoreUint64(&stat.GroupInGroupsFails, atomic.LoadUint64(&stat.GroupInGroupsFails)/i)
	atomic.StoreUint64(&stat.GroupInfos, atomic.LoadUint64(&stat.GroupInfos)/i)
	atomic.StoreUint64(&stat.GroupInfoFails, atomic.LoadUint64(&stat.GroupInfoFails)/i)
	atomic.StoreUint64(&stat.GroupCreatedList, atomic.LoadUint64(&stat.GroupCreatedList)/i)
	atomic.StoreUint64(&stat.GroupCreatedListFails, atomic.LoadUint64(&stat.GroupCreatedListFails)/i)
	atomic.StoreUint64(&stat.GroupJoinCounts, atomic.LoadUint64(&stat.GroupJoinCounts)/i)
	atomic.StoreUint64(&stat.GroupJoinCountFails, atomic.LoadUint64(&stat.GroupJoinCountFails)/i)
	atomic.StoreUint64(&stat.GroupMessageLists, atomic.LoadUint64(&stat.GroupMessageLists)/i)
	atomic.StoreUint64(&stat.GroupMessageListFails, atomic.LoadUint64(&stat.GroupMessageListFails)/i)

	return stat.QpsString()
}

const centerStatQpsFormat = "{ \"QPS\" : { " +
	"\"im\" : %v, \"peer\" : %v, \"public/hot\" : %v, \"retrieve\" : %v, \"recall\" : %v, " +
	"\"robot join chatroom\" : %v, \"robot quit chatroom\" : %v, \"send chatroom message\" : %v, \"chatroom broadcast\" : %v, \"online broadcast\" : %v, " +
	"\"query chatroom member count\" : %v, \"query chatroom member detail\" : %v, " +
	"\"query user in room\" : %v, \"query user session in room\" : %v }, " +
	"\"group qps\" : { " +
	"\"GroupJoins\" : %v, \"GroupCreates\" : %v, \"GroupQuits\" : %v, \"GroupDismiss\" : %v, \"GroupUserLists\" : %v, " +
	"\"GroupIsMemebers\" : %v, \"GroupInGroups\" : %v, \"GroupInfos\" : %v, \"GroupCreatedList\" : %v, " +
	"\"GroupJoinCounts\" : %v, \"GroupSends\" : %v, \"GroupMessageLists\" : %v }, " +
	"\"average response time(ms)\" : { \"stat\" : %v, \"im\" : %.3f, \"peer\" : %.3f, " +
	"\"robot join chatroom\" : %.3f, \"robot quit chatroom\" : %.3f, \"send chatroom message\" : %.3f } }"

func (stat *CenterStat) QpsString() string {
	// compute peer average response time
	peerAverage := float64(stat.AtomicGetPeerRespTime())
	peerReqs := stat.AtomicGetPeers() + stat.AtomicGetPeerFails()
	if peerReqs == 0 {
		peerAverage = 0
	} else {
		peerAverage = peerAverage / float64(peerReqs) / float64(time.Millisecond)
	}

	// compute im request average response time
	imAverage := float64(stat.AtomicGetImRespTime())
	imReqs := stat.AtomicGetIms() + stat.AtomicGetImFails()
	if imReqs == 0 {
		imAverage = 0
	} else {
		imAverage = imAverage / float64(imReqs) / float64(time.Millisecond)
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

	// compute send chatroom message request average response time
	sendAverage := float64(stat.AtomicGetSendRespTime())
	sendReqs := stat.AtomicGetSends() + stat.AtomicGetSendFails()
	if sendReqs == 0 {
		sendAverage = 0
	} else {
		sendAverage = sendAverage / float64(sendReqs) / float64(time.Millisecond)
	}

	return fmt.Sprintf(centerStatQpsFormat,
		imReqs, peerReqs, (stat.AtomicGetPublicHots() + stat.AtomicGetPublicHotFails()),
		(stat.AtomicGetRetrieves() + stat.AtomicGetRetrieveFails()),
		(stat.AtomicGetRecalls() + stat.AtomicGetRecallFails()),
		joinReqs, quitReqs, sendReqs, (stat.AtomicGetBroadcasts() + stat.AtomicGetBroadcastFails()), (stat.AtomicGetOnlineBroadcasts() + stat.AtomicGetOnlineBroadcastFails()),
		stat.AtomicGetQueryMemberCounts(), stat.AtomicGetQueryMemberDetails(),
		stat.AtomicGetQueryUserInRoom(), stat.AtomicGetQuerySessionInRoom(),
		(stat.AtomicGetGroupJoins() + stat.AtomicGetGroupJoinsFails()),
		(stat.AtomicGetGroupCreates() + stat.AtomicGetGroupCreateFails()),
		(stat.AtomicGetGroupQuits() + stat.AtomicGetGroupQuitFails()),
		(stat.AtomicGetGroupDismiss() + stat.AtomicGetGroupDismissFails()),
		(stat.AtomicGetGroupUserLists() + stat.AtomicGetGroupUserListFails()),
		(stat.AtomicGetGroupIsMemebers() + stat.AtomicGetGroupIsMemebersFails()),
		(stat.AtomicGetGroupInGroups() + stat.AtomicGetGroupInGroupsFails()),
		(stat.AtomicGetGroupInfos() + stat.AtomicGetGroupInfoFails()),
		(stat.AtomicGetGroupCreatedList() + stat.AtomicGetGroupCreatedListFails()),
		(stat.AtomicGetGroupJoinCounts() + stat.AtomicGetGroupJoinCountFails()),
		(stat.AtomicGetGroupSends() + stat.AtomicGetGroupSendFails()),
		(stat.AtomicGetGroupMessageLists() + stat.AtomicGetGroupMessageListFails()),
		stat.AtomicGetStatResponse(), imAverage, peerAverage,
		joinAverage, quitAverage, sendAverage)
}

// atomic get functions
// ---------------------------------------------------------------------------------------------------------

// response time stat related

func (stat *CenterStat) AtomicGetStatResponse() bool {
	if 0 == atomic.LoadInt32(&stat.StatResponse) {
		return false
	}

	return true
}

func (stat *CenterStat) AtomicSetStatResponse(value bool) {
	if value {
		atomic.StoreInt32(&stat.StatResponse, 1)
	} else {
		atomic.StoreInt32(&stat.StatResponse, 0)
	}
}

func (stat *CenterStat) AtomicGetPeerRespTime() uint64 {
	return atomic.LoadUint64(&stat.PeerRespTime)
}

func (stat *CenterStat) AtomicGetImRespTime() uint64 {
	return atomic.LoadUint64(&stat.ImRespTime)
}

func (stat *CenterStat) AtomicGetJoinRespTime() uint64 {
	return atomic.LoadUint64(&stat.JoinRespTime)
}

func (stat *CenterStat) AtomicGetQuitRespTime() uint64 {
	return atomic.LoadUint64(&stat.QuitRespTime)
}

func (stat *CenterStat) AtomicGetSendRespTime() uint64 {
	return atomic.LoadUint64(&stat.SendRespTime)
}

// p2p chat related

func (stat *CenterStat) AtomicGetTotalP2pRequests() uint64 {
	return atomic.LoadUint64(&stat.PublicHots) + atomic.LoadUint64(&stat.PublicHotFails) +
		atomic.LoadUint64(&stat.Peers) + atomic.LoadUint64(&stat.PeerFails) +
		atomic.LoadUint64(&stat.IMs) + atomic.LoadUint64(&stat.IMFails) +
		atomic.LoadUint64(&stat.Retrieves) + atomic.LoadUint64(&stat.RetrieveFails) +
		atomic.LoadUint64(&stat.Recalls) + atomic.LoadUint64(&stat.RecallFails)
}

func (stat *CenterStat) AtomicGetPublicHots() uint64 {
	return atomic.LoadUint64(&stat.PublicHots)
}

func (stat *CenterStat) AtomicGetPublicHotFails() uint64 {
	return atomic.LoadUint64(&stat.PublicHotFails)
}

func (stat *CenterStat) AtomicGetPeers() uint64 {
	return atomic.LoadUint64(&stat.Peers)
}

func (stat *CenterStat) AtomicGetPeerFails() uint64 {
	return atomic.LoadUint64(&stat.PeerFails)
}

func (stat *CenterStat) AtomicGetIms() uint64 {
	return atomic.LoadUint64(&stat.IMs)
}

func (stat *CenterStat) AtomicGetImFails() uint64 {
	return atomic.LoadUint64(&stat.IMFails)
}

func (stat *CenterStat) AtomicGetRetrieves() uint64 {
	return atomic.LoadUint64(&stat.Retrieves)
}

func (stat *CenterStat) AtomicGetRetrieveFails() uint64 {
	return atomic.LoadUint64(&stat.RetrieveFails)
}

func (stat *CenterStat) AtomicGetRecalls() uint64 {
	return atomic.LoadUint64(&stat.Recalls)
}

func (stat *CenterStat) AtomicGetRecallFails() uint64 {
	return atomic.LoadUint64(&stat.RecallFails)
}

// chatroom related

func (stat *CenterStat) AtomicGetJoins() uint64 {
	return atomic.LoadUint64(&stat.Joins)
}

func (stat *CenterStat) AtomicGetJoinFails() uint64 {
	return atomic.LoadUint64(&stat.JoinFails)
}

func (stat *CenterStat) AtomicGetQuits() uint64 {
	return atomic.LoadUint64(&stat.Quits)
}

func (stat *CenterStat) AtomicGetQuitFails() uint64 {
	return atomic.LoadUint64(&stat.QuitFails)
}

func (stat *CenterStat) AtomicGetSends() uint64 {
	return atomic.LoadUint64(&stat.Sends)
}

func (stat *CenterStat) AtomicGetSendFails() uint64 {
	return atomic.LoadUint64(&stat.SendFails)
}

func (stat *CenterStat) AtomicGetBroadcasts() uint64 {
	return atomic.LoadUint64(&stat.Broadcast)
}

func (stat *CenterStat) AtomicGetBroadcastFails() uint64 {
	return atomic.LoadUint64(&stat.BroadcastFails)
}

func (stat *CenterStat) AtomicGetOnlineBroadcasts() uint64 {
	return atomic.LoadUint64(&stat.OnlineBroadcast)
}

func (stat *CenterStat) AtomicGetOnlineBroadcastFails() uint64 {
	return atomic.LoadUint64(&stat.OnlineBroadcastFails)
}

func (stat *CenterStat) AtomicGetQueryMemberCounts() uint64 {
	return atomic.LoadUint64(&stat.QueryMemberCounts)
}

func (stat *CenterStat) AtomicGetQueryUserInRoom() uint64 {
	return atomic.LoadUint64(&stat.QueryUserInRoom)
}

func (stat *CenterStat) AtomicGetQuerySessionInRoom() uint64 {
	return atomic.LoadUint64(&stat.QuerySessionInRoom)
}

func (stat *CenterStat) AtomicGetQueryMemberDetails() uint64 {
	return atomic.LoadUint64(&stat.QueryMemberDetails)
}

// Group get function
func (stat *CenterStat) AtomicGetGroupJoins() uint64 {
	return atomic.LoadUint64(&stat.GroupJoins)
}

func (stat *CenterStat) AtomicGetGroupJoinsFails() uint64 {
	return atomic.LoadUint64(&stat.GroupJoinFails)
}

func (stat *CenterStat) AtomicGetGroupCreates() uint64 {
	return atomic.LoadUint64(&stat.GroupCreates)
}

func (stat *CenterStat) AtomicGetGroupCreateFails() uint64 {
	return atomic.LoadUint64(&stat.GroupCreateFails)
}

func (stat *CenterStat) AtomicGetGroupQuits() uint64 {
	return atomic.LoadUint64(&stat.GroupQuits)
}

func (stat *CenterStat) AtomicGetGroupQuitFails() uint64 {
	return atomic.LoadUint64(&stat.GroupQuitFails)
}

func (stat *CenterStat) AtomicGetGroupDismiss() uint64 {
	return atomic.LoadUint64(&stat.GroupDismiss)
}

func (stat *CenterStat) AtomicGetGroupDismissFails() uint64 {
	return atomic.LoadUint64(&stat.GroupDismissFails)
}

func (stat *CenterStat) AtomicGetGroupSends() uint64 {
	return atomic.LoadUint64(&stat.GroupSends)
}

func (stat *CenterStat) AtomicGetGroupSendFails() uint64 {
	return atomic.LoadUint64(&stat.GroupSendFails)
}

func (stat *CenterStat) AtomicGetGroupInfos() uint64 {
	return atomic.LoadUint64(&stat.GroupInfos)
}

func (stat *CenterStat) AtomicGetGroupInfoFails() uint64 {
	return atomic.LoadUint64(&stat.GroupInfoFails)
}

func (stat *CenterStat) AtomicGetGroupUserLists() uint64 {
	return atomic.LoadUint64(&stat.GroupUserLists)
}

func (stat *CenterStat) AtomicGetGroupUserListFails() uint64 {
	return atomic.LoadUint64(&stat.GroupUserListFails)
}

func (stat *CenterStat) AtomicGetGroupIsMemebers() uint64 {
	return atomic.LoadUint64(&stat.GroupIsMemebers)
}

func (stat *CenterStat) AtomicGetGroupIsMemebersFails() uint64 {
	return atomic.LoadUint64(&stat.GroupIsMemebersFails)
}

func (stat *CenterStat) AtomicGetGroupInGroups() uint64 {
	return atomic.LoadUint64(&stat.GroupInGroups)
}

func (stat *CenterStat) AtomicGetGroupInGroupsFails() uint64 {
	return atomic.LoadUint64(&stat.GroupInGroupsFails)
}

func (stat *CenterStat) AtomicGetGroupJoinCounts() uint64 {
	return atomic.LoadUint64(&stat.GroupJoinCounts)
}

func (stat *CenterStat) AtomicGetGroupJoinCountFails() uint64 {
	return atomic.LoadUint64(&stat.GroupJoinCountFails)
}

func (stat *CenterStat) AtomicGetGroupMessageLists() uint64 {
	return atomic.LoadUint64(&stat.GroupMessageLists)
}

func (stat *CenterStat) AtomicGetGroupMessageListFails() uint64 {
	return atomic.LoadUint64(&stat.GroupMessageListFails)
}

func (stat *CenterStat) AtomicGetGroupCreatedList() uint64 {
	return atomic.LoadUint64(&stat.GroupCreatedList)
}

func (stat *CenterStat) AtomicGetGroupCreatedListFails() uint64 {
	return atomic.LoadUint64(&stat.GroupCreatedListFails)
}

// atomic add functions
// ---------------------------------------------------------------------------------------------------------

// response time stat related

func (stat *CenterStat) AtomicAddPeerRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.PeerRespTime, i)
}

func (stat *CenterStat) AtomicAddImRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.ImRespTime, i)
}

func (stat *CenterStat) AtomicAddJoinRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.JoinRespTime, i)
}

func (stat *CenterStat) AtomicAddQuitRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuitRespTime, i)
}

func (stat *CenterStat) AtomicAddSendRespTime(i uint64) uint64 {
	return atomic.AddUint64(&stat.SendRespTime, i)
}

// p2p chat related

func (stat *CenterStat) AtomicAddPublicHots(i uint64) uint64 {
	return atomic.AddUint64(&stat.PublicHots, i)
}

func (stat *CenterStat) AtomicAddPublicHotFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.PublicHotFails, i)
}

func (stat *CenterStat) AtomicAddPeers(i uint64) uint64 {
	return atomic.AddUint64(&stat.Peers, i)
}

func (stat *CenterStat) AtomicAddPeerFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.PeerFails, i)
}

func (stat *CenterStat) AtomicAddIms(i uint64) uint64 {
	return atomic.AddUint64(&stat.IMs, i)
}

func (stat *CenterStat) AtomicAddImFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.IMFails, i)
}

func (stat *CenterStat) AtomicAddRetrieves(i uint64) uint64 {
	return atomic.AddUint64(&stat.Retrieves, i)
}

func (stat *CenterStat) AtomicAddRetrieveFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.RetrieveFails, i)
}

func (stat *CenterStat) AtomicAddRecalls(i uint64) uint64 {
	return atomic.AddUint64(&stat.Recalls, i)
}

func (stat *CenterStat) AtomicAddRecallFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.RecallFails, i)
}

// chatroom related

func (stat *CenterStat) AtomicAddJoins(i uint64) uint64 {
	return atomic.AddUint64(&stat.Joins, i)
}

func (stat *CenterStat) AtomicAddJoinFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.JoinFails, i)
}

func (stat *CenterStat) AtomicAddQuits(i uint64) uint64 {
	return atomic.AddUint64(&stat.Quits, i)
}

func (stat *CenterStat) AtomicAddQuitFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuitFails, i)
}

func (stat *CenterStat) AtomicAddSends(i uint64) uint64 {
	return atomic.AddUint64(&stat.Sends, i)
}

func (stat *CenterStat) AtomicAddSendFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.SendFails, i)
}

func (stat *CenterStat) AtomicAddBroadcast(i uint64) uint64 {
	return atomic.AddUint64(&stat.Broadcast, i)
}

func (stat *CenterStat) AtomicAddBroadcastFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.BroadcastFails, i)
}

func (stat *CenterStat) AtomicAddOnlineBroadcast(i uint64) uint64 {
	return atomic.AddUint64(&stat.OnlineBroadcast, i)
}

func (stat *CenterStat) AtomicAddOnlineBroadcastFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.OnlineBroadcastFails, i)
}

func (stat *CenterStat) AtomicAddQueryMemberCounts(i uint64) uint64 {
	return atomic.AddUint64(&stat.QueryMemberCounts, i)
}

func (stat *CenterStat) AtomicAddQueryUserInRoom(i uint64) uint64 {
	return atomic.AddUint64(&stat.QueryUserInRoom, i)
}

func (stat *CenterStat) AtomicAddQuerySessionInRoom(i uint64) uint64 {
	return atomic.AddUint64(&stat.QuerySessionInRoom, i)
}

func (stat *CenterStat) AtomicAddQueryMemberDetails(i uint64) uint64 {
	return atomic.AddUint64(&stat.QueryMemberDetails, i)
}

// chat group related
func (stat *CenterStat) AtomicAddGroupJoin(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupJoins, i)
}

func (stat *CenterStat) AtomicAddGroupJoinsFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupJoinFails, i)
}

func (stat *CenterStat) AtomicAddGroupCreate(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupCreates, i)
}

func (stat *CenterStat) AtomicAddGroupCreateFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupCreateFails, i)
}

func (stat *CenterStat) AtomicAddGroupQuit(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupQuits, i)
}

func (stat *CenterStat) AtomicAddGroupQuitFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupQuitFails, i)
}

func (stat *CenterStat) AtomicAddGroupDismiss(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupDismiss, i)
}

func (stat *CenterStat) AtomicAddGroupDismissFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupDismissFails, i)
}

func (stat *CenterStat) AtomicAddGroupSend(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupSends, i)
}

func (stat *CenterStat) AtomicAddGroupSendFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupSendFails, i)
}

func (stat *CenterStat) AtomicAddGroupInfo(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupInfos, i)
}

func (stat *CenterStat) AtomicAddGroupInfoFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupInfoFails, i)
}

func (stat *CenterStat) AtomicAddGroupUserList(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupUserLists, i)
}

func (stat *CenterStat) AtomicAddGroupUserListFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupUserListFails, i)
}

func (stat *CenterStat) AtomicAddGroupIsMemeber(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupIsMemebers, i)
}

func (stat *CenterStat) AtomicAddGroupIsMemebersFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupIsMemebersFails, i)
}

func (stat *CenterStat) AtomicAddGroupInGroups(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupInGroups, i)
}

func (stat *CenterStat) AtomicAddGroupInGroupsFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupInGroupsFails, i)
}

func (stat *CenterStat) AtomicAddGroupJoinCount(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupJoinCounts, i)
}

func (stat *CenterStat) AtomicAddGroupJoinCountFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupJoinCountFails, i)
}

func (stat *CenterStat) AtomicAddGroupMessageList(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupMessageLists, i)
}

func (stat *CenterStat) AtomicAddGroupMessageListFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupMessageListFails, i)
}

func (stat *CenterStat) AtomicAddGroupCreatedList(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupCreatedList, i)
}

func (stat *CenterStat) AtomicAddGroupCreatedListFails(i uint64) uint64 {
	return atomic.AddUint64(&stat.GroupCreatedListFails, i)
}
