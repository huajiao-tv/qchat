package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/network"
)

type Client struct {
	userId     string
	server     string
	sessionKey []byte

	conn     *network.TcpConnection
	stopChan chan bool
}

func NewClient(uid string, svr string) *Client {
	client := &Client{
		userId:   uid,
		server:   svr,
		stopChan: make(chan bool),
	}

	go client.main()
	return client
}

func (c *Client) main() {
	for {
		// 1. connect
		var err error
		ts := time.Now()
		c.conn, err = network.TcpConnect("", c.server, ConnectTimeout)
		if err != nil {
			Logger.Debug(c.server, c.userId, "connect gateway failed", err)
			time.Sleep(time.Duration(10) * time.Second)
			continue
		}
		Logger.Debug(c.server, c.userId, "connect gateway cost", time.Now().Sub(ts).Nanoseconds())

		// 2. login
		ts = time.Now()
		err = c.login()
		if err != nil {
			Logger.Debug(c.server, c.userId, "login failed", err.Error())
			c.conn.Close()
			time.Sleep(time.Duration(10) * time.Second)
			continue
		}
		Logger.Debug(c.server, c.userId, "login cost", time.Now().Sub(ts).Nanoseconds())

		// 3. start heartbeat
		go c.heartbeat()

		// 4. get infos
		// im
		ts = time.Now()
		err = c.getInfo("im")
		if err != nil {
			Logger.Debug(c.server, c.userId, "get im info failed", err.Error())
			goto Restart
		}
		Logger.Debug(c.server, c.userId, "get im info cost", time.Now().Sub(ts).Nanoseconds())
		// peer
		ts = time.Now()
		err = c.getInfo("peer")
		if err != nil {
			Logger.Debug(c.server, c.userId, "get peer info failed", err.Error())
			goto Restart
		}
		Logger.Debug(c.server, c.userId, "get peer info cost", time.Now().Sub(ts).Nanoseconds())
		// public
		ts = time.Now()
		err = c.getInfo("public")
		if err != nil {
			Logger.Debug(c.server, c.userId, "get public info failed", err.Error())
			goto Restart
		}
		Logger.Debug(c.server, c.userId, "get public info cost", time.Now().Sub(ts).Nanoseconds())

		// 5. join room
		ts = time.Now()
		c.join(roomId)
		Logger.Debug(c.server, c.userId, "join chatroom cost", time.Now().Sub(ts).Nanoseconds())

		// 6. receive messages
		for {
			msg, err := c.readPacket(HeartBeatTimeout * 2)
			if err != nil {
				Logger.Debug(c.server, c.userId, "read packet failed", err.Error())
				goto Restart
			}
			c.unpackMessage(msg)
		}

	Restart:
		c.conn.Close()
		c.stopChan <- true
		Logger.Debug(c.server, c.userId, "reconnecting")
		time.Sleep(time.Minute)
	}
}

func (c *Client) heartbeat() {
	t := time.NewTicker(HeartBeatTimeout)

	for {
		select {
		case <-t.C:
			pingXimp := &network.XimpBuffer{
				IsHeartbeat: true,
			}
			if err := pingXimp.WriteTo(c.conn, WriteTimeout); err != nil {
				Logger.Debug(c.server, c.userId, "ping failed", err)
				continue
			}
		case <-c.stopChan:
			return
		}
	}
}

func (c *Client) readPacket(timeout time.Duration) (*network.XimpBuffer, error) {
	for {
		pkt := network.NewXimpBuffer()
		if err := pkt.ReadFrom(false, c.conn, timeout); err != nil {
			return nil, err
		}
		if pkt.IsHeartbeat {
			continue
		}
		return pkt, nil
	}
}

