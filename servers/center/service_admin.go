package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"reflect"

	"github.com/huajiao-tv/qchat/client/center"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
)

type Stat interface {
	Add(add *Stat) *Stat
	Sub(sub *Stat) *Stat
	String() string
	QpsString() string
}

/*
 * get statistics data handler
 * @param req is http.Request who has called ParseForm
 * @return result string formatted as json
 */
func GetStat(req *http.Request) string {
	centersParam := req.FormValue("centers")
	gatewaysParam := req.FormValue("gateways")
	saversParam := req.FormValue("savers")
	sessionsParam := req.FormValue("sessions")
	routersParam := req.FormValue("routers")

	results := make([]string, 0, 3)

	// query modules
	if centersParam == "" && gatewaysParam == "" && saversParam == "" && sessionsParam == "" && routersParam == "" {
		results = append(results, GetCenterStat(nil), GetGatewayStat(nil), GetSaverStat(nil), GetSessionStat(nil), GetRouterStat(nil))
	} else {
		// query center modules
		if centersParam != "" {
			if centersParam == "all" {
				results = append(results, GetCenterStat(nil))
			} else {
				results = append(results, GetCenterStat(strings.Split(centersParam, ",")))
			}
		}

		// query gateway modules
		if gatewaysParam != "" {
			if gatewaysParam == "all" {
				results = append(results, GetGatewayStat(nil))
			} else {
				results = append(results, GetGatewayStat(strings.Split(gatewaysParam, ",")))
			}
		}

		// query saver modules
		if saversParam != "" {
			if saversParam == "all" {
				results = append(results, GetSaverStat(nil))
			} else {
				results = append(results, GetSaverStat(strings.Split(saversParam, ",")))
			}
		}

		// query session modules
		if sessionsParam != "" {
			if sessionsParam == "all" {
				results = append(results, GetSessionStat(nil))
			} else {
				results = append(results, GetSessionStat(strings.Split(sessionsParam, ",")))
			}
		}

		// query router modules
		if routersParam != "" {
			if routersParam == "all" {
				results = append(results, GetRouterStat(nil))
			} else {
				results = append(results, GetRouterStat(strings.Split(routersParam, ",")))
			}
		}
	}

	return "{ \"code\" : 0, \"reason\" : \"\", \"data\" : [ " +
		strings.Join(results, ", ") + " ] }"
}

/*
 * this callget specified modules statistics data
 * @param rpcs are modules' rpc address
 * @param module indicate which type modules will be queried statistic data
 * @param t is reflect type of correspond statistic struct
 * @return json formatted statistic data
 */
