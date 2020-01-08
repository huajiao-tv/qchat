/**
 * 处理各种包的逻辑
 */
package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/cryption"
	"github.com/huajiao-tv/qchat/utility/network"
)

func dealInitLoginReq(gwp *router.GwPackage, m *pb.Message, gwr *router.GwResp) (*pb.Message, *Error) {
	if m.Req.InitLoginReq == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "expect a initlogin package, but get nil")
	}
	if _, ok := gwp.Property["Appid"]; ok {
		return nil, NewError(pb.ERR_BAD_SEQUENCE, "init login should send as first package")
	}

	sig := ""
	if m.Req.InitLoginReq.Sig != nil {
		sig = *m.Req.InitLoginReq.Sig
	}

	if m.Req.InitLoginReq.GetClientRam() == "" {
		return nil, NewError(pb.ERR_BAD_PARAM, "some field is nil: clientRam | sn | sender")
	}

	if ok := logic.GetAppids()[strconv.Itoa(int(gwp.XimpBuff.Appid))]; !ok {
		return nil, NewError(pb.ERR_BAD_PARAM, "invalid appid:"+strconv.Itoa(int(gwp.XimpBuff.Appid)))
	}

	serverRam := ServerRam()
	initLoginRes := &pb.Message{
		Msgid: proto.Uint32(pb.INIT_LOGIN_RESP),
		Sn:    m.Sn,
		Resp: &pb.Response{
			InitLoginResp: &pb.InitLoginResp{
				ClientRam: m.Req.InitLoginReq.ClientRam,
				ServerRam: &serverRam,
			},
		},
	}
	ds, err := proto.Marshal(initLoginRes)
	if err != nil {
		return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "pb marshal error"+err.Error())
	}
	gwr.XimpBuff = &network.XimpBuffer{
		IsDecrypt:  true,
		HasHeader:  true,
		DataStream: ds,
		Version:    gwp.XimpBuff.Version, // 用来拼回包
	}

	gwr.Property = map[string]string{
		"ClientRam": *m.Req.InitLoginReq.ClientRam,
		"Sig":       sig,
		"ServerRam": serverRam,
		"CVersion":  strconv.Itoa(int(gwp.XimpBuff.CVersion)),
		"Version":   strconv.Itoa(int(gwp.XimpBuff.Version)),
		"Appid":     strconv.Itoa(int(gwp.XimpBuff.Appid)),
		"Sender":    *m.Sender,
	}

	return initLoginRes, nil
}

func dealLogoutReq(gwp *router.GwPackage, m *pb.Message, gwr *router.GwResp) (*pb.Message, *Error) {
	if m.Req.Logout == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "respect a logout package, but get nil")
	}

	appid := logic.StringToUint16(gwp.Property["Appid"])
	if appid == 0 {
		return nil, NewError(pb.ERR_BAD_SEQUENCE, "appid is not setted, did you send a init login?")
	}

	closeSessionResp, err := session.Close(gwp.Property["Sender"], appid, gwp.GatewayAddr, gwp.ConnectionId, gwp.Property)
	if err != nil {
		Logger.Error(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealLogoutReq", "router.closeSession error", err)
		return nil, GenError(pb.ERR_SESSION_GENERIC, err)
	}

	// 生成回到客户端的包
	var result uint32 = 0
	logoutRes := &pb.Message{
		Msgid: proto.Uint32(pb.LOGOUT_RESP),
		Sn:    m.Sn,
		Resp: &pb.Response{
			Logout: &pb.LogoutResp{
				Result: &result,
			},
		},
	}
	ds, err := proto.Marshal(logoutRes)
	if err != nil {
		return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "pb marshal error:"+err.Error())
	}
	ximpBuff := &network.XimpBuffer{
		DataStream: ds,
		IsDecrypt:  true,
	}

	gwr.XimpBuff = ximpBuff
	if len(closeSessionResp.Tags) != 0 {
		gwr.Tags = make(map[string]bool, len(closeSessionResp.Tags))
		for _, v := range closeSessionResp.Tags {
			gwr.Tags[v] = false
		}
	}
	gwr.Property = map[string]string{"Open": "0"}
	Logger.Trace(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealLogoutReq", "closeSession", gwr.Tags)
	return logoutRes, nil
}

