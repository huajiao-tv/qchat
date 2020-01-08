package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/qchat/client/gateway"
	"github.com/huajiao-tv/qchat/client/saver"
	"github.com/huajiao-tv/qchat/client/session"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/data"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/network"
)

var KeeperAddr string = "127.0.0.1:7000"
var NodeID string = "test"
var Sections []string = []string{"global.conf", "saver.conf"}
var Domain string = "qchat_test"
var Component string = "dev_test"

func init() {
	keeperCli := gokeeper.New(KeeperAddr, Domain, NodeID, Component, Sections, nil)
	keeperCli.LoadData(data.ObjectsContainer).RegisterCallback(logic.UpdateDynamicConfType)
	if err := keeperCli.Work(); err != nil {
		panic(err)
	}
}

func main() {
	client := &XimpClient{
		Server:         "127.0.0.1:8080",
		ConnTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    10 * time.Second,
		ReceiveTimeout: 10 * time.Second,

		User:            "",
		Sig:             "",
		DefaultKey:      []byte("894184791415baf5c113f83eaff360f0"),
		ProtocolVersion: 1,
		ClientVersion:   100,
		Appid:           1080,
		NetType:         3,
		Password:        "",
		UserType:        "jid",
		MobileType:      "android",
		DeviceId:        "device_id",
		Platform:        "mobile",
		HeartFeq:        300,
		MessagePool:     NewMsgPool(),
		CloseRead:       make(chan bool, 1),
		LastId:          make(map[string]int64),
	}
	if err := client.Login(); err != nil {
		fmt.Println("Connect with error:", err)
		return
	}
	if err := client.GetInfo("peer"); err != nil {
		fmt.Println("Connect with error:", err)
		return
	}
	if err := client.GetInfo("public"); err != nil {
		fmt.Println("Connect with error:", err)
		return
	}
	if err := checkSession(client); err != nil {
		fmt.Println("checkSession error:", err)
		return
	}
	if err := checkGateway(client); err != nil {
		fmt.Println("checkGateway error:", err)
		return
	}
	if err := client.JoinChatroom("12345"); err != nil {
		fmt.Println("JoinChatroom error:", err)
		return
	}

	if err := client.Logout(); err != nil {
		fmt.Println("logout error:", err)
		return
	}
	if err := checkLogoutSession(client); err != nil {
		fmt.Println("checkLogoutSession error:", err)
		return
	}
	client.Close()
	time.Sleep(time.Second)

	if err := checkLogoutGateway(client); err != nil {
		fmt.Println("CheckLogoutGateway error:", err)
		return
	}

	fmt.Println("PASS!!!!")
	select {}

}
func (this *XimpClient) JoinChatroom(roomId string) error {
	ximpBuff, err := this.MakeJoinChatroom(roomId)
	if err != nil {
		return err
	}
	if err := ximpBuff.WriteTo(this.Conn, this.WriteTimeout); err != nil {
		return err
	}
	m, err := this.MessagePool.Receive(this.CurrentSn, this.ReceiveTimeout)
	if err != nil {
		return err
	}
	fmt.Println("Get get service resp", m.String())
	// 校验结果
	if pb.SERVICE_RESP != m.GetMsgid() {
		return errors.New("not a serice_resp")
	}
	if pb.CR_PAYLOAD_JOIN_RESP != m.Resp.ServiceResp.GetServiceId() {
		return errors.New("not a join_resp")
	}
	if m.GetSn() != this.CurrentSn {
		return errors.New("sn not match")
	}
	p := &pb.ChatRoomPacket{}
	if err := proto.Unmarshal(m.Resp.ServiceResp.GetResponse(), p); err != nil {
		return err
	}
	fmt.Println("Get get chatroom_join resp", p.String())

	return nil
}

