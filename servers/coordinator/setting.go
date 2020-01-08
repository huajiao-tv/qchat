package main

import (
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/logger"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
)

// 组件名称
const Component = "coordinator"

// 包全局变量
var (
	KeeperAddr string
	Domain     string
	NodeID     string

	Logger *logger.Logger

	staticConf = newStaticConfType()
)

func initSetting() error {
	flag.StringVar(&KeeperAddr, "k", "", "keeper address ip:port")
	flag.StringVar(&Domain, "d", "", "domain name")
	flag.StringVar(&NodeID, "n", "", "current node id")
	flag.Parse()

	sections := logic.GetAllSubscribeSection(Component, NodeID)
	keeperCli := gokeeper.New(KeeperAddr, Domain, NodeID, Component, sections, logic.ComponentTags)
	keeperCli.LoadData(data.ObjectsContainer).RegisterCallback(logic.UpdateDynamicConfType, UpdateDynamicConfType)
	if err := keeperCli.Work(); err != nil {
		return err
	}

	logic.StaticConf.Init()
	staticConf.init()
	if err := initGlobal(); err != nil {
		return err
	}
	return nil
}

type DynamicConfType struct {
	// 属于我负责的机房
	MyServerNames map[string]bool

	// 基础过滤规则
	CommonFilter map[int][]CommonFilterTypePriority

	// 普通聊天室过滤规则
	NormalChatRoomDiscardMessagePolicy map[string]map[int]int

	// 流量机房配置
	ServerRoomFlowConfig map[string]*ServerRoomFlowConfigItem
	StaticMsgSend        map[string]map[string]int
	CoordinatorRpcs      []string
}

var DynamicConfPtr unsafe.Pointer = unsafe.Pointer(&DynamicConfType{})
var MsgPoliceConfPtr unsafe.Pointer = unsafe.Pointer(&ServerRoomConf{})

// 获取动态配置
func DynamicConf() *DynamicConfType {
	return (*DynamicConfType)(atomic.LoadPointer(&DynamicConfPtr))
}

func MsgPoliceConf() *ServerRoomConf {
	return (*ServerRoomConf)(atomic.LoadPointer(&MsgPoliceConfPtr))
}

func UpdateDynamicConfType() {
	newDynamic := &DynamicConfType{
		MyServerNames:   make(map[string]bool),
		CoordinatorRpcs: logic.NetGlobalConf().CoordinatorRpcs,
	}
	// 如果不分地区，那么配置的server_rooms不再起作用

	if len(logic.NetGlobalConf().CoordinatorArea) != 0 {
		for _, sr := range netConf().ServerRooms {
			newDynamic.MyServerNames[sr] = true
		}
	}

	newDynamic.StaticMsgSend = updateStaticMsgSend()
	newDynamic.CommonFilter = parseCommonFilter()
	newDynamic.NormalChatRoomDiscardMessagePolicy = parseNormalChatRoomDiscardMessagePolicy()
	newDynamic.ServerRoomFlowConfig = parseServerRoomFlowConfig()

	atomic.StorePointer(&DynamicConfPtr, unsafe.Pointer(newDynamic))
	if Logger != nil {
		Logger.SetLevel(netConf().Loglevel)
	}

	// 更新房间配比配置
	msgPoliceConf := initMsgPolicyConf()
	atomic.StorePointer(&MsgPoliceConfPtr, unsafe.Pointer(msgPoliceConf))
}

func compareGorpc(new, old []string) bool {
	if len(new) != len(old) {
		return false
	}
	for k, s := range old {
		if new[k] != s {
			return false
		}
	}
	return true
}

func updateStaticMsgSend() map[string]map[string]int {
	result := make(map[string]map[string]int, len(netConf().StaticMsgSend))
	for roomidSr, count := range netConf().StaticMsgSend {
		arr := strings.Split(roomidSr, "-")
		if len(arr) != 2 {
			continue
		}
		if _, ok := result[arr[0]]; !ok {
			result[arr[0]] = make(map[string]int)
		}
		result[arr[0]][arr[1]] = count
	}
	return result
}

func initGlobal() error {
	var err error
	logTag := ""
	if netConf().AdminListen != "" && len(netConf().AdminListen) > 1 {
		logTag = netConf().AdminListen[1:]
	} else {
		logTag = NodeID
	}
	filename := filepath.Join(logic.StaticConf.LogDir, fmt.Sprintf("%s-%s", Component, logTag))
	Logger, err = logger.NewLogger(filename, Component+"|"+NodeID, logic.StaticConf.BackupLogDir)
	if err != nil {
		return err
	}
	Logger.SetLevel(netConf().Loglevel)
	return nil
}

