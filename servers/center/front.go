package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/adminJson"
)

func FrontServer() {
	//	r := mux.NewRouter()
	r := http.NewServeMux()

	r.HandleFunc("/", NotFound)
	r.HandleFunc("/status.php", LvsCheckHandler)

	// APIs for chat message
	r.HandleFunc("/huajiao/all", HuajiaoPublicAndHotHandler)
	r.HandleFunc("/huajiao/hot", HuajiaoPublicAndHotHandler)
	r.HandleFunc("/huajiao/users", HuajiaoUsersHandler)
	r.HandleFunc("/huajiao/users/unread", HuajiaoUsersUnreadHandler)
	r.HandleFunc("/huajiao/chat", HuajiaoUsersHandler)
	r.HandleFunc("/push/chat", HuajiaoUsersHandler)
	r.HandleFunc("/huajiao/recall", HuajiaoRecallHandler)
	r.HandleFunc("/push/recall", HuajiaoRecallHandler)
	r.HandleFunc("/push/notification", PushNoticationHandler)
	r.HandleFunc("/public/notification", PushPublicNoticationHandler)

	r.HandleFunc("/operation/retrieve", HuajiaoRetrieveHandler)

	r.HandleFunc("/user/info", UserInfoHandler)
	r.HandleFunc("/online/len", OnlineLenHandler)

	// 向所有连接发push, (不存mongo)
	r.HandleFunc("/online/broadcast", BroadcastOnlineUsersHandler)

	// APIs for chat room
	r.HandleFunc("/chatroom/degrade", ChatRoomDegradeHandler)
	r.HandleFunc("/chatroom/degradedlist", ChatRoomDegradedListHandler)
	r.HandleFunc("/chatroom/len", ChatRoomLenHandler)
	r.HandleFunc("/chatroom/create", CreateChatRoomHandler)
	r.HandleFunc("/chatroom/join", JoinChatRoomHandler)
	r.HandleFunc("/chatroom/quit", QuitChatRoomHandler)
	r.HandleFunc("/chatroom/send", SendChatRoomMessageHandler)
	r.HandleFunc("/chatroom/send_raw", SendChatRoomMessageRawHandler)
	r.HandleFunc("/chatroom/send/high/priority", SendChatRoomHighPriorityHandler)
	r.HandleFunc("/chatroom/broadcast", BroadcastChatRoomHandler)
	r.HandleFunc("/chatroom/broadcast_raw", BroadcastChatRoomRawHandler)
	r.HandleFunc("/chatroom/query/member_count", QueryChatRoomMemberCountHandler)
	r.HandleFunc("/chatroom/query/member_detail", QueryChatRoomMemberDetailHandler)
	r.HandleFunc("/chatroom/query/userinroom", QueryUserInRoomHandler)
	r.HandleFunc("/chatroom/query/usersessioninroom", QueryUserSessionInRoomHandler)
	r.HandleFunc("/chatroom/private/send", PrivateSendChatRoomMessageHandler)
	r.HandleFunc("/chatroom/private/send_raw", PrivateSendChatRoomMessageRawHandler)
	// 兼容接口
	r.HandleFunc("/joinroom", JoinChatRoomHandler)
	r.HandleFunc("/quitroom", QuitChatRoomHandler)
	r.HandleFunc("/send", SendChatRoomMessageHandler)
	r.HandleFunc("/querylist", QueryChatRoomMemberCountHandler)
	r.HandleFunc("/querymemcount", QueryChatRoomMemberDetailHandler)

	// APIS for live notify
	r.HandleFunc("/live/start", LiveHandler)
	r.HandleFunc("/live/stop", LiveHandler)

	if len(netConf().MultiListen) != 0 {
		for _, addr := range netConf().MultiListen {
			Logger.Trace("front listen(m)", addr)
			go func(a string) {
				http.ListenAndServe(a, r)
				panic("invalid front listen " + a)
			}(addr)
		}
	} else if netConf().Listen != "" {
		Logger.Trace("front listen", netConf().Listen)
		http.ListenAndServe(netConf().Listen, r)
		panic("invalid front listen" + netConf().Listen)
	} else {
		panic("empty listen")
	}
}

// 如果把stopLvsLock设置成1，表示从lvs上下线
var stopLvsLock int32 = 0

func isStopLvs() bool {
	return atomic.LoadInt32(&stopLvsLock) != 0
}

func downLvs() {
	atomic.StoreInt32(&stopLvsLock, 1)
}

func upLvs() {
	atomic.StoreInt32(&stopLvsLock, 0)
}

func LvsCheckHandler(w http.ResponseWriter, req *http.Request) {
	if isStopLvs() {
		io.WriteString(w, "fail\n")
	} else {
		io.WriteString(w, "ok\n")
	}
}

func NotFound(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, adminJson.FmtJson(404, "page not found", ""))
}