func (this *XimpClient) MakeJoinChatroom(roomId string) (*network.XimpBuffer, error) {
	rid := []byte(roomId)
	this.CurrentSn = logic.GetSn()
	payload := uint32(pb.CR_PAYLOAD_JOIN)
	appid := uint32(this.Appid)
	p := &pb.ChatRoomPacket{
		Roomid:   rid,
		Appid:    &appid,
		ClientSn: proto.Uint64(this.CurrentSn),
		ToServerData: &pb.ChatRoomUpToServer{
			Applyjoinchatroomreq: &pb.ApplyJoinChatRoomRequest{
				Roomid: rid,
			},
			Payloadtype: &payload,
		},
	}
	fmt.Println("make join req:", p.String())

	return this.PackChatroomMessage(p)

}

func (this *XimpClient) PackChatroomMessage(p *pb.ChatRoomPacket) (*network.XimpBuffer, error) {
	var ds []byte
	if b, err := proto.Marshal(p); err != nil {
		return nil, err
	} else {
		ds = b
	}
	var sn = this.CurrentSn
	srvID := uint32(pb.CHATROOM_SERVICE_ID)
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.SERVICE_REQ),
		Req: &pb.Request{
			ServiceReq: &pb.Service_Req{
				ServiceId: &srvID,
				Request:   ds,
			},
		},
	}
	fmt.Println("make join req:", m.String())
	return this.PackMessage(false, false, m)

}

func checkLogoutGateway(c *XimpClient) error {
	if info, err := gateway.GetConnectionInfo(c.Session.GatewayAddr, c.Session.ConnectionId); err != nil {
		if !strings.Contains(err.Error(), "connection not found") {
			return err
		} else {
			fmt.Println("gateway info not found", err)
		}
	} else {
		fmt.Println("=====get close gateinfo====")
		showJson(info)
		return errors.New("gate info is not null after logout")
	}
	return nil
}
func (this *XimpClient) Close() {
	select {
	case this.CloseRead <- true:
	default:
	}
	if this.Conn != nil {
		this.Conn.Close()
	}
}

func checkGateway(c *XimpClient) error {
	if info, err := gateway.GetConnectionInfo(c.Session.GatewayAddr, c.Session.ConnectionId); err != nil {
		return err
	} else {
		fmt.Println("=====get gateinfo====")
		showJson(info)
		if info["Rkey"] != string(c.SessionKey) {
			return errors.New("SessionKey not match")
		}
		if info["Sender"] != c.User {
			return errors.New("User not match")
		}
	}

	if tags, err := gateway.GetConnectionTags(c.Session.GatewayAddr, c.Session.ConnectionId); err != nil {
		return err
	} else {
		fmt.Println("======get gateinfo tags ==== ")
		showJson(tags)
		if len(tags) == 0 {
			return errors.New("not found appid tag")
		}
		found := false
		for _, t := range tags {
			if t == c.Session.Property["Appid"] {
				found = true
				break
			}
		}
		if !found {
			return errors.New("not found appid tags")
		}
	}

	return nil
}

func checkLogoutSession(c *XimpClient) error {
	req := &session.UserSession{
		AppId:  c.Appid,
		UserId: c.User,
	}
	news, err := saver.QueryUserSession([]*session.UserSession{req})
	if err != nil {
		return errors.New("QueryUserSession with err:" + err.Error())
	}
	fmt.Println("=====get sessions after logout====")
	showJson(news)
	for _, s := range news {
		if s.Platform == c.Platform && s.Deviceid == c.DeviceId {
			return errors.New("get a  session after logout!!")
		}
	}
	return nil
}

func checkSession(c *XimpClient) error {
	req := &session.UserSession{
		AppId:  c.Appid,
		UserId: c.User,
	}
	news, err := saver.QueryUserSession([]*session.UserSession{req})
	if err != nil {
		return errors.New("QueryUserSession with err:" + err.Error())
	}
	fmt.Println("=====get sessions====")
	showJson(news)
	for _, s := range news {
		if s.Platform == c.Platform && s.Deviceid == c.DeviceId {
			if c.Session != nil {
				fmt.Println("warning :there one more same session!!")
			}
			c.Session = s
		}
	}
	if c.Session == nil {
		return errors.New("error:check session faild:no session found!!")
	}
	return nil
}

type MsgPool struct {
	sync.Mutex
	Msgs      map[uint64]*pb.Message
	MsgNotify chan bool
}

