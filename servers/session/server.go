package main

import (
	"errors"
	"fmt"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

type GorpcService struct{}

var rpcServer *gorpc.Server

func GorpcServer() {
	if netConf().GorpcListen == "" {
		panic("empty gorpc_listen")
	}
	Logger.Trace("gorpc listen", netConf().GorpcListen)
	rpcServer = gorpc.NewServer(netConf().GorpcListen)
	rpcServer.Register(new(GorpcService))
	rpcServer.Serve()
	panic("invalid gorpc listen" + netConf().GorpcListen)
}

// 记录日志的时候记录下用户的property
func formatProperty(s *session.UserSession) string {
	return fmt.Sprintf("%s,%s,%s,%s,%v,%s", s.Platform, s.Deviceid, s.Property["MobileType"], s.Property["NetType"], s.IsLoginUser, s.Property["ClientIp"])
}

func (this *GorpcService) Open(req *session.OpenSessionReq, resp *session.OpenSessionResp) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countOpenRespTime(req.UserId, req.TraceId, "OpenSession", req.AppId)
		defer countFunc()
	}

	Logger.Trace(req.UserId, req.AppId, req.TraceId, "session.Open", req.Property["ConnectionType"], formatProperty(req.UserSession))
	oldSessions, err := saver.SaveSession(req.UserSession)
	if err != nil {
		// count failed operation
		requestStat.AtomicAddOpenSessionFails(1)

		Logger.Error(req.UserId, req.AppId, req.TraceId, "session.Open", "saver.SaveSession error", err.Error())
		return errors.New("saver.SaverSession error:" + err.Error())
	}

	// count successful operation
	requestStat.AtomicAddOpenSessions(1)

	resp.Tags = []string{fmt.Sprintf("%d", req.UserSession.AppId)}
	resp.OldUserSessions = oldSessions
	onlineCache.Del(req.AppId, req.UserId)
	return nil
}

func (this *GorpcService) Close(req *session.CloseSessionReq, resp *session.CloseSessionResp) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countCloseRespTime(req.UserId, req.TraceId, "CloseSession", req.AppId)
		defer countFunc()
	}

	if err := saver.RemoveSession(req.UserSession); err != nil {
		Logger.Error(req.UserId, req.AppId, req.TraceId, "session.Close", "saver.RemoveSession error", err.Error())
	}

	tags, err := cleanUserSessionTags(req.UserSession, true) //note:只有正式用户才会返回聊天室id
	if err != nil {
		Logger.Error(req.UserId, req.AppId, req.TraceId, "session.Close", "session.cleanUserSessionTags error", err.Error())
	}
	Logger.Trace(req.UserId, req.AppId, req.TraceId, "session.Close", req.Property["CloseReason"], tags, req.Property["ClientIp"], req.Property["ConnectionType"])

	// here we always treat operation as successful
	requestStat.AtomicAddCloseSessions(1)

	resp.Tags = []string{fmt.Sprintf("%d", req.UserSession.AppId)}
	resp.Tags = append(resp.Tags, tags...)
	onlineCache.Del(req.AppId, req.UserId)
	return nil
}

func (this *GorpcService) Query(req *session.QuerySessionReq, resp *session.QuerySessionResp) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countQuerySessionRespTime("", "", "QuerySession", logic.DEFAULT_APPID)
		defer countFunc()
	}

	result, err := saver.QueryUserSession(req.QueryUserSessions)
	if err != nil {
		// count failed operation
		requestStat.AtomicAddQuerySessionFails(1)

		Logger.Error("center", "", "", "Query", "saver.QueryUserSession error", err.Error())
		return errors.New("saver.RemoveSession error:" + err.Error())
	}
	Logger.Trace("", "", "", "Query", len(req.QueryUserSessions), len(result))

	resp.ResultUserSessions = result

	// count successful operation
	requestStat.AtomicAddQuerySessions(1)

	// 做一些像加载离线之类的事情go something more
	return nil
}