func GetStatFunc(rpcs []string, module string, t reflect.Type) string {
	wg := new(sync.WaitGroup)
	wg.Add(len(rpcs))

	qpsSum := reflect.New(t)
	allOpSum := reflect.New(t)
	nodes := make([]string, 0, len(rpcs))

	qpsChan := make(chan reflect.Value, len(rpcs))
	allOpChan := make(chan reflect.Value, len(rpcs))
	nodesChan := make(chan string, len(rpcs))

	// start counter thread
	go func() {
		for {
			select {
			case qps, ok := <-qpsChan:
				if ok {
					f := qpsSum.MethodByName("Add")
					f.Call([]reflect.Value{qps})
				}
			case allOp, ok := <-allOpChan:
				if ok {
					f := allOpSum.MethodByName("Add")
					f.Call([]reflect.Value{allOp})
				}
			case node, ok := <-nodesChan:
				if ok {
					nodes = append(nodes, node)
				} else {
					return
				}
			}
		}
	}()

	// start query threads according to module
	switch module {
	case "centers":
		for _, rpc := range rpcs {
			go func(rpcAddr string) {
				qps, _ := center.GetCenterQps(rpcAddr)
				allOp, _ := center.GetCenterTotalOps(rpcAddr)
				node := "{ \"node\" : \"" + rpcAddr + "\", \"statistics\" : " + qps.QpsString() +
					", \"all operstions\" : " + allOp.String() + " }"

				qpsChan <- reflect.ValueOf(qps)
				allOpChan <- reflect.ValueOf(allOp)
				nodesChan <- node
				wg.Done()
			}(rpc)
		}
	case "gateways":
		for _, rpc := range rpcs {
			go func(rpcAddr string) {
				qps, _ := gateway.GetGatewayQps(rpcAddr)
				allOp, _ := gateway.GetGatewayTotalOps(rpcAddr)
				node := "{ \"node\" : \"" + rpcAddr + "\", \"statistics\" : " + qps.QpsString() +
					", \"all operstions\" : " + allOp.String() + " }"

				qpsChan <- reflect.ValueOf(qps)
				allOpChan <- reflect.ValueOf(allOp)
				nodesChan <- node
				wg.Done()
			}(rpc)
		}
	case "savers":
		for _, rpc := range rpcs {
			go func(rpcAddr string) {
				qps, _ := saver.GetSaverQps(rpcAddr)
				allOp, _ := saver.GetSaverTotalOps(rpcAddr)
				node := "{ \"node\" : \"" + rpcAddr + "\", \"statistics\" : " + qps.QpsString() +
					", \"all operstions\" : " + allOp.String() + " }"

				qpsChan <- reflect.ValueOf(qps)
				allOpChan <- reflect.ValueOf(allOp)
				nodesChan <- node
				wg.Done()
			}(rpc)
		}
	case "sessions":
		for _, rpc := range rpcs {
			go func(rpcAddr string) {
				qps, _ := session.GetSessionQps(rpcAddr)
				allOp, _ := session.GetSessionTotalOps(rpcAddr)
				node := "{ \"node\" : \"" + rpcAddr + "\", \"statistics\" : " + qps.QpsString() +
					", \"all operstions\" : " + allOp.String() + " }"

				qpsChan <- reflect.ValueOf(qps)
				allOpChan <- reflect.ValueOf(allOp)
				nodesChan <- node
				wg.Done()
			}(rpc)
		}
	case "routers":
		for _, rpc := range rpcs {
			go func(rpcAddr string) {
				qps, _ := router.GetRouterQps(rpcAddr)
				allOp, _ := router.GetRouterTotalOps(rpcAddr)
				node := "{ \"node\" : \"" + rpcAddr + "\", \"statistics\" : " + qps.QpsString() +
					", \"all operstions\" : " + allOp.String() + " }"

				qpsChan <- reflect.ValueOf(qps)
				allOpChan <- reflect.ValueOf(allOp)
				nodesChan <- node
				wg.Done()
			}(rpc)
		}
	default:
		close(qpsChan)
		close(allOpChan)
		// notify counter thread to close
		close(nodesChan)
		return ""
	}

	wg.Wait()
	// wait 100 milliseconds for ensure that counter thread handled all data
	time.Sleep(time.Millisecond * 100)
	close(qpsChan)
	close(allOpChan)
	// notify counter thread to close
	close(nodesChan)

	f := qpsSum.MethodByName("QpsString")
	s := allOpSum.MethodByName("String")

	return fmt.Sprintf("{ \"%s\" : {\"qps sum\" : %v, \"all operstions sum\" : %s, \"nodes\" : [ %s ] } }",
		module, f.Call([]reflect.Value{})[0].Interface(), s.Call([]reflect.Value{})[0].Interface(),
		strings.Join(nodes, ", "))
}

/*
 * get center modules statistic data
 * @param rpcs are center modules' rpc address
 * @return json formatted statistic data
 */
func GetCenterStat(rpcs []string) string {
	if len(rpcs) == 0 {
		rpcs = logic.NetGlobalConf().CenterRpcs
	}

	return GetStatFunc(rpcs, "centers", reflect.TypeOf(center.CenterStat{}))
}

/*
 * get gateway modules statistic data
 * @param rpcs are gateway modules' rpc address
 * @return json formatted statistic data
 */
func GetGatewayStat(rpcs []string) string {
	if len(rpcs) == 0 {
		rpcs = logic.NetGlobalConf().GatewayRpcs
	}

	return GetStatFunc(rpcs, "gateways", reflect.TypeOf(gateway.GatewayStat{}))
}

/*
 * get saver modules statistic data
 * @param rpcs are saver modules' rpc address
 * @return json formatted statistic data
 */
func GetSaverStat(rpcs []string) string {
	if len(rpcs) == 0 {
		rpcs = logic.NetGlobalConf().SaverRpcs
	}

	return GetStatFunc(rpcs, "savers", reflect.TypeOf(saver.SaverStat{}))
}

/*
 * get session modules statistic data
 * @param rpcs are session modules' rpc address
 * @return json formatted statistic data
 */
func GetSessionStat(rpcs []string) string {
	if len(rpcs) == 0 {
		rpcs = logic.NetGlobalConf().SessionRpcs
	}

	return GetStatFunc(rpcs, "sessions", reflect.TypeOf(session.SessionStat{}))
}

/*
 * get router modules statistic data
 * @param rpcs are router modules' rpc address
 * @return json formatted statistic data
 */
func GetRouterStat(rpcs []string) string {
	if len(rpcs) == 0 {
		rpcs = logic.NetGlobalConf().RouterRpcs
	}

	return GetStatFunc(rpcs, "routers", reflect.TypeOf(router.RouterStat{}))
}
