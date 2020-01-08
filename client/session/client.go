package session

import (
	"errors"
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)

	statNetOption := gorpc.NewNetOptions(1*time.Second, 1*time.Second, 1*time.Second)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetSessionQps", statNetOption)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetSessionTotalOps", statNetOption)
}

func Close(userId string, appId uint16, gatewayAddr string, connectionId logic.ConnectionId, property map[string]string) (*CloseSessionResp, error) {

	platform, ok := property["Platform"]
	if !ok {
		return nil, errors.New("lack Platform field")
	}

	loginTime, ok := property["LoginTime"]
	if !ok {
		return nil, errors.New("lack LoginTime field")
	}

	deviceId, ok := property["Deviceid"]
	if !ok {
		return nil, errors.New("lack Deviceid field")
	}

	traceid := logic.GetTraceId(gatewayAddr, connectionId)

	usersess := &UserSession{
		UserId:       userId,
		UserIp:       property["ClientIp"],
		LoginTime:    loginTime,
		AppId:        appId,
		Deviceid:     deviceId,
		Platform:     platform,
		ConnectionId: connectionId,
		GatewayAddr:  gatewayAddr,
		TraceId:      traceid,
		Property:     property,
	}

	req := &CloseSessionReq{
		UserSession: usersess,
	}

	resp := &CloseSessionResp{}

	if err := GorpcClient.CallWithAddress(logic.GetStatedSessionGorpc(userId), "GorpcService", "Close", req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func Query(querys []*UserSession) ([]*UserSession, error) {

	req := &QuerySessionReq{
		QueryUserSessions: querys,
	}

	resp := &QuerySessionResp{}

	if err := GorpcClient.CallWithAddress(logic.GetSessionGorpc(), "GorpcService", "Query", req, resp); err != nil {
		return nil, err
	}
	return resp.ResultUserSessions, nil
}

func Open(userId string, appId uint16, isLoginUser bool, sessionKey string,
	connectionId logic.ConnectionId, gatewayAddr string, property map[string]string) (*OpenSessionResp, error) {

	sessionId, ok := property["ServerRam"]
	if !ok {
		return nil, errors.New("lack serverram field")
	}

	senderType, ok := property["SenderType"]
	if !ok {
		return nil, errors.New("lack SenderType field")
	}

	platform, ok := property["Platform"]
	if !ok {
		return nil, errors.New("lack Platform field")
	}

	userIp, ok := property["ClientIp"]
	if !ok {
		return nil, errors.New("lack ClientIp field")
	}

	loginTime, ok := property["LoginTime"]
	if !ok {
		return nil, errors.New("lack LoginTime field")
	}

	deviceId, ok := property["Deviceid"]
	if !ok {
		return nil, errors.New("lack Deviceid field")
	}

	cversion, ok := property["CVersion"]
	if !ok {
		return nil, errors.New("lack CVersion field")
	}

	traceid := logic.GetTraceId(gatewayAddr, connectionId)

	usersess := &UserSession{
		UserId:     userId,
		SessionId:  sessionId,
		SessionKey: sessionKey,
		LoginTime:  loginTime,
		AppId:      appId,
		Deviceid:   deviceId,
		ClientVer:  cversion,
		SenderType: senderType, //jid or phone ?
		Platform:   platform,
		UserIp:     userIp,

		TraceId:      traceid,
		IsLoginUser:  isLoginUser,
		ConnectionId: connectionId,
		GatewayAddr:  gatewayAddr,
		Property:     property,
	}

	req := &OpenSessionReq{
		UserSession: usersess,
	}

	resp := &OpenSessionResp{}

	if err := GorpcClient.CallWithAddress(logic.GetStatedSessionGorpc(userId), "GorpcService", "Open", req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

//@todo 返回值 还不知道 是什么
func GetInfo(user string, appid uint16, connId logic.ConnectionId, gatewayAddr string, infoType string, infoId int64, infoOffset int32, sParameter []byte) (int, error) {
	req := &GetMsgInfo{
		user,
		appid,
		connId,
		gatewayAddr,
		infoType,
		infoId,
		infoOffset,
		sParameter,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSessionGorpc(), "GorpcService", "GetInfo", req, &resp); err != nil {
		return 0, err
	}
	return resp, nil
}

func JoinChatRoom(user string, appid string, gateway string, connectionId logic.ConnectionId, connectionType string, room string, props map[string]string) (*JoinChatRoomResponse, error) {
	appID := logic.StringToUint16(appid)
	if appID == 0 {
		appID = logic.DEFAULT_APPID
	}
	req := &JoinChatRoomRequest{
		UserChatRoomRequest: &UserChatRoomRequest{
			User:         user,
			UserType:     GetUserType(user, appID, connectionType),
			Appid:        appID,
			ConnectionId: connectionId,
			GatewayAddr:  gateway,
			RoomID:       room,
			IsJoin:       true,
		},
		Properties: props,
	}
	resp := &JoinChatRoomResponse{}
	if err := GorpcClient.CallWithAddress(logic.GetStatedSessionGorpc(user), "GorpcService", "JoinChatRoom", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func QuitChatRoom(user string, appid string, gateway string, connectionId logic.ConnectionId, connectionType string, room string, props map[string]string) (*QuitChatRoomResponse, error) {
	appID := logic.StringToUint16(appid)
	if appID == 0 {
		appID = logic.DEFAULT_APPID
	}
	req := &QuitChatRoomRequest{
		UserChatRoomRequest: &UserChatRoomRequest{
			User:         user,
			UserType:     GetUserType(user, appID, connectionType),
			Appid:        appID,
			ConnectionId: connectionId,
			GatewayAddr:  gateway,
			RoomID:       room,
			IsJoin:       false,
		},
		Properties: props,
	}
	resp := &QuitChatRoomResponse{
		UserChatRoomResponse: &UserChatRoomResponse{
			Code:   0,
			Reason: "",
		},
	}
	if err := GorpcClient.CallWithAddress(logic.GetStatedSessionGorpc(user), "GorpcService", "QuitChatRoom", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

//properties用于保存机器人其他请求参数（除去uid, appid, rid之外）
func AddRobotIntoChatRoom(user string, appid string, room string, properties map[string]string) (*JoinChatRoomResponse, error) {
	appID := logic.StringToUint16(appid)
	if appID == 0 {
		appID = logic.DEFAULT_APPID
	}
	req := &JoinChatRoomRequest{
		UserChatRoomRequest: &UserChatRoomRequest{
			User:     user,
			UserType: RobotChatRoomUser,
			Appid:    appID,
			RoomID:   room,
			IsJoin:   true,
		},
		Properties: properties,
	}
	resp := &JoinChatRoomResponse{}
	if err := GorpcClient.CallWithAddress(logic.GetStatedSessionGorpc(user), "GorpcService", "JoinChatRoom", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func RemoveRobotFromChatRoom(user string, appid string, room string) (*QuitChatRoomResponse, error) {
	appID := logic.StringToUint16(appid)
	if appID == 0 {
		appID = logic.DEFAULT_APPID
	}
	req := &QuitChatRoomRequest{
		UserChatRoomRequest: &UserChatRoomRequest{
			User:     user,
			UserType: RobotChatRoomUser,
			Appid:    appID,
			RoomID:   room,
			IsJoin:   false,
		},
	}
	resp := &QuitChatRoomResponse{
		UserChatRoomResponse: &UserChatRoomResponse{
			Code:   0,
			Reason: "",
		},
	}
	if err := GorpcClient.CallWithAddress(logic.GetStatedSessionGorpc(user), "GorpcService", "QuitChatRoom", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil

}

func QueryChatRoom(user string, appid string, room string) (*ChatRoomDetail, error) {
	req := &UserChatRoomRequest{
		User:   user,
		Appid:  logic.StringToUint16(appid),
		RoomID: room,
	}
	resp := &ChatRoomDetail{}
	if err := GorpcClient.CallWithAddress(logic.GetSessionGorpc(), "GorpcService", "QueryChatRoom", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func QueryUserOnlineCache(addr string, appid uint16, users []string) (map[string][]*logic.UserGateway, error) {
	req := &UserOnlineCache{
		AppId:   appid,
		UserIds: users,
	}
	resp := map[string][]*logic.UserGateway{}
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "QueryUserOnlineCache", req, &resp); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

// 获取online_cache的统计信息
func OnlineCacheStat(addr string) (map[string][]map[string]uint64, error) {
	var req int
	var resp map[string][]map[string]uint64
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "OnlineCacheStat", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
