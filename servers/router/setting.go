package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/logger"
	"github.com/johntech-o/gorpc"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
)

// 组件名称
const Component = "router"

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

func ProcessGorpcSaverClientTimeout() map[string]*gorpc.NetOptions {
	rs := make(map[string]*gorpc.NetOptions)
	for method, rawTimeout := range netConf().GorpcSaverClientTimeout {
		timeout := strings.Split(rawTimeout, "|")
		if len(timeout) < 2 {
			continue
		}
		connTimeout, _ := strconv.Atoi(timeout[0])
		socketTimeout, _ := strconv.Atoi(timeout[1])
		readTimeout := socketTimeout / 2
		writeTimeout := socketTimeout / 2
		netOptions := gorpc.NewNetOptions(time.Duration(connTimeout)*time.Second, time.Duration(readTimeout)*time.Second, time.Duration(writeTimeout)*time.Second)
		rs[method] = netOptions
		if Logger != nil {
			Logger.Debug("ProcessGorpcSaverClientTimeout", "method", method, "connTimeout", connTimeout, "socketTimeout", socketTimeout)
		}
	}
	return rs

}

func UpdateDynamicConfType() {
	if Logger != nil {
		Logger.SetLevel(netConf().Loglevel)
	}

	// GorpcSaverClient
	netOptionsList := ProcessGorpcSaverClientTimeout()
	if saver.GorpcClient != nil {
		saver.SetMethodNetOptinons("GorpcService", netOptionsList)
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
}

func newStaticConfType() *staticConfType {
	return &staticConfType{}
}

func netConf() *data.Router {
	return data.CurrentRouter()
}

func (this *staticConfType) init() {
	this.PidFile = filepath.Join(logic.StaticConf.TmpDir, fmt.Sprintf("%s-%s.pid", Component, NodeID))

	fmt.Printf("\nstaticConf %s\n %#v \n", time.Now().String(), staticConf)
}
