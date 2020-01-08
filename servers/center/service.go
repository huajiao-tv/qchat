package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
)

// For chat room
type ErrorCode int

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5741
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

const (
	errorSuccess ErrorCode = iota
	errorInvalidRequest
	errorLostArgument
	errorBadArguments
	errorUnsupportedHttpMethod

	errorInternalError
	errorMessageTooLong
	errorAccessTooFrequent
)

const (
	errorReasonSuccess               = "successful"
	errorReasonInvalidRequest        = "invalid request"
	errorReasonLostArgument          = "no arg: %s"
	errorReasonBadArguments          = "bad arguments"
	errorReasonUnsupportedHttpMethod = "unsupported http method"
	errorReasonChatRoomNotExist      = "room not exist"
	errorReasonMessageTooLong        = "too long message"
	errorReasonRequestFail           = "internal request fail"
	errorReasonUserInRoom            = "user is in room"
	errorReasonUserNotInRoom         = "user not in room"
	errorReasonAccessTooFrequent     = "access too frequent"
)

const (
	ResUserIsNotInRoom = iota
	ResUserIsInRoom
)

const (
	PriorityNormal = 0
	PriorityChat   = 1
	PriorityFaceU  = 900 // faceu礼物从其它参数里获取
)

type errorResponse struct {
	Code   ErrorCode   `json:"code"`
	Reason string      `json:"reason"`
	Data   interface{} `json:"data"`
}

func (e *errorResponse) Error() string {
	if data, err := json.Marshal(*e); err != nil {
		return fmt.Sprintf("{\"code\":%d, \"reason\":\"%s\"}", errorInternalError, err.Error())
	} else {
		return string(data)
	}
}

func checkArguments(form *url.Values, res *errorResponse, args ...string) (*errorResponse, bool) {
	for _, arg := range args {
		if _, ok := (*form)[arg]; !ok {
			res.Code, res.Reason = errorLostArgument, fmt.Sprintf(errorReasonLostArgument, arg)
			return res, false
		}
	}
	return res, true
}

func newChatRoomMessage(sender string, room *session.ChatRoomDetail, msgtype int, content string, importance bool) *logic.ChatRoomMessage {
	return &logic.ChatRoomMessage{
		RoomID:      room.RoomID,
		Appid:       room.AppID,
		Sender:      sender,
		MsgType:     0, // msgtype,
		MsgContent:  []byte(content),
		RegMemCount: room.Registered(),
		MemCount:    room.MemberCount(),
		TimeStamp:   time.Now().UnixNano() / 1e6,
		MaxID:       room.MaxID,
		Priority:    importance,
	}
}

func forwardChatRoomJoinNotify(user string, room string, appid uint16, content string) error {
	detail, err := saver.QueryChatRoomDetail(room, appid)
	if err != nil {
		return err
	} else if detail == nil {
		Logger.Trace(user, appid, "", "forwardChatRoomJoinNotify", room+" not exist")
		return nil
	}
	message := newChatRoomMessage(user, detail, pb.CR_PAYLOAD_JOIN, content, false)
	router.SendChatRoomNotify(message, detail.GatewayAddrs, "FORWARD_NOTIFY")
	return nil
}

func newBroadcastRoomMessage(sender string, appid uint16, msgtype int, content string, importance bool) *logic.ChatRoomMessage {
	return &logic.ChatRoomMessage{
		RoomID:      strconv.Itoa(int(appid)),
		Appid:       appid,
		Sender:      sender,
		MsgType:     msgtype,
		MsgContent:  []byte(content),
		RegMemCount: 0,
		MemCount:    0,
		TimeStamp:   time.Now().UnixNano() / 1e6,
		MaxID:       0,
		Priority:    importance,
	}
}

