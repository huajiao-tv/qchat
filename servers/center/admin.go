package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/huajiao-tv/qchat/client/coordinator"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
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

	http.HandleFunc("/down/lvs", DownLvs)
	http.HandleFunc("/up/lvs", UpLvs)
	http.HandleFunc("/chatroom/userlist", ChatRoomUserList)
	http.HandleFunc("/chatroom/member_count", UpdateChatRoomMemberCount)
	http.HandleFunc("/gateway/connlen", GatewayConnlenHandler)
	http.HandleFunc("/gateway/tagstat", GatewayTagstatHandler)

	http.HandleFunc("/stat", GetAllStatHandler)
	http.HandleFunc("/user/online_cache", UserOnlineCache)

	http.HandleFunc("/reconnect/notify", ReConnectNotify)

	http.HandleFunc("/onlinecache/stat", OnlinecacheStat)
	http.HandleFunc("/coordinator/adapter/stat", CoordinatorAdapterStat)
	http.HandleFunc("/coordinator/adapter/stat/sr", CoordinatorAdapterStatSr)

	http.HandleFunc("/flow", FlowStat)

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
func CoordinatorAdapterStatSr(w http.ResponseWriter, r *http.Request) {
	resp := errorResponse{errorSuccess, "", nil}
	var allLock sync.Mutex
	all := map[string]map[string]*coordinator.AdapterStat{}
	unionNode(func(g string) interface{} {
		if stat, err := coordinator.GetAdapterStat(g); err != nil {
			Logger.Error("", "", "", "CoordinatorAdapterStat", "coordinator.GetAdapterStat", err)
			return interface{}(nil)
		} else {
			allLock.Lock()
			for rid, v := range stat {
				for sr, as := range v {
					if _, ok := all[sr]; !ok {
						all[sr] = map[string]*coordinator.AdapterStat{}
					}
					as.Host = g
					as.FlowHum = logic.HumSize(as.FlowCount)
					all[sr][rid] = as
				}
			}
			allLock.Unlock()

			return interface{}(stat)
		}
	}, logic.NetGlobalConf().CoordinatorRpcs)
	resp.Data = all
	w.Write([]byte(resp.Error()))
}

func CoordinatorAdapterStat(w http.ResponseWriter, r *http.Request) {
	resp := errorResponse{errorSuccess, "", nil}
	var allLock sync.Mutex
	all := map[string]map[string]*coordinator.AdapterStat{}
	unionNode(func(g string) interface{} {
		if stat, err := coordinator.GetAdapterStat(g); err != nil {
			Logger.Error("", "", "", "CoordinatorAdapterStat", "coordinator.GetAdapterStat", err)
			return interface{}(nil)
		} else {
			allLock.Lock()
			for rid, v := range stat {
				for sr, as := range v {
					if _, ok := all[rid]; !ok {
						all[rid] = map[string]*coordinator.AdapterStat{}
					}
					all[rid][sr] = as
				}
			}
			allLock.Unlock()

			return interface{}(stat)
		}
	}, logic.NetGlobalConf().CoordinatorRpcs)
	resp.Data = all
	w.Write([]byte(resp.Error()))
}

func ReConnectNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		return
	}
	if !logic.CheckRequestIp(r) {
		fmt.Fprintf(w, adminJson.FmtJson(500, "", "fail"))
		Logger.Warn("", "", "", "ReConnectNotify", "ip forbidden", logic.ClientIp(r))
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, "parse form error", err))
		return
	}
	appid, _ := strconv.Atoi(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.APPID_HUAJIAO
	}
	var userGateways []*logic.UserGateway
	var moreIps, gateways, tags []string
	var ip string
	if r.FormValue("ips") == "" {
		fmt.Fprintf(w, adminJson.FmtJson(500, "ips is empty", ""))
		return
	}

	ips := strings.Split(r.FormValue("ips"), ",")
	for _, i := range ips {
		if i == "" {
			continue
		}
		if ip == "" {
			ip = i
		} else {
			moreIps = append(moreIps, i)
		}
	}
	if r.FormValue("gateways") != "" {
		gateways = strings.Split(r.FormValue("gateways"), ",")
	}
	if r.FormValue("tags") != "" {
		tags = strings.Split(r.FormValue("tags"), ",")
	}

	port, err := strconv.Atoi(r.FormValue("port"))
	if (port != 443 && port != 80) || err != nil {
		fmt.Fprintf(w, adminJson.FmtJson(500, "port is invalid", err))
		return
	}

	if r.FormValue("uid") != "" {
		querySession := []*session.UserSession{
			&session.UserSession{
				UserId: r.FormValue("uid"),
				AppId:  uint16(appid),
			},
		}
		resp, err := session.Query(querySession)
		if err != nil {
			Logger.Error(r.FormValue("uid"), appid, "", "ReConnectNotify", "session.Query error", err)
		} else {
			for _, session := range resp {
				userGateways = append(userGateways, &logic.UserGateway{
					session.GatewayAddr,
					session.ConnectionId,
				})
			}
		}
	}

	if err := router.SendReConnectNotify(ip, uint32(port), moreIps, gateways, tags, userGateways); err != nil {
		Logger.Error(ip, port, moreIps, "ReConnectNotify", "SendReConnectNotify error", err)
		fmt.Fprintf(w, adminJson.FmtJson(500, "SendReConnectNotify error", err))
	} else {
		fmt.Fprintf(w, adminJson.FmtJson(0, "success", userGateways))
	}
}

