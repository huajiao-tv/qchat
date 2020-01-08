package main

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/client/router"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/network"
)

type Error struct {
	Id          uint32
	Description string
}

func NewError(id uint32, desc string) *Error {
	return &Error{
		id,
		desc,
	}
}

func GenError(id uint32, err error) *Error {
	return &Error{
		id,
		err.Error(),
	}
}

func (this *Error) Error() string {
	return fmt.Sprintf("%d,%s", this.Id, this.Description)
}

// @todo 2000以内的消息定义为严重错误 ，并断开连接
func (this *Error) IsFatal() bool {
	return this.Id < 2000
}

// 生成一个返回到客户端的ximp
func genErrorXimpBuff(sn uint64, rErr *Error) (*network.XimpBuffer, error) {
	errRes := &pb.Message{
		Msgid: proto.Uint32(0),
		Sn:    &sn,
		Resp: &pb.Response{
			Error: &pb.Error{
				Id:          &rErr.Id,
				Description: []byte(rErr.Description),
			},
		},
	}
	ds, err := proto.Marshal(errRes)
	if err != nil {
		return nil, errors.New("pb marshal error:" + err.Error())
	}
	ximpBuff := &network.XimpBuffer{
		DataStream: ds,
		IsDecrypt:  true,
	}
	return ximpBuff, nil
}

func packErrorResp(gwr *router.GwResp, rErr *Error, m *pb.Message) {
	if gwr.XimpBuff != nil {
		Logger.Warn("", "", "", "packErrorResp", "ximp can't not send to gateway because of error occured", gwr.XimpBuff.String())
	}
	var err error
	gwr.XimpBuff, err = genErrorXimpBuff(m.GetSn(), rErr)
	if err != nil {
		Logger.Warn("", "", "", "packErrorResp", "genErrorXimpBuff error", err)
	}
	if m.GetMsgid() == pb.INIT_LOGIN_REQ {
		gwr.XimpBuff.HasHeader = true
	}
	if rErr.IsFatal() {
		gwr.Actions = append(gwr.Actions, router.DisconnectAction)
	}
}
