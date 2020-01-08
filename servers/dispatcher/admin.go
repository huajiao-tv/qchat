package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/adminJson"
)

func AdminServer() {
	s := &http.Server{
		Addr:         netConf().AdminListen,
		ReadTimeout:  logic.StaticConf.ExternalReadTimeout,
		WriteTimeout: logic.StaticConf.ExternalWriteTimeout,
	}

	http.HandleFunc("/hello/world", HelloWorld)
	http.HandleFunc("/rpcserver/status", RpcServerStatus)

	http.HandleFunc("/down/lvs", DownLvs)
	http.HandleFunc("/up/lvs", UpLvs)
	http.HandleFunc("/stat", QpsStat)

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "hello world"))
}

func RpcServerStatus(w http.ResponseWriter, r *http.Request) {
	if rpcServer == nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, "rpc server is nil", ""))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", rpcServer.Status()))
	}
}

func DownLvs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "DownLvs", "ip forbidden", logic.ClientIp(r))
		return
	}
	downLvs()
	Logger.Warn("", "", r.RemoteAddr, "DownLvs", "", "")
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "success"))
}

func UpLvs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "UpLvs", "ip forbidden", logic.ClientIp(r))
		return
	}
	upLvs()
	Logger.Warn("", "", r.RemoteAddr, "UpLvs", "", "")
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "success"))
}

func QpsStat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "", Stats.GetQps()))
}
