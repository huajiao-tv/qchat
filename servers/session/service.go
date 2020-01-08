package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

const (
	MESSAGE_TYPE_JOIN = 10
	MESSAGE_TYPE_QUIT = 16
)

type partnerResponse struct {
	Code    int    `json:"errno"`
	Reason  string `json:"errmsg"`
	Consume int    `json:"consume"`
	Time    uint64 `json:"time"`
	Md5     string `json:"md5"`
	Data    struct {
		Priority int    `json:"priority"`
		Response string `json:"singlecast"`
		Notify   string `json:"multicast"`
	} `json:"data"`
}

func getCallbackUrl(req *session.UserChatRoomRequest, properties map[string]string) string {
	if req.IsJoin {
		args := make(url.Values, len(properties)+2)
		args["rid"] = []string{req.RoomID}
		args["uid"] = []string{req.User}
		for k, v := range properties {
			args[k] = []string{v}
		}
		return fmt.Sprintf("%s/join?%s", netConf().ChatroomCallbackUrl[strconv.Itoa(int(req.Appid))], args.Encode())
	} else {
		return fmt.Sprintf("%s/quit?rid=%s&uid=%s", netConf().ChatroomCallbackUrl[strconv.Itoa(int(req.Appid))], req.RoomID, req.User)
	}
}

func forwardChatRoomRequest(req *session.UserChatRoomRequest, properties map[string]string) (res *partnerResponse, err error) {
	// 针对聊天室限制回调业务的qps
	if _, ok := logic.NetGlobalConf().BigRoom[req.RoomID]; ok {
		if !CallbackThreshold.Incr() {
			// 超过阈值，丢弃回调请求
			Logger.Debug(req.User, req.Appid, req.RoomID, "session.forwardChatRoomRequest", "degraded")
			resp := &partnerResponse{
				Reason: "success",
			}
			// 兼容旧版客户端
			if properties != nil && properties["audienceflag"] == "1" {
				resp.Data.Response = "response"
			}
			return resp, nil
		}
	}

	url := getCallbackUrl(req, properties)
	start := time.Now()
	resp, err := HttpClient().Client.Get(url)
	cost := time.Now().Sub(start)
	if err != nil {
		Logger.Error(req.User, req.Appid, req.RoomID, "session.forwardChatRoomRequest err:"+err.Error(), cost)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		Logger.Error(req.User, req.Appid, req.RoomID, "session.forwardChatRoomRequest http code err:", resp.StatusCode, cost)
		return nil, errors.New("http code is not ok ")
	}
	body, err := ioutil.ReadAll(resp.Body)

	response := &partnerResponse{}
	if err = json.Unmarshal(body, response); err != nil {
		Logger.Error(req.User, req.Appid, req.RoomID, "session.forwardChatRoomRequest resp err(json,unmarshal):", err.Error(), body)
		return nil, err
	}
	Logger.Debug(req.User, req.Appid, req.RoomID, "forwardChatRoomRequest", req.IsJoin, cost, response)
	return response, nil
}

func getChatRoomDetail(room string, appid uint16) (*session.ChatRoomDetail, error) {
	detail, err := saver.QueryChatRoomDetail(room, appid)
	if err != nil {
		return nil, errors.New("saver.QueryChatRoomDetail error:" + err.Error())
	} else if detail == nil {
		return nil, errors.New("room not exist")
	}
	return detail, nil
}

func generateChatRoomNotify(user string, appid uint16, room *session.ChatRoomDetail, notify *string, priority bool) *logic.ChatRoomMessage {
	return &logic.ChatRoomMessage{
		RoomID:      room.RoomID,
		Sender:      user,
		Appid:       appid,
		MsgContent:  []byte(*notify),
		RegMemCount: room.Registered(),
		MemCount:    room.MemberCount(),
		//MsgID:       room.MaxID,
		MaxID:     room.MaxID,
		TimeStamp: time.Now().UnixNano() / 1e6,
		Priority:  priority,
	}
}