func OnlinecacheStat(w http.ResponseWriter, r *http.Request) {
	resp := errorResponse{errorSuccess, "", nil}
	keys := []string{"hit", "miss", "len"}
	var allLock sync.Mutex
	all := map[string]uint64{}
	result := unionNode(func(g string) interface{} {
		if stat, err := session.OnlineCacheStat(g); err != nil {
			Logger.Error("", "", "", "OnlinecacheStat", "session.OnlinecacheStat", err)
			return interface{}(nil)
		} else {
			fstat := map[string]uint64{}
			for _, v := range stat {
				for _, vv := range v {
					for _, k := range keys {
						fstat[k] += vv[k]
					}
				}
			}

			allLock.Lock()
			for _, v := range keys {
				all[v] += fstat[v]
			}
			allLock.Unlock()
			return interface{}(fstat)
		}
	}, logic.NetGlobalConf().SessionRpcs)
	result["all"] = interface{}(all)
	resp.Data = result
	w.Write([]byte(resp.Error()))

}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, adminJson.FmtJson(0, "", "hello world"))
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

func ChatRoomUserList(w http.ResponseWriter, r *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		w.Write([]byte(res.Error()))
	}()
	api := r.URL.Path
	limit := netConf().QpsLimits[api]
	if limit == 0 {
		limit = 1
	}
	ok, err := saver.AddQpsCount("center", api, limit)
	if err != nil {
		res.Code, res.Reason = errorInternalError, err.Error()
		return
	}
	if !ok {
		res.Code, res.Reason = errorAccessTooFrequent, errorReasonAccessTooFrequent
		return
	}
	switch r.Method {
	case MethodGet, MethodPost:
		if err := r.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		appid := logic.StringToUint16(r.FormValue("appid"))
		if appid == 0 {
			appid = logic.APPID_HUAJIAO
		}
		room := r.FormValue("rid")
		if room == "" {
			res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
			break
		}
		users, err := saver.QueryChatRoomUsers(room, appid)
		if err != nil {
			res.Code, res.Reason = errorInternalError, err.Error()
			break
		}
		res.Data = users
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
	}
}

func UpdateChatRoomMemberCount(w http.ResponseWriter, r *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch r.Method {
	case MethodPost:
		if err := r.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, errorReasonInvalidRequest
			break
		}
		req := &saver.ChatRoomMemberUpdate{
			Appid: logic.DEFAULT_APPID,
		}
		if req.RoomID = r.FormValue("rid"); req.RoomID == "" {
			res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
			break
		}
		if appid := r.FormValue("appid"); appid != "" {
			req.Appid = logic.StringToUint16(appid)
		}
		if uid := r.FormValue("uid"); uid != "" {
			if msg := r.FormValue("msg"); msg == "" {
				res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
				break
			} else {
				// 转发旧系统加入通知
				forwardChatRoomJoinNotify(uid, req.RoomID, req.Appid, msg)
			}
		} else if count := r.FormValue("count"); count != "" {
			req.Count, _ = strconv.Atoi(count)
			if !session.CheckUserType(r.FormValue("type")) {
				res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
				break
			} else {
				req.Type = r.FormValue("type")
				if err := saver.UpdateChatRoomMember(req); err != nil {
					res.Code, res.Reason = errorInternalError, err.Error()
				}
			}
		} else {
			res.Code, res.Reason = errorBadArguments, errorReasonBadArguments
			break
		}
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
	}
	w.Write([]byte(res.Error()))
}

