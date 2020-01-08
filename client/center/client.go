package center

import (
	"time"

	"github.com/johntech-o/gorpc"
)

var GorpcClient *gorpc.Client

func init() {
	netOptions := gorpc.NewNetOptions(gorpc.DefaultConnectTimeout, gorpc.DefaultReadTimeout, gorpc.DefaultWriteTimeout)
	GorpcClient = gorpc.NewClient(netOptions)

	statNetOption := gorpc.NewNetOptions(1*time.Second, 1*time.Second, 1*time.Second)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetCenterTotalOps", statNetOption)
	GorpcClient.SetMethodNetOptinons("GorpcService", "GetCenterQps", statNetOption)
}
