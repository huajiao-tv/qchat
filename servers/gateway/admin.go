package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

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
	http.HandleFunc("/stat", GatewayQps)
	http.HandleFunc("/flow", GatewayFlow)

	http.HandleFunc("/conn/info", ConnInfo)
	http.HandleFunc("/conn/len", ConnLen)
	http.HandleFunc("/service/switch", ServiceSwitch)
	http.HandleFunc("/service/status", ServiceStatus)
	http.HandleFunc("/monitor/data", MonitorData)
	http.HandleFunc("/tag/stat", TagStat)
	http.HandleFunc("/del/tag", DelTag)
	http.HandleFunc("/stop/gateway", StopGateway)
	http.HandleFunc("/tag/closeconn", TagCloseConn)
	http.HandleFunc("/tag/delcloseconn", TagDelCloseConn)
	http.HandleFunc("/tag/allconn", TagAllConn)

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func TagAllConn(w http.ResponseWriter, r *http.Request) {
	tag := r.FormValue("tag")
	if len(tag) == 0 {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "need tag"))
		return
	}
	p := tagPools.GetPool(tag)
	if p == nil {
		fmt.Fprintf(w, adminJson.FmtJson(1, "tag not found", ""))
	} else {
		result := p.GetAllConn()
		fmt.Fprintf(w, adminJson.FmtJson(0, "", result))
	}
}

func TagCloseConn(w http.ResponseWriter, r *http.Request) {
	tag := r.FormValue("tag")
	if len(tag) == 0 {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "need tag"))
		return
	}
	p := tagPools.GetPool(tag)
	if p == nil {
		fmt.Fprintf(w, adminJson.FmtJson(0, "tag not found", ""))
		return
	}
	result := p.GetCloseConn()
	fmt.Fprintf(w, adminJson.FmtJson(0, "", result))
}

func TagDelCloseConn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "TagDelCloseConn", "ip forbidden", logic.ClientIp(r))
		return
	}
	tag := r.FormValue("tag")
	if len(tag) == 0 {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "need tag"))
		return
	}
	Logger.Trace("", "", "", "TagDelCloseConn", tag, "")
	p := tagPools.GetPool(tag)
	if p == nil {
		fmt.Fprintf(w, adminJson.FmtJson(0, "tag not found", ""))
		return
	}
	result := p.DelCloseConn()
	fmt.Fprintf(w, adminJson.FmtJson(0, "", result))
}

// 停掉此gateway，一旦停掉，只能重启才能接受新服务
func StopGateway(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "StopGateway", "ip forbidden", logic.ClientIp(r))
		return
	}
	i, _ := strconv.Atoi(r.FormValue("interval"))
	interval := time.Duration(i) * time.Millisecond
	StopAccept()
	cs := connectionPool.Connections()
	Logger.Warn("", "", "", "StopGateway", interval, cs)
	for _, c := range cs {
		if !c.IsClose() {
			c.Close()
		}
		if interval != 0 {
			time.Sleep(interval)
		}
	}
	StopListen()
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "success"))
}

func TagStat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "", tagPools.Stat(false)))
}

func DelTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "DelTag", "ip forbidden", logic.ClientIp(r))
		return
	}
	tag := r.FormValue("tag")
	Logger.Trace("", "", "", "DelTag", tag, "")
	tagPools.DelTag(tag)
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "success"))
}

func MonitorData(w http.ResponseWriter, r *http.Request) {
	var write, read []uint64
	count := 0
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			readCount, writeCount := monitor.MonitorData()
			read = append(read, readCount)
			write = append(write, writeCount)
		}
		count++
		if count == 4 {
			break
		}
	}
	ticker.Stop()
	result := make(map[string]uint64)
	result["reading"] = (read[len(read)-1] - read[0]) / uint64(len(read)-1)
	result["writing"] = (write[len(write)-1] - write[0]) / uint64(len(read)-1)
	fmt.Fprintf(w, adminJson.FmtJson(0, "", result))
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

func ConnInfo(w http.ResponseWriter, r *http.Request) {
	i, _ := strconv.Atoi(r.FormValue("id"))
	connId := logic.ConnectionId(i)
	connection := connectionPool.Connection(connId)
	if connection == nil {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", "connection not found"))
		return
	}
	conn, ok := connection.(*XimpConnection)
	if !ok {
		fmt.Fprintf(w, adminJson.FmtJson(0, "", "not a XimpConnection"))
		return
	}
	prop := conn.GetPropCopy()

	fmt.Fprintf(w, adminJson.FmtJson(0, "", prop))
}

func ConnLen(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "", connectionPool.GetLen()))
}

func ServiceSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "ServiceSwitch", "ip forbidden", logic.ClientIp(r))
		return
	}
	s := r.FormValue("switch")
	if s == "1" {
		StopAccept()
	} else if s == "0" {
		StartAccept()
	}
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "success"))
}

func ServiceStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "accept stopped?", IsStopAccept()))
}

/*
 * gateway QPS request handler
 * this will write two json string to client
 */
func GatewayQps(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, fmt.Sprintf("{ \"Gateway\" : [ %s, %s ] }", qpsData.QpsString(),
		fmt.Sprintf("{\"total requests\" : %s}", requestStat.String())))
}

func GatewayFlow(w http.ResponseWriter, r *http.Request) {
	resultFmt := "{\"LastSecondFlow\": %v, \"ThisSecondFlow\": %v}"
	fmt.Fprintf(w, resultFmt, requestStat.AtomicGetLastSecondFlow(), requestStat.AtomicGetThisSecondFlow())
}
