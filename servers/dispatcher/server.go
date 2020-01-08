package main

import "github.com/johntech-o/gorpc"

type GorpcService struct{}

var rpcServer *gorpc.Server

func GorpcServer() {
	if netConf().GorpcListen == "" {
		panic("empty gorpc_listen")
	}
	Logger.Trace("gorpc listen", netConf().GorpcListen)
	rpcServer = gorpc.NewServer(netConf().GorpcListen)
	rpcServer.Register(new(GorpcService))
	rpcServer.Serve()
	panic("invalid gorpc listen" + netConf().GorpcListen)
}

// 获取qps
func (this *GorpcService) GetQps(req int, resp *map[string]float64) error {
	*resp = Stats.GetQps()
	return nil
}
