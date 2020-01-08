package gateway

import (
	"time"

	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)

	statNetOption := gorpc.NewNetOptions(1*time.Second, 1*time.Second, 1*time.Second)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetGatewayQps", statNetOption)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetGatewayTotalOps", statNetOption)

	flowNetOption := gorpc.NewNetOptions(3*time.Second, 3*time.Second, 3*time.Second)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetLastSecondFlow", flowNetOption)
}

// 发送一些行为(gwr)到gateway，对应的connIds（如果有）或者tags（如果有）的连接将执行这些操作
func DoOperations(gatewayAddr string, tags []string, prefixTag string, connIds []logic.ConnectionId, gwr *router.GwResp) error {
	req := &Operations{
		Gwr:           gwr,
		ConnectionIds: connIds,
		Tags:          tags,
		PrefixTag:     prefixTag,
	}
	var resp int
	if err := GorpcClient.CallWithAddress(gatewayAddr, "GorpcService", "DoOperations", req, &resp); err != nil {
		return err
	}
	return nil
}

func DoOperationsBatch(gatewayAddr string, req []*Operations) error {
	var resp int
	if err := GorpcClient.CallWithAddress(gatewayAddr, "GorpcService", "DoOperationsBatch", req, &resp); err != nil {
		return err
	}
	return nil
}

func CheckTags(addr string, connId logic.ConnectionId, tags []string) (map[string]bool, error) {
	req := &ConnTags{
		connId,
		tags,
	}
	var resp map[string]bool
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "CheckTags", req, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func GetConnectionTags(addr string, connId logic.ConnectionId) ([]string, error) {
	var resp []string
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "GetConnectionTags", connId, &resp); err != nil {
		return nil, err
	}
	return resp, nil

}

func DelTag(addr string, tag string) error {
	var resp int
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "DelTag", tag, &resp); err != nil {
		return err
	}
	return nil
}

func GetConnectionInfo(addr string, connId logic.ConnectionId) (map[string]string, error) {
	var resp map[string]string
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "GetConnectionInfo", connId, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func ConnLen(addr string) (int, error) {
	var resp int
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "ConnLen", 0, &resp); err != nil {
		return 0, err
	}
	return resp, nil
}

func TagStat(addr string) (map[string]int, error) {
	resp := map[string]int{}
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "TagStat", 0, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func TagStatAll(addr string) (map[string]int, error) {
	resp := map[string]int{}
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "TagStatAll", 0, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func GetLastSecondFlow(addr string) (uint64, error) {
	var resp uint64
	if err := GorpcClient.CallWithAddress(addr, "GorpcService", "GetLastSecondFlow", 0, &resp); err != nil {
		return 0, err
	}
	return resp, nil
}
