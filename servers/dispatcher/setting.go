package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"
	"unsafe"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/utility/logger"
	"github.com/huajiao-tv/qchat/utility/stat"
)

// 组件名称
const Component = "dispatcher"

// 包全局变量
var (
	KeeperAddr string
	Domain     string
	NodeID     string

	Logger          *logger.Logger
	Stats           *stat.Stat
	staticConf      = newStaticConfType()
	DefaultInterval = 10

	SignKey           = unsafe.Pointer(&rsa.PrivateKey{})
	Gateways          = unsafe.Pointer(&GatewayConf{})
	WhiteList         = unsafe.Pointer(&GatewayConf{})
	ZoneGatewayWeight = unsafe.Pointer(&GatewayZoneConf{})
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
	if Stats != nil {
		Stats.SetInterval(netConf().StatInterval)
	}
	updatePrivateKey()
	updateGatewayConf()
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
	Stats = stat.NewStat(netConf().StatInterval)
	Logger.SetLevel(netConf().Loglevel)
	return nil
}

type staticConfType struct {
	PidFile string
}

func newStaticConfType() *staticConfType {
	return &staticConfType{}
}

func netConf() *data.Dispatcher {
	return data.CurrentDispatcher()
}

func (this *staticConfType) init() {
	this.PidFile = filepath.Join(logic.StaticConf.TmpDir, fmt.Sprintf("%s-%s.pid", Component, NodeID))

	fmt.Printf("\nstaticConf %s\n %#v \n", time.Now().String(), staticConf)
}

func updatePrivateKey() {
	buf := make([]byte, base64.StdEncoding.DecodedLen(len(netConf().PrivateKey)))
	n, err := base64.StdEncoding.Decode(buf, []byte(netConf().PrivateKey))
	if err != nil {
		Logger.Error("", "", "", "updatePrivateKey", "error", err.Error())
		return
	}
	pk, err := x509.ParsePKCS1PrivateKey(buf[:n])
	if err != nil {
		fmt.Println(netConf().PrivateKey, buf, n, buf[:n], err)
		Logger.Error("", "", "", "updatePrivateKey", "error", err.Error())
		return
	}
	atomic.StorePointer(&SignKey, unsafe.Pointer(pk))
}

func updateGatewayConf() {
	conf := NewGatewayConf(netConf().GatewayWeight)
	atomic.StorePointer(&Gateways, unsafe.Pointer(conf))
	wl := NewGatewayConf(netConf().WhiteListAddr)
	atomic.StorePointer(&WhiteList, unsafe.Pointer(wl))
	zgw := NewGatewayZoneConf(netConf().ZoneGatewayweight)
	atomic.StorePointer(&ZoneGatewayWeight, unsafe.Pointer(zgw))
}
