package saver

import (
	"time"

	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client = func() *gorpc.Client {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	g := gorpc.NewClient(netOptions)
	statNetOption := gorpc.NewNetOptions(1*time.Second, 1*time.Second, 1*time.Second)
	g.SetMethodNetOptinons("GorpcService", "GetSaverQps", statNetOption)
	g.SetMethodNetOptinons("GorpcService", "GetSaverTotalOps", statNetOption)
	return g
}()

func SetMethodNetOptinons(service string, netOptionsList map[string]*gorpc.NetOptions) error {

	for method, netOptions := range netOptionsList {
		if err := GorpcClient.SetMethodNetOptinons(service, method, netOptions); err != nil {
			return err
		}
	}
	return nil
}

func SaveSession(userSession *session.UserSession) ([]*session.UserSession, error) {
	resp := []*session.UserSession{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "SaveSession", userSession, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func RemoveSession(userSession *session.UserSession) error {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "RemoveSession", userSession, &resp); err != nil {
		return err
	}
	return nil
}

func QueryUserSession(query []*session.UserSession) ([]*session.UserSession, error) {
	resp := []*session.UserSession{}

	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryUserSession", query, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func QueryUserSessionSummary(appid uint16, user string) (map[string]int, error) {
	req := &session.UserSession{
		AppId:  appid,
		UserId: user,
	}
	var resp map[string]int

	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryUserSessionSummary", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func GetActiveUserNum(appid uint16) (int, error) {
	var num int
	req := &GetActiveReq{
		AppId: appid,
	}

	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "GetActiveUserNum", req, &num); err != nil {
		return 0, err
	}
	return num, nil
}

func GetActiveUserSessions(appid uint16, zoneid uint16) ([]*session.UserSession, error) {
	resp := []*session.UserSession{}
	req := &GetActiveReq{
		AppId:  appid,
		ZoneId: zoneid,
	}

	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "GetActiveUserSessions", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func CheckUserSessionInRoom(req *session.UserChatRoomRequest) (bool, error) {
	var resp bool
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "CheckUserSessionInRoom", req, &resp); err != nil {
		return false, err
	}
	return resp, nil
}

func QueryUserSessionChatRoomList(req *session.UserChatRoomRequest) ([]string, error) {
	var resp []string
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryUserSessionChatRoomList", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func GetActiveChatRoomNum(appid uint16) (int, error) {
	var num int
	req := &GetActiveReq{
		AppId: appid,
	}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "GetActiveChatRoomNum", req, &num); err != nil {
		return 0, err
	}
	return num, nil
}

func AddChatRoomUser(req *session.UserChatRoomRequest) (int, error) {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "AddChatRoomUser", req, &resp); err != nil {
		return 0, err
	}
	return resp, nil
}

func RemoveChatRoomUser(req *session.UserChatRoomRequest) (int, error) {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "RemoveChatRoomUser", req, &resp); err != nil {
		return 0, err
	}
	return resp, nil
}

func CreateChatRoom(req *session.UserChatRoomRequest) error {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "CreateChatRoom", req, &resp); err != nil {
		return err
	}
	return nil
}

func AddChatRoomRobot(req *session.UserChatRoomRequest) error {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "AddChatRoomRobot", req, &resp); err != nil {
		return err
	}
	return nil
}

func RemoveChatRoomRobot(req *session.UserChatRoomRequest) error {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "RemoveChatRoomRobot", req, &resp); err != nil {
		return err
	}
	return nil
}

func UpdateChatRoomMember(req *ChatRoomMemberUpdate) error {
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "UpdateChatRoomMember", req, &resp); err != nil {
		return err
	}
	return nil
}

func QueryChatRoomMemberCount(rooms []string, appid uint16) (map[string]map[string]int, error) {
	req := &ChatRoomsAppid{
		Appid:   appid,
		RoomIDs: make([]string, len(rooms)),
	}
	for i, roomid := range rooms {
		req.RoomIDs[i] = roomid
	}
	var resp = map[string]map[string]int{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryChatRoomMemberCount", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func QueryChatRoomDetail(room string, appid uint16) (*session.ChatRoomDetail, error) {
	req := &ChatRoomsAppid{
		Appid:   appid,
		RoomIDs: []string{room},
	}
	var resp = map[string]*session.ChatRoomDetail{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryChatRoomDetail", req, &resp); err != nil {
		return nil, err
	}
	return resp[room], nil
}

func QueryChatRoomsByZone(appid uint16, zoneid uint16) (map[string]*session.ChatRoomDetail, error) {
	req := &GetActiveReq{
		AppId:  appid,
		ZoneId: zoneid,
	}
	var resp = map[string]*session.ChatRoomDetail{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryChatRoomsByZone", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func QueryChatRoomUsers(room string, appid uint16) ([]string, error) {
	req := &ChatRoomsAppid{
		Appid:   appid,
		RoomIDs: []string{room},
	}
	var resp = map[string][]string{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryChatRoomUsers", req, &resp); err != nil {
		return nil, err
	}
	return resp[room], nil
}

//判断用户是否在某直播间
func QueryUserInRoom(room string, user string, appid uint16) (bool, error) {
	req := &UserChatRoom{
		User:   user,
		RoomId: room,
		Appid:  appid,
	}
	var resp bool
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "QueryUserInRoom", req, &resp); err != nil {
		return false, err
	}
	return resp, nil
}

func CleanChatRoom(room string, appid uint16) error {
	req := &ChatRoomsAppid{
		Appid:   appid,
		RoomIDs: []string{room},
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "CleanChatRoom", req, &resp); err != nil {
		return err
	}
	return nil
}

func CacheChatRoomMessage(req *logic.ChatRoomMessage) (uint, error) {
	var resp uint
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "CacheChatRoomMessage", req, &resp); err != nil {
		return 0, err
	}
	return resp, nil
}

