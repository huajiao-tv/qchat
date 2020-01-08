package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/adminJson"
)

func AdminServer() {
	// start stat thread
	go StatQps()

	s := &http.Server{
		Addr:         netConf().AdminListen,
		ReadTimeout:  logic.StaticConf.ExternalReadTimeout,
		WriteTimeout: logic.StaticConf.ExternalWriteTimeout,
	}

	http.HandleFunc("/hello/world", HelloWorld)
	http.HandleFunc("/rpcserver/status", RpcServerStatus)
	http.HandleFunc("/stat", SessionStatHandler)

	http.HandleFunc("/clean/session", CleanUserSession)
	http.HandleFunc("/clean/chatroom", CleanChatRoom)

	http.HandleFunc("/callback/room/join", JoinChatRoomCallback)
	http.HandleFunc("/callback/room/quit", QuitChatRoomCallback)

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

func JoinChatRoomCallback(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("uid")
	room := r.FormValue("rid")
	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if user == "" || room == "" {
		fmt.Fprintf(w, adminJson.FmtJson(1, "invalid request", ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "success", struct {
			Response string `json:"singlecast"`
			Notify   string `json:"multicast"`
		}{
			"", "",
		}))
	}
}

func QuitChatRoomCallback(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("uid")
	room := r.FormValue("rid")
	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.DEFAULT_APPID
	}
	if user == "" || room == "" {
		fmt.Fprintf(w, adminJson.FmtJson(1, "invalid request", ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "success", nil))
	}
}

func CleanUserSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "CleanUserSession", "ip forbidden", logic.ClientIp(r))
		return
	}
	iterateUserSession()
	fmt.Fprintf(w, adminJson.FmtJson(0, "success", nil))
}

func CleanChatRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "CleanChatRoom", "ip forbidden", logic.ClientIp(r))
		return
	}
	iterateChatRoom()
	fmt.Fprintf(w, adminJson.FmtJson(0, "success", nil))
}

/*
 * Session stat request handler
 * this will write two json string to client
 */
func SessionStatHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, fmt.Sprintf("{ \"Session\" : [ %s, %s ] }", qpsData.QpsString(),
		fmt.Sprintf("{ \"total requests\" : %s }", requestStat.String())))
}
