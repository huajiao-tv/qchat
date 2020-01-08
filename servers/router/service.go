package main

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/network"
)

const (
	DEFAULT_QUERY_AFTER_SECONDS = 60

	// 设置在记录resp时如果这个回包是service，需要记录日志的最大长度，因为service记录了
	DEFAULT_LOG_SERVICE_RESP_LEN = 200
	// 设置返回包时普通包记录日志的最大长度
	DEFAULT_LOG_NORMAL_RESP_LEN = 2000

	// 设置保存service的resp的最大日志长度
	DEFAULT_LOG_INNER_SERVICE_RESP_LEN = 2000
)

/***
* state of the connection
 */
const (
	stateIdle      = iota // not yet started
	stateRunning   = iota
	stateInitiated = iota
	stateLoggedIn  = iota
)

const (
	degrade_push = 1
	degrade_pull = 2
)

func getState(gwp *router.GwPackage) int32 {
	if len(gwp.Property["Sender"]) == 0 || len(gwp.Property["Appid"]) == 0 {
		return stateRunning
	} else if len(gwp.Property["IsLoginUser"]) != 0 {
		return stateLoggedIn
	} else {
		return stateInitiated
	}
}

// 校验包括sn, req, sender, 和是否已经登录过了
func checkReq(gwp *router.GwPackage, m *pb.Message) *Error {
	if m.Msgid == nil {
		return NewError(pb.ERR_BAD_PARAM, "msgid should not be nil")
	}
	state := getState(gwp)
	if *m.Msgid == pb.INIT_LOGIN_REQ {
		if state != stateRunning {
			return NewError(pb.ERR_BAD_SEQUENCE, "allready send init login")
		}
		if m.Sender == nil || len(*m.Sender) == 0 {
			return NewError(pb.ERR_BAD_PARAM, "sender is empty")
		}

	} else {
		if *m.Msgid == pb.LOGIN_REQ {
			if state != stateInitiated {
				return NewError(pb.ERR_BAD_SEQUENCE, "do not send login this time")
			}
		} else if state != stateLoggedIn {
			return NewError(pb.ERR_BAD_SEQUENCE, "need login first")
		}
		if m.Sender != nil && gwp.Property["Sender"] != *m.Sender {
			return NewError(pb.ERR_BAD_PARAM, "sender not match")
		}
	}
	if m.Req == nil || m.Sn == nil {
		return NewError(pb.ERR_BAD_PARAM, "some field is nil: req, sn")
	}
	return nil
}

