package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	Success = iota
	ErrorInvalidArguments
	ErrorBadRequest
	ErrorInternalError
)

type DispatcherResp struct {
	ErrNo int    `json:"errno"`
	Err   string `json:"errmsg"`
	Data  string `json:"data"`
	Sign  string `json:"sign"`
}

func (r *DispatcherResp) String() string {
	if data, err := json.Marshal(*r); err != nil {
		return fmt.Sprintf("{\"data\":\"\", \"reason\":\"%s\"}", err.Error())
	} else {
		return string(data)
	}
}

type GatewayConf struct {
	max      int
	cursor   uint64
	gateways []string
}

func NewGatewayConf(c map[string]int) *GatewayConf {
	gwys := make([]string, 0, len(c))
	for gw, weight := range c {
		for i := 0; i < weight; i++ {
			gwys = append(gwys, gw)
		}
	}
	return &GatewayConf{
		max:      len(c),
		cursor:   0,
		gateways: gwys,
	}
}

type GatewayZoneConf struct {
	zonenum         int
	zonegatewayconf map[string]*GatewayConf
}

func NewGatewayZoneConf(c map[string][]string) *GatewayZoneConf {
	zoneGatewayconfMap := make(map[string]*GatewayConf, len(c))

	for zone, addrWeights := range c {
		addrWeigthMap := make(map[string]int, len(addrWeights))
		for _, addrWeigth := range addrWeights {
			addrWeightPair := strings.Split(addrWeigth, "-")
			if len(addrWeightPair) > 1 {
				weight, err := strconv.ParseInt(addrWeightPair[1], 10, 64)
				if err != nil {
					continue
				}
				addrWeigthMap[addrWeightPair[0]] = int(weight)
			}
		}

		gatewayConf := NewGatewayConf(addrWeigthMap)
		zoneGatewayconfMap[zone] = gatewayConf
	}

	return &GatewayZoneConf{
		zonenum:         len(zoneGatewayconfMap),
		zonegatewayconf: zoneGatewayconfMap,
	}
}

func (gzc *GatewayZoneConf) Get(zone string, num int) []string {
	if conf, ok := gzc.zonegatewayconf[zone]; ok {
		return conf.Get(num)
	}

	return nil
}

func (g *GatewayConf) Get(num int) []string {
	if num > g.max {
		num = g.max
	}
	ret := make([]string, 0, num)
	start := int((atomic.AddUint64(&g.cursor, 1) % uint64(len(g.gateways))))
	last := ""
	for i := 0; len(ret) < num; i++ {
		index := (start + i) % len(g.gateways)
		if index == start && i != 0 {
			break
		}
		if last != g.gateways[index] {
			last = g.gateways[index]
			ret = append(ret, last)
		}
	}
	return ret
}

func GetWhiteListGateways() string {
	conf := (*GatewayConf)(atomic.LoadPointer(&WhiteList))
	gws := conf.Get(netConf().MaxGateways)
	return strings.Join(gws, ";")
}

//zone: 地域名称例如 北京市
func GetNormalGateways(zone string) string {
	zoneconf := (*GatewayZoneConf)(atomic.LoadPointer(&ZoneGatewayWeight))
	gws := zoneconf.Get(zone, netConf().MaxGateways)
	if gws == nil {
		conf := (*GatewayConf)(atomic.LoadPointer(&Gateways))
		gws = conf.Get(netConf().MaxGateways)
	}
	return strings.Join(gws, ";")
}
