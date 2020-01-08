package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"strconv"

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
	// start public message reload handler
	go reloadPublicMsgCacheHandler()

	Logger.Trace("gorpc listen", netConf().GorpcListen)
	rpcServer = gorpc.NewServer(netConf().GorpcListen)
	rpcServer.Register(new(GorpcService))
	rpcServer.Serve()
	panic("invalid gorpc listen" + netConf().GorpcListen)
}

func (this *GorpcService) Helloworld(foo string, resp *int) error {
	// @todo 需要处理退出动作
	fmt.Println("helloworld")
	return nil
}

func (this *GorpcService) SaveSession(userSession *session.UserSession, resp *[]*session.UserSession) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countSessionResponseTime(userSession.UserId, userSession.TraceId,
			"saver.SaveSession", userSession.AppId)
		defer countFunc()
	}

	Logger.Trace(userSession.UserId, userSession.AppId, userSession.TraceId, "SaveSession", "req", "")

	r := []*session.UserSession{}

	prop := map[string]string{
		"IsLoginUser":  strconv.FormatBool(userSession.IsLoginUser),
		"SessionKey":   userSession.SessionKey,
		"ConnectionId": fmt.Sprintf("%d", userSession.ConnectionId),
		"GatewayAddr":  userSession.GatewayAddr,
		"AppId":        fmt.Sprintf("%d", userSession.AppId),
		"UserId":       userSession.UserId,
		"TraceId":      userSession.TraceId,
		"SessionId":    userSession.SessionId,
		"LoginTime":    userSession.LoginTime,
		"Deviceid":     userSession.Deviceid,
		"ClientVer":    userSession.ClientVer,
		"SenderType":   userSession.SenderType, //jid or phone ?
		"Platform":     userSession.Platform,
		"UserIp":       userSession.UserIp,
	}
	for k, v := range userSession.Property {
		prop[k] = v
	}
	oldSessionProp, err := saveUserSession(prop)
	if err != nil {
		requestStat.AtomicAddOpenSessionFails(1) // count failed operation
		Logger.Error(userSession.UserId, userSession.AppId, userSession.TraceId, "saver.SaveSession", "saveUserSession error", err.Error())
		return errors.New("saveUserProperty error:" + err.Error())
	}

	if 0 != len(oldSessionProp) {
		oldSession, err := prop2usersession(oldSessionProp)
		if err == nil {
			r = append(r, oldSession)
		}
	}

	requestStat.AtomicAddOpenSessions(1) // count successful request
	*resp = r
	return nil
}

func (this *GorpcService) RemoveSession(userSession *session.UserSession, resp *int) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countSessionResponseTime(userSession.UserId, userSession.TraceId,
			"saver.RemoveSession", userSession.AppId)
		defer countFunc()
	}

	Logger.Trace(userSession.UserId, userSession.AppId, userSession.TraceId, "RemoveSession", "req", "")

	prop := map[string]string{
		"ConnectionId": fmt.Sprintf("%d", userSession.ConnectionId),
		"GatewayAddr":  userSession.GatewayAddr,
		"AppId":        fmt.Sprintf("%d", userSession.AppId),
		"UserId":       userSession.UserId,
		"Platform":     userSession.Platform,
		"Deviceid":     userSession.Deviceid,
		"TraceId":      userSession.TraceId,
	}

	if err := removeUserSession(prop); err != nil {
		requestStat.AtomicAddCloseSessionFails(1) // count failed operation
		*resp = -1
		Logger.Error(userSession.UserId, userSession.AppId, userSession.TraceId, "saver.RemoveSession", "removeUserSession error", err.Error())
		return errors.New("save removeUserSession error:" + err.Error())
	}

	requestStat.AtomicAddCloseSessions(1) // count successful request
	*resp = 0
	return nil
}

func (this *GorpcService) QueryUserSessionSummary(req *session.UserSession, resp *map[string]int) error {
	ret := make(map[string]int, 4)
	prop := map[string]string{
		"AppId":  fmt.Sprintf("%d", req.AppId),
		"UserId": req.UserId,
	}

	masterKey, _ := getUserSessionKey(prop)
	if masterKey == "" {
		return errors.New("lack userid or appid field")
	}
	secondKeys, err := SessionPool.Call(getSessionAddr(masterKey)).SMEMBERS(masterKey)
	if err != nil {
		return err
	}
	for i, skey := range secondKeys {
		if i >= netConf().MaxSessionSummary {
			break
		}
		mobileType, err := SessionPool.Call(getSessionAddr(masterKey)).HGET(string(skey), "MobileType")
		if err != nil {
			continue
		}
		// mobileType
		ret[string(mobileType)]++
	}
	if other := len(secondKeys) - netConf().MaxSessionSummary; other > 0 {
		ret["other"] = other
	}
	*resp = ret
	return nil
}

