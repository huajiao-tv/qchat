package main

import (
	//"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/msgRedis"
)

const (
	userStatSetKey       = "userstat:%d:%d:set" //appid, sn:0-99
	userstatSetsCount    = 100
	userStatSetVal       = "%s" //userid
	usersessionmasterkey = "usersession:%s:%s:sets"
	usersessionsecondkey = "usersession:%s:%s:%s:%s:hset"
)

func getUserSessionKey(property map[string]string) (mkey string, skey string) {

	if _, ok := property["UserId"]; !ok {
		mkey = ""
		skey = ""
		return
	}

	if _, ok := property["AppId"]; !ok {
		mkey = ""
		skey = ""
		return
	}

	mkey = fmt.Sprintf(usersessionmasterkey, property["UserId"], property["AppId"])

	if _, ok := property["Platform"]; !ok {
		skey = ""
		return
	}

	if _, ok := property["Deviceid"]; !ok {
		skey = ""
		return
	}

	skey = fmt.Sprintf(usersessionsecondkey, property["UserId"], property["AppId"], property["Platform"], property["Deviceid"])
	return
}

func addUserStat(masterkey, userid, appid string) error {
	statkey, statval := getActiveUserKey(userid, appid)
	num, err := SessionPool.Call(getSessionAddr(masterkey)).SCARD(masterkey)
	if err != nil {
		return errors.New("call masterkey scard fail: " + err.Error())
	}

	if num == 1 {
		if _, err := SessionPool.Call(getSessionAddr(statkey)).SADD(statkey, []string{statval}); err != nil {
			return errors.New("call statkey sadd fail: " + err.Error())
		}
	}

	return nil
}

func remUserStat(masterkey, userid, appid string) error {
	statkey, statval := getActiveUserKey(userid, appid)
	num, err := SessionPool.Call(getSessionAddr(masterkey)).SCARD(masterkey)
	if err != nil {
		return errors.New("call masterkey scard fail: " + err.Error())
	}

	if num == 0 {
		if _, err := SessionPool.Call(getSessionAddr(statkey)).SREM(statkey, []string{statval}); err != nil {
			return errors.New("call statkey srem fail: " + err.Error())
		}
	}

	return nil
}

func getActiveUserKey(userid, appid string) (string, string) {
	usersum := logic.Sum(userid)
	appidn, _ := strconv.Atoi(appid)
	return fmt.Sprintf(userStatSetKey, appidn, usersum%userstatSetsCount), fmt.Sprintf(userStatSetVal, userid)
}

func getActiveUserNum(appid uint16) (int, error) {

	var num int
	for i := 0; i < userstatSetsCount; i++ {
		keyname := fmt.Sprintf(userStatSetKey, appid, i)
		cnum, err := SessionPool.Call(getSessionAddr(keyname)).SCARD(keyname)
		if err != nil {
			return num, errors.New(keyname + ":" + err.Error())
		}
		num += int(cnum)
	}

	return num, nil
}

//zone: 0~userstatSetsCount
func getActiveUserInzone(appid uint16, zoneid uint16) ([]*session.UserSession, error) {

	var resp []*session.UserSession

	keyname := fmt.Sprintf(userStatSetKey, appid, zoneid)

	userids, err := SessionPool.Call(getSessionAddr(keyname)).SMEMBERS(keyname)
	if err != nil {
		return resp, errors.New(keyname + ":" + err.Error())
	}

	appidstr := strconv.FormatUint(uint64(appid), 10)

	for _, userid := range userids {
		masterkey := fmt.Sprintf(usersessionmasterkey, userid, appidstr)

		secondkeys, err := SessionPool.Call(getSessionAddr(masterkey)).SMEMBERS(masterkey)
		if err != nil {
			Logger.Error(userid, appidstr, "", "getActiveUserInzone", "SMEMBERS masterkey error", "")
			continue
		} else {
			for _, secondkey := range secondkeys {
				props, err := SessionPool.Call(getSessionAddr(masterkey)).HGETALLMAP(string(secondkey))
				if err != nil && err != msgRedis.ErrKeyNotExist {
					continue //must not be here
				}

				if err == nil {
					userSession, err := prop2usersession(props)
					if err == nil {
						resp = append(resp, userSession)
					}
				}
			}
		}

		//remove duty masterkey
		if len(secondkeys) == 0 {
			SessionPool.Call(getSessionAddr(keyname)).SREM(keyname, []string{string(userid)})
		}
	}

	return resp, nil
}