// priority 可能取值范围 加入退出消息（0），聊天消息（1），礼物消息（101），红包消息（201）
// ChatroomDegradePolicy 配置项目 DegradeToZero(0)，DegradeChatToZero（1），DegradeGiftToZero（101），DegradeAllToZero（201）
func checkPolicy(room *session.ChatRoomDetail, priority int) bool {
	switch netConf().ChatroomDegradePolicy {
	case PriorityNormal:
		if room.ConnCount() < netConf().ChatroomDegradeMembers {
			return priority > PriorityNormal
		} else if priority > PriorityChat {
			return true
		}
	default:
		if priority > netConf().ChatroomDegradePolicy {
			return true
		}
	}
	return false
}

type MessagesPolicyStruct struct {
	DiscardPolicy map[string]map[int]int
	DelayPolicy   map[string]map[int]int
	mutex         *sync.RWMutex
}

var MessagesPolicy = &MessagesPolicyStruct{
	DiscardPolicy: make(map[string]map[int]int),
	DelayPolicy:   make(map[string]map[int]int),
	mutex:         &sync.RWMutex{},
}

// parse configures like following format
// eg: cr_drop_msgs_dt_policy string = 30-102:1000|5,5000|35;30-103:1000|10,5000|25
func parsePolicyDetail(conf string) map[string]map[int]int {
	result := make(map[string]map[int]int)

	// example: 30-102:1000|5,5000|35;30-103:1000|10,5000|25

	configItemStrArray := strings.Split(conf, ";")

	// []string{"30-102:1000|5,5000|35", "30-103:1000|10,5000|25"}
	for _, configItemStr := range configItemStrArray {
		configKvArray := strings.Split(configItemStr, ":")
		if len(configKvArray) != 2 {
			continue
		}

		// tp: 30-102; tpcvs: 1000|5,5000|35
		tp, tpcs := configKvArray[0], configKvArray[1]

		// []string{"1000|5", "5000|35"}
		tpca := strings.Split(tpcs, ",")
		if len(tpca) == 0 {
			continue
		}

		for _, tpc := range tpca {
			tmp := strings.Split(tpc, "|")
			if len(tmp) != 2 {
				continue
			}

			numofppl, _ := strconv.Atoi(tmp[0])
			percentage, _ := strconv.Atoi(tmp[1])

			if _, ok := result[tp]; !ok {
				result[tp] = make(map[int]int)
			}

			result[tp][numofppl] = percentage
		}
	}

	return result
}

// true 丢弃；false 放行
func checkDiscardMessagesDetailPolicy(msgtype, priority int, members int) bool {
	MessagesPolicy.mutex.RLock()
	policy := MessagesPolicy.DiscardPolicy
	MessagesPolicy.mutex.RUnlock()

	var useDefault bool
	key := fmt.Sprintf("%v-%v", msgtype, priority)
	if _, ok := policy[key]; !ok {
		// 默认配置
		useDefault = true
	}

	var percentage int
	switch useDefault {
	case true:
		for numofppl, p := range netConf().CrDropMsgsDtDefault {
			if members >= numofppl && p > percentage {
				percentage = p
			}
		}

	case false:
		var percentageKey int
		for numofppl, _ := range policy[key] {
			if members >= numofppl && numofppl > percentageKey {
				percentageKey = numofppl
			}
		}

		percentage = policy[key][percentageKey]
	}

	// 按千分比丢弃
	ri := rand.Intn(1000)
	if ri < percentage {
		return true
	}

	return false
}

func getDelayInterval(msgtype, priority, members int) int {
	MessagesPolicy.mutex.RLock()
	policy := MessagesPolicy.DelayPolicy
	MessagesPolicy.mutex.RUnlock()

	key := fmt.Sprintf("%v-%v", msgtype, priority)
	if _, ok := policy[key]; !ok {
		// 不在延迟发送列表中，走原有的降级配置
		return 0
	}
	var delayKey int
	for numofppl, _ := range policy[key] {
		if members >= numofppl && numofppl > delayKey {
			delayKey = numofppl
		}
	}
	return policy[key][delayKey]
}

func isSpecialMsg(msgtype, priority int) bool {
	key := fmt.Sprintf("%v-%v", msgtype, priority)
	_, ok := netConf().SpecialMessageType[key]
	return ok
}