func (this *GorpcService) QueryUserSession(querys []*session.UserSession, resp *[]*session.UserSession) error {
	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countSessionResponseTime("", "", "saver.QueryUserSession", logic.DEFAULT_APPID)
		defer countFunc()
	}

	r := []*session.UserSession{}
	var lasterr error = nil
	var totalSessions int = 0

	for _, query := range querys {
		prop := map[string]string{
			"AppId":  fmt.Sprintf("%d", query.AppId),
			"UserId": query.UserId,
		}

		sessions, err := getUserSession(prop)

		if err == nil {
			totalSessions += len(sessions)
			for _, usersession := range sessions {
				isLoginUser, _ := strconv.ParseBool(usersession["IsLoginUser"])
				connId, _ := strconv.ParseUint(usersession["ConnectionId"], 10, 64)

				s := &session.UserSession{
					UserId:       query.UserId,
					AppId:        query.AppId,
					IsLoginUser:  isLoginUser,
					SessionKey:   usersession["SessionKey"],
					ConnectionId: logic.ConnectionId(connId),
					GatewayAddr:  usersession["GatewayAddr"],
					Property:     usersession,
					LoginTime:    usersession["LoginTime"],
					SessionId:    usersession["ServerRam"],
					SenderType:   usersession["SenderType"],
					Platform:     usersession["Platform"],
					Deviceid:     usersession["Deviceid"],
					UserIp:       usersession["ClientIp"],
					ClientVer:    usersession["CVersion"],
				}
				r = append(r, s)
			}

			requestStat.AtomicAddQuerySessionSuccess(1) // count successful query
		} else {
			requestStat.AtomicAddQuerySessionFails(1) // count failed query
			Logger.Error(query.UserId, query.AppId, query.TraceId, "saver.QueryUserSession", "saver.getUserSession error", err.Error())
			lasterr = err
		}
	}
	Logger.Trace("", "", "", "saver.QueryUserSession", len(querys), len(r))

	*resp = r

	requestStat.AtomicAddQuerySessions(1) // count query session request
	if len(r) == 0 {
		return lasterr
	} else {
		return nil
	}
}

func (this *GorpcService) GetActiveUserNum(req *saver.GetActiveReq, resp *int) error {
	Logger.Trace("", "", "", "saver.GetActiveUserNum", "recv GetActiveUserNum req", req.AppId)
	num, err := getActiveUserNum(req.AppId)
	if err != nil {
		Logger.Error("", "", "", "saver.GetActiveUserNum", "saver.GetActiveUserNum error", err.Error())
		return errors.New("saver.GetActiveUserNum error:" + err.Error())
	}

	*resp = num
	return nil
}

func (this *GorpcService) GetActiveUserSessions(req *saver.GetActiveReq, resp *[]*session.UserSession) error {
	Logger.Trace("", "", "", "saver.GetActiveUserSessions", "recv GetActiveUserSessions req", fmt.Sprintf("%d:%d", req.AppId, req.ZoneId))

	usersessions, err := getActiveUserInzone(req.AppId, req.ZoneId)
	if err != nil {
		Logger.Error("", "", "", "saver.GetActiveUserSessions", "saver.GetActiveUserSessions error", err.Error())
		return errors.New("saver.GetActiveUserSessions error:" + err.Error())
	}

	*resp = usersessions
	return nil
}

func (this *GorpcService) GetActiveChatRoomNum(req *saver.GetActiveReq, resp *int) error {
	Logger.Trace("", "", "", "saver.GetActiveChatRoomNum", "recv GetActiveChatRoomNum req", req.AppId)
	num, err := getActiveChatRoomNum(req.AppId)
	if err != nil {
		Logger.Error("", "", "", "saver.GetActiveChatRoomNum", "saver.GetActiveChatRoomNum error", err.Error())
		return errors.New("saver.GetActiveChatRoomNum error:" + err.Error())
	}

	*resp = num
	return nil
}