func (this *MsgPool) Save(sn uint64, m *pb.Message) {
	this.Lock()
	defer this.Unlock()
	if _, ok := this.Msgs[sn]; ok {
		fmt.Println("receive message(raw) with same sn:", m.String())
	} else {
		fmt.Println("receive message(raw):", m.String())
	}
	this.Msgs[sn] = m
	select {
	case this.MsgNotify <- true:
	default:
	}
}

func (this *MsgPool) Receive(sn uint64, timeout time.Duration) (*pb.Message, error) {
	timer := time.NewTimer(timeout)
	for {
		this.Lock()
		if m, ok := this.Msgs[sn]; ok {
			delete(this.Msgs, sn)
			this.Unlock()
			return m, nil
		}
		this.Unlock()
		select {
		case <-timer.C:
			return nil, errors.New("msg not received")
		case <-this.MsgNotify:
		}
	}
}

func NewMsgPool() *MsgPool {
	return &MsgPool{
		Msgs:      make(map[uint64]*pb.Message),
		MsgNotify: make(chan bool, 1),
	}
}

type XimpClient struct {
	Server         string
	ConnTimeout    time.Duration
	Conn           *network.TcpConnection
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	ReceiveTimeout time.Duration

	User            string
	Sig             string
	DefaultKey      []byte
	ProtocolVersion uint16
	ClientVersion   uint16
	Appid           uint16
	NetType         uint32
	Password        string
	UserType        string
	MobileType      string
	Platform        string
	DeviceId        string
	HeartFeq        uint32

	SessionKey []byte
	ClientRam  string
	CurrentSn  uint64
	ServerRam  string
	Session    *session.UserSession

	MessagePool *MsgPool
	CloseRead   chan bool
	LastId      map[string]int64
}

func (this *XimpClient) Login() error {
	var err error
	if this.Conn, err = network.TcpConnect("", this.Server, this.ConnTimeout); err != nil {
		return err
	}
	if err := this.DealInitLoginPacket(); err != nil {
		return errors.New("DealInitLoginPacket with err:" + err.Error())
	}
	if err := this.DealLoginPacket(); err != nil {
		return errors.New("DealLoginPacket with err:" + err.Error())
	}
	go this.Read()
	return nil
}

func (this *XimpClient) MakeGetInfo(channel string) (*network.XimpBuffer, error) {
	this.CurrentSn = logic.GetSn()
	start := this.LastId[channel] + 1
	size := int32(binary.MaxVarintLen32)
	m := &pb.Message{
		Sn:    proto.Uint64(this.CurrentSn),
		Msgid: proto.Uint32(pb.GET_INFO_REQ),
		Req: &pb.Request{
			GetInfo: &pb.GetInfoReq{
				InfoType:      &channel,
				GetInfoId:     &start,
				GetInfoOffset: &size,
			},
		},
	}
	fmt.Println("MakeGetInfo:", m.String())

	return this.PackMessage(false, false, m)
}
func (this *XimpClient) Logout() error {
	ximpBuff, err := this.MakeLogoutPacket()
	if err != nil {
		return err
	}

	if err := ximpBuff.WriteTo(this.Conn, this.WriteTimeout); err != nil {
		return err
	}

	m, err := this.MessagePool.Receive(this.CurrentSn, this.ReceiveTimeout)
	if err != nil {
		return err
	}

	fmt.Println("Get get logout resp", m.String())
	// 校验结果
	if pb.LOGOUT_RESP != m.GetMsgid() {
		return errors.New("not a logout resp")
	}
	if m.GetSn() != this.CurrentSn {
		return errors.New("sn not match")
	}
	return nil

}
func (this *XimpClient) MakeLogoutPacket() (*network.XimpBuffer, error) {
	this.CurrentSn = logic.GetSn()
	m := &pb.Message{
		Sender:     proto.String(this.User),
		SenderType: proto.String(this.UserType),
		Sn:         proto.Uint64(this.CurrentSn),
		Msgid:      proto.Uint32(pb.LOGOUT_REQ),
		Req: &pb.Request{
			Logout: &pb.LogoutReq{
				Reason: proto.String("want to logout"),
			},
		},
	}

	fmt.Println("MakeLogoutPacket:", m.String())

	return this.PackMessage(false, false, m)
}