func GetCachedChatRoomMessages(req *FetchChatRoomMessageReq) (map[uint]*logic.ChatRoomMessage, error) {
	resp := map[uint]*logic.ChatRoomMessage{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "GetCachedChatRoomMessages", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

/*
 * Stores chat messages to storage
 * @param req is a saver.StoreMessagesRequest point which include messages information need to store
 * @return (*StoreMessagesResponse, nil) if no error occurs, otherwise (nil, error) is returned
 */
func StoreChatMessages(req *StoreMessagesRequest) (*StoreMessagesResponse, error) {
	resp := &StoreMessagesResponse{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "StoreChatMessages", req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

/*
 * Retrieves chat messages from storage
 * @param req is a saver.RetrieveMessagesRequest point which include retrieving request information
 * @param resp is a saver.RetrieveMessagesResponse point which includes response
 * @return (*RetrieveMessagesResponse, nil) if no error occurs, otherwise (nil, error) is returned
 */
func RetrieveChatMessages(req *RetrieveMessagesRequest) (*RetrieveMessagesResponse, error) {
	resp := &RetrieveMessagesResponse{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "RetrieveChatMessages", req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// 获取未读消息数
func RetrieveUnreadCount(appid uint16, userIds, channels []string) ([]*RetrieveMessagesResponse, error) {
	channel := make(map[string]*RetrieveChannel, len(channels))
	for _, ch := range channels {
		channel[ch] = &RetrieveChannel{
			Channel: ch,
		}
	}
	req := make([]*RetrieveMessagesRequest, 0, len(userIds))
	for _, uid := range userIds {
		req = append(req, &RetrieveMessagesRequest{
			Appid:        appid,
			Owner:        uid,
			ChatChannels: channel,
		})
	}

	var resp []*RetrieveMessagesResponse
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "RetrieveUnreadCount", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

/*
 * Set chat messages recall flag to storage
 * @param req is a saver.RecallMessagesRequest point which include recall request information
 * @param resp is a saver.StoreMessagesResponse point which includes response
 * @return (*StoreMessagesResponse, nil) if no error occurs, otherwise (nil, error) is returned
 */
func RecallChatMessages(req *RecallMessagesRequest) (*StoreMessagesResponse, error) {
	resp := &StoreMessagesResponse{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "RecallChatMessages", req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func AddChatroomCountKeys(appid uint16, roomids []string) error {
	req := &ChatroomCountKeysRequest{
		Appid:   appid,
		Roomids: roomids,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "AddChatroomCountKeys", req, &resp); err != nil {
		return err
	}
	return nil
}

func DelChatroomCountKeys(appid uint16, roomids []string) error {
	req := &ChatroomCountKeysRequest{
		Appid:   appid,
		Roomids: roomids,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "DelChatroomCountKeys", req, &resp); err != nil {
		return err
	}
	return nil
}

func GetChatroomCountKeys(appid uint16, rpcIndex, rpcLength int, clientIP string) (*ChatroomCountKeysResponse, error) {
	req := &GetChatroomCountKeysRequest{
		Appid:     appid,
		ClientIP:  clientIP,
		RpcIndex:  rpcIndex,
		RpcLength: rpcLength,
	}
	resp := &ChatroomCountKeysResponse{}
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "GetChatroomCountKeys", req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func AddQpsCount(module, api string, limit int) (bool, error) {
	req := &QpsCount{
		Module: module,
		API:    api,
		Limit:  limit,
	}
	var resp bool
	if err := GorpcClient.CallWithAddress(logic.GetSaverGorpc(), "GorpcService", "AddQpsCount", req, &resp); err != nil {
		return false, err
	}
	return resp, nil
}