func sentByCoordinator(roomid string) bool {
	_, ok := logic.NetGlobalConf().BigRoom[roomid]
	return (ok || logic.NetGlobalConf().NewChatroomSend)
}

func sendChatRoomMessage(req *http.Request, res *errorResponse) {
	uid := req.FormValue("sender")
	traceid := req.FormValue("traceid")
	if traceid == "" { // 经济系统调用使用的是 sn
		traceid = req.FormValue("sn")
	}
	roomid := req.FormValue("roomid")
	content := req.FormValue("content")
	msgid, _ := strconv.Atoi(req.FormValue("msgid"))
	msgtype, _ := strconv.Atoi(req.FormValue("type"))
	delay, _ := strconv.Atoi(req.FormValue("delay"))
	// 最高延迟 300 秒发送
	if delay > netConf().MessageMaxDelay {
		delay = netConf().MessageMaxDelay
	}
	// 消息类型和优先级，类型暂时不传，优先级：加入退出消息（0），聊天消息（1），礼物消息（101），红包消息（201）
	priority, _ := strconv.Atoi(req.FormValue("priority"))
	appid := logic.StringToUint16(req.FormValue("appid"))
	if req.FormValue("isFaceYGift") == "1" {
		priority = PriorityFaceU
	}
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if len(content) > netConf().MaxMsgSize {
		res.Code, res.Reason = errorMessageTooLong, errorReasonMessageTooLong
		Logger.Error(uid, appid, traceid, "center.SendChatRoomMessageHandler", roomid, msgtype, priority, content, "content too long")
		return
	}

	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countSendRespTime(uid+"@"+roomid, traceid, "SendChatRoomMessageHandler", appid)
		defer countFunc()
	}

	if isSpecialMsg(msgtype, priority) {
		if detail, err := saver.QueryChatRoomDetail(roomid, appid); err == nil && detail != nil {
			interval := getDelayInterval(msgtype, priority, detail.ConnCount())
			message := newChatRoomMessage(uid, detail, msgtype, content, true)
			if err := router.SendChatRoomNotifyWithDelay(message, detail.GatewayAddrs, traceid, time.Duration(delay)*time.Second, time.Duration(interval)*time.Millisecond); err != nil {
				Logger.Error(uid, appid, traceid, "router.SendChatRoomNotify error", roomid, msgtype, priority, err)
				res.Code, res.Reason = errorInternalError, errorReasonRequestFail
			}
			Logger.Trace(uid, appid, traceid, "center.SendChatRoomMessageHandler delay", detail.ConnCount(), roomid, priority, content, msgtype, delay, logic.ClientIp(req))
			return
		} else {
			// 没取到聊天室信息，继续原来逻辑
		}
	}

	if sentByCoordinator(roomid) {
		err := coordinator.ChatRoomMsg(appid, roomid, uid, content, msgtype, priority, uint(msgid), traceid)
		if err != nil {
			Logger.Error(roomid, appid, traceid, "SendChatRoomMessageHandler", "coordinator.ChatRoomMsg", err.Error())
		} else {
			Logger.Trace(uid, appid, traceid, "center.SendChatRoomMessageHandler", "coordinator", roomid, priority, content, msgtype, logic.ClientIp(req))
		}
		return
	}

	detail, err := saver.QueryChatRoomDetail(roomid, appid)
	if err != nil {
		Logger.Error(uid, appid, traceid, "saver.QueryChatRoomDetail error", roomid, msgtype, priority, err.Error())
		res.Code, res.Reason = errorInternalError, err.Error()
	} else if detail == nil {
		Logger.Error(uid, appid, traceid, "saver.QueryChatRoomDetail error", roomid, msgtype, priority, "not exist")
		res.Code, res.Reason = errorInternalError, errorReasonChatRoomNotExist
	} else {
		members := detail.ConnCount()

		// 降级策略，根据人数丢弃优先级低于某个值的消息
		p0 := 0
		for ts, p := range netConf().ChatroomDiscardMessages {
			if members > ts && p > p0 {
				p0 = p
			}
		}
		if priority < p0 {
			Logger.Warn(uid, appid, traceid, "center.sendChatRoomMessage", roomid, msgtype, priority, content, "discard msg level 1")
			return
		}

		// 降级策略：根据类型－优先级－人数，按千分比配置丢弃消息
		if checkDiscardMessagesDetailPolicy(msgtype, priority, members) {
			Logger.Warn(uid, appid, traceid, "center.sendChatRoomMessage", roomid, msgtype, priority, content, "discard msg level 2")
			return
		}

		// 缓存消息
		importance := checkPolicy(detail, priority)
		message := newChatRoomMessage(uid, detail, msgtype, content, importance)
		if importance {
			message.MsgID = uint(msgid) // 新旧系统切换双发兼容
			if msgid, err := saver.CacheChatRoomMessage(message); err != nil {
				Logger.Error(uid, appid, traceid, "saver.CacheChatRoomMessage error", roomid, msgtype, priority, msgid, err)
				res.Reason = err.Error()
			} else {
				message.MsgID, message.MaxID = msgid, msgid
			}
		}
		detail.GatewayAddrs = logic.FilterGatewayAddrs(netConf().ChatroomMessagePolicy, detail.ConnCount(), detail.GatewayAddrs, importance)
		if len(detail.GatewayAddrs) > 0 {
			if err := router.SendChatRoomNotify(message, detail.GatewayAddrs, traceid); err != nil {
				Logger.Error(uid, appid, traceid, "router.SendChatRoomNotify error", roomid, err)
				res.Code, res.Reason = errorInternalError, errorReasonRequestFail
				return
			}
		}
		Logger.Trace(uid, appid, traceid, "center.SendChatRoomMessageHandler", detail.ConnCount(), roomid, priority, content, msgtype, logic.ClientIp(req))
	}
}