//根据配置中的消息配比获得最简不可再约形式,并生成消息序列,返回配置
func initMsgPolicyConf() *ServerRoomConf {
	//取消息类型中的最小比例
	minRatio := 100
	for _, ratio := range netConf().MsgTypeRatio {
		if ratio < minRatio {
			minRatio = ratio
		}
	}
	//计算最大公约数
	greatestDivisor := 1
	for i := 2; i <= minRatio; i += 1 {
		for _, ratio := range netConf().MsgTypeRatio {
			if ratio%i == 0 {
				greatestDivisor = i
			}
		}
	}
	//根据最大公约数得到取消息序列
	simpleRatio := map[string]int{}
	for msgCluster, ratio := range netConf().MsgTypeRatio {
		simpleRatio[msgCluster] = ratio / greatestDivisor
	}
	//初始化消息类型序列
	order := []string{}
	for msgCluster, ratio := range simpleRatio {
		for i := 0; i < ratio; i += 1 {
			order = append(order, msgCluster)
		}
	}
	//打乱顺序，防止同消息类型聚集在一起(此部分只能在填补策略为指定分组的时候使用)
	disOrder := []string{}
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, v := range rr.Perm(len(order)) {
		disOrder = append(disOrder, order[v])
	}

	msgTypeMap := map[string]string{}
	for cluster, typeList := range netConf().MsgTypeMap {
		for _, msgType := range typeList {
			msgTypeMap[msgType] = cluster
		}
	}
	fmt.Println(netConf().MsgTypeMap)
	if _, ok := msgTypeMap["--"]; !ok {
		panic("msg complex filter conf has no msg type \"--\"")
	}

	msgPoliceConf := &ServerRoomConf{
		MaxCacheWhiteListSec: netConf().MaxCacheWhiteListSec,
		MaxCacheSec:          netConf().MaxCacheSec,
		MaxCacheNumPerSec:    netConf().MaxCacheNumPerSec,
		MsgClusterRatio:      simpleRatio,
		MsgTypeMap:           msgTypeMap,
		MsgTypeOrder:         disOrder,
		OrderLen:             len(disOrder),
		WhiteListCluster:     netConf().WhiteListCluster,
	}
	return msgPoliceConf
}

type staticConfType struct {
	PidFile string
}

func newStaticConfType() *staticConfType {
	return &staticConfType{}
}

func netConf() *data.Coordinator {
	return data.CurrentCoordinator()
}

func (this *staticConfType) init() {
	this.PidFile = filepath.Join(logic.StaticConf.TmpDir, fmt.Sprintf("%s-%s.pid", Component, NodeID))

	fmt.Printf("\nstaticConf %s\n %#v \n", time.Now().String(), staticConf)
}

type ServerRoomFlowConfigItem struct {
	Bandwidth uint64
	Limit     int
}

func parseServerRoomFlowConfig() map[string]*ServerRoomFlowConfigItem {
	result := make(map[string]*ServerRoomFlowConfigItem)

	for sr, cis := range netConf().SrFlow {
		tmp := strings.Split(cis, "|")
		if len(tmp) != 2 {
			continue
		}

		bws, _ := strconv.Atoi(tmp[0])
		limit, _ := strconv.Atoi(tmp[1])

		bw := uint64(bws) * uint64(1000000000)

		configItem := &ServerRoomFlowConfigItem{
			Bandwidth: bw,
			Limit:     limit,
		}

		result[sr] = configItem
	}

	return result
}

type CommonFilterTypePriority struct {
	Type     string
	Priority string
}

func parseCommonFilter() map[int][]CommonFilterTypePriority {
	data := make(map[int][]CommonFilterTypePriority)
	for ppl, configstr := range netConf().CommonFilter {
		configItems := strings.Split(configstr, "|")
		cm := make([]CommonFilterTypePriority, 0, len(configItems))
		for _, configItem := range configItems {
			tmp := strings.Split(configItem, "-")
			if len(tmp) != 2 {
				continue
			}

			filter := CommonFilterTypePriority{
				Type:     tmp[0],
				Priority: tmp[1],
			}

			cm = append(cm, filter)
		}

		data[ppl] = cm
	}

	return data
}

func parseNormalChatRoomDiscardMessagePolicy() map[string]map[int]int {
	result := make(map[string]map[int]int)

	// example: 30-102:1000|5,5000|35;30-103:1000|10,5000|25

	configItemStrArray := strings.Split(netConf().CrDropMsgsDtPolicy, ";")

	// []string{"30-102:1000|5,5000|35", "30-103:1000|10,5000|25"}
	for _, configItemStr := range configItemStrArray {
		configKvArray := strings.Split(configItemStr, ":")
		if len(configKvArray) != 2 {
			continue
		}

		// tp: 30-102; tpcvs: 1000|5,5000|35
		tp, tpcs := configKvArray[0], configKvArray[1]

		// []string{"1000|5", "5000|35"}
		tpca := strings.Split(tpcs, ",")
		if len(tpca) == 0 {
			continue
		}

		for _, tpc := range tpca {
			tmp := strings.Split(tpc, "|")
			if len(tmp) != 2 {
				continue
			}

			numofppl, _ := strconv.Atoi(tmp[0])
			percentage, _ := strconv.Atoi(tmp[1])

			if _, ok := result[tp]; !ok {
				result[tp] = make(map[int]int)
			}

			result[tp][numofppl] = percentage
		}
	}

	return result
}
