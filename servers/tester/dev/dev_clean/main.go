package main

import (
	"errors"
	"fmt"
	"strconv"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/msgRedis"
)

var KeeperAddr string = "127.0.0.1:7000"
var NodeID string = "dev_clean"
var Sections []string = []string{"global.conf", "saver.conf"}
var Domain string = "qchat_online"
var Component string = "dev_clean"
var Appid int = 2080
var SessionPool *msgRedis.MultiPool

func init() {
	keeperCli := gokeeper.New(KeeperAddr, Domain, NodeID, Component, Sections, nil)
	keeperCli.LoadData(data.ObjectsContainer).RegisterCallback(logic.UpdateDynamicConfType)
	if err := keeperCli.Work(); err != nil {
		panic(err)
	}
	SessionPool = msgRedis.NewMultiPool(
		data.CurrentSaver().SessionAddrs,
		msgRedis.DefaultMaxConnNumber+20,
		msgRedis.DefaultMaxIdleNumber+95,
		msgRedis.DefaultMaxIdleSeconds)
}

// 获取某个key hash到的redis
func getHashRedisAddr(key string) string {
	return data.CurrentSaver().SessionAddrs[logic.Sum(key)%len(data.CurrentSaver().SessionAddrs)]
}

// 获取指定块的在线用户列表
func getActiveUsers(seq int) ([]string, error) {
	key := fmt.Sprintf("userstat:%d:%d:set", Appid, seq)
	users, err := SessionPool.Call(getHashRedisAddr(key)).SMEMBERS(key)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(users))
	for _, u := range users {
		result = append(result, string(u))
	}
	return result, nil
}

// 获取用户session的主key名称
func getMasterKey(user string) string {
	return fmt.Sprintf("usersession:%s:%d:sets", user, Appid)
}

// 从用户的主key里获取所有用户的二级key
func getSecondsKeys(user string) ([]string, error) {
	masterKey := getMasterKey(user)
	skeys, err := SessionPool.Call(getHashRedisAddr(masterKey)).SMEMBERS(masterKey)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(skeys))
	for _, sk := range skeys {
		result = append(result, string(sk))
	}
	return result, nil
}

// 获取用户某一个二级key 的session信息
func getUserProperty(masterKey, secondKey string) (map[string]string, error) {
	r := getHashRedisAddr(masterKey)
	prop, err := SessionPool.Call(r).HGETALLMAP(secondKey)
	if err != nil {
		return nil, err
	}
	return prop, nil
}

func dealUser(user string) {
	skeys, err := getSecondsKeys(user)
	if err != nil {
		fmt.Println("ERROR, getSecondsKeys ", err)
	}
	mkey := getMasterKey(user)
	flag := false
	for _, sk := range skeys {
		prop, err := getUserProperty(mkey, sk)
		if err != nil {
			fmt.Println("ERROR, getUserProperty ", err)
			return
		}
		isDirty, err := checkDirty(prop)
		if err != nil {
			fmt.Println("ERROR, checkDirty ", err)
			return
		}
		if isDirty {
			flag = true
			if err := removeUser(mkey, sk, prop); err != nil {
				fmt.Println("ERROR, removeUser ", err)
				return
			}
			if prop["Platform"] == "" {
				prop["Platform"] = ""
			}
			fmt.Printf("remove user,%s,%s,%s\n", prop["UserId"], prop["Deviceid"], prop["Platform"])
		}
	}
	// 如果没有删除任务脏数据，不做清理
	if !flag {
		return
	}
	if err := removeActive(user); err != nil {
		fmt.Println("removeActive user error", user)

	}

}

