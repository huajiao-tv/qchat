package main

import (
	"bytes"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

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
	http.HandleFunc("/adapter/stat", AdapterStatHandler)
	http.HandleFunc("/cpu/info", CpuInfo)
	http.HandleFunc("/config", ConfigHandler)

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, fmt.Sprintf("%#v", netConf()))
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

func AdapterStatHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, "parseform error", err))
		return
	}
	roomid := r.FormValue("roomid")
	if roomid != "" {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", adapterStats.GetByRoomid(roomid)))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", adapterStats.GetAll()))
	}
}

// 查看cpu占用信息 curl -G http://127.0.0.1:19201/cpu/info -d seconds=30&file=ppp
func CpuInfo(w http.ResponseWriter, r *http.Request) {
	var buff bytes.Buffer
	seconds, _ := strconv.Atoi(r.FormValue("seconds"))
	fileName := r.FormValue("file")
	if seconds == 0 {
		fmt.Fprintf(w, adminJson.FmtJson(400, "bad args!", ""))
		return
	}
	if fileName == "" {
		pprof.StartCPUProfile(&buff)
		time.Sleep(time.Duration(seconds) * 1e9)
		pprof.StopCPUProfile()
		w.Write(buff.Bytes())
		return
	}

	profileFile, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer profileFile.Close()
	pprof.StartCPUProfile(profileFile)
	time.Sleep(time.Duration(seconds) * 1e9)
	pprof.StopCPUProfile()

	output := "Please get " + fileName + " pprof file from " + netConf().AdminListen +
		" after " + strconv.Itoa(seconds) + " seconds!\n"
	fmt.Fprintf(w, adminJson.FmtJson(0, "", output))
}
