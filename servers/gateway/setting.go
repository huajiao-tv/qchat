package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/logger"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
)

// 组件名称
const Component = "gateway"

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
	return nil
}

type staticConfType struct {
	PidFile string

	XimpReadTimeout      time.Duration // 第一个包刚开始读取protobuf的超时
	HeartBeatTimeout     time.Duration // 默认的心跳超时
	HeartBeatBaseTimeout time.Duration // 心跳时间基础上需要增加的时间

	ClientQueueLen int // 队列长度
}

func newStaticConfType() *staticConfType {
	return &staticConfType{}
}

func netConf() *data.Gateway {
	return data.CurrentGateway()
}

func (this *staticConfType) init() {
	this.PidFile = filepath.Join(logic.StaticConf.TmpDir, fmt.Sprintf("%s-%s.pid", Component, NodeID))
	this.XimpReadTimeout = 60 * time.Second
	this.HeartBeatTimeout = 5 * time.Minute
	this.HeartBeatBaseTimeout = 20 * time.Second

	this.ClientQueueLen = 500

	fmt.Printf("\nstaticConf %s\n %#v \n", time.Now().String(), staticConf)
}