func (this *GorpcService) JoinChatRoom(req *session.JoinChatRoomRequest, resp *session.JoinChatRoomResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countJoinRespTime(req.User, req.RoomID, "session.JoinChatRoom", req.Appid)
		defer countFunc()
	}
	// 回调业务 join
	partner, err := forwardChatRoomRequest(req.UserChatRoomRequest, req.Properties)
	if err != nil {
		Logger.Error(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "JoinChatRoom", "forwardChatRoomRequest error:"+err.Error(), req.RoomID)

		resp.UserChatRoomResponse = &session.UserChatRoomResponse{
			Code:     session.CallbackFailed,
			Reason:   "callback failed",
			Response: []byte(""),
		}
		return nil
	}
	clientIP := req.Properties["ClientIp"]
	connType := req.Properties["ConnectionType"]
	platform := req.Properties["Platform"]
	deviceID := req.Properties["Deviceid"]
	resp.UserChatRoomResponse = &session.UserChatRoomResponse{
		Code:     partner.Code,
		Reason:   partner.Reason,
		Response: []byte(partner.Data.Response),
	}
	ret := 0
	if partner.Code == session.Success {
		ret = doJoinChatRoom(req, resp, &partner.Data.Notify, partner.Data.Priority)
	} else {
		ret = session.AddUserFailed
	}
	Logger.Trace(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "JoinChatRoom", req.RoomID, partner.Code, clientIP, connType, platform, ret, deviceID)

	// count operations
	if resp.Code == session.Success {
		requestStat.AtomicAddJoins(1)
	} else {
		requestStat.AtomicAddJoinFails(1)
	}

	return nil
}

func (this *GorpcService) QuitChatRoom(req *session.QuitChatRoomRequest, resp *session.QuitChatRoomResponse) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countQuitRespTime(req.User, req.RoomID, "session.QuitChatRoom", req.Appid)
		defer countFunc()
	}

	callback := false // 是否回调业务方
	if req.UserType == session.RobotChatRoomUser {
		if err := saver.RemoveChatRoomRobot(req.UserChatRoomRequest); err != nil {
			Logger.Error(req.User, req.Appid, req.RoomID, "QuitChatRoom", "saver.RemoveChatRoomRobot error", err)

			// count failed operation
			requestStat.AtomicAddQuitFails(1)
		}
	} else {
		if code, err := saver.RemoveChatRoomUser(req.UserChatRoomRequest); err != nil {
			Logger.Error(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "QuitChatRoom", "saver.RemoveChatRoomUser error:"+err.Error(), req.RoomID)

			// count failed operation
			requestStat.AtomicAddQuitFails(1)
		} else if code == session.AllSessionQuitedChatRoom {
			addQuitCallback(&quitChatRoomCallback{
				UserChatRoomRequest: req.UserChatRoomRequest,
				Quited:              true,
				RoomIDs:             []string{req.RoomID},
				ClientIP:            req.Properties["ClientIp"],
				ConnType:            req.Properties["ConnectionType"],
				Platform:            req.Properties["Platform"],
				DeviceID:            req.Properties["Deviceid"],
			})

			// count successful operation
			requestStat.AtomicAddQuits(1)
			callback = true
		}
	}
	clientIP := req.Properties["ClientIp"]
	connType := req.Properties["ConnectionType"]
	platform := req.Properties["Platform"]
	deviceID := req.Properties["Deviceid"]
	Logger.Trace(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "QuitChatRoom", req.RoomID, callback, clientIP, connType, platform, deviceID)
	return nil
}

func (this *GorpcService) QueryChatRoom(req *session.UserChatRoomRequest, resp *session.ChatRoomDetail) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countQueryRespTime(req.User, req.RoomID, "session.QueryChatRoom", req.Appid)
		defer countFunc()
	}

	detail, err := getChatRoomDetail(req.RoomID, req.Appid)
	if err != nil {
		// count failed operation
		requestStat.AtomicAddQueryFails(1)
		Logger.Error(req.User, req.Appid, req.RoomID, "session.QueryChatRoom", "getChatRoomDetail err", err)

		return err
	}
	Logger.Trace(req.User, req.Appid, req.RoomID, "QueryChatRoom", "", "")
	*resp = *detail

	// count successful operation
	requestStat.AtomicAddQuerys(1)
	return nil
}

func (this *GorpcService) QueryUserOnlineCache(req *session.UserOnlineCache, resp *map[string][]*logic.UserGateway) error {
	ret := onlineCache.CheckOnline(req.AppId, req.UserIds)
	*resp = ret
	return nil
}

// 获取online_cache的统计信息
func (this *GorpcService) OnlineCacheStat(req int, resp *map[string][]map[string]uint64) error {
	*resp = onlineCache.Stat()
	return nil
}