func saveUserSession(property map[string]string) (map[string]string, error) {

	masterkey, secondkey := getUserSessionKey(property)

	if masterkey == "" || secondkey == "" {
		return nil, errors.New("lack MUST field")
	}

	var old map[string]string

	if id, err := SessionPool.Call(getSessionAddr(masterkey)).SADD(masterkey, []string{secondkey}); err != nil {
		return nil, err
	} else if id == 0 {
		Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.saveUserSesiion", "same key", fmt.Sprintf("%s-%s", masterkey, secondkey))

		if props, err := SessionPool.Call(getSessionAddr(masterkey)).HGETALLMAP(secondkey); err != nil && err != msgRedis.ErrKeyNotExist {
			return nil, err
		} else {
			old = props
		}
	}

	values := make(map[string]interface{})
	for k, v := range property {
		values[k] = v
	}

	if _, err := SessionPool.Call(getSessionAddr(masterkey)).HMSET(secondkey, values); err != nil {
		return nil, err
	}

	if err := addUserStat(masterkey, property["UserId"], property["AppId"]); err != nil {
		Logger.Error(property["UserId"], property["AppId"], property["TraceId"], "saver.saveUserSesiion", "addUserStat", err.Error())
	}

	return old, nil
}

func removeUserSession(property map[string]string) error {
	masterkey, secondkey := getUserSessionKey(property)

	if masterkey == "" {
		return errors.New("lack userid or appid field")
	}

	if secondkey == "" {
		if secondkeys, err := SessionPool.Call(getSessionAddr(masterkey)).SMEMBERS(masterkey); err != nil {
			return err
		} else {
			for _, oldsecondkey := range secondkeys {
				if num, err := SessionPool.Call(getSessionAddr(masterkey)).DEL(string(oldsecondkey)); err != nil {
					Logger.Error(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "del skey error", oldsecondkey)
					return errors.New(string(oldsecondkey) + err.Error())
				} else if num == 0 {
					Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "skey not exist", oldsecondkey)
				}
			}

			if num, err := SessionPool.Call(getSessionAddr(masterkey)).DEL(masterkey); err != nil {
				return errors.New(masterkey + err.Error())
			} else if num == 0 {
				Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "mkey not exist", masterkey)
			}
		}
	} else {
		oldprop, err := SessionPool.Call(getSessionAddr(masterkey)).HGETALLMAP(secondkey)
		if err == msgRedis.ErrKeyNotExist {
			Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "secondkey not exist", secondkey)
			return nil
		}

		if err != nil {
			Logger.Error(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "redis error: "+err.Error(), secondkey)
			return errors.New("redis error: " + err.Error())
		}

		if oldprop["ConnectionId"] != property["ConnectionId"] {
			Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "current usersession have differenet connectid ", property["ConnectionId"])
			return nil
		}

		if oldprop["GatewayAddr"] != property["GatewayAddr"] {
			Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "current usersession have differenet gatewayAddr ", property["GatewayAddr"])
			return nil
		}

		if id, err := SessionPool.Call(getSessionAddr(masterkey)).DEL(secondkey); err != nil {
			return errors.New(secondkey + err.Error())
		} else if id == 0 {
			Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "skey not exist", secondkey)
		}

		if id, err := SessionPool.Call(getSessionAddr(masterkey)).SREM(masterkey, []string{secondkey}); err != nil {
			return errors.New(secondkey + err.Error())
		} else if id == 0 {
			Logger.Warn(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "mkey havenot skey", masterkey, secondkey)
		}
	}

	if err := remUserStat(masterkey, property["UserId"], property["AppId"]); err != nil {
		Logger.Error(property["UserId"], property["AppId"], property["TraceId"], "saver.removeUserSesiion", "remUserStat", err.Error())
	}

	return nil
}