func (c *Client) login() error {
	initXimp, err := c.makeInitLogin()
	if err != nil {
		errMsg := fmt.Sprintf("makeInitLogin error: %s", err.Error())
		return errors.New(errMsg)
	}
	if err := initXimp.WriteTo(c.conn, WriteTimeout); err != nil {
		errMsg := fmt.Sprintf("write initXimp error: %s", err.Error())
		return errors.New(errMsg)
	}

	initResp, err := c.readPacket(ReadTimeout)
	if err != nil {
		errMsg := fmt.Sprintf("initResp readfrom error: %s", err.Error())
		return errors.New(errMsg)
	}
	serverRam, err := c.unpackInitResp(initResp)
	if err != nil {
		errMsg := fmt.Sprintf("unpackInitResp error: %s", err.Error())
		return errors.New(errMsg)
	}

	loginXimp, err := c.makeLogin(serverRam)
	if err != nil {
		errMsg := fmt.Sprintf("makeLogin error: %s", err.Error())
		return errors.New(errMsg)
	}
	if err := loginXimp.WriteTo(c.conn, WriteTimeout); err != nil {
		errMsg := fmt.Sprintf("write loginXimp error: %s", err.Error())
		return errors.New(errMsg)
	}

	loginResp, err := c.readPacket(ReadTimeout)
	if err != nil {
		errMsg := fmt.Sprintf("loginResp readfrom error: %s", err.Error())
		return errors.New(errMsg)
	}
	key, err := c.unpackLoginResp(loginResp)
	if err != nil {
		errMsg := fmt.Sprintf("unpackLoginResp error: %s", err.Error())
		return errors.New(errMsg)
	}

	c.sessionKey = []byte(key)
	return nil
}

func (c *Client) makeInitLogin() (*network.XimpBuffer, error) {
	clientRan := logic.RandString(8)
	sn := logic.GetSn()
	m := &pb.Message{
		Sender: &c.userId,
		Sn:     &sn,
		Msgid:  proto.Uint32(pb.INIT_LOGIN_REQ),
		Req: &pb.Request{
			InitLoginReq: &pb.InitLoginReq{
				ClientRam: &clientRan,
			},
		},
	}

	pbData, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	buf := &network.XimpBuffer{
		IsDecrypt:  true,
		IsClient:   true,
		HasHeader:  true,
		Version:    1,
		CVersion:   102,
		Appid:      AppId,
		DataStream: pbData,
	}
	if err := buf.Encrypt(DefaultKey); err != nil {
		return nil, err
	}

	return buf, err
}

func (c *Client) unpackInitResp(initResp *network.XimpBuffer) (string, error) {
	if err := initResp.Decrypt(DefaultKey); err != nil {
		return "", err
	}

	m := &pb.Message{}
	if len(initResp.DataStream) > 0 {
		if err := proto.Unmarshal(initResp.DataStream, m); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("no data return")
	}

	if *m.Msgid != uint32(pb.INIT_LOGIN_RESP) {
		return "", errors.New("m.Msgid is not ok")
	}

	return *(m.Resp.InitLoginResp.ServerRam), nil
}

func (c *Client) makeLogin(serverRam string) (*network.XimpBuffer, error) {
	sn := logic.GetSn()
	var netType uint32 = 3
	var verfCode = logic.MakeVerfCode(c.userId)
	var secretRan []byte = logic.MakeSecretRan(c.userId, serverRam)
	m := &pb.Message{
		Sender:     proto.String(c.userId),
		SenderType: proto.String("jid"),
		Sn:         &sn,
		Msgid:      proto.Uint32(pb.LOGIN_REQ),
		Req: &pb.Request{
			Login: &pb.LoginReq{
				VerfCode:   &verfCode,
				AppId:      proto.Uint32(uint32(AppId)),
				NetType:    &netType,
				MobileType: proto.String("android"),
				ServerRam:  proto.String(serverRam),
				SecretRam:  secretRan,
				HeartFeq:   proto.Uint32(uint32(60)),
				Deviceid:   proto.String(logic.RandString(10)),
				Platform:   proto.String(""),
				NotEncrypt: proto.Bool(true),
			},
		},
	}
	pbData, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	buf := &network.XimpBuffer{
		IsDecrypt:  true,
		IsClient:   true,
		Version:    1,
		CVersion:   102,
		Appid:      AppId,
		DataStream: pbData,
	}
	if err := buf.Encrypt(DefaultKey); err != nil {
		return nil, err
	}

	return buf, err
}