// 清理用户session，postpone表示是否延迟清理 & 回调业务
func cleanUserSessionTags(req *session.UserSession, postpone bool) ([]string, error) {
	crReq := &session.UserChatRoomRequest{
		User:         req.UserId,
		UserType:     session.GetUserType(req.UserId, req.AppId, req.Property["ConnectionType"]),
		Appid:        req.AppId,
		GatewayAddr:  req.GatewayAddr,
		ConnectionId: req.ConnectionId,
		IsJoin:       false,
	}
	crs, err := saver.QueryUserSessionChatRoomList(crReq)
	if err != nil {
		return []string{}, err
	}
	ret := make([]string, 0, len(crs))
	for _, room := range crs {
		ret = append(ret, logic.GenerateChatRoomTag(req.AppId, room))
		crReq.RoomID = room
	}
	callbackData := &quitChatRoomCallback{
		UserChatRoomRequest: crReq,
		Quited:              false,
		RoomIDs:             crs,
		ClientIP:            req.Property["ClientIp"],
		ConnType:            req.Property["ConnectionType"],
		Platform:            req.Property["Platform"] + "::" + req.Property["MobileType"],
		DeviceID:            req.Property["Deviceid"],
	}
	if postpone && netConf().PostponeCallbackDuration > 0 {
		go func() {
			time.Sleep(time.Duration(netConf().PostponeCallbackDuration) * time.Millisecond)
			addQuitCallback(callbackData)
		}()
	} else {
		addQuitCallback(callbackData)
	}
	return ret, nil
}

func doJoinChatRoom(req *session.JoinChatRoomRequest, resp *session.JoinChatRoomResponse, notify *string, priority int) int {
	code := session.UserIsRobot
	if req.UserType == session.RobotChatRoomUser {
		if err := saver.AddChatRoomRobot(req.UserChatRoomRequest); err != nil {
			resp.Code = session.AddUserFailed
			resp.Reason = err.(*gorpc.Error).Reason
			Logger.Error("saver.AddChatRoomRobot error:", err.Error())
		}
	} else {
		if ret, err := saver.AddChatRoomUser(req.UserChatRoomRequest); err != nil {
			Logger.Error("saver.AddChatRoomUser error:", err.Error())
			resp.Code = session.AddUserFailed
			resp.Reason = err.(*gorpc.Error).Reason
		} else {
			code = ret
		}
	}
	if resp.Code != session.Success || *notify == "" {
		return resp.Code
	}
	switch code {
	// 重复加入不发通知
	case session.UserAlreadyInChatRoom:
		return session.Success
	case session.SessionAlreadyInChatRoom:
		return session.SessionAlreadyInChatRoom
	default:
		if _, ok := logic.NetGlobalConf().BigRoom[req.RoomID]; ok || logic.NetGlobalConf().NewChatroomSend {
			err := coordinator.ChatRoomMsg(req.Appid, req.RoomID, req.User, *notify, MESSAGE_TYPE_JOIN, priority, 0, "JOIN_NOTIFY-"+req.RoomID)
			if err != nil {
				Logger.Error(req.User, req.Appid, "JOIN_NOTIFY", "coordinator.ChatRoomMsg", req.RoomID, err.Error())
			} else {
				Logger.Trace(req.User, req.Appid, "JOIN_NOTIFY", "coordinator.ChatRoomMsg", req.RoomID, priority)
			}
			return session.Success
		}
		detail, err := getChatRoomDetail(req.RoomID, req.Appid)
		if err != nil {
			Logger.Error(req.User, req.Appid, req.RoomID, "doJoinChatRoom", "session.getChatRoomDetail", err.Error())
			resp.Code = session.AddUserFailed
			return session.AddUserFailed
		} else {
			resp.ChatRoomDetail = detail
			Logger.Debug(req.User, req.Appid, req.RoomID, "ChatRoomMembers", priority, detail.CreateTime, detail.MaxID, detail.Members)
		}
		importance := (priority == session.HighLevelUser)
		msg := generateChatRoomNotify(req.User, req.Appid, detail, notify, importance)
		detail.GatewayAddrs = logic.FilterGatewayAddrs(netConf().ChatroomNotifyPolicy, detail.ConnCount(), detail.GatewayAddrs, importance)
		if len(detail.GatewayAddrs) != 0 {
			addJoinNotify(msg, detail.GatewayAddrs)
		}
		return session.Success
	}
}