//请求参数需要md5校验
func CreateChatRoomHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		w.Write([]byte(res.Error()))
	}()

	parameters, resn := GetChatRequestParameter(req)
	if resn.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(resn.Code), resn.Reason
		Logger.Error("", "", "", "center.CreateChatRoomHandler", res, logic.ClientIp(req))
		return
	}
	if err, ok := checkArguments(parameters, res, "rid"); !ok {
		res = err
		Logger.Error("", "", "", "center.CreateChatRoomHandler", res, logic.ClientIp(req))
		return
	}
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	request := &session.UserChatRoomRequest{
		Appid:  appid,
		RoomID: parameters.Get("rid"),
	}

	err := saver.CreateChatRoom(request)
	if err != nil {
		res.Code, res.Reason = errorInternalError, err.Error()
	}
	Logger.Trace("", req.FormValue("appid"), "", "center.CreateChatRoomHandler", req.FormValue("rid"), res.Code, logic.ClientIp(req))
	return
}

func JoinChatRoomHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "rid", "uid"); !ok {
			res = err
			Logger.Error("", "", "", "center.JoinChatRoomHandler", res)
			break
		}

		// count request response time if need
		if netConf().StatResponseTime {
			appid, err := strconv.ParseUint(req.FormValue("appid"), 10, 16)
			if err != nil || appid == 0 {
				appid = uint64(logic.DEFAULT_APPID)
			}
			countFunc := countJoinRespTime(req.FormValue("uid"), "", "JoinChatRoomHandler", uint16(appid))
			defer countFunc()
		}

		//机器人其他请求参数
		properties := map[string]string{}
		for k, values := range req.Form {
			if k == "uid" || k == "appid" || k == "rid" {
				continue
			}
			v := ""
			if len(values) > 0 {
				v = values[0]
			}
			properties[k] = v
		}

		resp, err := session.AddRobotIntoChatRoom(req.FormValue("uid"), req.FormValue("appid"), req.FormValue("rid"), properties)
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
		} else {
			res.Code, res.Reason = ErrorCode(resp.Code), resp.Reason
		}
		Logger.Trace(req.FormValue("uid"), req.FormValue("appid"), "", "center.JoinChatRoomHandler", req.FormValue("rid"), resp.Code, logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.JoinChatRoomHandler", res, logic.ClientIp(req))
	}

	// count request
	if res.Code == errorSuccess {
		requestStat.AtomicAddJoins(1)
	} else {
		requestStat.AtomicAddJoinFails(1)
	}

	w.Write([]byte(res.Error()))
}

func QuitChatRoomHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "rid", "uid"); !ok {
			res = err
			Logger.Error("", "", "", "center.QuitChatRoomHandler", res)
			break
		}

		// count request response time if need
		if netConf().StatResponseTime {
			appid, err := strconv.ParseUint(req.FormValue("appid"), 10, 16)
			if err != nil || appid == 0 {
				appid = uint64(logic.DEFAULT_APPID)
			}
			countFunc := countQuitRespTime(req.FormValue("uid"), "", "QuitChatRoomHandler", uint16(appid))
			defer countFunc()
		}

		if resp, err := session.RemoveRobotFromChatRoom(req.FormValue("uid"), req.FormValue("appid"), req.FormValue("rid")); err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
		} else {
			res.Code, res.Reason = ErrorCode(resp.Code), resp.Reason
		}
		Logger.Trace(req.FormValue("uid"), req.FormValue("appid"), "", "center.QuitChatRoomHandler", req.FormValue("rid"), logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.QuitChatRoomHandler", res, logic.ClientIp(req))
	}

	// count request
	if res.Code == errorSuccess {
		requestStat.AtomicAddQuits(1)
	} else {
		requestStat.AtomicAddQuitFails(1)
	}

	w.Write([]byte(res.Error()))
}

func SendChatRoomMessageHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		// 原系统使用的是sn，但业务一直只传了traceid
		if err, ok := checkArguments(&req.Form, res, "sender", "content", "roomid"); !ok {
			res = err
			Logger.Error("", "", "", "center.SendChatRoomMessageHandler", res, logic.ClientIp(req))
			break
		}
		sendChatRoomMessage(req, res)
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.SendChatRoomMessageHandler", res, logic.ClientIp(req))
	}

	// count request
	if res.Code == errorSuccess {
		requestStat.AtomicAddSends(1)
	} else {
		requestStat.AtomicAddSendFails(1)
	}

	w.Write([]byte(res.Error()))
}

func SendChatRoomMessageRawHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodPost:
		sendChatRoomMessageRaw(req, res)
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.SendChatRoomPBMessageHandler", res, logic.ClientIp(req))
	}

	// count request
	if res.Code == errorSuccess {
		requestStat.AtomicAddSends(1)
	} else {
		requestStat.AtomicAddSendFails(1)
	}

	w.Write([]byte(res.Error()))
}

func BroadcastChatRoomHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		// count request
		if res.Code == errorSuccess {
			requestStat.AtomicAddBroadcast(1)
		} else {
			requestStat.AtomicAddBroadcastFails(1)
		}

		w.Write([]byte(res.Error()))
	}()
	parameters, resn := GetChatRequestParameter(req)
	if resn.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(resn.Code), resn.Reason
		Logger.Error("", "", "", "center.BroadcastChatRoomHandler", res)
		return
	}
	// @参数列表与 send 接口一致
	if err, ok := checkArguments(parameters, res, "sender", "content"); !ok {
		res = err
		Logger.Error("", "", "", "center.BroadcastChatRoomHandler", res)
		return
	}
	uid := parameters.Get("sender")
	roomid := parameters.Get("roomid")
	traceid := parameters.Get("traceid")
	if traceid == "" { // 经济系统调用使用的是 sn
		traceid = parameters.Get("sn")
	}
	content := parameters.Get("content")
	// 消息类型和优先级，类型暂时不传，优先级：加入退出消息（0），聊天消息（1），礼物消息（101），红包消息（201）
	msgtype, _ := strconv.Atoi(parameters.Get("type"))
	priority, _ := strconv.Atoi(parameters.Get("priority"))
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if len(content) > netConf().MaxMsgSize {
		res.Code, res.Reason = errorMessageTooLong, errorReasonMessageTooLong
		Logger.Error(uid, appid, traceid, "center.BroadcastChatRoomHandler", roomid, priority, content, "content too long")
		return
	}
	importance := true
	gws := logic.NetGlobalConf().GatewayRpcs
	if netConf().ChatroomBroadcastPolicy >= 100 {
		gws = []string{}
	} else if netConf().ChatroomBroadcastPolicy != 0 {
		start := rand.Intn(len(gws))
		end := (len(gws)*netConf().ChatroomBroadcastPolicy/100 + start) % len(gws)
		if start < end {
			res := gws[0:start]
			gws = append(res, gws[end:len(gws)]...)
		} else {
			gws = gws[end:start]
		}
	}
	gwAddrs := make(map[string]int, len(gws))
	for _, gw := range gws {
		gwAddrs[gw] = 1
	}
	if len(gwAddrs) != 0 {
		message := newBroadcastRoomMessage(uid, appid, msgtype, content, importance)
		if err := router.SendChatRoomBroadcast(message, gwAddrs, traceid); err != nil {
			Logger.Error(uid, appid, traceid, "router.SendChatRoomBroadcast", roomid, priority, content, err.Error(), logic.ClientIp(req))
			res.Code, res.Reason = errorInternalError, errorReasonRequestFail
		} else {
			Logger.Trace(uid, appid, traceid, "center.BroadcastChatRoomHandler", roomid, priority, content, logic.ClientIp(req))
		}
	}
	return
}

func BroadcastChatRoomRawHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		// count request
		if res.Code == errorSuccess {
			requestStat.AtomicAddBroadcast(1)
		} else {
			requestStat.AtomicAddBroadcastFails(1)
		}

		w.Write([]byte(res.Error()))
	}()

	query := req.URL.Query()

	uid := query.Get("sender")
	traceid := query.Get("traceid")
	appid := logic.StringToUint16(query.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	// 消息类型和优先级，用于之后的降级处理
	msgtype, _ := strconv.Atoi(query.Get("type"))
	priority, _ := strconv.Atoi(query.Get("priority"))

	if uid == "" {
		res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
		return
	}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
		Logger.Error(uid, appid, traceid, "center.BroadcastChatRoomRawHandler", priority, len(buf), "read body failed")
		return
	}
	content := string(buf)

	if len(content) > netConf().MaxMsgSize {
		res.Code, res.Reason = errorMessageTooLong, errorReasonMessageTooLong
		Logger.Error(uid, appid, traceid, "center.BroadcastChatRoomRawHandler", priority, len(content), "content too long")
		return
	}
	importance := true
	gws := logic.NetGlobalConf().GatewayRpcs
	if netConf().ChatroomBroadcastPolicy >= 100 {
		gws = []string{}
	} else if netConf().ChatroomBroadcastPolicy != 0 {
		start := rand.Intn(len(gws))
		end := (len(gws)*netConf().ChatroomBroadcastPolicy/100 + start) % len(gws)
		if start < end {
			res := gws[0:start]
			gws = append(res, gws[end:len(gws)]...)
		} else {
			gws = gws[end:start]
		}
	}
	gwAddrs := make(map[string]int, len(gws))
	for _, gw := range gws {
		gwAddrs[gw] = 1
	}
	if len(gwAddrs) != 0 {
		message := newBroadcastRoomMessage(uid, appid, msgtype, content, importance)
		if err := router.SendChatRoomBroadcast(message, gwAddrs, traceid); err != nil {
			Logger.Error(uid, appid, traceid, "router.SendChatRoomBroadcast", priority, content, err.Error(), logic.ClientIp(req))
			res.Code, res.Reason = errorInternalError, errorReasonRequestFail
		} else {
			Logger.Trace(uid, appid, traceid, "center.BroadcastChatRoomRawHandler", priority, len(content), fmt.Sprintf("%X", content), logic.ClientIp(req))
		}
	}
	return
}

func BroadcastOnlineUsersHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		// count request
		if res.Code == errorSuccess {
			requestStat.AtomicAddOnlineBroadcast(1)
		} else {
			requestStat.AtomicAddOnlineBroadcastFails(1)
		}
		w.Write([]byte(res.Error()))
	}()
	parameters, resn := GetChatRequestParameter(req)
	if resn.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(resn.Code), resn.Reason
		Logger.Error("", "", "", "center.BroadcastChatRoomHandler", res)
		return
	}
	// @参数列表与 send 接口一致
	if err, ok := checkArguments(parameters, res, "sender", "content"); !ok {
		res = err
		Logger.Error("", "", "", "center.BroadcastChatRoomHandler", res)
		return
	}
	uid := parameters.Get("sender")
	roomid := parameters.Get("roomid")
	traceid := parameters.Get("traceid")
	if traceid == "" { // 经济系统调用使用的是 sn
		traceid = parameters.Get("sn")
	}
	content := parameters.Get("content")
	// 消息类型和优先级，类型暂时不传，优先级：加入退出消息（0），聊天消息（1），礼物消息（101），红包消息（201）
	msgtype, _ := strconv.Atoi(parameters.Get("type"))
	priority, _ := strconv.Atoi(parameters.Get("priority"))
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if len(content) > netConf().MaxMsgSize {
		res.Code, res.Reason = errorMessageTooLong, errorReasonMessageTooLong
		Logger.Error(uid, appid, traceid, "center.BroadcastChatRoomHandler", roomid, priority, content, "content too long")
		return
	}
	importance := true
	gws := logic.NetGlobalConf().GatewayRpcs
	if netConf().OnlineBroadcastPolicy >= 100 {
		gws = []string{}
	} else if netConf().OnlineBroadcastPolicy != 0 {
		start := rand.Intn(len(gws))
		end := (len(gws)*netConf().OnlineBroadcastPolicy/100 + start) % len(gws)
		if start < end {
			res := gws[0:start]
			gws = append(res, gws[end:len(gws)]...)
		} else {
			gws = gws[end:start]
		}
	}
	gwAddrs := make(map[string]int, len(gws))
	for _, gw := range gws {
		gwAddrs[gw] = 1
	}
	if len(gwAddrs) != 0 {
		message := newBroadcastRoomMessage(uid, appid, msgtype, content, importance)
		if err := router.SendOnlineBroadcast(message, gwAddrs, traceid); err != nil {
			Logger.Error(uid, appid, traceid, "router.SendOnlineBroadcast", roomid, priority, content, err.Error(), logic.ClientIp(req))
			res.Code, res.Reason = errorInternalError, errorReasonRequestFail
		} else {
			Logger.Trace(uid, appid, traceid, "center.BroadcastOnlineUsersHandler", roomid, priority, content, logic.ClientIp(req))
		}
	}
}

func QueryChatRoomMemberCountHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "content"); !ok {
			res = err
			Logger.Error("", "", "", "center.QueryChatRoomMemberCountHandler", res)
			break
		}
		rooms := strings.Split(req.FormValue("content"), ",")
		appid := logic.StringToUint16(req.FormValue("appid"))
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		if resp, err := saver.QueryChatRoomMemberCount(rooms, appid); err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
		} else {
			ret := make(map[string]int, len(resp))
			for room, v := range resp {
				ret[room] = v[session.RegisteredChatRoomUser] + v[session.NonRegisteredChatRoomUser] + v[session.RobotChatRoomUser] +
					v[session.ChatRoomWebUserPrefix+session.RegisteredChatRoomUser] + v[session.ChatRoomWebUserPrefix+session.NonRegisteredChatRoomUser]
			}
			res.Data = ret
		}
		Logger.Trace("", appid, req.FormValue("sn"), "center.QueryChatRoomMemberCountHandler", res, logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.QueryChatRoomMemberCountHandler", res, logic.ClientIp(req))
	}

	// count request
	requestStat.AtomicAddQueryMemberCounts(1)

	w.Write([]byte(res.Error()))
}

func QueryChatRoomMemberDetailHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "content"); !ok {
			res = err
			Logger.Error("", "", "", "center.QueryChatRoomMemberDetailHandler", res)
			break
		}
		rooms := strings.Split(req.FormValue("content"), ",")
		appid := logic.StringToUint16(req.FormValue("appid"))
		v2, _ := strconv.Atoi(req.FormValue("v2"))
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		if resp, err := saver.QueryChatRoomMemberCount(rooms, appid); err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
		} else {
			r := make(map[string]map[string]int, len(resp))
			if v2 != 0 {
				for room, v := range resp {
					r[room] = map[string]int{
						"reg_app":   v[session.RegisteredChatRoomUser],
						"noreg_app": v[session.NonRegisteredChatRoomUser],
						"reg_web":   v[session.ChatRoomWebUserPrefix+session.RegisteredChatRoomUser],
						"noreg_web": v[session.ChatRoomWebUserPrefix+session.NonRegisteredChatRoomUser],
						"fake":      v[session.RobotChatRoomUser],
					}
				}
			} else {
				for room, v := range resp {
					r[room] = map[string]int{
						"reg":   v[session.RegisteredChatRoomUser] + v[session.ChatRoomWebUserPrefix+session.RegisteredChatRoomUser],
						"noreg": v[session.NonRegisteredChatRoomUser] + v[session.ChatRoomWebUserPrefix+session.NonRegisteredChatRoomUser],
						"fake":  v[session.RobotChatRoomUser],
						"web":   v[session.ChatRoomWebUserPrefix+session.RegisteredChatRoomUser] + v[session.ChatRoomWebUserPrefix+session.NonRegisteredChatRoomUser],
					}
				}
			}
			res.Data = r
		}
		Logger.Trace("", appid, req.FormValue("sn"), "center.QueryChatRoomMemberDetailHandler", res, logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.QueryChatRoomMemberDetailHandler", res, logic.ClientIp(req))
	}

	// count request
	requestStat.AtomicAddQueryMemberDetails(1)

	w.Write([]byte(res.Error()))
}