func GetAllStatHandler(w http.ResponseWriter, r *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch r.Method {
	case MethodGet, MethodPost:
		if err := r.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, err.Error()
			break
		}

		w.Write([]byte(GetStat(r)))
		res = nil
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
	}

	if res != nil {
		w.Write([]byte(res.Error()))
	}
}

func UserOnlineCache(w http.ResponseWriter, r *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	switch r.Method {
	case MethodGet, MethodPost:
		if err := r.ParseForm(); err != nil {
			res.Code, res.Reason = errorInvalidRequest, err.Error()
			break
		}

		appid := logic.StringToUint16(r.FormValue("appid"))
		if appid == 0 {
			appid = logic.APPID_HUAJIAO
		}
		users := strings.Split(r.FormValue("userids"), ",")
		m := make(map[string][]string)
		for _, uid := range users {
			key := logic.GetStatedSessionGorpc(uid)
			s, ok := m[key]
			if !ok {
				s = make([]string, 0, len(users))
			}
			m[key] = append(s, uid)
		}
		ret := make(map[string]map[string][]*logic.UserGateway)
		for addr, uList := range m {
			resp, err := session.QueryUserOnlineCache(addr, appid, uList)
			if err != nil {
				res.Code, res.Reason = errorInternalError, err.Error()
				goto Exit
			}
			ret[addr] = resp
		}
		res.Data = ret
	default:
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
	}

Exit:
	if res != nil {
		w.Write([]byte(res.Error()))
	}
}

func getFlow(flowBase float64, managers []string) (uint64, map[string]uint64) {
	var srFlowAll uint64
	result := make(map[string]uint64)
	locker := &sync.Mutex{}

	var wg sync.WaitGroup
	for _, manager := range managers {
		wg.Add(1)
		go func(m string) {
			var flowTrue uint64
			flow, err := gateway.GetLastSecondFlow(m)
			if err != nil {
				Logger.Error("", "", "", "srf.GetFlow", "gateway.GetLastSecondFlow", manager, "rpc call failed.")
			} else {
				flowTrue = flow
				if flowBase != 0 {
					flowTrue = uint64(float64(flow) * (flowBase / 1000))
				}
			}
			locker.Lock()
			result[m] = flowTrue
			locker.Unlock()

			atomic.AddUint64(&srFlowAll, flowTrue)
			wg.Done()
		}(manager)
	}
	wg.Wait()

	return srFlowAll, result
}

func FlowStat(w http.ResponseWriter, r *http.Request) {
	res := &errorResponse{Code: errorSuccess, Reason: errorReasonSuccess}
	defer func() {
		w.Write([]byte(res.Error()))
	}()
	if r.Method != MethodGet {
		res.Code, res.Reason = errorUnsupportedHttpMethod, errorReasonUnsupportedHttpMethod
		return
	}

	if err := r.ParseForm(); err != nil {
		res.Code, res.Reason = errorInvalidRequest, err.Error()
		return
	}

	appid := logic.StringToUint16(r.FormValue("appid"))
	if appid == 0 {
		appid = logic.APPID_HUAJIAO
	}

	flowBaseTmp, _ := strconv.Atoi(r.FormValue("flowbase"))
	flowBase := float64(flowBaseTmp)

	idc := r.FormValue("idc")
	switch idc {
	case "":
		// all idc
		ret := make(map[string]map[string]interface{})

		for idc, managers := range logic.NetGlobalConf().GatewayRpcsSr {
			srFlowAll, data := getFlow(flowBase, managers)
			if _, ok := ret[idc]["all"]; !ok {
				ret[idc] = make(map[string]interface{})
			}
			ret[idc]["all"] = srFlowAll
			ret[idc]["detail"] = data
		}

		res.Data = ret
	default:
		ret := make(map[string]map[string]interface{})

		managers, ok := logic.NetGlobalConf().GatewayRpcsSr[idc]
		if !ok {
			res.Code, res.Reason = errorInvalidRequest, "idc not found"
			return
		}

		srFlowAll, data := getFlow(flowBase, managers)
		if _, ok := ret[idc]["all"]; !ok {
			ret[idc] = make(map[string]interface{})
		}
		ret[idc]["all"] = srFlowAll
		ret[idc]["detail"] = data

		res.Data = ret
	}
}
