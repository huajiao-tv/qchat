package logic

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/process"
)

// 从keeper取回来的配置
func NetGlobalConf() *data.Global {
	return data.CurrentGlobal()
}

func LogicSubscribe(nodeId string) []string {
	srn := GetServerRoomNameSection()
	if srn != "" {
		return []string{srn + "/global.conf/" + nodeId}
	} else {
		return []string{"global.conf/" + nodeId}
	}
}

// 初始化项目根目录
var RootDir = func() string {
	binDir, err := process.GetProcessBinaryDir()
	if err != nil {
		panic(err.Error())
	}
	return binDir + "/.."
}()

// 动态配置结构更新，由keeper/client触发调用
func UpdateDynamicConfType() {
	newDynamic := &DynamicConfType{
		GatewaySrMap: make(map[string]string, len(NetGlobalConf().GatewayRpcs)),
	}
	newDynamic.setAppids(NetGlobalConf().Appids)
	innerIp := GetInnerIp()
	if len(innerIp) != 0 {
		for _, addr := range NetGlobalConf().SessionRpcs {
			if strings.Split(addr, ":")[0] == innerIp {
				newDynamic.LocalSessionRpc = addr
			}
		}
		for _, addr := range NetGlobalConf().SaverRpcs {
			if strings.Split(addr, ":")[0] == innerIp {
				newDynamic.LocalSaverRpc = addr
			}
		}
		for _, addr := range NetGlobalConf().RouterRpcs {
			if strings.Split(addr, ":")[0] == innerIp {
				newDynamic.LocalRouterRpc = addr
			}
		}
		for _, addr := range NetGlobalConf().CoordinatorRpcs {
			if strings.Split(addr, ":")[0] == innerIp {
				newDynamic.LocalCoordinatorRpc = addr
			}
		}
	}
	for sr, rpcs := range NetGlobalConf().GatewayRpcsSr {
		for _, rpc := range rpcs {
			newDynamic.GatewaySrMap[rpc] = sr
		}
	}
	atomic.StorePointer(&DynamicConfPtr, unsafe.Pointer(newDynamic))
}

func (this *DynamicConfType) setAppids(appids []string) {
	this.Appids = make(map[string]bool)
	for _, v := range appids {
		this.Appids[v] = true
	}
}

// 需要再处理的动态配置，每次从keeper取回来的配置有更新的话，会同步更新该结构
type DynamicConfType struct {
	Appids              map[string]bool
	LocalSaverRpc       string
	LocalSessionRpc     string
	LocalRouterRpc      string
	LocalCoordinatorRpc string
	GatewaySrMap        map[string]string
}

var DynamicConfPtr unsafe.Pointer = unsafe.Pointer(&DynamicConfType{})

// 获取动态配置
func DynamicConf() *DynamicConfType {
	return (*DynamicConfType)(atomic.LoadPointer(&DynamicConfPtr))
}

var StaticConf = newStaticConfType()

// 静态配置，只在程序启动时初始化一次
type staticConfType struct {
	InternalConnectTimeout time.Duration
	InternalReadTimeout    time.Duration
	InternalWriteTimeout   time.Duration

	ExternalConnectTimeout time.Duration
	ExternalReadTimeout    time.Duration
	ExternalWriteTimeout   time.Duration

	IpStoreFile  string
	CertFile     string
	KeyFile      string
	LogDir       string
	TmpDir       string
	BackupLogDir string
}

func newStaticConfType() *staticConfType {
	return &staticConfType{
		LogDir:       filepath.Join(RootDir, "log"),
		BackupLogDir: filepath.Join(RootDir, "log/backup"),
		TmpDir:       filepath.Join(RootDir, "tmp"),
	}
}

// 静态配置初始化
func (this *staticConfType) Init() {
	this.InternalConnectTimeout = time.Duration(64) * time.Second
	this.InternalReadTimeout = time.Duration(64) * time.Second
	this.InternalWriteTimeout = time.Duration(64) * time.Second

	this.ExternalConnectTimeout = time.Duration(64) * time.Second
	this.ExternalReadTimeout = time.Duration(64) * time.Second
	this.ExternalWriteTimeout = time.Duration(64) * time.Second

	this.IpStoreFile = filepath.Join(RootDir, "data/ip.dat")
	this.CertFile = filepath.Join(RootDir, "ssl/certs/*.com.pm")
	this.KeyFile = filepath.Join(RootDir, "ssl/private/*.com.key")

	fmt.Printf("\nlogic.StaticConf %s\n %#v \n", time.Now().String(), StaticConf)
}

func GetRouterGorpc() string {
	s := DynamicConf().LocalRouterRpc
	if s != "" {
		return s
	}
	routerRpcs := NetGlobalConf().RouterRpcs
	if len(routerRpcs) > 0 {
		return routerRpcs[rand.Intn(len(routerRpcs))]
	} else {
		return ""
	}
}

func GetStatedCenterGorpc(key string) string {
	centerRpcs := NetGlobalConf().CenterRpcs
	if len(centerRpcs) > 0 {
		return centerRpcs[Sum(key)%len(centerRpcs)]
	} else {
		return ""
	}
}

