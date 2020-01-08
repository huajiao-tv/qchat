package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/logger"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/qchat/utility/stat"
)

// 组件名称
const Component = "session"

// 包全局变量
var (
	KeeperAddr string
	Domain     string
	NodeID     string

	Logger            *logger.Logger
	CallbackThreshold *stat.Threshold

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
func UpdateDynamicConfType() {
	if Logger != nil {
		Logger.SetLevel(netConf().Loglevel)
	}
	if CallbackThreshold != nil {
		CallbackThreshold.SetLimit(netConf().CallbackDegradeQps / len(logic.NetGlobalConf().SessionRpcs))
	}

	// Update Http Client
	UpdateHttpClientConfig()
	newDynamic := &DynamicConfType{}

	old := (*DynamicConfType)(atomic.LoadPointer(&DynamicConfPtr))
	if !compareGorpc(logic.NetGlobalConf().SessionRpcs, old.SessionRpcs) {
		// 如果session的rpc发生变化
		newDynamic.SessionRpcs = logic.NetGlobalConf().SessionRpcs
		if onlineCache != nil {
			onlineCache.Clean()
		}
		atomic.StorePointer(&DynamicConfPtr, unsafe.Pointer(newDynamic))
	}

	initOnlineCache()
	onlineCache.UpdateConfig(netConf().OnlineCacheSlot, netConf().OnlineCacheCap, time.Duration(netConf().OnlineCacheExpire)*time.Second)
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

type DynamicConfType struct {
	SessionRpcs []string
}

var DynamicConfPtr unsafe.Pointer = unsafe.Pointer(&DynamicConfType{})

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
	CallbackThreshold = stat.NewThreshold(netConf().CallbackDegradeQps / len(logic.NetGlobalConf().SessionRpcs))
	return nil
}

type staticConfType struct {
	PidFile string
}

func newStaticConfType() *staticConfType {
	return &staticConfType{}
}

func netConf() *data.Session {
	return data.CurrentSession()
}

func (this *staticConfType) init() {
	this.PidFile = filepath.Join(logic.StaticConf.TmpDir, fmt.Sprintf("%s-%s.pid", Component, NodeID))

	fmt.Printf("\nstaticConf %s\n %#v \n", time.Now().String(), staticConf)
}