func (this *XimpClient) GetInfo(channel string) error {
	ximpBuff, err := this.MakeGetInfo(channel)
	if err != nil {
		return err
	}
	if err := ximpBuff.WriteTo(this.Conn, this.WriteTimeout); err != nil {
		return err
	}
	m, err := this.MessagePool.Receive(this.CurrentSn, this.ReceiveTimeout)
	if err != nil {
		return err
	}

	fmt.Println("Get get info resp", m.String())
	// 校验结果
	if pb.GET_INFO_RESP != m.GetMsgid() {
		return errors.New("not a get_info_resp")
	}
	if m.GetSn() != this.CurrentSn {
		return errors.New("sn not match")
	}
	return nil
}

func (this *XimpClient) Read() {
	for {
		select {
		case <-this.CloseRead:
			return
		default:
		}
		ximpBuff := &network.XimpBuffer{}
		if err := ximpBuff.ReadFrom(false, this.Conn, time.Duration(this.HeartFeq)*time.Second+this.ReadTimeout); err != nil {
			fmt.Println("readfrom with error and exit:", err)
			return
		}
		if ximpBuff.IsHeartbeat {
			this.MessagePool.Save(0, &pb.Message{})
		} else {
			if m, err := this.UnPackMessage(this.SessionKey, ximpBuff); err != nil {
				fmt.Println("UnPackMessage with err:", err)
			} else {
				this.MessagePool.Save(m.GetSn(), m)
			}
		}
	}
}

func (this *XimpClient) DealLoginPacket() error {
	ximpBuff, err := this.MakeLoginPackage()
	if err != nil {
		return err
	}
	if err := ximpBuff.WriteTo(this.Conn, this.WriteTimeout); err != nil {
		return err
	}
	resp := &network.XimpBuffer{}
	if err := resp.ReadFrom(false, this.Conn, this.ReadTimeout); err != nil {
		return nil
	}
	if resp.IsHeartbeat || len(resp.DataStream) == 0 {
		return errors.New("expect a initlogin, but is heartbeat or datastream is empty")
	}
	backDs := make([]byte, len(resp.DataStream))
	copy(backDs, resp.DataStream)
	var m *pb.Message
	if t, err := this.UnPackMessage([]byte(this.Password), resp); err != nil {
		resp.DataStream = backDs
		resp.IsDecrypt = false
		if t, err := this.UnPackMessage(this.DefaultKey, resp); err != nil {
			return err
		} else {
			m = t
		}
	} else {
		m = t
	}
	fmt.Println("Get login resp", m.String())
	// 校验结果
	if pb.LOGIN_RESP != m.GetMsgid() {
		return errors.New("not a login_resp")
	}
	if m.GetSn() != this.CurrentSn {
		return errors.New("sn not match")
	}
	this.SessionKey = []byte(m.Resp.Login.GetSessionKey())
	if len(this.SessionKey) == 0 {
		return errors.New("empty session key")
	}
	return nil
}
func (this *XimpClient) MakeLoginPackage() (*network.XimpBuffer, error) {
	this.CurrentSn = logic.GetSn()
	var verfCode = logic.MakeVerfCode(this.User)
	var secretRam []byte = logic.MakeSecretRan(this.Password, this.ServerRam)

	m := &pb.Message{
		Sender:     proto.String(this.User),
		SenderType: proto.String(this.UserType),
		Sn:         proto.Uint64(this.CurrentSn),
		Msgid:      proto.Uint32(pb.LOGIN_REQ),
		Req: &pb.Request{
			Login: &pb.LoginReq{
				AppId:      proto.Uint32(uint32(this.Appid)),
				NetType:    proto.Uint32(this.NetType),
				MobileType: proto.String(this.MobileType),
				Deviceid:   proto.String(this.DeviceId),
				Platform:   proto.String(this.Platform),
				ServerRam:  proto.String(this.ServerRam),
				SecretRam:  secretRam,
				HeartFeq:   proto.Uint32(this.HeartFeq),
			},
		},
	}

	if len(this.Sig) == 0 {
		m.Req.Login.VerfCode = &verfCode
	}

	fmt.Println("MakeLoginPacket:", m.String())

	return this.PackMessage(false, false, m)
}

