package main

import (
	"runtime"

	"github.com/huajiao-tv/qchat/utility/process"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := initSetting(); err != nil {
		panic(err)
	}

	chatRoomAdapterPool = NewChatRoomAdapterPool()
	adapterStats = NewAdapterStats()
	serverRoomFlowService = NewServerRoomFlowService()
	go serverRoomFlowService.getFlowCron()
}

func main() {
	process.SavePid(staticConf.PidFile)
	go GorpcServer()
	go AdminServer()

	select {}
}