// 获取用户的信息，如果后面两个参数为空，那么将把所有以这个用户登录的用户
func getUserSession(property map[string]string) (map[string]map[string]string, error) {
	masterkey, secondkey := getUserSessionKey(property)
	result := make(map[string]map[string]string)

	if masterkey == "" {
		return result, errors.New("lack userid or appid field")
	}

	if secondkey == "" {
		if secondkeys, err := SessionPool.Call(getSessionAddr(masterkey)).SMEMBERS(masterkey); err != nil {
			return result, err
		} else {
			for _, origsecondkey := range secondkeys {
				props, err := SessionPool.Call(getSessionAddr(masterkey)).HGETALLMAP(string(origsecondkey))
				if err != nil && err != msgRedis.ErrKeyNotExist {
					return result, err
				}

				if err == nil {
					result[string(origsecondkey)] = props
				}
			}
		}
	} else {
		props, err := SessionPool.Call(getSessionAddr(masterkey)).HGETALLMAP(secondkey)
		if err != nil && err != msgRedis.ErrKeyNotExist {
			return result, err
		}

		if err == nil {
			result[secondkey] = props
		}
	}

	return result, nil
}

//
const (
	roomSetsCount = 100
)

const (
	chatRoomIDSetKey           = "chatrooms:%d:%d:set"
	chatRoomPropertyHashKey    = "chatroom:property:%s:%d:hset"
	chatRoomGateWayHashKey     = "chatroom:gateway:%s:%d:hset"
	chatRoomMembersSetKey      = "chatroom:members:%s:%d:set"
	chatRoomMessageKey         = "chatroom:message:%s:%d:id:%d"
	chatRoomUserSessionHashKey = "chatroom:user:session:%s:%d:%s:%d:hset"

	createTimeField   = "time:created"
	cacheMessageField = "cache:message"
)

func getChatRoomSetKey(room string, appid uint16) string {
	sum := logic.Sum(room)
	return fmt.Sprintf(chatRoomIDSetKey, appid, sum%roomSetsCount)
}

func getChatRoomMembersSetKey(room string, appid uint16) string {
	return fmt.Sprintf(chatRoomMembersSetKey, room, appid)
}

func getChatRoomPropertyHashKey(room string, appid uint16) string {
	return fmt.Sprintf(chatRoomPropertyHashKey, room, appid)
}

func getChatRoomGateWayHashKey(room string, appid uint16) string {
	return fmt.Sprintf(chatRoomGateWayHashKey, room, appid)
}

func getChatRoomUserSessionHashKey(user string, appid uint16, gateway string, connectionId logic.ConnectionId) string {
	return fmt.Sprintf(chatRoomUserSessionHashKey, user, appid, gateway, connectionId)
}

func getChatRoomMessageKey(room string, appid uint16, id uint) string {
	return fmt.Sprintf(chatRoomMessageKey, room, appid, id)
}

func getRoomIdsByZone(appid, zone uint16) ([]string, error) {
	key := fmt.Sprintf(chatRoomIDSetKey, appid, zone)
	rooms, err := SessionPool.Call(getSessionAddr(key)).SMEMBERS(key)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(rooms))
	for _, rid := range rooms {
		ret = append(ret, string(rid))
	}
	return ret, nil
}

func getUserAllSessionsKeyForChatRoom(user string, appid uint16) ([]string, error) {
	prop := map[string]string{
		"AppId":  fmt.Sprintf("%d", appid),
		"UserId": user,
	}
	sessions, err := getUserSession(prop)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(sessions))
	for _, v := range sessions {
		connectionId, err := strconv.ParseInt(v["ConnectionId"], 10, 0)
		if err != nil {
			Logger.Error(user, appid, "saver.getUserSession", "ConnectionId err")
		} else {
			keys = append(keys, fmt.Sprintf(chatRoomUserSessionHashKey, user, appid, v["GatewayAddr"], logic.ConnectionId(connectionId)))
		}
	}
	return keys, nil
}