func (this *GorpcService) QueryChatRoomMemberCount(req *saver.ChatRoomsAppid, resp *map[string]map[string]int) error {
	r := make(map[string]map[string]int, len(req.RoomIDs))
	for _, room := range req.RoomIDs {
		if count, err := queryChatRoomMemberCount(room, req.Appid); err != nil {
			return err
		} else {
			r[room] = count
		}
	}
	Logger.Trace("", req.Appid, "", "saver.QueryChatRoomMemberDetails", "QueryChatRoomMemberDetails resp", r)
	*resp = r
	return nil
}

func (this *GorpcService) QueryChatRoomDetail(req *saver.ChatRoomsAppid, resp *map[string]*session.ChatRoomDetail) error {
	r := make(map[string]*session.ChatRoomDetail, len(req.RoomIDs))
	appid := req.Appid
	for _, room := range req.RoomIDs {
		if detail, err := queryChatRoomDetail(room, appid); err != nil {
			return err
		} else if detail != nil {
			r[room] = detail
		}
	}
	Logger.Trace("", req.Appid, req.RoomIDs, "saver.QueryChatRoomDetails", "", len(r))
	*resp = r
	return nil
}

func (this *GorpcService) CheckUserSessionInRoom(req *session.UserChatRoomRequest, resp *bool) error {
	inRoom, err := checkUserSessionInChatRoom(req.User, req.RoomID, req.Appid, req.GatewayAddr, req.ConnectionId)
	if err != nil {
		return err
	}
	Logger.Trace(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "saver.CheckUserSessionInRoom", req.RoomID, inRoom)
	*resp = inRoom
	return nil
}

func (this *GorpcService) QueryUserSessionChatRoomList(req *session.UserChatRoomRequest, resp *[]string) error {
	Logger.Trace(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "QueryUserChatRoomList", "", "")
	if ret, err := queryUserSessionChatRooms(req.User, req.Appid, req.GatewayAddr, req.ConnectionId); err != nil {
		return err
	} else {
		*resp = ret
		return nil
	}
}

func (this *GorpcService) AddChatRoomUser(req *session.UserChatRoomRequest, resp *int) error {
	Logger.Trace(req.User, req.Appid, "", "saver.AddChatRoomUser", req.RoomID, req.GatewayAddr, req.ConnectionId)
	// 查看聊天室是否存在，不存在则创建
	if _, err := checkChatRoomExists(req.RoomID, req.Appid, true); err != nil {
		return err
	}
	if code, err := addUserIntoChatRoom(req.User, req.RoomID, req.Appid, req.GatewayAddr, req.ConnectionId, req.UserType); err != nil {
		return err
	} else {
		*resp = code
		return nil
	}
}

func (this *GorpcService) RemoveChatRoomUser(req *session.UserChatRoomRequest, resp *int) error {
	Logger.Trace(req.User, req.Appid, "", "saver.RemoveChatRoomUser", req.RoomID, req.GatewayAddr, req.ConnectionId)
	if code, err := removeUserFromChatRoom(req.User, req.RoomID, req.Appid, req.GatewayAddr, req.ConnectionId, req.UserType); err != nil {
		return err
	} else {
		*resp = code
		return nil
	}
}

func (this *GorpcService) CreateChatRoom(req *session.UserChatRoomRequest, resp *int) error {
	Logger.Trace("", req.Appid, "", "saver.CreateChatRoom", req.RoomID, "")
	if _, err := checkChatRoomExists(req.RoomID, req.Appid, true); err != nil {
		return err
	}
	return nil
}

func (this *GorpcService) AddChatRoomRobot(req *session.UserChatRoomRequest, resp *int) error {
	Logger.Trace(req.User, req.Appid, "", "saver.AddChatRoomRobot", req.RoomID, "")
	// 查看聊天室是否存在,聊天室不存在则不执行加入操作
	if exists, err := checkChatRoomExists(req.RoomID, req.Appid, false); err != nil {
		return err
	} else if !exists {
		return errors.New("Chatroom not exists")
	}
	if err := addRobotIntoChatRoom(req.User, req.RoomID, req.Appid); err != nil {
		return err
	}
	return nil
}

func (this *GorpcService) RemoveChatRoomRobot(req *session.UserChatRoomRequest, resp *int) error {
	Logger.Trace(req.User, req.Appid, "", "saver.RemoveChatRoomRobot", req.RoomID, "")
	if err := removeRobotFromChatRoom(req.User, req.RoomID, req.Appid); err != nil {
		return err
	}
	return nil
}

