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
}

func main() {
	process.SavePid(staticConf.PidFile)
	go GorpcServer()
	go AdminServer()
	go FrontServer()

	select {}
}
