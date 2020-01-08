package main

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/cpool"
	"github.com/johntech-o/gorpc"
)

const (
	JoinNotifyConsumerChanLen   = 50000
	JoinNotifyConsumerCount     = 50
	QuitCallbackConsumerChanLen = 50000
	QuitCallbackConsumerCount   = 50
	MaxZoneID                   = 100
)

type quitChatRoomCallback struct {
	*session.UserChatRoomRequest
	RoomIDs  []string
	Quited   bool
	ClientIP string
	ConnType string
	Platform string
	DeviceID string
}

var joinNotifyPool *cpool.ConsumerPool
var quitCallbackPool *cpool.ConsumerPool
var sessionTimeCronId chan bool
var chatRoomTimeCronId chan bool

func initCron() {
	joinNotifyPool = cpool.NewConsumerPool(JoinNotifyConsumerCount, JoinNotifyConsumerChanLen, JoinNotifyFn)
	quitCallbackPool = cpool.NewConsumerPool(QuitCallbackConsumerCount, QuitCallbackConsumerChanLen, QuitCallbackFn)
	//create gc usersession pool
	sessionTimeCronId = createTimeCron(netConf().SessionTimerHour, netConf().SessionTimerMinute, iterateUserSession)
	chatRoomTimeCronId = createTimeCron(netConf().ChatRoomTimerHour, netConf().ChatRoomTimerMinute, iterateChatRoom)
}

func JoinNotifyFn(d interface{}) {
	j, ok := d.(*logic.ChatRoomMessageNotify)
	if !ok {
		Logger.Error("", "", "", "JoinNotifyFn", "Consumer error:type not match", d)
		return
	}
	if j.ChatRoomMessage == nil {
		Logger.Warn("", "", "", "JoinNotifyFn", "ChatRoomMessageNotify.ChatRoomMessage is nil", j)
		return
	}
	if err := router.SendChatRoomNotify(j.ChatRoomMessage, j.GatewayAddrs, "JOIN-"+j.ChatRoomMessage.Sender+"-"+j.ChatRoomMessage.RoomID); err != nil {
		Logger.Error(j.Sender, j.Appid, j.GatewayAddrs, "JoinNotifyCron", "SendChatRoomNotify error", err.Error())
	}
}

func QuitCallbackFn(d interface{}) {
	req, ok := d.(*quitChatRoomCallback)
	if !ok {
		Logger.Error("", "", "", "QuitCallbackFn", "Consumer error:type not match", d)
		return
	}
	for _, room := range req.RoomIDs {
		req.RoomID = room
		callback := true
		if !req.Quited {
			if code, err := saver.RemoveChatRoomUser(req.UserChatRoomRequest); err != nil {
				Logger.Error(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "QuitChatRoomCron", "saver.RemoveChatRoomUser error:"+err.Error(), req.RoomID)
				continue // retry?
			} else {
				callback = (code == session.AllSessionQuitedChatRoom)
			}
			Logger.Trace(req.User, req.Appid, logic.GetTraceId(req.GatewayAddr, req.ConnectionId), "QuitChatRoomCron", req.RoomID, callback, req.ClientIP, req.ConnType, req.Platform, req.DeviceID)
		}
		if callback {
			if resp, err := forwardChatRoomRequest(req.UserChatRoomRequest, nil); err != nil {
				Logger.Error(req.User, req.Appid, req.RoomID, "QuitCallbackCron", "forwardChatRoomRequest error", err.Error())
			} else if resp.Code != session.Success || resp.Data.Notify == "" {
				continue
			} else {
				Logger.Debug(req.User, req.Appid, req.RoomID, "QuitCallbackCron", "forwardChatRoomRequest", resp.Code, resp)
				if _, ok := logic.NetGlobalConf().BigRoom[req.RoomID]; ok || logic.NetGlobalConf().NewChatroomSend {
					err := coordinator.ChatRoomMsg(req.Appid, req.RoomID, req.User, resp.Data.Notify, MESSAGE_TYPE_QUIT, resp.Data.Priority, 0, "QUIT_NOTIFY-"+req.RoomID)
					if err != nil {
						Logger.Error(req.User, req.Appid, "QUIT_NOTIFY", "coordinator.ChatRoomMsg", req.RoomID, err.Error())
					} else {
						Logger.Trace(req.User, req.Appid, "QUIT_NOTIFY", "coordinator.ChatRoomMsg", req.RoomID, resp.Data.Priority)
					}
					continue
				}
				detail, err := getChatRoomDetail(req.RoomID, req.Appid)
				if err != nil {
					Logger.Error(req.User, req.Appid, req.RoomID, "QuitCallbackFn", "session.getChatRoomDetail", err.Error())
					continue
				}
				importance := (resp.Data.Priority == session.HighLevelUser)
				msg := generateChatRoomNotify(req.User, req.Appid, detail, &resp.Data.Notify, importance)
				gws := logic.FilterGatewayAddrs(netConf().ChatroomNotifyPolicy, detail.ConnCount(), detail.GatewayAddrs, importance)
				if len(gws) != 0 {
					if err := router.SendChatRoomNotify(msg, detail.GatewayAddrs, "QUIT-"+msg.Sender+"-"+msg.RoomID); err != nil {
						Logger.Error(msg.Sender, msg.Appid, gws, "QuitCallbackFn", "SendChatRoomNotify error", err.Error())
					}
				}
			}
		}
	}
}

