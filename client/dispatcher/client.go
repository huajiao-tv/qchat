package dispatcher

import (
	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)
}

func GetQps(address string) (map[string]float64, error) {
	var resp map[string]float64
	if err := GorpcClient.CallWithAddress(address, "GorpcService", "GetQps", 0, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