// 无条件删除用户在在线用户列表里的记录
func removeActive(user string) error {
	masterkey := getMasterKey(user)
	num, err := SessionPool.Call(getHashRedisAddr(masterkey)).SCARD(masterkey)
	if err != nil {
		return errors.New("removeActiove scard error:" + err.Error())
	}

	if num == 0 {
		key := fmt.Sprintf("userstat:%d:%d:set", Appid, logic.Sum(user)%100)
		fmt.Println("remove active", key, user, getHashRedisAddr(masterkey), masterkey, num)
		if _, err := SessionPool.Call(getHashRedisAddr(key)).SREM(key, []string{user}); err != nil {

			return errors.New("SRem active error:" + err.Error())
		}
	}
	return nil
}

// 删除这个用户相关的信息
func removeUser(mkey, sk string, prop map[string]string) error {
	r := getHashRedisAddr(mkey)
	if _, err := SessionPool.Call(r).DEL(sk); err != nil {
		return err
	}

	if _, err := SessionPool.Call(r).SREM(mkey, []string{sk}); err != nil {
		return err
	}

	if err := removeChatroom(prop); err != nil {
		return errors.New("removeChatroom error:" + err.Error())
	}
	return nil
}

// 删除对应连接的聊天室相关信息
func removeChatroom(prop map[string]string) error {
	crSessionKey := fmt.Sprintf("chatroom:user:session:%s:%s:%s:%s:hset", prop["UserId"], prop["AppId"], prop["Gateway"], prop["ConnId"])
	m, err := SessionPool.Call(getHashRedisAddr(prop["UserId"])).HGETALLMAP(crSessionKey)
	if err == msgRedis.ErrKeyNotExist {
		return nil
	} else if err != nil {
		return err
	}
	fmt.Println("remove chatroom:", m)
	if _, err := SessionPool.Call(getHashRedisAddr(prop["UserId"])).DEL(crSessionKey); err != nil {
		return errors.New("del chatroom session error:" + err.Error())
	}
	for r, _ := range m {
		fmt.Println("del chatroom ", r, prop["UserId"])
		gwKey := fmt.Sprintf("chatroom:gateway:%s:%s:hset", r, prop["AppId"])
		if v, err := SessionPool.Call(getHashRedisAddr(r)).HINCRBY(gwKey, prop["Gateway"], -1); err != nil {
			return err
		} else if v <= 0 {
			return errors.New("gateway上没有用户了")
		}
		// 这块就直接把这个用户从这房间给清理掉了，但是其实这里面不家可能有其它session在这个房间，此时不应该这样处理，此处只是清理数据时方便而已
		// 用户所有连接都已经退出
		crMemberKey := fmt.Sprintf("chatroom:members:%s:%s:set", r, prop["AppId"])
		if v, err := SessionPool.Call(getHashRedisAddr(r)).SREM(crMemberKey, []string{prop["UserId"]}); err != nil {
			return err
		} else if v != 0 {
			crPropKey := fmt.Sprintf("chatroom:property:%s:%s:hset", r, prop["AppId"])
			if _, err := SessionPool.Call(getHashRedisAddr(r)).HINCRBY(crPropKey, session.GetUserType(prop["UserId"], uint16(Appid), prop["ConnectionType"]), -1); err != nil {
				return err
			}
		}

	}
	return nil
}

func checkDirty(prop map[string]string) (bool, error) {
	connId, err := strconv.ParseUint(prop["ConnectionId"], 10, 64)
	if err != nil {
		return false, err
	}
	if t, err := gateway.GetConnectionInfo(prop["GatewayAddr"], logic.ConnectionId(connId)); err == nil && prop["Sender"] == t["Sender"] {
		return false, nil
	} else {
		fmt.Println("check dirty", prop["Sender"], t["Sender"], true, err, connId, prop["GatewayAddr"])
		return true, nil
	}
}

func main() {
	total := 0
	for i := 0; i < 100; i++ {
		users, err := getActiveUsers(i)
		if err != nil {
			fmt.Println("ERROR, getActiveUsers ", err)
			return
		}
		total += len(users)
		for _, u := range users {
			dealUser(u)
		}
		fmt.Println("deal ", i, "over")
	}
	fmt.Println("deal user num:", total)

}
