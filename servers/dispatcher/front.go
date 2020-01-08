package main

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/adminJson"
	"github.com/johntech-o/iphelper"
)

func FrontServer() {
	//	r := mux.NewRouter()
	r := http.NewServeMux()

	r.HandleFunc("/", NotFound)
	r.HandleFunc("/status.php", LvsCheckHandler)

	// APIs
	r.HandleFunc("/get", QueryIps)

	//init ipstore
	_, err := os.Open(logic.StaticConf.IpStoreFile)
	if err != nil {
		fmt.Println("error opening file: %v\n" + err.Error())
	} else {
		ipStore = iphelper.NewIpStore(logic.StaticConf.IpStoreFile)
	}

	if netConf().Listen != "" {
		Logger.Trace("front listen", netConf().Listen)
		go http.ListenAndServe(netConf().Listen, r)
	} else {
		panic("empty listen")
	}
	if netConf().SslListen != "" {
		Logger.Trace("front ssl listen", netConf().SslListen)
		go http.ListenAndServeTLS(netConf().SslListen, logic.StaticConf.CertFile, logic.StaticConf.KeyFile, r)
	}
}

var ipStore *iphelper.IpStore

// 如果把stopLvsLock设置成1，表示从lvs上下线
var stopLvsLock int32 = 0

func getIP(req *http.Request) string {
	if ip := req.Header.Get("HTTP_CLIENT_IP"); ip != "" {
		return ip
	}

	ipport := strings.Split(req.RemoteAddr, ":")
	if len(ipport) > 0 {
		return ipport[0]
	}
	return "nullip"
}

func isStopLvs() bool {
	return atomic.LoadInt32(&stopLvsLock) != 0
}

func NotFound(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, adminJson.FmtJson(404, "page not found", ""))
}

func downLvs() {
	atomic.StoreInt32(&stopLvsLock, 1)
}

func upLvs() {
	atomic.StoreInt32(&stopLvsLock, 0)
}

func LvsCheckHandler(w http.ResponseWriter, req *http.Request) {
	if isStopLvs() {
		io.WriteString(w, "fail\n")
	} else {
		io.WriteString(w, "ok\n")
	}
}

func QueryIps(w http.ResponseWriter, req *http.Request) {
	Stats.Incr(req.URL.Path)
	resp := &DispatcherResp{}

	switch req.Method {
	case http.MethodGet, http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			resp.ErrNo = ErrorInvalidArguments
			resp.Err = "invalid args"
			Logger.Error("", "", "", "dispatcher.QueryIps", "ParseForm err", err.Error())
			break
		}
		mobile := req.FormValue("mobiletype")
		appid := req.FormValue("appid")
		if appid == "" {
			appid = logic.APPID_HUAJIAO_STR
		}
		random := req.FormValue("random")
		if len(random) > 16 {
			random = random[:16]
		}

		zone := ""
		clientip := getIP(req)

		if ipStore != nil {
			if clientip != "" {
				geo, err := ipStore.GetGeoByIp(clientip)
				if err == nil {
					country := geo["country"]
					if country == "中国" {
						zone = geo["city"]
					} else if country == "未知" {
						zone = geo["city"]
					} else {
						zone = "外国"
					}
				}
			}
		} else {
			zone = "null ip store"
		}

		uid := req.FormValue("uid")
		if _, ok := netConf().WhiteListUser[uid]; ok {
			resp.Data = GetWhiteListGateways()
		} else {
			resp.Data = GetNormalGateways(zone)
		}
		Logger.Trace(uid, appid, "", "dispatcher.QueryIps", mobile, random, zone, clientip, resp.Data)
		if netConf().NeedSign {
			buf := make([]byte, 0, 32)
			buf = append(buf, []byte(random)...)
			hashed := md5.Sum([]byte(resp.Data))
			buf = append(buf, hashed[0:]...)
			pk := (*rsa.PrivateKey)(atomic.LoadPointer(&SignKey))
			s, err := rsa.SignPKCS1v15(nil, pk, 0, hashed[:len(hashed)])
			if err != nil {
				resp.ErrNo = ErrorInternalError
				resp.Err = "server error"
				break
			} else {
				ret := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
				base64.StdEncoding.Encode(ret, s)
				resp.Sign = string(ret)
			}
		}
	default:
		resp.ErrNo = ErrorBadRequest
		resp.Err = "bad request"
	}

	fmt.Fprintf(w, resp.String())
}