/*
 * handler of path huajiao/all, huajiao/hot
 */
func HuajiaoPublicAndHotHandler(w http.ResponseWriter, req *http.Request) {
	// get decoded chat request parameters
	parameters, response := GetChatRequestParameter(req)
	res := response.ToSimple()
	if res.Code == ChatSuccess {
		Logger.Debug("", "", "", "HuajiaoPublicAndHot", "Request information", parameters, logic.ClientIp(req))
		res = StorePublicAndHot(parameters, strings.HasSuffix(req.RequestURI, "hot"), logic.StringToUint16(parameters.Get("appid")))
	} else {
		Logger.Error("", "", "", "HuajiaoPublicAndHot", "Invalid request",
			fmt.Sprint("from: ", req.RemoteAddr, " psot form: ", req.PostForm))
	}

	Logger.Debug("", "", "", "HuajiaoPublicAndHot", "Response to callser", res.Response(), logic.ClientIp(req))
	if res.Code == ChatParameterError {
		res.Reason = "client not send required parameter"
	}

	// count request
	if res.Code == ChatSuccess {
		requestStat.AtomicAddPublicHots(1)
	} else {
		requestStat.AtomicAddPublicHotFails(1)
	}

	w.Write([]byte(res.Response()))
}

/*
 * handler of path huajiao/users, push/chat
 */
func HuajiaoUsersHandler(w http.ResponseWriter, req *http.Request) {
	// get decoded chat request parameters
	parameters, res := GetChatRequestParameter(req)
	if res.Code == ChatSuccess {
		appid := logic.StringToUint16(parameters.Get("appid"))
		Logger.Debug("", "", "", "HuajiaoUsers", "Request information", parameters, logic.ClientIp(req))
		if req.RequestURI == "/push/chat" || req.RequestURI == "/huajiao/chat" {
			// count request response time if need
			if netConf().StatResponseTime {
				countFunc := countImRespTime(parameters.Get(ChatReceivers), parameters.Get(ChatTraceId),
					"HuajiaoUsersHandler", appid)
				defer countFunc()
			}
			// store IM message and then notify receivers a notification if they are online
			res = StoreMessageAndPushNotification(parameters, saver.ChatChannelIM, appid)

			// count request
			if res.Code == ChatSuccess {
				requestStat.AtomicAddIms(1)
			} else {
				requestStat.AtomicAddImFails(1)
			}
		} else {
			// count request response time if need
			if netConf().StatResponseTime {
				countFunc := countPeerRespTime(parameters.Get(ChatReceivers), parameters.Get(ChatTraceId),
					"HuajiaoUsersHandler", appid)
				defer countFunc()
			}
			// store notification message and then notify receivers a notification if they are online
			res = StoreMessageAndPushNotification(parameters, saver.ChatChannelNotify, appid)

			// count request
			if res.Code == ChatSuccess {
				requestStat.AtomicAddPeers(1)
			} else {
				requestStat.AtomicAddPeerFails(1)
			}
		}
	} else {
		Logger.Error("", "", "", "HuajiaoUsers", "Invalid request",
			fmt.Sprint("from: ", req.RemoteAddr, " psot form: ", req.PostForm))
	}

	Logger.Debug("", "", "", "HuajiaoUsers", "Response to callser", res.Response(), logic.ClientIp(req))
	if res.Code == ChatParameterError {
		res.Reason = "client not send required parameter"
	}
	w.Write([]byte(res.Response()))
}

func HuajiaoUsersUnreadHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		w.Write([]byte(res.Error()))
	}()
	parameters, resn := GetChatRequestParameter(req)
	if resn.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(resn.Code), resn.Reason
		Logger.Error("", "", "", "center.HuajiaoUsersUnreadHandler", res)
		return
	}
	if err, ok := checkArguments(parameters, res, "uids", "channel"); !ok {
		res = err
		Logger.Error("", "", "", "center.HuajiaoUsersUnreadHandler", res)
		return
	}
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	uids := parameters.Get("uids")
	chs := parameters.Get("channel")
	userIds := strings.Split(uids, ",")
	channels := strings.Split(chs, ",")

	resp, err := saver.RetrieveUnreadCount(appid, userIds, channels)
	if err != nil {
		res.Code, res.Reason = ChatSaverFailed, err.Error()
		Logger.Error(uids, appid, "", "center.HuajiaoUsersUnreadHandler", chs, res)
		return
	}
	type retDataStruct struct {
		Unread     uint64
		LastReadID uint64
		LatestID   uint64
	}
	data := make(map[string]map[string]retDataStruct, len(resp))
	for _, r := range resp {
		resData := make(map[string]retDataStruct, len(r.LatestID))
		for ch, id := range r.LastReadID {
			if c := r.LatestID[ch] - id; c < 0 {
				resData[ch] = retDataStruct{
					Unread:     0,
					LastReadID: id,
					LatestID:   r.LatestID[ch],
				}
			} else if ch == saver.ChatChannelIM {
				resData[ch] = retDataStruct{
					Unread:     c / 2,
					LastReadID: id,
					LatestID:   r.LatestID[ch],
				}
			} else {
				resData[ch] = retDataStruct{
					Unread:     c,
					LastReadID: id,
					LatestID:   r.LatestID[ch],
				}
			}
		}
		data[r.Owner] = resData
	}

	res.Data = data
	Logger.Trace(uids, appid, "", "center.HuajiaoUsersUnreadHandler", chs, logic.ClientIp(req))
}