func dealLoginReq(gwp *router.GwPackage, m *pb.Message, gwr *router.GwResp) (*pb.Message, *Error) {
	if m.Req.Login == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "expect a login package, but get nil")
	}
	if len(m.Req.Login.SecretRam) == 0 {
		return nil, NewError(pb.ERR_BAD_PARAM, "empty secretRam or empty sender")
	}
	if m.Req.Login.GetServerRam() != gwp.Property["ServerRam"] {
		return nil, NewError(pb.ERR_BAD_PARAM, "server ram not match")
	}

	appid := logic.StringToUint16(gwp.Property["Appid"])
	if appid == 0 {
		return nil, NewError(pb.ERR_BAD_SEQUENCE, "appid is not setted, did you send a init login?")
	}

	platform := m.Req.Login.GetPlatform()
	mobileType := m.Req.Login.GetMobileType()
	//判断mobileType是否为web类型，如果符合，则连接类型改为network.WebSocketNetwork
	if _, ok := netConf().WebConnectionTypes[mobileType]; ok {
		// pc端Flash标记为websocket
		gwp.Property["ConnectionType"] = network.WebSocketNetwork
	}

	// 验证sig, secret_ram, verf_code
	isLoginUser, key, sessionKey, rErr := verify(m, gwp)
	if rErr != nil {
		return nil, rErr
	}

	// 关闭加密
	if (netConf().DisableEncrypt || gwp.Property["ConnectionType"] == network.WebSocketNetwork) && m.Req.Login.GetNotEncrypt() {
		sessionKey = ""
	}

	netType := m.Req.Login.GetNetType()
	heartFeq := m.Req.Login.GetHeartFeq()
	// 原客户端心跳为2分钟，但是因为有些机型无法按照指定时间唤醒而是会对其5分钟，典型的手机：小米，所以如果心跳设置小于5分钟，会出现被断开的情况，不合理，所以在这里服务端将超时时间x3
	gwr.HeartBeatTimeout = time.Duration(heartFeq) * 3 * time.Second
	if m.Req.Login.AppId != nil && *m.Req.Login.AppId != uint32(appid) {
		return nil, NewError(pb.ERR_BAD_PARAM, "appid is not match,login.appid != header.appid")
	}
	deviceid := m.Req.Login.GetDeviceid()
	if deviceid == "" {
		deviceid = ServerRam()
	}

	gwp.Property["NetType"] = fmt.Sprintf("%d", netType)
	gwp.Property["HeartFeq"] = fmt.Sprintf("%d", heartFeq)
	gwp.Property["Deviceid"] = deviceid
	gwp.Property["Platform"] = platform
	gwp.Property["MobileType"] = mobileType
	gwp.Property["LoginTime"] = time.Now().String()
	gwp.Property["SenderType"] = m.GetSenderType()
	openSessionResp, err := session.Open(gwp.Property["Sender"], appid, isLoginUser, sessionKey, gwp.ConnectionId, gwp.GatewayAddr, gwp.Property)
	if err != nil {
		Logger.Error(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealLoginReq", "session.OpenSession error", err)
		return nil, GenError(pb.ERR_SESSION_REFUSED, err)
	}

	// 此设置为了防止下面报错导致没有将open设置到gateway，从而使closeSession无法调用
	gwr.Property = map[string]string{
		"Open": "1", // 用来标记这个用户是否在session里写成功,在logout的时候 用到
	}

	// 生成回到客户端的包
	loginRes := &pb.Message{
		Msgid: proto.Uint32(pb.LOGIN_RESP),
		Sn:    m.Sn,
		Resp: &pb.Response{
			Login: &pb.LoginResp{
				SessionKey:    proto.String(sessionKey),
				ClientLoginIp: proto.String(gwp.Property["ClientIp"]),
				Serverip:      proto.String(logic.GetIp()),
				SessionId:     proto.String(""),
				Timestamp:     proto.Uint32(uint32(time.Now().Unix())),
			},
		},
	}
	ds, err := proto.Marshal(loginRes)
	if err != nil {
		return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "pb marshal error:"+err.Error())
	}
	ximpBuff := &network.XimpBuffer{
		DataStream: ds,
	}
	// 由于这次使用的加密key是password(即token), 所以在这块就加密了
	if err := ximpBuff.Encrypt(key); err != nil {
		return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "ximp encrypt error:"+err.Error())
	}

	gwr.XimpBuff = ximpBuff
	if len(openSessionResp.Tags) != 0 {
		gwr.Tags = make(map[string]bool)
		for _, v := range openSessionResp.Tags {
			gwr.Tags[v] = true
		}
	}
	gwr.Property = map[string]string{
		"NetType":     fmt.Sprintf("%d", netType),
		"HeartFeq":    fmt.Sprintf("%d", heartFeq),
		"Deviceid":    deviceid,
		"Platform":    platform,
		"MobileType":  mobileType,
		"LoginTime":   time.Now().String(),
		"IsLoginUser": strconv.FormatBool(isLoginUser), // 做为走到这个步骤的标识
		"Open":        "1",                             // 用来标记这个用户是否在session里写成功,在logout的时候 用到
	}
	// @将Flash客户端标记未websocket
	if mobileType == "pc" {
		gwr.Property["ConnectionType"] = network.WebSocketNetwork
	}
	// 后续的加密都走这个key
	if sessionKey == "" {
		gwr.Property["NotEncrypt"] = "1"
	}
	gwr.Rkey = []byte(sessionKey)

	//处理旧的的usersession
	if len(openSessionResp.OldUserSessions) != 0 {
		Logger.Warn(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealLoginReq", "find oldSessions", len(openSessionResp.OldUserSessions))
		go handleOldSession(openSessionResp.OldUserSessions)
	}
	Logger.Trace(gwp.Property["Sender"], appid, logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealLoginReq", gwp.Property["ConnectionType"], gwp.Property["ClientIp"])

	return loginRes, nil
}

func dealGetInfoReq(gwp *router.GwPackage, m *pb.Message, gwr *router.GwResp) (*pb.Message, *Error) {
	if m.Req.GetInfo == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "field getinfo is nil")
	}
	if m.Req.GetInfo.GetInfoType() == "" {
		return nil, NewError(pb.ERR_BAD_PARAM, "info_type is nil")
	}
	if m.Req.GetInfo.GetInfoId == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "get_info_id is nil")
	}
	cversion, _ := strconv.Atoi(gwp.Property["CVersion"])
	infoType := m.Req.GetInfo.GetInfoType()
	infos := []*pb.Info{}
	lastid := int64(0)
	if infoType == saver.ChatChannelIM && cversion <= 101 {
	} else {
		if checkPushDegrade(infoType, gwp.Property["Sender"], true) {
			lastid = 0
			Logger.Warn(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealGetInfoReq", "pull disabled", infoType)
		} else {
			retrieveReq := &saver.RetrieveMessagesRequest{
				logic.StringToUint16(gwp.Property["Appid"]),
				gwp.Property["Sender"],
				map[string]*saver.RetrieveChannel{
					infoType: &saver.RetrieveChannel{
						infoType,
						m.Req.GetInfo.GetGetInfoId(),
						int(m.Req.GetInfo.GetGetInfoOffset()),
					},
				},
				logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId),
			}
			msgs, err := saver.RetrieveChatMessages(retrieveReq)
			if err != nil {
				Logger.Error(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealGetInfoReq", "saver.RetrieveChatMessages error", err)
				lastid = 0
			} else {
				if len(msgs.Inbox[infoType]) != 0 {
					for _, m := range msgs.Inbox[infoType] {
						// msg_type(4) + msg_sn(8) + info_id(8) + time_sent(8) + expire_time(8) + msg_box(4)
						buffer := make([]byte, 40)
						binary.BigEndian.PutUint32(buffer[0:4], m.Type)
						binary.BigEndian.PutUint64(buffer[4:12], uint64(m.TraceSN))
						binary.BigEndian.PutUint64(buffer[12:20], uint64(m.MsgId))
						binary.BigEndian.PutUint64(buffer[20:28], uint64(m.Creation.UnixNano()/1e6))
						binary.BigEndian.PutUint64(buffer[28:36], uint64(m.ExpireTime.UnixNano()/1e6))
						binary.BigEndian.PutUint32(buffer[36:40], uint32(m.Box))
						propPairs := []*pb.Pair{
							&pb.Pair{Key: []byte("info_id"), Value: buffer[12:20]},
							&pb.Pair{Key: []byte("msg_type"), Value: buffer[0:4]},
							&pb.Pair{Key: []byte("chat_body"), Value: []byte(m.Content)},
							&pb.Pair{Key: []byte("time_sent"), Value: buffer[20:28]},
							&pb.Pair{Key: []byte("expire_time"), Value: buffer[28:36]},
							&pb.Pair{Key: []byte("msg_box"), Value: buffer[36:40]},
							&pb.Pair{Key: []byte("msg_sn"), Value: buffer[4:12]},
							&pb.Pair{Key: []byte("msg_to"), Value: []byte(m.To)},
							&pb.Pair{Key: []byte("msg_from"), Value: []byte(m.From)},
						}
						infos = append(infos, &pb.Info{PropertyPairs: propPairs})
					}
				}
				lastid = int64(msgs.LatestID[infoType])
			}
		}
	}
	getInfoResp := &pb.Message{
		Msgid: proto.Uint32(pb.GET_INFO_RESP),
		Sn:    m.Sn,
		Resp: &pb.Response{
			GetInfo: &pb.GetInfoResp{
				InfoType:   m.Req.GetInfo.InfoType,
				Infos:      infos,
				LastInfoId: &lastid,
			},
		},
	}
	ds, err := proto.Marshal(getInfoResp)
	if err != nil {
		return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "pb marshal error:"+err.Error())
	}
	gwr.XimpBuff = &network.XimpBuffer{
		IsDecrypt:  true,
		DataStream: ds,
	}

	return getInfoResp, nil
}

