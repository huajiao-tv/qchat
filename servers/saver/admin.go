package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync/atomic"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/adminJson"
)

var MongoClean = int32(0)

func AdminServer() {
	s := &http.Server{
		Addr:         netConf().AdminListen,
		ReadTimeout:  logic.StaticConf.ExternalReadTimeout,
		WriteTimeout: logic.StaticConf.ExternalWriteTimeout,
	}

	// start stat thread
	go StatQps()

	http.HandleFunc("/hello/world", HelloWorld)
	http.HandleFunc("/rpcserver/status", RpcServerStatus)
	http.HandleFunc("/stat", SaverStat)
	http.HandleFunc("/stats", SaverStats)
	http.HandleFunc("/stat/redis", RedisStat)

	http.HandleFunc("/user/property", UserProperty)
	http.HandleFunc("/room/query/members", QueryMembers)
	http.HandleFunc("/room/query/detail", QueryDetail)
	http.HandleFunc("/room/message/get", GetCachedRoomMessages)
	http.HandleFunc("/user/room/list", GetUserChatRoomList)
	http.HandleFunc("/mongo_clean", MongoCleanHandler)

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
func RpcServerStatus(w http.ResponseWriter, r *http.Request) {
	if rpcServer == nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, "rpc server is nil", ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", rpcServer.Status()))
	}
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "hello world"))
}

func MongoCleanHandler(w http.ResponseWriter, r *http.Request) {
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "MongoCleanHandler", "ip forbidden", logic.ClientIp(r))
		return
	}

	atomic.StoreInt32(&MongoClean, 1)
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "mongo cleaned"))
}

func UserProperty(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	appid := r.FormValue("appid")
	if appidUint := logic.StringToUint16(appid); appidUint == 0 {
		fmt.Fprintf(w, adminJson.FmtJson(500, "appid invalid", ""))
		return
	} else {
		query := map[string]string{"UserId": user, "AppId": appid}
		if props, err := getUserSession(query); err != nil {
			fmt.Fprintf(w, adminJson.FmtJson(500, err.Error(), ""))
		} else {
			fmt.Fprintf(w, adminJson.FmtJson(0, "", props))
		}
	}
}

func QueryMembers(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("rid")
	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if resp, err := queryChatRoomMemberCount(room, appid); err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, err.Error(), ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "success", resp))
	}
}

func QueryDetail(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("rid")
	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if resp, err := queryChatRoomDetail(room, appid); err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, err.Error(), ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "success", resp))
	}
}

func GetCachedRoomMessages(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("rid")
	msgid, _ := strconv.Atoi(r.FormValue("msgid"))
	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if data, err := getCachedRoomMessage(room, appid, uint(msgid)); err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, err.Error(), ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "success", string(data)))
	}
}

func GetUserChatRoomList(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("uid")
	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	sessions, err := getUserAllSessionsKeyForChatRoom(user, appid)
	if err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, err.Error(), ""))
		return
	}
	resp := make(map[string][]string)
	for _, key := range sessions {
		rooms, err := queryUserSessionChatRoomsBySessionKey(user, key)
		if err != nil {
			fmt.Fprintf(w, adminJson.FmtJson(500, err.Error(), ""))
			return
		}
		resp[key] = rooms
	}
	fmt.Fprintf(w, adminJson.FmtJson(0, "success", resp))
}

/*
 * Saver QPS request handler
 * this will write two json string to client
 */
func SaverStat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, fmt.Sprintf("{ \"Saver\" : [ %s, %s ] }", qpsData.QpsString(),
		fmt.Sprintf("{ \"total requests\" : %s }", requestStat.String())))
}

func SaverStats(w http.ResponseWriter, r *http.Request) {
	ret := make(map[string]interface{}, 4)
	ret["component"] = Component
	ret["node"] = NodeID
	ret["qps"] = Stats.GetQps()
	fmt.Fprintf(w, adminJson.FmtJson(0, "success", ret))
}

func RedisStat(w http.ResponseWriter, r *http.Request) {
	switch r.FormValue("inst") {
	case "session":
		fmt.Fprintf(w, SessionPool.Info())
	default:
		fmt.Fprintf(w, `{"ErrNo":500,"Err":"invalid args","Data":""}`)
	}
}