/*
 * handler of path push/recall, huajiao/recall
 */
func HuajiaoRecallHandler(w http.ResponseWriter, req *http.Request) {
	// get decoded chat request parameters
	parameters, response := GetChatRequestParameter(req)
	res := response.ToSimple()
	if res.Code == ChatSuccess {
		Logger.Debug("", "", "", "HuajiaoRecallHandler", "Request information", parameters)
		res = RecallImMessage(parameters, logic.StringToUint16(parameters.Get("appid")))
	} else {
		Logger.Error("", "", "", "HuajiaoRecallHandler", "Invalid request",
			fmt.Sprint("from: ", req.RemoteAddr, " psot form: ", req.PostForm))
	}

	Logger.Debug("", "", "", "HuajiaoRecallHandler", "Response to callser", res.Response(), logic.ClientIp(req))
	if res.Code == ChatParameterError {
		res.Reason = "client not send required parameter"
	}

	// count request
	if res.Code == ChatSuccess {
		requestStat.AtomicAddRecalls(1)
	} else {
		requestStat.AtomicAddRecallFails(1)
	}

	w.Write([]byte(res.Response()))
}

/*
 * handler of path operation/retrieve
 */
func HuajiaoRetrieveHandler(w http.ResponseWriter, req *http.Request) {
	// get decoded chat request parameters
	parameters, response := GetChatRequestParameter(req)
	res := &RetrieveChatResponse{Code: response.Code, Reason: response.Reason}
	if res.Code == ChatSuccess {
		Logger.Debug("", "", "", "HuajiaoRetrieveHandler", "Request information", parameters, logic.ClientIp(req))
		res = RetrieveMessages(parameters, logic.StringToUint16(parameters.Get("appid")))
	} else {
		Logger.Error("", "", "", "HuajiaoRetrieveHandler", "Invalid request",
			fmt.Sprint("from: ", req.RemoteAddr, " psot form: ", req.PostForm))
	}

	Logger.Debug("", "", "", "HuajiaoRetrieveHandler", "Response to callser", res.Response(), logic.ClientIp(req))
	if res.Code == ChatParameterError {
		res.Reason = "client not send required parameter"
	}

	// count request
	if res.Code == ChatSuccess {
		requestStat.AtomicAddRetrieves(1)
	} else {
		requestStat.AtomicAddRetrieveFails(1)
	}

	w.Write([]byte(res.Response()))
}

func PushNoticationHandler(w http.ResponseWriter, req *http.Request) {
	// get decoded chat request parameters
	parameters, response := GetChatRequestParameter(req)
	res := response.ToSimple()
	if res.Code == ChatSuccess {
		appid, err := strconv.Atoi(parameters.Get("appid"))
		if err != nil {
			res.Code, res.Reason = ChatParameterError, err.Error()
			goto end
		}
		sender := parameters.Get("sender")
		channel := parameters.Get("ch")
		traceId := parameters.Get("sn")
		Logger.Trace(sender, appid, traceId, "PushNoticationHandler", "Distribute forward request", parameters)
		msgStr := parameters.Get("msgs")
		msgs := []interface{}{}
		dec := json.NewDecoder(bytes.NewBuffer([]byte(msgStr)))
		dec.UseNumber()
		if err := dec.Decode(&msgs); err != nil {
			res.Code, res.Reason = ChatParameterError, err.Error()
			goto end
		}
		inbox := map[string]uint64{}
		outbox := map[string]uint64{}
		reqSess := []*session.UserSession{}
		for _, m := range msgs {
			mt, ok := m.(map[string]interface{})
			if !ok {
				res.Code, res.Reason = ChatParameterError, "invalid request"
				goto end
			}
			owner, ok1 := mt["owner"].(string)
			idNumber, ok2 := mt["id"].(json.Number)
			box, ok3 := mt["box"].(string)
			if !ok1 || !ok2 || !ok3 {
				res.Code, res.Reason = ChatParameterError, "invalid request"
				goto end
			}
			id, err := idNumber.Int64()
			if err != nil {
				res.Code, res.Reason = ChatParameterError, err.Error()
				goto end
			}
			if box == "inbox" {
				inbox[owner] = uint64(id)
			} else if box == "outbox" {
				outbox[owner] = uint64(id)
			} else {
				res.Code, res.Reason = ChatParameterError, "invalid request"
				goto end
			}
			reqSess = append(reqSess, &session.UserSession{UserId: owner, AppId: uint16(appid)})
		}
		Logger.Trace(sender, appid, traceId, "PushNoticationHandler", "Request information", msgStr)
		if onlineUsers, err := session.Query(reqSess); err != nil {
			Logger.Error(sender, appid, traceId, "PushNoticationHandler", "session.Query error", err, logic.ClientIp(req))
		} else {
			PushNotication(appid, sender, onlineUsers, channel, inbox, outbox, traceId)
			Logger.Trace(sender, appid, traceId, "PushNoticationHandler", "Query Session Result", onlineUsers, logic.ClientIp(req))
		}

	}
end:
	Logger.Debug("", "", "", "PushNoticationHandler", "Response to callser", res.Response(), logic.ClientIp(req))
	w.Write([]byte(res.Response()))
}