func dealGetMultiInfoReq(gwp *router.GwPackage, m *pb.Message, gwr *router.GwResp) (*pb.Message, *Error) {
	if m.Req.GetMultiInfos == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "field getmultiinfos is nil")
	}
	if m.Req.GetMultiInfos.GetInfoType() == "" {
		return nil, NewError(pb.ERR_BAD_PARAM, "info_type is nil")
	}
	if m.Req.GetMultiInfos.GetGetInfoIds() == nil {
		return nil, NewError(pb.ERR_BAD_PARAM, "get_info_ids is nil")
	}
	switch m.Req.GetMultiInfos.GetInfoType() {
	case "chatroom":
		requestStat.AtomicAddMltiInfos()
		detail, err := session.QueryChatRoom(gwp.Property["Sender"], gwp.Property["Appid"], string(m.Req.GetMultiInfos.SParameter))
		if err != nil {
			return nil, NewError(pb.ERR_BAD_PARAM, err.Error())
		}
		infos, err := getCachedChatRoomMessage(m.Req.GetMultiInfos, gwp.Property["Appid"])
		if err != nil {
			return nil, NewError(pb.ERR_BAD_PARAM, err.Error())
		}
		getInfoResp := &pb.Message{
			Msgid: proto.Uint32(pb.GET_MULTI_INFOS_RESP),
			Sn:    m.Sn,
			Resp: &pb.Response{
				GetMultiInfos: &pb.GetMultiInfosResp{
					InfoType:   m.Req.GetMultiInfos.InfoType,
					SParameter: m.Req.GetMultiInfos.SParameter,
					Infos:      infos,
					LastInfoId: proto.Int64(int64(detail.MaxID)),
				},
			},
		}
		ds, err := proto.Marshal(getInfoResp)
		if err != nil {
			return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "pb marshal error:"+err.Error())
		}
		gwr.XimpBuff = &network.XimpBuffer{
			IsDecrypt:  true,
			DataStream: ds,
		}
		return getInfoResp, nil

	default:
		return nil, NewError(pb.ERR_BAD_PARAM, "info_type not supported")
	}
}