func (c *Client) unpackLoginResp(loginResp *network.XimpBuffer) (string, error) {
	if err := loginResp.Decrypt([]byte(c.userId)); err != nil {
		return "", errors.New("decrypt error")
	}

	m := &pb.Message{}
	if len(loginResp.DataStream) > 0 {
		if err := proto.Unmarshal(loginResp.DataStream, m); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("no data return")
	}

	if *m.Msgid != uint32(pb.LOGIN_RESP) {
		return "", errors.New("m.Msgid is not ok")
	}
	if m.Resp.Login.SessionKey == nil {
		return "", nil
	} else {
		return *(m.Resp.Login.SessionKey), nil
	}
}

func (c *Client) makeService(serviceId uint32, data []byte) (*network.XimpBuffer, error) {
	var sn = logic.GetSn()
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.SERVICE_REQ),
		Req: &pb.Request{
			ServiceReq: &pb.Service_Req{
				ServiceId: &serviceId,
				Request:   data,
			},
		},
	}
	pbData, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}

	buf := &network.XimpBuffer{
		IsDecrypt:  true,
		IsClient:   true,
		Version:    3,
		CVersion:   102,
		Appid:      AppId,
		DataStream: pbData,
	}
	if err := buf.Encrypt(c.sessionKey); err != nil {
		return nil, err
	}
	return buf, nil
}

func (c *Client) unpackService(resp *network.XimpBuffer) (*pb.Message, error) {
	if err := resp.Decrypt(c.sessionKey); err != nil {
		return nil, err
	}
	m := &pb.Message{}
	if len(resp.DataStream) > 0 {
		if err := proto.Unmarshal(resp.DataStream, m); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no data return")
	}
	if *m.Msgid != uint32(pb.SERVICE_RESP) {
		return nil, errors.New("m.Msgid is not ok")
	}
	return m, nil
}

func (c *Client) join(roomId string) error {
	joinXimp, err := c.makeJoin(roomId)
	if err != nil {
		errMsg := fmt.Sprintf("makeJoin error: %s", err.Error())
		return errors.New(errMsg)
	}
	if err := joinXimp.WriteTo(c.conn, WriteTimeout); err != nil {
		errMsg := fmt.Sprintf("write joinXimp  error: %s", err.Error())
		return errors.New(errMsg)
	}

	joinResp, err := c.readPacket(ReadTimeout)
	if err != nil {
		errMsg := fmt.Sprintf("joinResp readfrom error: %s", err.Error())
		return errors.New(errMsg)
	}
	if err := c.unpackJoinResp(joinResp); err != nil {
		errMsg := fmt.Sprintf("unpackJoinResp error: %s", err.Error())
		return errors.New(errMsg)
	}

	return nil
}

func (c *Client) makeJoin(roomId string) (*network.XimpBuffer, error) {
	rid := []byte(roomId)
	payload := uint32(pb.CR_PAYLOAD_JOIN)
	appid := uint32(AppId)
	sn := logic.GetSn()
	p := &pb.ChatRoomPacket{
		Roomid:   rid,
		Appid:    &appid,
		ClientSn: &sn,
		ToServerData: &pb.ChatRoomUpToServer{
			Applyjoinchatroomreq: &pb.ApplyJoinChatRoomRequest{
				Roomid: rid,
			},
			Payloadtype: &payload,
		},
	}

	data, err := proto.Marshal(p)

	if err != nil {
		return nil, err
	}
	return c.makeService(pb.CHATROOM_SERVICE_ID, data)
}

func (c *Client) unpackJoinResp(joinResp *network.XimpBuffer) error {
	m, err := c.unpackService(joinResp)
	if err != nil {
		return err
	}

	if *m.Resp.ServiceResp.ServiceId != uint32(pb.CHATROOM_SERVICE_ID) {
		return errors.New("not a join resp")
	}
	var p pb.ChatRoomPacket
	if err := proto.Unmarshal(m.Resp.ServiceResp.Response, &p); err != nil {
		return err
	}
	if *p.ToUserData.Result != 0 {
		return errors.New("result is not 0")
	}
	return nil
}