func addJoinNotify(msg *logic.ChatRoomMessage, gwAddrs map[string]int) {
	j := &logic.ChatRoomMessageNotify{msg, gwAddrs, "", 0, 0}
	if ok := joinNotifyPool.Add(j); !ok {
		Logger.Warn(msg.Sender, msg.Appid, "", "addJoinNotify", "chan full", "")
	}
}

func addQuitCallback(req *quitChatRoomCallback) {
	if ok := quitCallbackPool.Add(req); !ok {
		Logger.Warn(req.User, req.Appid, "", "addQuitCallback", "chan full", "")
	}
}

func gcUserSession(us *session.UserSession) {

	if us.GatewayAddr == "" {
		Logger.Warn("", "", "", "GcUserSessionFn", "Usersession.GatewayAddr is nil", us)
		return
	}

	cleanfun := func() {
		//below similarr session.Close
		if err := saver.RemoveSession(us); err != nil {
			Logger.Error(us.UserId, us.AppId, us.TraceId, "GcUserSessionFn", "saver.RemoveSession error", err.Error())
		}

		cleanUserSessionTags(us, false)
	}

	props, err := gateway.GetConnectionInfo(us.GatewayAddr, us.ConnectionId)
	if err != nil {
		gorpcerr, ok := err.(*gorpc.Error)
		if ok {
			if gorpcerr.Reason == "connection not found" || gorpcerr.Reason == "not a ximpconnection" {
				Logger.Trace(us.UserId, us.AppId, us.GatewayAddr, us.ConnectionId, "is invalid user, no connection")
				cleanfun()
			} else {
				Logger.Error(us.UserId, us.AppId, us.TraceId, "GcUserSessionFn", "gatway.GetConnectionInfo error", gorpcerr.Reason)
			}
		}
		return
	}

	appid := logic.StringToUint16(props["Appid"])
	if props["Sender"] != us.UserId || appid != us.AppId {
		Logger.Trace(us.UserId, us.AppId, us.GatewayAddr, us.ConnectionId, "is invalid user", props["Sender"], appid)
		cleanfun()
		return
	}

	Logger.Debug(us.UserId, us.AppId, us.TraceId, "is valid user")
}

//iterate all online usersession
func iterateUserSession() {
	sessionRpcs := logic.NetGlobalConf().SessionRpcs
	Logger.Trace("", "", "", "iterateUserSession", "check SessionRpcs len", len(sessionRpcs))
	if len(sessionRpcs) <= 0 {
		return
	}

	Logger.Trace("", "", "", "iterateUserSession", "check head session", sessionRpcs[0])
	if strings.Split(sessionRpcs[0], ":")[0] != strings.Split(NodeID, ":")[0] {
		return
	}

	for _, appidStr := range logic.NetGlobalConf().Appids {
		appid := logic.StringToUint16(appidStr)
		Logger.Trace("", appid, "", "iterateUserSession", "begin exec iterate all usersessions")

		for i := 0; i < MaxZoneID; i++ {
			usersessions, err := saver.GetActiveUserSessions(appid, uint16(i))
			if err != nil {
				Logger.Error("", appid, "", "iterateUserSession", "saver.GetActiveUserSessions error: ", err.Error())
				continue
			}

			Logger.Trace("", appid, "", "iterateUserSession", fmt.Sprintf("UserSessions[%d]=%d", i, len(usersessions)))

			for _, usersession := range usersessions {
				gcUserSession(usersession)
			}

			Logger.Trace("", appid, "", "iterateUserSession", fmt.Sprintf("UserSessions[%d] done", i))

			//time.Sleep(1 * time.Second)
		}
	}
}

