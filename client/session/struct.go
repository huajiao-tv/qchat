package session

import (
	"strconv"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/network"
)

type UserSession struct {
	UserId     string
	SessionId  string
	SessionKey string
	LoginTime  string //uint64
	AppId      uint16
	Deviceid   string
	ClientVer  string //uint16
	SenderType string
	Platform   string
	UserIp     string
	TraceId    string

	IsLoginUser  bool
	ConnectionId logic.ConnectionId
	GatewayAddr  string
	Property     map[string]string
	//Gatewayname string //move to property
	//UserPhone   string
	//MobileType  string
	//UserQid     string
}

type OpenSessionReq struct {
	*UserSession
}

type OpenSessionResp struct {
	OldUserSessions []*UserSession
	Tags            []string
}

type CloseSessionReq struct {
	*UserSession
}

type CloseSessionResp struct {
	Tags []string
}

type QuerySessionReq struct {
	QueryUserSessions []*UserSession
}

type QuerySessionResp struct {
	ResultUserSessions []*UserSession
}

type UserChatRoomRequest struct {
	User         string
	UserType     string
	Appid        uint16
	ConnectionId logic.ConnectionId
	GatewayAddr  string
	RoomID       string
	IsJoin       bool
	AudienceList bool
}

type UserChatRoomResponse struct {
	Code     int
	Reason   string
	Response []byte
}

type GetMsgInfo struct {
	User         string
	Appid        uint16
	ConnectionId logic.ConnectionId
	GatewayAddr  string
	InfoType     string
	InfoId       int64 // lastId+1
	InfoOffset   int32 // 最多返回多少条
	SParameter   []byte
}

const (
	RegisteredChatRoomUser    = "reg"
	NonRegisteredChatRoomUser = "noreg"
	RobotChatRoomUser         = "fake"

	ChatRoomWebUserPrefix = "web_"
)

const (
	Success = iota
	CallbackFailed
	AddUserFailed

	UserIsRobot              int = 100
	UserAlreadyInChatRoom        = 101
	SessionAlreadyInChatRoom     = 102
	AllSessionQuitedChatRoom     = 103
	SessionNotInChatRoom         = 104
)

const (
	NormalUser    int = 0
	HighLevelUser     = 1
)

const (
	GroupOwner      uint8 = 255
	GroupNormalUser uint8 = 0
)

type ChatRoomDetail struct {
	AppID        uint16
	RoomID       string
	CreateTime   int64
	Version      uint
	MaxID        uint
	Members      map[string]int
	GatewayAddrs map[string]int
}

// pb协议中的注册用户数，包括普通注册用户 + 机器人
func (this *ChatRoomDetail) Registered() int {
	return this.Members[RegisteredChatRoomUser] + this.Members[RobotChatRoomUser] + this.Members[ChatRoomWebUserPrefix+RegisteredChatRoomUser]
}

// pb协议中的成员数，包括所有用户
func (this *ChatRoomDetail) MemberCount() int {
	return this.Members[RegisteredChatRoomUser] + this.Members[RobotChatRoomUser] + this.Members[NonRegisteredChatRoomUser] +
		this.Members[ChatRoomWebUserPrefix+RegisteredChatRoomUser] + this.Members[ChatRoomWebUserPrefix+NonRegisteredChatRoomUser]
}

// 长连用户数
func (this *ChatRoomDetail) ConnCount() int {
	return this.Members[RegisteredChatRoomUser] + this.Members[NonRegisteredChatRoomUser] +
		this.Members[ChatRoomWebUserPrefix+RegisteredChatRoomUser] + this.Members[ChatRoomWebUserPrefix+NonRegisteredChatRoomUser]
}

func GetUserType(user string, appid uint16, connectionType string) string {
	ut := RegisteredChatRoomUser
	if len(user) > 12 { // id 大于 12 位的是游客
		ut = NonRegisteredChatRoomUser
	} else if _, err := strconv.Atoi(user); err != nil {
		ut = NonRegisteredChatRoomUser
	}
	switch connectionType {
	case network.WebSocketNetwork:
		return ChatRoomWebUserPrefix + ut
	default:
		return ut
	}
}

func CheckUserType(userType string) bool {
	switch userType {
	case RegisteredChatRoomUser, NonRegisteredChatRoomUser, RobotChatRoomUser:
		return true
	case ChatRoomWebUserPrefix + RegisteredChatRoomUser, ChatRoomWebUserPrefix + NonRegisteredChatRoomUser:
		return true
	default:
		return false
	}
}

type JoinChatRoomRequest struct {
	*UserChatRoomRequest
	Properties map[string]string
}

type JoinChatRoomResponse struct {
	*UserChatRoomResponse
	*ChatRoomDetail
}

type QuitChatRoomRequest struct {
	*UserChatRoomRequest
	Properties map[string]string
}

type QuitChatRoomResponse struct {
	*UserChatRoomResponse
}

// 构造加入/退出成功返回
var (
	JoinChatRoomSuccessResp = &JoinChatRoomResponse{
		UserChatRoomResponse: &UserChatRoomResponse{
			Code:     Success,
			Reason:   "success",
			Response: []byte("response"),
		},
		ChatRoomDetail: &ChatRoomDetail{
			AppID:   logic.APPID_HUAJIAO,
			Members: make(map[string]int),
			//GatewayAddrs: logic.GetGatewayGorpcMap(),
		},
	}
	QuitChatRoomSuccessResp = &QuitChatRoomResponse{
		UserChatRoomResponse: &UserChatRoomResponse{
			Code:     Success,
			Reason:   "success",
			Response: []byte(""),
		},
	}
)

type UserOnlineCache struct {
	AppId   uint16
	UserIds []string
}