func queryUserSessionChatRooms(user *session.UserSession) error {
	req := &session.UserChatRoomRequest{
		User:         user.UserId,
		Appid:        user.AppId,
		GatewayAddr:  user.GatewayAddr,
		ConnectionId: user.ConnectionId,
	}
	rooms, err := saver.QueryUserSessionChatRoomList(req)
	if err != nil {
		Logger.Error(user.UserId, user.AppId, "", "queryUserSessionChatRooms", "saver.QueryUserSessionChatRoomList error"+err.Error())
		return err
	} else {
		user.Property["ChatRooms"] = strings.Join(rooms, ",")
	}
	return nil
}

func sendChatRoomMessageRaw(req *http.Request, res *errorResponse) {
	query := req.URL.Query()

	uid := query.Get("sender")
	roomid := query.Get("roomid")
	traceid := query.Get("traceid")
	msgid, _ := strconv.Atoi(query.Get("msgid"))
	appid := logic.StringToUint16(query.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	// 消息类型和优先级，用于之后的降级处理
	msgtype, _ := strconv.Atoi(query.Get("type"))
	priority, _ := strconv.Atoi(query.Get("priority"))

	if uid == "" || roomid == "" {
		res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
		return
	}

	// 读取 body
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
		Logger.Error(uid, appid, traceid, "center.SendChatRoomMessageRawHandler", roomid, msgtype, priority, len(buf), "read body failed")
		return
	}
	content := string(buf)

	// 判断长度
	if len(content) > netConf().MaxMsgSize {
		res.Code, res.Reason = errorMessageTooLong, errorReasonMessageTooLong
		Logger.Error(uid, appid, traceid, "center.SendChatRoomMessageRawHandler", roomid, msgtype, priority, len(content), "content too long")
		return
	}

	// 发送到 corrdinator
	if sentByCoordinator(roomid) {
		err := coordinator.ChatRoomMsg(appid, roomid, uid, content, msgtype, priority, uint(msgid), traceid)
		if err != nil {
			Logger.Error(roomid, appid, traceid, "SendChatRoomMessageRawHandler", "coordinator.ChatRoomMsg", err.Error())
		} else {
			Logger.Trace(uid, appid, traceid, "center.SendChatRoomMessageRawHandler", "coordinator", roomid, priority, len(content), fmt.Sprintf("%X", content), msgtype, logic.ClientIp(req))
		}
		return
	}

	detail, err := saver.QueryChatRoomDetail(roomid, appid)
	if err != nil {
		Logger.Error(uid, appid, traceid, "saver.QueryChatRoomDetail error", roomid, msgtype, priority, err.Error())
		res.Code, res.Reason = errorInternalError, err.Error()
	} else if detail == nil {
		Logger.Error(uid, appid, traceid, "saver.QueryChatRoomDetail error", roomid, msgtype, priority, "not exist")
		res.Code, res.Reason = errorInternalError, errorReasonChatRoomNotExist
	} else {
		message := newChatRoomMessage(uid, detail, msgtype, content, false)
		if len(detail.GatewayAddrs) > 0 {
			if err := router.SendChatRoomNotify(message, detail.GatewayAddrs, traceid); err != nil {
				Logger.Error(uid, appid, traceid, "router.SendChatRoomNotify error", roomid, err)
				res.Code, res.Reason = errorInternalError, errorReasonRequestFail
				return
			}
		}
		Logger.Trace(uid, appid, traceid, "center.SendChatRoomMessageRawHandler", detail.ConnCount(), roomid, priority, len(content), fmt.Sprintf("%X", content), msgtype, logic.ClientIp(req))
	}
}