func scanChatRoom(room *session.ChatRoomDetail) {
	users, err := saver.QueryChatRoomUsers(room.RoomID, room.AppID)
	if err != nil {
		Logger.Error("", "", "", "scanChatRoom", "saver.QueryChatRoomUsers error", err.Error())
		return
	}

	for _, uid := range users {
		req := &session.UserSession{
			AppId:  room.AppID,
			UserId: uid,
		}
		sessions, err := saver.QueryUserSession([]*session.UserSession{req})
		if err != nil {
			Logger.Error("", "", "", "scanChatRoom", "saver.QueryUserSession error", err.Error())
			return
		}
		for _, s := range sessions {
			r := &session.UserChatRoomRequest{
				User:         uid,
				Appid:        room.AppID,
				ConnectionId: s.ConnectionId,
				GatewayAddr:  s.GatewayAddr,
				RoomID:       room.RoomID,
			}
			b, err := saver.CheckUserSessionInRoom(r)
			if err != nil {
				Logger.Error("", "", "", "scanChatRoom", "saver.CheckUserSessionInRoom error", err.Error())
				return
			} else if b {
				Logger.Trace(s.UserId, room.AppID, room.RoomID, "saver.CheckUserSessionInRoom", "still in room")
				return
			}
		}
	}

	err = saver.CleanChatRoom(room.RoomID, room.AppID)
	if err != nil {
		Logger.Error("", "", "", "scanChatRoom", "saver.CleanChatRoom error", err.Error())
	} else {
		Logger.Trace("", room.AppID, room.RoomID, "saver.CleanChatRoom", "cleaned")
	}
}

func iterateChatRoom() {
	sessionRpcs := logic.NetGlobalConf().SessionRpcs
	Logger.Trace("", "", "", "iterateChatRoom", "check SessionRpcs len", len(sessionRpcs))
	if len(sessionRpcs) <= 0 {
		return
	}

	Logger.Trace("", "", "", "iterateChatRoom", "check head session", sessionRpcs[0])
	if strings.Split(sessionRpcs[0], ":")[0] != strings.Split(NodeID, ":")[0] {
		return
	}

	for _, appidStr := range logic.NetGlobalConf().Appids {
		appid := logic.StringToUint16(appidStr)
		Logger.Trace("", appid, "", "iterateChatRoom", "begin exec iterate all chat rooms")

		for i := 0; i < MaxZoneID; i++ {
			rooms, err := saver.QueryChatRoomsByZone(appid, uint16(i))
			if err != nil {
				Logger.Error("", appid, "", "iterateChatRoom", "saver.QueryChatRoomIds error: ", err.Error())
				continue
			}

			Logger.Trace("", appid, "", "iterateChatRoom", fmt.Sprintf("ChatRoom[%d]=%d", i, len(rooms)))

			for _, room := range rooms {
				// 聊天室建立超过24小时的
				if time.Duration(time.Now().Unix()-room.CreateTime)*time.Second > time.Duration(netConf().ChatRoomMinDuration)*time.Second {
					scanChatRoom(room)
				}
			}

			Logger.Trace("", appid, "", "iterateChatRoom", fmt.Sprintf("ChatRoom[%d] done", i))

			//time.Sleep(1 * time.Second)
		}
	}
}

func createTimeCron(hour int, minute int, callback func()) chan bool {

	timeid := make(chan bool)

	go func(h int, m int, cb func(), quit chan bool) {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				if (now.Hour() == h || -1 == h) && (now.Minute() == m || -1 == m) {
					Logger.Trace("", "", "", "createTimeCron", "TimeCron Run ", runtime.FuncForPC(reflect.ValueOf(cb).Pointer()).Name())
					cb()
				}
			case <-quit:
				return
			}
		}
	}(hour, minute, callback, timeid)

	return timeid
}