// 将service的回包切分一下记录日志, 只保留一定长度
func cutServiceResp(s string) string {
	if len(s) > DEFAULT_LOG_INNER_SERVICE_RESP_LEN {
		return s[0:DEFAULT_LOG_INNER_SERVICE_RESP_LEN]
	}
	return s
}

func dealServiceReq(gwp *router.GwPackage, m *pb.Message, gwr *router.GwResp) (*pb.Message, *Error) {
	if m.Req.ServiceReq == nil || m.Req.ServiceReq.ServiceId == nil || len(m.Req.ServiceReq.Request) == 0 {
		return nil, NewError(pb.ERR_BAD_PARAM, "field service_req|service_id|request is nil")
	}
	switch *m.Req.ServiceReq.ServiceId {
	case pb.CHATROOM_SERVICE_ID:
		crm := &pb.ChatRoomPacket{}
		if err := proto.Unmarshal(m.Req.ServiceReq.Request, crm); err != nil {
			return nil, NewError(pb.ERR_BAD_PARAM, "unmarshal chatroom service:"+err.Error())
		}
		if len(crm.Roomid) == 0 || crm.ToServerData == nil || crm.ToServerData.Payloadtype == nil {
			return nil, NewError(pb.ERR_BAD_PARAM, "field roomid or to_server_data is nil")
		}
		crmResp, rErr := dealChatroom(gwp, crm, gwr)
		if rErr != nil {
			Logger.Error(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealServiceReq", "dealCahtroom error", rErr)
			return nil, rErr
		}
		Logger.Debug(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealServiceReq", crm.String(), cutServiceResp(crmResp.String()))
		crmStream, err := proto.Marshal(crmResp)
		if err != nil {
			return nil, NewError(pb.ERR_ROUTER_EXCEPTION, "marshal crm error:"+err.Error())
		}
		respM, ds, err := genMessageResp(m.GetSn(), pb.CHATROOM_SERVICE_ID, crmStream)
		if err != nil {
			return nil, GenError(pb.ERR_ROUTER_EXCEPTION, err)
		}
		gwr.XimpBuff = &network.XimpBuffer{
			IsDecrypt:  true,
			DataStream: ds,
		}
		return respM, nil
	default:
		return nil, NewError(pb.ERR_BAD_PARAM, "unknown service id")
	}
}

// 校验verf_code字段
func verifyCode(sender, verfCode string) bool {
	sum := md5.Sum([]byte(sender + logic.VERF_CODE_SALT))
	if fmt.Sprintf("%x", sum)[24:32] != verfCode {
		return false
	}
	return true
}

// 校验secret_ram，包括调用业务方的接口（登录时）, verfCode
func verify(m *pb.Message, gwp *router.GwPackage) (bool, []byte, string, *Error) {
	sender := gwp.Property["Sender"]
	connectionType := gwp.Property["ConnectionType"]
	appid := logic.StringToUint16(gwp.Property["Appid"])
	//secretRam := m.Req.Login.SecretRam
	//serverRam := gwp.Property["ServerRam"]

	// 用户的 token
	var key []byte

	// 是否是登录用户/游客
	isLoginUser := false

	// 后续传输的加密 key，当不加密时，会被置空
	sessionKey := GenSessionKey()

	if sig, ok := gwp.Property["Sig"]; ok && len(sig) > 0 {
		cversion, _ := strconv.Atoi(gwp.Property["CVersion"])
		defaultKey := logic.GetDefaultKey(appid, uint16(cversion))

		// 登陆用户
		key = defaultKey
		isLoginUser = true

		// passport 降级开启，使用 default key 加密
		if netConf().CallbackDisable {
			return isLoginUser, key, sessionKey, nil
		}

		// passport 降级关闭
		if token, err := checkSig(gwp.Property["Appid"], sender, sig); err != nil {
			_ = token
			return isLoginUser, nil, "", NewError(pb.ERR_USER_INVALID, "check sig error:"+err.Error())
			//} else {
			//	key = token
		}
	} else {
		switch session.GetUserType(sender, appid, connectionType) {
		case session.ChatRoomWebUserPrefix + session.RegisteredChatRoomUser:
			if netConf().CheckWebUser {
				return isLoginUser, nil, "", NewError(pb.ERR_USER_INVALID, "check sig error: param not contains sig")
			}
		case session.RegisteredChatRoomUser:
			if netConf().CheckAppUser {
				return isLoginUser, nil, "", NewError(pb.ERR_USER_INVALID, "check sig error: param not contains sig")
			}
		}

		// 游客

		// 如果是游客，那么这个值就是用户id
		key = []byte(sender)

		// 只有不登录用户才校验verf_code
		if !verifyCode(sender, m.Req.Login.GetVerfCode()) {
			return isLoginUser, nil, "", NewError(pb.ERR_BAD_PARAM, "invalid verf_code")
		}
	}

	return isLoginUser, key, sessionKey, nil
	//if verifySecretRam(secretRam, []byte(serverRam), key) {
	//	return isLoginUser, key, sessionKey, nil
	//} else {
	//	return isLoginUser, nil, "", NewError(pb.ERR_USER_INVALID, "secret_ram verify failed")
	//}
}

func verifySecretRam(secretRam, serverRam, key []byte) bool {
	sr, err := cryption.Rc4Decrypt(secretRam, key)
	if err != nil || len(sr) == 0 || len(sr) < len(serverRam) || bytes.Compare(sr[:len(serverRam)], serverRam) != 0 {
		Logger.Debug(secretRam, serverRam, key, "verifySecretRam", "fail", sr, err)
		return false
	}
	return true
}

func GenSessionKey() string {
	return ServerRam()
}

func ServerRam() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.FormatInt(rand.Int63(), 10)
}

func handleOldSession(s []*session.UserSession) {
	kickUsers := make(map[string][]logic.ConnectionId)
	for _, v := range s {
		kickUsers[v.GatewayAddr] = append(kickUsers[v.GatewayAddr], v.ConnectionId)
	}
	if len(kickUsers) == 0 {
		return
	}
	gwr := &router.GwResp{
		Actions: []int{router.DisconnectAction},
	}
	for k, v := range kickUsers {
		if err := gateway.DoOperations(k, nil, "", v, gwr); err != nil {
			Logger.Error(v, k, "", "handleOldSession", "DoOperations err", err)
		}
	}
}

func genMessageResp(sn uint64, serviceId uint32, response []byte) (*pb.Message, []byte, error) {
	resp := &pb.Message{
		Msgid: proto.Uint32(pb.SERVICE_RESP),
		Sn:    &sn,
		Resp: &pb.Response{
			ServiceResp: &pb.Service_Resp{
				ServiceId: &serviceId,
				Response:  response,
			},
		},
	}
	if ds, err := proto.Marshal(resp); err != nil {
		return nil, nil, err
	} else {
		return resp, ds, nil
	}

}