func (this *GorpcService) UpdateChatRoomMember(req *saver.ChatRoomMemberUpdate, resp *int) error {
	Logger.Trace("", req.Appid, "", "saver.UpdateChatRoomMember", req.RoomID)
	if err := updateChatRoomMemberCount(req.RoomID, req.Appid, req.Type, req.Count); err != nil {
		return err
	}
	return nil
}

func (this *GorpcService) QueryChatRoomUsers(req *saver.ChatRoomsAppid, resp *map[string][]string) error {
	Logger.Trace("", req.Appid, "", "saver.GetChatRoomUsers", req.RoomIDs)
	res := make(map[string][]string, len(req.RoomIDs))
	for _, room := range req.RoomIDs {
		users, err := queryChatRoomUsers(room, req.Appid)
		if err != nil {
			return err
		}
		res[room] = users
	}
	*resp = res
	return nil
}

func (this *GorpcService) QueryChatRoomsByZone(req *saver.GetActiveReq, resp *map[string]*session.ChatRoomDetail) error {
	Logger.Trace("", req.AppId, "", "saver.QueryChatRoomsByZone", req.ZoneId)
	rooms, err := getRoomIdsByZone(req.AppId, req.ZoneId)
	if err != nil {
		return err
	}
	res := make(map[string]*session.ChatRoomDetail, len(rooms))
	for _, room := range rooms {
		if detail, err := queryChatRoomDetail(room, req.AppId); err != nil {
			return err
		} else if detail != nil {
			res[room] = detail
		}
	}
	*resp = res
	return nil
}

func (this *GorpcService) QueryUserInRoom(req *saver.UserChatRoom, resp *bool) error {
	Logger.Trace("", req.Appid, req.User, "saver.QueryUserInRoom", req.RoomId)
	isIn, err := queryUserInRoom(req.User, req.RoomId, req.Appid)
	*resp = isIn
	if err != nil {
		return err
	} else {
		return nil
	}
}

func (this *GorpcService) CleanChatRoom(req *saver.ChatRoomsAppid, resp *int) error {
	Logger.Trace("", req.Appid, "", "saver.CleanChatRoom", req.RoomIDs)
	for _, room := range req.RoomIDs {
		if err := cleanChatRoom(room, req.Appid); err != nil {
			Logger.Error("", req.Appid, "", "saver.cleanChatRoom", room, err.Error())
		}
	}
	return nil
}

func (this *GorpcService) CacheChatRoomMessage(req *logic.ChatRoomMessage, resp *uint) error {
	Logger.Trace(req.Sender, req.Appid, req.RoomID, "saver.CacheChatRoomMessage", req.TimeStamp, string(req.MsgContent))
	// req.MsgID != 0，是旧系统过来的消息
	if req.MsgID != 0 {
		req.MaxID = req.MsgID
		if err := setChatRoomMaxMessageID(req.RoomID, req.Appid, req.MsgID); err != nil {
			return err
		}
	} else {
		if id, err := generateChatRoomMessageID(req.RoomID, req.Appid); err != nil {
			return err
		} else {
			req.MaxID, req.MsgID = id, id
		}
	}
	addCacheChatRoomMessage(req)
	*resp = req.MsgID
	return nil
}

func (this *GorpcService) GetCachedChatRoomMessages(req *saver.FetchChatRoomMessageReq, resp *map[uint]*logic.ChatRoomMessage) error {
	room := req.RoomID
	appid := req.Appid
	ret := make(map[uint]*logic.ChatRoomMessage, len(req.MsgIDs))
	// @todo, 复用decoder对象
	for _, msgid := range req.MsgIDs {
		if data, err := getCachedRoomMessage(room, appid, msgid); err != nil {
			return err
		} else if data != nil {
			var message logic.ChatRoomMessage
			buf := bytes.NewBuffer(data)
			if err = gob.NewDecoder(buf).Decode(&message); err != nil {
				return err
			}
			ret[msgid] = &message
		}
	}
	Logger.Trace("", req.Appid, req.RoomID, "saver.GetCachedChatRoomMessages", "resp", req.MsgIDs)
	*resp = ret
	return nil
}

const (
	MessageStoreMongo = iota
	MessageStoreRedis
)
const (
	SaverErrorNotSupportStorage = "get not support storage %d from config"
)