func getActiveChatRoomNum(appid uint16) (int, error) {
	var num int
	for i := 0; i < roomSetsCount; i++ {
		keyname := fmt.Sprintf(chatRoomIDSetKey, appid, i)
		cnum, err := SessionPool.Call(getSessionAddr(keyname)).SCARD(keyname)
		if err != nil {
			return num, errors.New(keyname + ":" + err.Error())
		}
		num += int(cnum)
	}

	return num, nil
}

func queryChatRoomMemberCount(room string, appid uint16) (map[string]int, error) {
	key := getChatRoomPropertyHashKey(room, appid)
	m, err := SessionPool.Call(getSessionAddr(room)).HGETALLMAP(key)
	if err == msgRedis.ErrKeyNotExist {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	members := make(map[string]int, len(m))
	for key, value := range m {
		switch key {
		case createTimeField:
		case cacheMessageField:
		default:
			if v, _ := strconv.Atoi(value); v < 0 {
				members[key] = 0
			} else {
				members[key] = v
			}
		}
	}
	return members, nil
}

func queryChatRoomDetail(room string, appid uint16) (*session.ChatRoomDetail, error) {
	key := getChatRoomPropertyHashKey(room, appid)
	m, err := SessionPool.Call(getSessionAddr(room)).HGETALLMAP(key)
	if err == msgRedis.ErrKeyNotExist {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	resp := &session.ChatRoomDetail{
		AppID:  appid,
		RoomID: room,
	}
	resp.Members = make(map[string]int, len(m))
	for key, value := range m {
		switch key {
		case createTimeField:
			resp.CreateTime, _ = strconv.ParseInt(value, 10, 0)
		case cacheMessageField:
			id, _ := strconv.Atoi(value)
			resp.MaxID = uint(id)
		default:
			resp.Members[key], _ = strconv.Atoi(value)
		}
	}
	if netConf().ChatroomGatewaysDegrade {
		resp.GatewayAddrs = logic.GetGatewayGorpcMap()
	} else {
		key = getChatRoomGateWayHashKey(room, appid)
		if m, err := SessionPool.Call(getSessionAddr(key)).HGETALLMAP(key); err == nil {
			resp.GatewayAddrs = make(map[string]int, len(m))
			for key, value := range m {
				resp.GatewayAddrs[key], _ = strconv.Atoi(value)
			}
		} else if err != msgRedis.ErrKeyNotExist {
			return nil, err
		}
	}
	return resp, nil
}

func updateChatRoomMemberCount(room string, appid uint16, userType string, count int) error {
	key := getChatRoomPropertyHashKey(room, appid)
	if _, err := SessionPool.Call(getSessionAddr(room)).HINCRBY(key, userType, count); err != nil {
		return err
	}
	return nil
}

func checkChatRoomExists(room string, appid uint16, create bool) (bool, error) {
	key := getChatRoomPropertyHashKey(room, appid)
	exists, err := SessionPool.Call(getSessionAddr(room)).EXISTS(key)
	if err != nil {
		return false, err
	}
	if !exists && create {
		// 聊天室不存在，新建property hash
		m := make(map[string]interface{}, 2)
		m[cacheMessageField] = 0
		m[createTimeField] = time.Now().Unix()
		if _, err := SessionPool.Call(getSessionAddr(room)).HMSET(key, m); err != nil {
			return false, err
		}
		// ChatRoom集合中增加索引
		key = getChatRoomSetKey(room, appid)
		if _, err := SessionPool.Call(getSessionAddr(key)).SADD(key, []string{room}); err != nil {
			return false, err
		}
		return true, nil
	}
	return exists, nil
}

func queryChatRoomUsers(room string, appid uint16) ([]string, error) {
	key := getChatRoomMembersSetKey(room, appid)
	ret, err := SessionPool.Call(getSessionAddr(key)).SMEMBERS(key)
	if err != nil {
		return nil, err
	}
	users := make([]string, len(ret))
	for i, uid := range ret {
		users[i] = string(uid)
	}
	return users, nil
}

func cleanChatRoom(room string, appid uint16) error {
	key := getChatRoomMembersSetKey(room, appid)
	for {
		v, err := SessionPool.Call(getSessionAddr(key)).SPOP(key)
		if err == msgRedis.ErrKeyNotExist {
			break
		} else if err != nil {
			Logger.Error("getChatRoomMembers error:", err.Error())
			// 退出，防止死循环
			break
		}
		sessions, err := getUserAllSessionsKeyForChatRoom(string(v), appid)
		if err != nil {
			continue
		}
		for _, key := range sessions {
			if _, err := SessionPool.Call(getSessionAddr(string(v))).HDEL(key, []string{room}); err != nil {
				Logger.Error("HDEL user chatroom error:", err.Error())
			}
		}
	}
	// 删除Room Property hash key
	key = getChatRoomPropertyHashKey(room, appid)
	if _, err := SessionPool.Call(getSessionAddr(room)).DEL(key); err != nil {
		Logger.Error("DEL chatroom property error:", err.Error())
	}
	// 删除Room gateway hash key
	key = getChatRoomGateWayHashKey(room, appid)
	if _, err := SessionPool.Call(getSessionAddr(key)).DEL(key); err != nil {
		Logger.Error("DEL chatroom gateway error:", err.Error())
	}
	// ChatRoom集合中删除RoomID
	key = getChatRoomSetKey(room, appid)
	if _, err := SessionPool.Call(getSessionAddr(key)).SREM(key, []string{room}); err != nil {
		Logger.Error("SREM chatroom id error:", err.Error())
	}
	return nil
}

func addRobotIntoChatRoom(user string, room string, appid uint16) error {
	key := getChatRoomMembersSetKey(room, appid)
	if v, err := SessionPool.Call(getSessionAddr(key)).SADD(key, []string{user}); err != nil {
		return err
	} else if v == 0 {
		// 已经有实例在该聊天室中
	} else {
		key = getChatRoomPropertyHashKey(room, appid)
		if _, err = SessionPool.Call(getSessionAddr(room)).HINCRBY(key, session.RobotChatRoomUser, 1); err != nil {
			return err
		}
	}
	return nil
}

func removeRobotFromChatRoom(user string, room string, appid uint16) error {
	key := getChatRoomMembersSetKey(room, appid)
	if v, err := SessionPool.Call(getSessionAddr(key)).SREM(key, []string{user}); err != nil {
		return err
	} else if v == 0 {
		// 不在聊天室中
	} else {
		key = getChatRoomPropertyHashKey(room, appid)
		if _, err = SessionPool.Call(getSessionAddr(room)).HINCRBY(key, session.RobotChatRoomUser, -1); err != nil {
			return err
		}
	}
	return nil
}

func checkUserSessionInChatRoom(user string, room string, appid uint16, gateway string, connectionId logic.ConnectionId) (bool, error) {
	key := getChatRoomUserSessionHashKey(user, appid, gateway, connectionId)
	return checkUserSessionInChatRoomBySessionKey(user, room, key)
}

func checkUserSessionInChatRoomBySessionKey(user, room, session string) (bool, error) {
	if v, err := SessionPool.Call(getSessionAddr(user)).HGET(session, room); err != nil {
		Logger.Debug(user, "", room, "saver.checkUserSessionInChatRoomBySessionKey", v, err)
		if err == msgRedis.ErrKeyNotExist {
			return false, nil
		}
		return false, err
	} else {
		return true, nil
	}
}

func queryUserSessionChatRooms(user string, appid uint16, gateway string, connectionId logic.ConnectionId) ([]string, error) {
	key := getChatRoomUserSessionHashKey(user, appid, gateway, connectionId)
	return queryUserSessionChatRoomsBySessionKey(user, key)
}

func queryUserSessionChatRoomsBySessionKey(user, key string) ([]string, error) {
	m, err := SessionPool.Call(getSessionAddr(user)).HGETALLMAP(key)
	if err == msgRedis.ErrKeyNotExist {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	resp := make([]string, 0, len(m))
	for room, _ := range m {
		resp = append(resp, room)
	}
	return resp, nil
}

func queryUserInRoom(user string, room string, appid uint16) (bool, error) {
	if netConf().ChatroomMembersDegrade {
		return true, nil
	}
	key := getChatRoomMembersSetKey(room, appid)
	if v, err := SessionPool.Call(getSessionAddr(key)).SISMEMBER(key, user); err != nil {
		return false, err
	} else if v == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func addUserIntoChatRoom(user string, room string, appid uint16, gateway string, connectionId logic.ConnectionId, userType string) (int, error) {
	key := getChatRoomUserSessionHashKey(user, appid, gateway, connectionId)
	if v, err := SessionPool.Call(getSessionAddr(user)).HSET(key, room, time.Now().Unix()); err != nil {
		return 0, err
	} else if v == 0 {
		return session.SessionAlreadyInChatRoom, nil
	}
	key = getChatRoomGateWayHashKey(room, appid)
	if _, err := SessionPool.Call(getSessionAddr(key)).HINCRBY(key, gateway, 1); err != nil {
		return 0, err
	}
	key = getChatRoomMembersSetKey(room, appid)
	if v, err := SessionPool.Call(getSessionAddr(key)).SADD(key, []string{user}); err != nil {
		return 0, err
	} else if v == 0 {
		return session.UserAlreadyInChatRoom, nil
	} else {
		key = getChatRoomPropertyHashKey(room, appid)
		if _, err := SessionPool.Call(getSessionAddr(room)).HINCRBY(key, userType, 1); err != nil {
			return 0, err
		}
	}
	return session.Success, nil
}

func removeUserFromChatRoom(user string, room string, appid uint16, gateway string, connectionId logic.ConnectionId, userType string) (int, error) {
	key := getChatRoomUserSessionHashKey(user, appid, gateway, connectionId)
	if v, err := SessionPool.Call(getSessionAddr(user)).HDEL(key, []string{room}); err != nil {
		return 0, err
	} else if v == 0 {
		// 不在聊天室
		return session.SessionNotInChatRoom, nil
	}
	key = getChatRoomGateWayHashKey(room, appid)
	if v, err := SessionPool.Call(getSessionAddr(key)).HINCRBY(key, gateway, -1); err != nil {
		return 0, err
	} else if v <= 0 {
		// todo: gateway上没有用户了
	}
	sessions, err := getUserAllSessionsKeyForChatRoom(user, appid)
	if err != nil {
		return 0, err
	}
	for _, key := range sessions {
		if b, err := checkUserSessionInChatRoomBySessionKey(user, room, key); err != nil {
			return 0, err
		} else if b {
			// 还有其他连接在聊天室
			return session.Success, nil
		}
	}
	// 用户所有连接都已经退出
	key = getChatRoomMembersSetKey(room, appid)
	if v, err := SessionPool.Call(getSessionAddr(key)).SREM(key, []string{user}); err != nil {
		return 0, err
	} else if v != 0 {
		key = getChatRoomPropertyHashKey(room, appid)
		if _, err := SessionPool.Call(getSessionAddr(room)).HINCRBY(key, userType, -1); err != nil {
			return 0, err
		}
	}
	return session.AllSessionQuitedChatRoom, nil
}

func setChatRoomMaxMessageID(room string, appid uint16, value uint) error {
	key := getChatRoomPropertyHashKey(room, appid)
	if _, err := SessionPool.Call(getSessionAddr(room)).HSET(key, cacheMessageField, value); err != nil {
		return err
	}
	return nil
}

func generateChatRoomMessageID(room string, appid uint16) (uint, error) {
	key := getChatRoomPropertyHashKey(room, appid)
	if id, err := SessionPool.Call(getSessionAddr(room)).HINCRBY(key, cacheMessageField, 1); err != nil {
		return 0, err
	} else {
		return uint(id), nil
	}
}

func cacheChatRoomMessage(room string, appid uint16, msgid uint, message []byte) error {
	key := getChatRoomMessageKey(room, appid, msgid)
	if _, err := SessionPool.Call(getChatRoomMessageAddr(key)).SET(key, string(message)); err != nil {
		return err
	}
	if _, err := SessionPool.Call(getChatRoomMessageAddr(key)).EXPIRE(key, netConf().ChatroomMessageTimeout); err != nil {
		return err
	}
	return nil
}

func getCachedRoomMessage(room string, appid uint16, msgid uint) ([]byte, error) {
	key := getChatRoomMessageKey(room, appid, msgid)
	if data, err := SessionPool.CallWithTimeout(getChatRoomMessageAddr(key), 3e9, 60e9).GET(key); err != nil {
		if err == msgRedis.ErrKeyNotExist {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return data, nil
	}
}

const (
	chatRoomCountKey = "chatroom:count:%d:roomid:set"
)

// 以roomid作hash取redis地址
func AddChatroomCountKeys(req *saver.ChatroomCountKeysRequest) error {
	setPool := map[string][]string{}
	key := fmt.Sprintf(chatRoomCountKey, req.Appid)
	for _, roomid := range req.Roomids {
		addr := getChatRoomMessageAddr(roomid)
		setPool[addr] = append(setPool[addr], roomid)
	}
	var err error
	for addr, roomids := range setPool {
		if _, e := SessionPool.Call(addr).SADD(key, roomids); e != nil {
			Logger.Error("", req.Appid, "", "AddChatroomCountKeys", "sadd count roomid error", addr, e.Error())
			err = e
		}
	}
	return err
}

func DelChatroomCountKeys(req *saver.ChatroomCountKeysRequest) error {
	delPool := map[string][]string{}
	key := fmt.Sprintf(chatRoomCountKey, req.Appid)
	for _, roomid := range req.Roomids {
		addr := getChatRoomMessageAddr(roomid)
		delPool[addr] = append(delPool[addr], roomid)
	}
	var err error
	for addr, roomids := range delPool {
		if _, e := SessionPool.Call(addr).SREM(key, roomids); e != nil {
			Logger.Error("", req.Appid, "", "DelChatroomCountKeys", "srem count roomid error", addr, e.Error())
			err = e
		}
	}
	return err
}

func GetChatroomCountKeys(req *saver.GetChatroomCountKeysRequest,
	resp *saver.ChatroomCountKeysResponse) error {
	key := fmt.Sprintf(chatRoomCountKey, req.Appid)
	resp.Roomids = []string{}
	for _, addr := range netConf().ChatroomMessageAddrs {
		data, err := SessionPool.Call(addr).SMEMBERS(key)
		if err != nil {
			Logger.Error("", req.Appid, "", "GetChatroomCountKeys", "smembers count roomid error", addr, err.Error())
			continue
		}

	innerLoop:
		for _, tmp := range data {
			roomid := string(tmp)
			sum := logic.Sum(roomid)
			if sum%req.RpcLength != req.RpcIndex {
				continue innerLoop
			}
			resp.Roomids = append(resp.Roomids, roomid)
		}
	}
	return nil
}

const qpsCountKey = "qps:count:%s:%s:%d"

func AddQpsCount(req *saver.QpsCount, resp *bool) error {
	*resp = false
	t := time.Now().Unix()
	key := fmt.Sprintf(qpsCountKey, req.Module, req.API, t)
	addr := getChatRoomMessageAddr(key)
	count, err := SessionPool.Call(addr).INCR(key)
	if err != nil {
		Logger.Error("", "", "", "AddQpsCount", "GET qps count error", key, err.Error())
		return err
	}
	if count > int64(req.Limit) {
		return nil
	}
	*resp = true
	if count == 1 {
		if _, err := SessionPool.Call(addr).EXPIRE(key, 5); err != nil {
			Logger.Error("", "", "", "AddQpsCount", "SET qps count expire error", key, err.Error())
			return err
		}
	}
	return nil
}
