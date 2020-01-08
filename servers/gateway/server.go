package main

import (
	"errors"
	"strings"

	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

type GorpcService struct{}

var rpcServer *gorpc.Server

func GorpcServer() {
	if netConf().Manager == "" || netConf().GorpcListen == "" {
		panic("empty manager or gorpc_listen")
	}
	Logger.Trace("", "", "", "GorpcServer", "gorpc listen", netConf().GorpcListen)
	rpcServer = gorpc.NewServer(netConf().GorpcListen)
	rpcServer.Register(new(GorpcService))
	rpcServer.Serve()
	panic("invalid gorpc listen" + netConf().GorpcListen)
}
func (this *GorpcService) Helloworld(foo string, response *int) error {
	return nil
}

func (this *GorpcService) GetConnectionInfo(connId logic.ConnectionId, response *map[string]string) error {
	conn := connectionPool.Connection(connId)
	if conn == nil {
		return errors.New("connection not found")
	} else if ximpConn, ok := conn.(*XimpConnection); !ok {
		return errors.New("not a ximpconnection")
	} else {
		*response = ximpConn.GetPropCopy()
		return nil
	}
}

func (this *GorpcService) ConnLen(req int, count *int) error {
	*count = connectionPool.GetLen()
	return nil
}

func (this *GorpcService) TagStat(req int, response *map[string]int) error {
	*response = tagPools.Stat(true)
	return nil
}

func (this *GorpcService) TagStatAll(req int, response *map[string]int) error {
	*response = tagPools.Stat(false)
	return nil
}

func (this *GorpcService) DelTag(req string, response *int) error {
	tagPools.DelTag(req)
	return nil
}

func (this *GorpcService) GetConnectionTags(connId logic.ConnectionId, response *[]string) error {
	*response = tagPools.GetTags(connId)
	return nil
}

func (this *GorpcService) CheckTags(connTags *gateway.ConnTags, resp *map[string]bool) error {
	result := map[string]bool{}
	for _, t := range connTags.Tags {
		result[t] = tagPools.Check(t, connTags.ConnId)
	}
	*resp = result
	return nil
}

func (this *GorpcService) DoOperationsBatch(req []*gateway.Operations, response *int) error {
	for _, o := range req {
		this.DoOperations(o, response)
	}
	return nil
}

func (this *GorpcService) DoOperations(operations *gateway.Operations, response *int) error {
	count := make(map[logic.ConnectionId]bool)
	fake := 0

	var traceId string

	if operations.Gwr != nil && operations.Gwr.XimpBuff != nil {
		// count handling time if need
		if netConf().StatResponseTime {
			countFunc := countGatewayOperation(operations.Gwr.XimpBuff.TimeStamp,
				"tags:"+strings.Join(operations.Tags, ","), "", traceId, "DoOperations")
			defer countFunc()
		}

		traceId = operations.Gwr.XimpBuff.TraceId
	}
	if len(operations.ConnectionIds) != 0 {
		for _, connId := range operations.ConnectionIds {
			conn := connectionPool.Connection(connId)
			if conn == nil || conn.IsClose() {
				fake += 1
				Logger.Warn(connId, "", traceId, "DoOperations", "connection is close", "")
			} else {
				conn.Operate(operations.Gwr, false)
				count[connId] = true
			}
		}
	}
	if len(count) != 0 || fake != 0 {
		Logger.Trace(operations.ConnectionIds, operations.Tags, traceId, "DoOperations", len(count), fake)
	}
	// tag处理放到chan里串行执行
	if len(operations.Tags) != 0 {
		for _, tag := range operations.Tags {
			tp := tagPools.GetPool(tag)
			if tp == nil {
				Logger.Warn(tag, "", traceId, "DoOperations", "GetPool", "tag not found")
			} else if ok := tp.PushOperation(operations.Gwr); !ok {
				Logger.Error(tag, "", traceId, "DoOperations", "PushOperation error", "tag channel full")
			}
		}
	}

	// 发指定前缀的消息
	if operations.PrefixTag != "" {
		tags := tagPools.Stat(true)
		for tag, _ := range tags {
			if strings.HasPrefix(tag, operations.PrefixTag) {
				tp := tagPools.GetPool(tag)
				if tp == nil {
					Logger.Error(tag, "", traceId, "SendPrefixTag", "GetPool", "tag not found")
				} else if ok := tp.PushOperation(operations.Gwr); !ok {
					Logger.Error(tag, "", traceId, "SendPrefixTag", "PushOperation error", "tag channel full")
				}
			}
		}
	}

	return nil
}

func (this *GorpcService) GetLastSecondFlow(req int, flow *uint64) error {
	*flow = requestStat.AtomicGetLastSecondFlow()
	return nil
}