/*
 * Stores chat messages to storage
 * @param req is a saver.StoreMessagesRequest point which include messages information need to store
 * @param resp is a saver.StoreMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func (this *GorpcService) StoreChatMessages(req *saver.StoreMessagesRequest,
	resp *saver.StoreMessagesResponse) error {

	switch netConf().MessageStore {
	case MessageStoreMongo:
		return StoreMongoChatMessages(req, resp)
	}

	err := fmt.Sprintf(SaverErrorNotSupportStorage, netConf().MessageStore)
	Logger.Error("", req.Appid, "", "StoreChatMessages", "config error", err)
	return errors.New(err)
}

/*
 * Retrieves chat messages from storage
 * @param req is a saver.RetrieveMessagesRequest point which include retrieving request information
 * @param resp is a saver.RetrieveMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func (this *GorpcService) RetrieveChatMessages(req *saver.RetrieveMessagesRequest,
	resp *saver.RetrieveMessagesResponse) error {
	switch netConf().MessageStore {
	case MessageStoreMongo:
		return RetrieveMongoChatMessages(req, resp)
	}

	err := fmt.Sprintf(SaverErrorNotSupportStorage, netConf().MessageStore)
	Logger.Error("", req.Appid, "", "RetrieveChatMessages", "config error", err)
	return errors.New(err)
}

// 读取未读消息数量
func (this *GorpcService) RetrieveUnreadCount(req []*saver.RetrieveMessagesRequest,
	resp *[]*saver.RetrieveMessagesResponse) error {
	*resp = make([]*saver.RetrieveMessagesResponse, 0, len(req))
	switch netConf().MessageStore {
	case MessageStoreMongo:
		for _, r := range req {
			res, err := RetrieveMongoUnreadCount(r)
			if err != nil {
				Logger.Error(r.Owner, r.Appid, "", "RetrieveUnreadCount err", err)
				return err
			} else {
				*resp = append(*resp, res)
			}
		}
	}
	return nil
}

/*
 * Set chat messages recall flag to storage
 * @param req is a saver.RecallMessagesRequest point which include retrieving request information
 * @param resp is a saver.StoreMessagesResponse point which includes response
 * @return nil if no error occurs, otherwise an error interface is returned
 */
func (this *GorpcService) RecallChatMessages(req *saver.RecallMessagesRequest,
	resp *saver.StoreMessagesResponse) error {
	switch netConf().MessageStore {
	case MessageStoreMongo:
		return RecallMongoChatMessages(req, resp)
	}

	err := fmt.Sprintf(SaverErrorNotSupportStorage, netConf().MessageStore)
	Logger.Error("", req.Appid, "", "RecallChatMessages", "config error", err)
	return errors.New(err)
}

func (this *GorpcService) AddChatroomCountKeys(req *saver.ChatroomCountKeysRequest, resp *int) error {
	Logger.Trace("", req.Appid, "", "AddChatroomCountKeys", req.Roomids)
	return AddChatroomCountKeys(req)
}

func (this *GorpcService) DelChatroomCountKeys(req *saver.ChatroomCountKeysRequest, resp *int) error {
	Logger.Trace("", req.Appid, "", "DelChatroomCountKeys", req.Roomids)
	return DelChatroomCountKeys(req)
}

func (this *GorpcService) GetChatroomCountKeys(req *saver.GetChatroomCountKeysRequest,
	resp *saver.ChatroomCountKeysResponse) error {
	Logger.Trace(req.ClientIP, req.Appid, "", "GetChatroomCountKeys", req.RpcIndex, req.RpcLength)
	return GetChatroomCountKeys(req, resp)
}

func (this *GorpcService) AddQpsCount(req *saver.QpsCount, resp *bool) error {
	Logger.Trace("", "", "", "AddQpsCount", *req)
	return AddQpsCount(req, resp)
}

func prop2usersession(prop map[string]string) (*session.UserSession, error) {

	isLoginUser, _ := strconv.ParseBool(prop["IsLoginUser"])
	connId, _ := strconv.ParseUint(prop["ConnectionId"], 10, 64)
	appId, _ := strconv.ParseUint(prop["AppId"], 10, 16)
	s := &session.UserSession{
		UserId:       prop["UserId"],
		UserIp:       prop["ClientIp"],
		AppId:        uint16(appId),
		IsLoginUser:  isLoginUser,
		SessionKey:   prop["SessionKey"],
		ConnectionId: logic.ConnectionId(connId),
		GatewayAddr:  prop["GatewayAddr"],
		Platform:     prop["Platform"],
		Deviceid:     prop["Deviceid"],
		TraceId:      prop["TraceId"],
		Property:     prop,
	}

	return s, nil
}