func PushPublicNoticationHandler(w http.ResponseWriter, req *http.Request) {
	parameters, response := GetChatRequestParameter(req)
	res := response.ToSimple()
	if res.Code == ChatSuccess {
		appidStr := parameters.Get("appid")
		traceId := parameters.Get("sn")
		Logger.Trace("", appidStr, traceId, "PushPublicNoticationHandler", "Distribute forward request", parameters)
		msgid, _ := strconv.ParseUint(parameters.Get("msgid"), 10, 64)
		if err := router.SendPushTags([]string{appidStr}, []string{}, ChatPublicSender, saver.ChatChannelPublic, traceId, uint64(msgid)); err != nil {
			// although we can not send notification for push failed,
			// we still send caller successful result for messages have been saved to storage yet
			// it does not matter for client will pull messages
			Logger.Error("", appidStr, traceId, "PushPublicNoticationHandler", "SendPushTags error", err)
		}
		Logger.Trace(appidStr, msgid, traceId, "PushPublicNoticationHandler", "req", "")
	}
	Logger.Debug("", "", "", "PushPublicNoticationHandler", "Response to callser", res.Response(), logic.ClientIp(req))
	w.Write([]byte(res.Response()))
}

func QueryUserInRoomHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonUserInRoom}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "uid", "rid"); !ok {
			res = err
			Logger.Error("", "", "", "center.QueryUserInRoomHandler", res)
			break
		}
		room := req.FormValue("rid")
		user := req.FormValue("uid")
		appid := logic.StringToUint16(req.FormValue("appid"))
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		if isIn, err := saver.QueryUserInRoom(room, user, appid); err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
		} else {
			if isIn {
				res.Reason = errorReasonUserInRoom
				res.Data = ResUserIsInRoom
			} else {
				res.Reason = errorReasonUserNotInRoom
				res.Data = ResUserIsNotInRoom
			}
		}
		Logger.Trace(user, appid, "", "center.QueryUserInRoomHandler", room, res, logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "UnspportHttpMethod", "center.QueryUserInRoomHandler", res, logic.ClientIp(req))
	}

	requestStat.AtomicAddQueryUserInRoom(1)
	w.Write([]byte(res.Error()))
}

func QueryUserSessionInRoomHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "uid", "rid"); !ok {
			res = err
			Logger.Error("", "", "", "center.QueryUserSessionInRoomHandler", res)
			break
		}
		room := req.FormValue("rid")
		user := req.FormValue("uid")
		appid := logic.StringToUint16(req.FormValue("appid"))
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		isIn, err := saver.QueryUserInRoom(room, user, appid)
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			break
		}
		if !isIn {
			res.Reason, res.Data = errorReasonUserNotInRoom, map[string]int{"inroom": ResUserIsNotInRoom}
			break
		}
		ret, err := saver.QueryUserSessionSummary(appid, user)
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
		} else {
			ret["inroom"] = ResUserIsInRoom
		}
		res.Reason, res.Data = errorReasonUserInRoom, ret

		Logger.Trace(user, appid, "", "center.QueryUserSessionInRoomHandler", room, res.Data, logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "UnspportHttpMethod", "center.QueryUserSessionInRoomHandler", res, logic.ClientIp(req))
	}

	requestStat.AtomicAddQuerySessionInRoom(1)
	w.Write([]byte(res.Error()))
}

func PrivateSendChatRoomMessageHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}

	switch req.Method {
	case MethodPost:
		privateSendChatRoomMessage(req, res)
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.PrivateSendChatRoomMessageHandler", res, logic.ClientIp(req))
	}

	w.Write([]byte(res.Error()))
}

func PrivateSendChatRoomMessageRawHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}

	switch req.Method {
	case MethodPost:
		privateSendChatRoomMessageRaw(req, res)
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.PrivateSendChatRoomMessageRawHandler", res, logic.ClientIp(req))
	}

	w.Write([]byte(res.Error()))
}

func SendChatRoomHighPriorityHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}

	switch req.Method {
	case MethodPost:
		// extract params
		parameters, extractRes := GetChatRequestParameter(req)
		if extractRes.Code != ChatSuccess {
			Logger.Error("", "", "", "SendChatRoomHighPriorityHandler", res)
			res.Code, res.Reason = ErrorCode(extractRes.Code), extractRes.Reason
			break
		}

		if err, ok := checkArguments(parameters, res, "sender", "traceid", "content", "roomid"); !ok {
			res = err
			Logger.Error("", "", "", "SendChatRoomHighPriorityHandler", res)
			break
		}

		traceid := parameters.Get("traceid")
		uid := parameters.Get("sender")
		//receiver := parameters.Get("receiver")
		appid := logic.StringToUint16(parameters.Get("appid"))
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		content := parameters.Get("content")
		roomid := parameters.Get("roomid")
		msgtype, _ := strconv.Atoi(parameters.Get("type"))
		priority, _ := strconv.Atoi(parameters.Get("priority"))

		if len(content) > netConf().MaxMsgSize {
			res.Code, res.Reason = errorMessageTooLong, errorReasonMessageTooLong
			Logger.Error(uid, appid, traceid, "SendChatRoomHighPriorityHandler", roomid, msgtype, priority, content, "content too long")
			break
		}

		if !isSpecialMsg(msgtype, priority) {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			Logger.Error(uid, appid, traceid, "SendChatRoomHighPriorityHandler", "msgtype forbidden", roomid, msgtype, priority, content, "")
			break
		}

		// send message by connectionid and gateway
		detail, err := saver.QueryChatRoomDetail(roomid, appid)
		if err != nil {
			Logger.Error(roomid, appid, traceid, "SendChatRoomHighPriorityHandler", "saver.QueryChatRoomDetail error", err.Error())
			res.Code, res.Reason = errorInternalError, err.Error()
			break
		}

		message := newChatRoomMessage(uid, detail, msgtype, content, true)

		if msgid, err := saver.CacheChatRoomMessage(message); err != nil {
			Logger.Error(uid, appid, traceid, "saver.SendChatRoomHighPriorityHandler error", roomid, msgid, err)
			res.Reason = err.Error()
			break
		} else {
			message.MsgID, message.MaxID = msgid, msgid
		}
		if len(detail.GatewayAddrs) > 0 {
			if err := router.SendChatRoomNotify(message, detail.GatewayAddrs, traceid); err != nil {
				Logger.Error(uid, appid, traceid, "router.SendChatRoomHighPriorityHandler error", roomid, msgtype, priority, err.Error())
				res.Code, res.Reason = errorInternalError, errorReasonRequestFail
			} else {
				Logger.Trace(uid, appid, traceid, "center.SendChatRoomHighPriorityHandler", detail.ConnCount(), roomid, priority, content, msgtype, logic.ClientIp(req))
			}
		}
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.SendChatRoomHighPriorityHandler", res, logic.ClientIp(req))
	}

	w.Write([]byte(res.Error()))
}

func ChatRoomDegradeHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		w.Write([]byte(res.Error()))
	}()
	parameters, resn := GetChatRequestParameter(req)
	if resn.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(resn.Code), resn.Reason
		Logger.Error("", "", "", "center.ChatRoomDegradeHandler", res)
		return
	}
	if err, ok := checkArguments(parameters, res, "roomid", "degrade"); !ok {
		res = err
		Logger.Error("", "", "", "center.ChatRoomDegradeHandler", res)
		return
	}
	roomid := parameters.Get("roomid")
	reason := parameters.Get("reason")
	traceid := parameters.Get("traceid")
	degrade, _ := strconv.Atoi(parameters.Get("degrade"))
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	c := sentByCoordinator(roomid)
	if c {
		if err := coordinator.DegradeChatRoom(appid, roomid, degrade > 0); err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			Logger.Error("", appid, traceid, "center.ChatRoomDegradeHandler", err.Error(), roomid, c, degrade, reason, logic.ClientIp(req))
			return
		}
	}
	Logger.Trace("", appid, traceid, "center.ChatRoomDegradeHandler", roomid, c, degrade, reason, logic.ClientIp(req))
}

func ChatRoomDegradedListHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		w.Write([]byte(res.Error()))
	}()
	parameters, resn := GetChatRequestParameter(req)
	if resn.Code != ChatSuccess {
		res.Code, res.Reason = ErrorCode(resn.Code), resn.Reason
		Logger.Error("", "", "", "center.ChatRoomDegradedListHandler", res)
		return
	}
	appid := logic.StringToUint16(parameters.Get("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	rooms, err := coordinator.GetDegradedChatRoomList(appid)
	if err != nil {
		res.Code, res.Reason = errorInternalError, err.Error()
		Logger.Error("", appid, "", "center.ChatRoomDegradedListHandler", err.Error(), appid, logic.ClientIp(req))
		return
	}
	res.Data = map[string][]string{
		"rooms": rooms,
	}
	Logger.Trace("", appid, "", "center.ChatRoomDegradedListHandler", rooms, logic.ClientIp(req))
}

func LiveHandler(w http.ResponseWriter, req *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch req.Method {
	case MethodGet, MethodPost:
		if err := req.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		if err, ok := checkArguments(&req.Form, res, "rid"); !ok {
			res = err
			Logger.Error("", "", "", "center.LiveHandler", res)
			break
		}

		appid := logic.StringToUint16(req.FormValue("appid"))
		if appid == 0 {
			appid = logic.DEFAULT_APPID
		}
		roomid := req.FormValue("rid")
		if len(roomid) == 0 {
			res.Code, res.Reason = errorLostArgument, "rid value empty"
			break
		}

		start := false
		if strings.Contains(req.URL.Path, "start") {
			start = true
		}
		if err := coordinator.LiveNotify(appid, roomid, start); err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			Logger.Error("", appid, "", "coordinator.LiveNotify", err.Error(), roomid, req, logic.ClientIp(req))
			break
		}

		Logger.Trace("", appid, "", "center.LiveHandler", res, logic.ClientIp(req))
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		Logger.Error("", "", "", "center.LiveHandler", res, logic.ClientIp(req))
	}
	w.Write([]byte(res.Error()))
}