func privateSendChatRoomMessage(req *http.Request, res *errorResponse) {
	// extract params
	parameters, extractRes := GetChatRequestParameter(req)
	if extractRes.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(extractRes.Code), extractRes.Reason
		Logger.Error("", "", "", "privateSendChatRoomMessage", res)
		return
	}

	if err, ok := checkArguments(parameters, res, "sender", "receiver", "content", "roomid"); !ok {
		res = err
		Logger.Error("", "", "", "privateSendChatRoomMessage", res)
		return
	}

	traceid := parameters.Get("traceid")
	sender := parameters.Get("sender")
	receiver := parameters.Get("receiver")
	if traceid == "" { // 经济系统调用使用的是 sn
		traceid = parameters.Get("sn")
	}
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	content := parameters.Get("content")
	roomid := parameters.Get("roomid")
	msgtype, _ := strconv.Atoi(parameters.Get("type"))

	// get all receiver's gateways in the room
	userGateways, err := getUserGatewaysInRoom(appid, roomid, receiver)
	if err != nil {
		Logger.Error("sender:", sender, "receiver:", receiver, appid, "privateSendChatRoomMessage", "getUserSessionsInRoom error", err)
		res.Code, res.Reason = errorInternalError, err.Error()
		return
	}
	if len(userGateways) == 0 {
		Logger.Debug("sender:", sender, "receiver:", receiver, appid, "privateSendChatRoomMessage", "getUserSessionsInRoom debug", "receiver not in room:", roomid)
		res.Code, res.Reason = errorBadArguments, errorReasonUserNotInRoom
		return
	}

	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countSendRespTime(sender+"@"+roomid, traceid, "SendChatRoomMessageHandler", appid)
		defer countFunc()
	}

	// send message by connectionid and gateway
	detail, err := saver.QueryChatRoomDetail(roomid, appid)
	if err != nil {
		Logger.Error(sender, appid, traceid, "saver.QueryChatRoomDetail error", roomid, err.Error())
		res.Code, res.Reason = errorInternalError, err.Error()
		return
	}

	message := newChatRoomMessage(sender, detail, msgtype, content, false)

	if err := router.PrivateSendChatRoomNotify(message, userGateways, traceid); err != nil {
		Logger.Error(sender, appid, traceid, "router.PrivateSendChatRoomNotify error", roomid, err)
		res.Code, res.Reason = errorInternalError, errorReasonRequestFail
		return
	} else {
		Logger.Trace(sender, appid, traceid, "center.PrivateSendChatRoomMessageHandler", roomid, content, logic.ClientIp(req))
	}
}

