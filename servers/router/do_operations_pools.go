package main

import (
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
)

var doOperationsPools *MultiConsumerPools

func initDoOperationsPools() error {
	doOperationsConsumerCount := uint(netConf().DoOperationsConsumerCount)
	doOperationsConsumerChanlen := uint64(netConf().DoOperationsConsumerChanlen)
	if doOperationsConsumerCount <= 0 || doOperationsConsumerChanlen <= 0 {
		doOperationsConsumerCount = 10
		doOperationsConsumerChanlen = 10000
	}
	doOperationsPools = NewMultiConsumerPools(doOperationsConsumerCount, doOperationsConsumerChanlen, doOperationsFn)
	Logger.Debug("", "", "", "initDoOperationsPools", "doOperationsConsumerCount", doOperationsConsumerCount, "doOperationsConsumerChanlen", doOperationsConsumerChanlen)
	return nil
}

type DoOperationsParams struct {
	GatewayAddr string
	Tags        []string
	PrefixTag   string
	ConnIds     []logic.ConnectionId
	Gwr         *router.GwResp
}

func addDoOperations(params *DoOperationsParams) {
	if params.GatewayAddr == "" {
		return
	}
	pool := doOperationsPools.GetPool(params.GatewayAddr)

	if ok := pool.Add(params); !ok {
		Logger.Error("", "", "", "addDoOperations", params.GatewayAddr, params.Tags, params.ConnIds, params.Gwr, "channel full")
	}
}

func doOperationsFn(d interface{}) {
	params, ok := d.(*DoOperationsParams)
	if !ok {
		Logger.Error("", "", "", "doOperationsFn", "Consumer error:type not match", d)
		return
	}

	err := gateway.DoOperations(params.GatewayAddr, params.Tags, params.PrefixTag, params.ConnIds, params.Gwr)
	if err != nil {
		Logger.Error("", "", "", "doOperationsFn", params.GatewayAddr, params.Tags, params.ConnIds, params.Gwr, err)
		return
	}
}
