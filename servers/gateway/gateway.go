package main

import (
	"runtime"

	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/process"
)

var (
	tagPools              *TagPools
	connectionIdGenerator *logic.ConnectionIdGenerator
	connectionPool        *ConnectionPool
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := initSetting(); err != nil {
		panic(err)
	}

	tagPools = NewTagPools()
	connectionIdGenerator = logic.NewConnectionIdGenerator()
	connectionPool = newConnectionPool()
}

func main() {
	process.SavePid(staticConf.PidFile)
	go GorpcServer()
	go FrontServer()
	go AdminServer()

	go StartCountFlow()

	select {}
}
