package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/utility/lru"
)

var serverRoomFlowService *ServerRoomFlowService

// 机房流量服务
type ServerRoomFlowService struct {
	mutex     *sync.RWMutex
	FlowCache *lru.Cache
}

func NewServerRoomFlowService() *ServerRoomFlowService {
	return &ServerRoomFlowService{
		mutex:     &sync.RWMutex{},
		FlowCache: lru.NewCache(1, 100, 3*time.Second),
	}
}

// 这个机房流量超出预期流量的比例,如果是负数表示还有剩下
func (srf *ServerRoomFlowService) OverFlow(sr string) float32 {
	flow := srf.GetFlow(sr)
	if flow == 0 {
		return 0
	}

	config, ok := DynamicConf().ServerRoomFlowConfig[sr]
	if !ok {
		return 0
	}

	bwLimit := float64(config.Bandwidth) * (float64(config.Limit) / 100)

	over := bwLimit - float64(flow)

	// 计算是否趋近于 limit
	floating := bwLimit * (float64(netConf().SrFlowFloatingRange) / 100)
	if float64(flow) <= (bwLimit+floating) && float64(flow) >= (bwLimit-floating) {
		return 0
	}

	if over > 0 {
		// 有剩余，提升
		return -1
	}

	// 没有剩余，降低
	return 1
}

func (srf *ServerRoomFlowService) getFlowCron() {
	for {
		var wg sync.WaitGroup
		for sr, managers := range logic.NetGlobalConf().GatewayRpcsSr {
			wg.Add(1)
			go func(sr string, managers []string) {
				var wgInside sync.WaitGroup
				var srFlow uint64
				var failed uint32
				for _, manager := range managers {
					wgInside.Add(1)
					go func(manager string) {
						flow, err := gateway.GetLastSecondFlow(manager)
						if err != nil {
							Logger.Error(manager, sr, "", "srf.GetFlow", "gateway.GetLastSecondFlow error", err)

							atomic.StoreUint32(&failed, 1)
						}
						atomic.AddUint64(&srFlow, flow)
						wgInside.Done()
					}(manager)
				}
				wgInside.Wait()
				var srFlowTrue uint64
				if failed == 0 {
					srFlowTrue = uint64(float64(srFlow) * (float64(netConf().SrFlowBase) / 1000))
				}
				srf.FlowCache.Add(sr, srFlowTrue)
				wg.Done()
			}(sr, managers)
		}
		wg.Wait()

		time.Sleep(1 * time.Second)
	}
}

// 获取指定机房的流量
func (srf *ServerRoomFlowService) GetFlow(sr string) uint64 {
	if value, ok := srf.FlowCache.Get(sr); ok {
		return value.(uint64)
	}
	return 0
}