func (this *XimpClient) UnPackMessage(key []byte, ximpBuff *network.XimpBuffer) (*pb.Message, error) {
	if err := ximpBuff.Decrypt(key); err != nil {
		return nil, err
	}
	m := &pb.Message{}
	if err := proto.Unmarshal(ximpBuff.DataStream, m); err != nil {
		return nil, errors.New("Unmarshal with err:" + err.Error())
	}
	return m, nil
}

func (this *XimpClient) DealInitLoginPacket() error {
	ximpBuff, err := this.MakeInitLoginPacket()
	if err != nil {
		return err
	}
	if err := ximpBuff.WriteTo(this.Conn, this.WriteTimeout); err != nil {
		return err
	}
	resp := &network.XimpBuffer{}
	if err := resp.ReadFrom(false, this.Conn, this.ReadTimeout); err != nil {
		return nil
	}
	if resp.IsHeartbeat || len(resp.DataStream) == 0 {
		return errors.New("expect a initlogin, but is heartbeat or datastream is empty")
	}
	m, err := this.UnPackMessage(this.DefaultKey, resp)
	if err != nil {
		return err
	}
	fmt.Println("Get init resp:", m.String())

	// 校验结果
	if pb.INIT_LOGIN_RESP != m.GetMsgid() {
		return errors.New("not a init_login_resp")
	}
	if m.GetSn() != this.CurrentSn {
		return errors.New("sn not match")
	}

	if m.Resp.InitLoginResp.GetClientRam() != this.ClientRam {
		return errors.New("client ram mismatches")
	}
	this.ServerRam = m.Resp.InitLoginResp.GetServerRam()
	if len(this.ServerRam) == 0 {
		return errors.New("server ram is nil")
	}
	return nil
}

func (this *XimpClient) MakeInitLoginPacket() (*network.XimpBuffer, error) {
	this.CurrentSn = logic.GetSn()
	this.ClientRam = logic.RandString(8)
	m := &pb.Message{
		Sender: proto.String(this.User),
		Sn:     proto.Uint64(this.CurrentSn),
		Msgid:  proto.Uint32(pb.INIT_LOGIN_REQ),
		Req: &pb.Request{
			InitLoginReq: &pb.InitLoginReq{
				ClientRam: proto.String(this.ClientRam),
			},
		},
	}

	if len(this.Sig) > 0 {
		m.Req.InitLoginReq.Sig = &this.Sig
	}
	fmt.Println("MakeInitLoginPacket:", m.String())

	return this.PackMessage(true, false, m)
}
func (this *XimpClient) PackMessage(hasHeader, isHearbeat bool, m *pb.Message) (*network.XimpBuffer, error) {
	var ds []byte
	if b, err := proto.Marshal(m); err != nil {
		return nil, err
	} else {
		ds = b
	}
	ximpBuff := &network.XimpBuffer{
		IsHeartbeat: isHearbeat,
		IsDecrypt:   true,
		IsClient:    true,
		HasHeader:   hasHeader,
		Version:     this.ProtocolVersion,
		CVersion:    this.ClientVersion,
		Appid:       this.Appid,
		DataStream:  ds,
	}
	if isHearbeat {
		return ximpBuff, nil
	}
	var key []byte
	if len(this.SessionKey) == 0 {
		key = this.DefaultKey
	} else {
		key = this.SessionKey
	}
	if err := ximpBuff.Encrypt(key); err != nil {
		return nil, err
	}
	return ximpBuff, nil
}
func showJson(j interface{}) {
	ds, err := json.Marshal(j)
	if err != nil {
		fmt.Println("json encode error:", err)
	}
	fmt.Println(string(ds))
}