func (c *Client) unpackMessage(msg *network.XimpBuffer) {
	if msg.IsHeartbeat {
		return
	}
	if len(msg.DataStream) <= 0 {
		return
	}
	if err := msg.Decrypt(c.sessionKey); err != nil {
		Logger.Debug(c.server, c.userId, "decrypt msg error", err.Error())
		return
	}

	m := &pb.Message{}
	if err := proto.Unmarshal(msg.DataStream, m); err != nil {
		Logger.Debug(c.server, c.userId, "unmarshal msg error", err.Error())
		return
	}
	if *m.Msgid != uint32(pb.NEW_MESSAGE_NOTIFY) {
		Logger.Debug(c.server, c.userId, "new msg notify", *m.Msgid)
		return
	}

	if *m.Notify.NewinfoNtf.InfoType == "chatroom" {
		c.unpackChatRoomMsg(m)
	}
}

func (c *Client) unpackChatRoomMsg(m *pb.Message) {
	var p pb.ChatRoomPacket
	if err := proto.Unmarshal(m.Notify.NewinfoNtf.InfoContent, &p); err != nil {
		Logger.Debug(c.server, c.userId, "unmarshal chat room notify error", err)
		return
	}

	switch p.ToUserData.GetPayloadtype() {
	case pb.CR_PAYLOAD_INCOMING_MSG:
		content := p.ToUserData.Newmsgnotify.Msgcontent
		c.calDelay(content)

	case pb.CR_PAYLOAD_COMPRESSED:
		for _, content := range p.ToUserData.Multinotify {
			if *content.Type != pb.CR_PAYLOAD_INCOMING_MSG {
				continue
			}
			data := ungzip(content.Data)
			msg := pb.ChatRoomNewMsg{}
			if err := proto.Unmarshal(data, &msg); err != nil {
				Logger.Debug(c.server, c.userId, "unmarshal compressed chat room notify error", err.Error())
				continue
			}
			c.calDelay(msg.Msgcontent)
		}
	}
}

func (c *Client) calDelay(data []byte) {
	var s PayloadJson
	if err := json.Unmarshal(data, &s); err != nil {
		return
	}
	if s.TimeStamp == 0 {
		return
	}
	d := time.Duration(time.Now().UnixNano()-s.TimeStamp) * time.Nanosecond
	Logger.Debug(c.server, c.userId, "chatroom message delay", d.Nanoseconds(), s.TimeStamp)
}

func (c *Client) getInfo(infoType string) error {
	getInfoXimp, err := c.makeGetInfo(infoType)
	if err != nil {
		errMsg := fmt.Sprintf("makeGetInfo error: %s", err.Error())
		return errors.New(errMsg)
	}
	if err := getInfoXimp.WriteTo(c.conn, WriteTimeout); err != nil {
		errMsg := fmt.Sprintf("write getInfoXimp error: %s", err.Error())
		return errors.New(errMsg)
	}

	getInfoResp, err := c.readPacket(ReadTimeout)
	if err != nil {
		errMsg := fmt.Sprintf("getInfoResp readfrom error: %s", err.Error())
		return errors.New(errMsg)
	}
	if err := c.unpackGetInfo(getInfoResp); err != nil {
		errMsg := fmt.Sprintf("unpackGetInfo error: %s", err.Error())
		return errors.New(errMsg)
	}

	return nil
}

func (c *Client) makeGetInfo(infoType string) (*network.XimpBuffer, error) {
	var sn = logic.GetSn()
	start := int64(0)
	size := int32(5)
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.GET_INFO_REQ),
		Req: &pb.Request{
			GetInfo: &pb.GetInfoReq{
				InfoType:      &infoType,
				GetInfoId:     &start,
				GetInfoOffset: &size,
			},
		},
	}
	pbData, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	buf := &network.XimpBuffer{
		DataStream: pbData,
	}
	if err := buf.Encrypt(c.sessionKey); err != nil {
		return nil, err
	}
	return buf, err
}

func (c *Client) unpackGetInfo(resp *network.XimpBuffer) error {
	if err := resp.Decrypt(c.sessionKey); err != nil {
		return errors.New("decrypt error")
	}

	m := &pb.Message{}
	if len(resp.DataStream) > 0 {
		if err := proto.Unmarshal(resp.DataStream, m); err != nil {
			return err
		}
	} else {
		return errors.New("no data return")
	}

	if *m.Msgid != uint32(pb.GET_INFO_RESP) {
		return errors.New("m.Msgid is not ok")
	}
	//for _, info := range m.Resp.GetInfo.Infos {
	//}
	return nil
}
