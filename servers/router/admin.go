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
	http.HandleFunc("/stat", RouterStatHandler)

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

/*
 * Router stat request handler
 * this will write two json string to client
 */
func RouterStatHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, fmt.Sprintf("{ \"Router\" : [ %s, %s ] }", qpsData.QpsString(),
		fmt.Sprintf("{ \"total requests\" : %s }", requestStat.String())))
}