func dealPackage(gwp *router.GwPackage) (*router.GwResp, error) {
	gwr := &router.GwResp{
		Priority: true, // 默认情况下，来自客户端请求的包都会走优先队列下发，除非各自的处理里做特殊处理
	} // 返回到gateway的包
	m := &pb.Message{} // 返回给客户端的pb
	if gwp.Property["ClientIp"] != "" && netConf().BlackIp[gwp.Property["ClientIp"]] {
		gwr.Actions = append(gwr.Actions, router.DisconnectAction)
		return gwr, nil
	}

	// 将pb转成数据结构，或者处理心跳包
	if len(gwp.XimpBuff.DataStream) > 0 {
		if err := proto.Unmarshal(gwp.XimpBuff.DataStream, m); err != nil {
			fmt.Println("dddd",err,gwp.XimpBuff.DataStream)
			Logger.Error(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealPackage", "unmarshal datastream error", err)
			// pb 解码失败，中断链接
			gwr.Actions = append(gwr.Actions, router.DisconnectAction)
			return gwr, err
		}
	} else if gwp.XimpBuff.IsHeartbeat {
		gwr.XimpBuff = &network.XimpBuffer{
			IsDecrypt:   false, // 心跳包不需要加解密
			IsHeartbeat: true,
		}
		return gwr, nil
	} else {
		Logger.Error(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealPackage", "datastream is empty", "")
		return gwr, errors.New("datastream is empty")
	}
	Logger.Debug(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealPackage", "req", m.String())

	if rErr := checkReq(gwp, m); rErr != nil {
		packErrorResp(gwr, rErr, m)
		return gwr, rErr
	}

	var rErr *Error
	var respM *pb.Message // 返回到gw的pb内容，做日志记录用
	switch *m.Msgid {
	case pb.INIT_LOGIN_REQ:
		respM, rErr = dealInitLoginReq(gwp, m, gwr)
	case pb.LOGIN_REQ:
		respM, rErr = dealLoginReq(gwp, m, gwr)
	case pb.GET_INFO_REQ:
		respM, rErr = dealGetInfoReq(gwp, m, gwr)
	case pb.SERVICE_REQ:
		respM, rErr = dealServiceReq(gwp, m, gwr)
	case pb.LOGOUT_REQ:
		respM, rErr = dealLogoutReq(gwp, m, gwr)
	case pb.GET_MULITI_INFOS_REQ:
		respM, rErr = dealGetMultiInfoReq(gwp, m, gwr)
	default:
		rErr = NewError(pb.ERR_BAD_PARAM, "unknown msgid!")
	}
	if rErr != nil {
		Logger.Error(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealPackage", "deal each package error", rErr.Error())
		packErrorResp(gwr, rErr, m)
		return gwr, rErr
	} else {
		Logger.Debug(gwp.Property["Sender"], gwp.Property["Appid"], logic.GetTraceId(gwp.GatewayAddr, gwp.ConnectionId), "dealPackage", "resp", formatRespMessage(respM))
	}
	return gwr, nil
}

func formatRespMessage(respM *pb.Message) string {
	if respM == nil {
		return ""
	}
	result := respM.String()
	if *respM.Msgid == pb.SERVICE_RESP {
		if len(result) > DEFAULT_LOG_SERVICE_RESP_LEN {
			result = result[0:DEFAULT_LOG_SERVICE_RESP_LEN]
		}
	} else if len(result) > DEFAULT_LOG_NORMAL_RESP_LEN {
		result = result[0:DEFAULT_LOG_NORMAL_RESP_LEN]
	}
	return result
}

func getQueryAfterSeconds() uint32 {
	if netConf().QueryAfterSeconds <= 0 {
		return uint32(rand.Int31n(DEFAULT_QUERY_AFTER_SECONDS))
	} else {
		return uint32(rand.Int31n(int32(netConf().QueryAfterSeconds)))
	}
}

func createReConnectNotify(ip string, port uint32, moreIps []string) (*network.XimpBuffer, error) {
	sn := logic.GetSn()
	reConnectNotify := &pb.Message{
		Msgid:    proto.Uint32(pb.RE_CONNECT_NOTIFY),
		Sn:       &sn,
		Sender:   proto.String(""),
		Receiver: proto.String(""),
		Notify: &pb.Notify{
			ReconnectNtf: &pb.ReConnectNotify{
				Ip:      &ip,
				Port:    &port,
				MoreIps: moreIps,
			},
		},
	}
	pbBytes, err := proto.Marshal(reConnectNotify)
	if err != nil {
		return nil, err
	}

	ximp := &network.XimpBuffer{
		IsDecrypt:  true,
		DataStream: pbBytes,
	}
	return ximp, nil
}

func sendNotify(gatewayAddr string, tags []string, prefix string, connIds []logic.ConnectionId, ximp *network.XimpBuffer, priority bool) error {
	gwr := &router.GwResp{
		XimpBuff: ximp,
		Priority: priority,
	}
	params := &DoOperationsParams{
		GatewayAddr: gatewayAddr,
		Tags:        tags,
		ConnIds:     connIds,
		Gwr:         gwr,
		PrefixTag:   prefix,
	}
	addDoOperations(params)
	return nil
}

func checkPushDegrade(infoType, userId string, pull bool) bool {
	_, degradeUser := netConf().PushDegradeUsers[userId]
	// @todo, public 放到 PushDegradeAll 里
	if degradeUser || netConf().PushDegradeAll || infoType == saver.ChatChannelPublic {
		if pull {
			return (netConf().PushDegrade[infoType] & degrade_pull) != 0
		} else {
			return (netConf().PushDegrade[infoType] & degrade_push) != 0
		}
	}
	return false
}