func GetSessionGorpc() string {
	s := DynamicConf().LocalSessionRpc
	if s != "" {
		return s
	}
	sessionRpcs := NetGlobalConf().SessionRpcs
	if len(sessionRpcs) > 0 {
		return sessionRpcs[rand.Intn(len(sessionRpcs))]
	} else {
		return ""
	}
}

// coordinator可能是多发，所以获取到的是一个[]string
func GetStatedCoordinatorGorpcs(key string) []string {
	if c, ok := NetGlobalConf().StaticRoomCoordinator[key]; ok && len(c) > 0 {
		return c
	}
	result := []string{}
	if len(NetGlobalConf().CoordinatorArea) > 0 {
		for _, coors := range NetGlobalConf().CoordinatorArea {
			if len(coors) == 0 {
				continue
			}
			result = append(result, coors[Sum(key)%len(coors)])
		}
	} else {
		result = append(result, GetStatedCoordinatorGorpc(key))
	}
	return result
}

func GetStatedCoordinatorGorpc(key string) string {
	coordinatorRpcs := NetGlobalConf().CoordinatorRpcs
	if len(coordinatorRpcs) > 0 {
		sum := Sum(key)
		return coordinatorRpcs[sum%len(coordinatorRpcs)]
	} else {
		return ""
	}
}

func GetStatedSessionGorpc(key string) string {
	sessionRpcs := NetGlobalConf().SessionRpcs
	if len(sessionRpcs) > 0 {
		sum := Sum(key)
		return sessionRpcs[sum%len(sessionRpcs)]
	} else {
		return ""
	}
}

func GetSaverGorpc() string {
	s := DynamicConf().LocalSaverRpc
	if s != "" {
		return s
	}
	saverRpcs := NetGlobalConf().SaverRpcs
	if len(saverRpcs) > 0 {
		return saverRpcs[rand.Intn(len(saverRpcs))]
	} else {
		return ""
	}
}

func GetAppids() map[string]bool {
	return DynamicConf().Appids
}
func GetDefaultKey(appid, cversion uint16) []byte {
	k := strconv.Itoa(int(appid)) + "-" + strconv.Itoa(int(cversion))
	if rkey, ok := NetGlobalConf().DefaultKeys[k]; ok {
		return []byte(rkey)
	} else {
		return []byte{}
	}
}
func GetInnerIp() string {
	info, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range info {
		ipMask := strings.Split(addr.String(), "/")
		// 排除ipv6，lo，和公网地址
		if ipMask[0] != "127.0.0.1" && ipMask[1] != "24" && strings.Contains(ipMask[0], ".") {
			return ipMask[0]
		}
	}
	return ""
}

// 如果无外网ip。那么返回内网ip
func GetIp() string {
	info, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range info {
		ipMask := strings.Split(addr.String(), "/")
		// 排除ipv6，lo，和公网地址 @todo 这块的判断有问题
		if ipMask[0] != "127.0.0.1" && ipMask[1] == "24" && strings.Contains(ipMask[0], ".") {
			return ipMask[0]
		}
	}
	return GetInnerIp()
}

// 判断这个ip是否在我们的白名单中
func CheckIp(ip string) bool {
	_, ok := NetGlobalConf().WhiteList[ip]
	return ok
}

func CheckRequestIp(r *http.Request) bool {
	return CheckIp(ClientIp(r))
}

func ClientIp(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		s := strings.Split(r.RemoteAddr, ":")
		ip = s[0]
	}
	return ip
}

func GetRandApnsGorpc() string {
	apnsRpcs := NetGlobalConf().ApnsRpcs
	if len(apnsRpcs) > 0 {
		return apnsRpcs[rand.Intn(len(apnsRpcs))]
	} else {
		return ""
	}
}

func GetGatewayGorpcMap() map[string]int {
	ret := make(map[string]int, len(NetGlobalConf().GatewayRpcs))
	for _, gateway := range NetGlobalConf().GatewayRpcs {
		ret[gateway] = 1
	}
	return ret
}

// 获取当前组件所在的机房
func GetServerRoomName() string {
	srn := ""
	host, e := os.Hostname()
	if e == nil {
		hostArr := strings.Split(host, ".")
		if len(hostArr) >= 3 {
			srn = hostArr[len(hostArr)-3]
		}
	}
	return srn
}

func GetServerRoomNameSection() string {
	srn := ""
	host, e := os.Hostname()
	if e == nil {
		hostArr := strings.Split(host, ".")
		if len(hostArr) >= 3 {
			srn = "sr-" + hostArr[len(hostArr)-3]
		}
	}
	return srn
}

func GetAllSubscribeSection(component, nodeId string) []string {
	srn := GetServerRoomNameSection()
	if srn != "" {
		return []string{
			srn + "/global.conf/" + nodeId,
			srn + "/" + component + ".conf/" + nodeId,
		}
	} else {
		return []string{
			"global.conf/" + nodeId,
			"/" + component + ".conf/" + nodeId,
		}
	}
}
