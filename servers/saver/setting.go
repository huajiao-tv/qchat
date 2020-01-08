package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/logger"
	"github.com/huajiao-tv/qchat/utility/stat"
)

// 组件名称
const Component = "saver"

// 包全局变量
var (
	KeeperAddr string
	Domain     string
	NodeID     string

	Logger *logger.Logger
	Stats  *stat.Stat

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
	Stats = stat.NewStat(10)
	return nil
}

type staticConfType struct {
	PidFile string
}

func newStaticConfType() *staticConfType {
	return &staticConfType{}
}

func netConf() *data.Saver {
	return data.CurrentSaver()
}

func (this *staticConfType) init() {
	this.PidFile = filepath.Join(logic.StaticConf.TmpDir, fmt.Sprintf("%s-%s.pid", Component, NodeID))

	fmt.Printf("\nstaticConf %s\n %#v \n", time.Now().String(), staticConf)
}

func getSessionAddr(key string) string {
	sum := logic.Sum(key)
	return netConf().SessionAddrs[sum%len(netConf().SessionAddrs)]
}

func getChatRoomMessageAddr(key string) string {
	sum := logic.Sum(key)
	return netConf().ChatroomMessageAddrs[sum%len(netConf().ChatroomMessageAddrs)]
}