func privateSendChatRoomMessageRaw(req *http.Request, res *errorResponse) {
	query := req.URL.Query()

	sender := query.Get("sender")
	roomid := query.Get("roomid")
	traceid := query.Get("traceid")
	receiver := query.Get("receiver")
	msgtype, _ := strconv.Atoi(query.Get("type"))
	appid := logic.StringToUint16(query.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}

	if sender == "" || roomid == "" || receiver == "" {
		res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
		return
	}

	// 读取 body
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
		Logger.Error(sender, appid, traceid, "center.PrivateSendChatRoomMessageRawHandler", roomid, msgtype, len(buf), "read body failed")
		return
	}
	content := string(buf)

	// get all receiver's gateways in the room
	userGateways, err := getUserGatewaysInRoom(appid, roomid, receiver)
	if err != nil {
		Logger.Error("sender:", sender, "receiver:", receiver, appid, "privateSendChatRoomMessageRaw", "getUserSessionsInRoom error", err)
		res.Code, res.Reason = errorInternalError, err.Error()
		return
	}
	if len(userGateways) == 0 {
		Logger.Debug("sender:", sender, "receiver:", receiver, appid, "privateSendChatRoomMessageRaw", "getUserSessionsInRoom debug", "receiver not in room:", roomid)
		res.Code, res.Reason = errorBadArguments, errorReasonUserNotInRoom
		return
	}

	// count request response time if need
	if netConf().StatResponseTime {
		countFunc := countSendRespTime(sender+"@"+roomid, traceid, "SendChatRoomMessageHandler", appid)
		defer countFunc()
	}

	// send message by connectionid and gateway
	detail, err := saver.QueryChatRoomDetail(roomid, appid)
	if err != nil {
		Logger.Error(sender, appid, traceid, "saver.QueryChatRoomDetail error", roomid, err.Error())
		res.Code, res.Reason = errorInternalError, err.Error()
		return
	}

	message := newChatRoomMessage(sender, detail, msgtype, content, false)

	if err := router.PrivateSendChatRoomNotify(message, userGateways, traceid); err != nil {
		Logger.Error(sender, appid, traceid, "router.PrivateSendChatRoomNotify error", roomid, err)
		res.Code, res.Reason = errorInternalError, errorReasonRequestFail
	} else {
		Logger.Trace(sender, appid, traceid, "center.PrivateSendChatRoomMessageRawHandler", roomid, len(content), fmt.Sprintf("%X", content), logic.ClientIp(req))
	}
}

func getUserGatewaysInRoom(appid uint16, roomid, uid string) ([]*logic.UserGateway, error) {
	// get all user's session
	querySession := []*session.UserSession{
		&session.UserSession{
			UserId: uid,
			AppId:  appid,
		},
	}
	sessions, err := session.Query(querySession)

	// filter the sessions that in the room
	res := make([]*logic.UserGateway, 0, 8)
	if err != nil {
		Logger.Error("", "receiver:", uid, appid, "", "getUserGatewaysInRoom", "session.Query error", err)
		return nil, err
	} else {
		for _, s := range sessions {
			req := &session.UserChatRoomRequest{
				User:         uid,
				Appid:        appid,
				ConnectionId: s.ConnectionId,
				GatewayAddr:  s.GatewayAddr,
				RoomID:       roomid,
			}
			isIn, err := saver.CheckUserSessionInRoom(req)

			if err != nil {
				Logger.Error("", "receiver:", uid, appid, "", "getUserSessionsInRoom", "saver.CheckUserSessionInRoom error", err)
				return nil, err
			} else if isIn {
				res = append(res, &logic.UserGateway{GatewayAddr: s.GatewayAddr, ConnId: s.ConnectionId})
			}
		}
	}

	return res, nil
}
