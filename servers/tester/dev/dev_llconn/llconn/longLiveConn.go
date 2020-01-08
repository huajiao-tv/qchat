package llconn

import (
	"bytes"
	"crypto/rc4"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"

	"sync/atomic"
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
	salt = logic.VERF_CODE_SALT

	chanCap        = 100
	connectTimeout = time.Duration(5) * time.Second
	readTimeout    = time.Duration(20) * time.Second
	writeTimeout   = time.Duration(20) * time.Second

	packetTimeout = time.Duration(20) * time.Second
)

type service struct {
	i         ServiceHandler
	serviceID int
}

type UserMessage struct {
	ID       uint64
	Content  []byte
	InfoType string
	MsgType  uint32
	Sn       uint64
	Valid    uint32
}

type ServiceHandler interface {
	HandleServiceMessage(data []byte, sn uint64, source int) bool
	HandleGetMessage(msg *UserMessage) bool
	Register(*LongLiveConn) (int, error)
}

type Stream interface {
	Close()
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type ClientConf struct {
	ConnType           string // Connection Type: "tcp", "websocket" or "wss"
	Addr               string // Server address
	AppID              int    // ApplicationID
	CV                 int    // ClientVersion
	PV                 int    // ProtocolVersion
	DK                 string // Default Key
	HeartItvl          int    // Heartbeat interval (unit second)
	SendHeartBeat      int    // Send heart beat or not. if 0, not send, else send.
	Autologin          int    // auto login after init login
	CheckLoginWhenJoin int    // check login when join chatroom
}

type AccountInfo struct {
	AccountType string // Account Type
	ID          string // User ID
	PWD         string // Corresponding password
	Sig         string // Signature to obtain token
	DID         string // Device ID
	Plf         string // Platform
	NE          bool   // Not Encrypt
}

// 代表一个消息盒子（槽）
type msgSlotInfo struct {
	infoType string // 消息盒子类型
	latestID int64  // 测试程序拿到的最新的ID
}

type packet struct {
	isPing bool        // 是否是心跳
	msg    *pb.Message // PB数据结构
	data   []byte      // 对应网络包数据， 包含处理过的magic code, length 和 Marshal/Unmarshal后的数据
}

type LongLiveConn struct {
	clientConf  ClientConf
	accountConf AccountInfo
	conn        Stream
	dumpStream  bool

	state int32
	key   string

	clientRan string
	serverRan string

	peer   *msgSlotInfo
	public *msgSlotInfo
	im     *msgSlotInfo

	sendChan chan *packet
	readChan chan *packet

	services map[int]ServiceHandler

	pendingMsg map[*packet]time.Time // used to store pending packet

	lastAction time.Time
}

var pingData = []byte{0, 0, 0, 0}

var idNameMap = map[int]string{
	pb.CHAT_REQ:           "ChatReq",
	pb.CHAT_RESP:          "ChatResp",
	pb.GET_INFO_REQ:       "GetInfoReq",
	pb.GET_INFO_RESP:      "GetInfoResp",
	pb.LOGIN_REQ:          "LoginReq",
	pb.LOGIN_RESP:         "LoginResp",
	pb.LOGOUT_REQ:         "LogoutReq",
	pb.LOGOUT_RESP:        "LogoutResp",
	pb.NEW_MESSAGE_NOTIFY: "NewMessageNotify",
	pb.RE_LOGIN_NOTIFY:    "ReLoginNotify",
	pb.RE_CONNECT_NOTIFY:  "ReConnectNotify",

	pb.INIT_LOGIN_REQ:  "InitLoginReq",
	pb.INIT_LOGIN_RESP: "InitLoginResp",

	pb.SERVICE_REQ:          "Service_Req",
	pb.SERVICE_RESP:         "Service_Resp",
	pb.GET_MULITI_INFOS_REQ: "GetMultipleInfoReq",
	pb.GET_MULTI_INFOS_RESP: "GetMultipleInfoResp"}

func (this *ClientConf) IsValid() bool {
	if nil == this {
		return false
	}

	if len(this.Addr) < 1 || len(this.DK) < 1 {
		return false
	}
	return true
}

func (this *AccountInfo) IsValid() bool {
	if nil == this {
		return false
	}

	if len(this.ID) < 1 || len(this.PWD) < 1 {
		return false
	}

	return true
}

func New(clientConf ClientConf, accountConf AccountInfo) (*LongLiveConn, *error) {
	conn := LongLiveConn{}
	if err := conn.init(clientConf, accountConf); err != nil {
		return &conn, &err
	}
	return &conn, nil
}

func (this *LongLiveConn) init(clientConf ClientConf, accountConf AccountInfo) (err error) {

	if nil == this {
		return errors.New("app internal error, pointer is null")
	}

	if !clientConf.IsValid() {
		return errors.New("client config is invalID")
	}

	if !accountConf.IsValid() {
		return errors.New("account info is invalID")
	}

	fmt.Println("userid", accountConf.ID, "password", accountConf.PWD, "signature", accountConf.Sig, "server addr", clientConf.Addr, "not encrypt", accountConf.NE)

	this.state = stateIdle
	this.conn = nil
	this.clientConf = clientConf
	this.accountConf = accountConf
	this.key = clientConf.DK
	this.dumpStream = true

	this.peer = &msgSlotInfo{infoType: "peer", latestID: 0}
	this.public = &msgSlotInfo{infoType: "public", latestID: 0}
	this.im = &msgSlotInfo{infoType: "im", latestID: 0}

	this.readChan = make(chan *packet, chanCap)
	this.sendChan = make(chan *packet, chanCap)

	this.pendingMsg = make(map[*packet]time.Time)

	this.services = make(map[int]ServiceHandler)

	return nil
}

func (this *LongLiveConn) connect() bool {

	fmt.Println("connecting to", this.clientConf.Addr)

	var conn Stream
	var err error

	if this.clientConf.ConnType == "websocket" {
		conn, err = NewWebSocketStream(this.clientConf.Addr, connectTimeout)
	} else if this.clientConf.ConnType == "wss" {
		conn, err = NewWebSocketSecurityStream(this.clientConf.Addr, connectTimeout)
	} else if this.clientConf.ConnType == "tcp" {
		conn, err = NewTcpStream(this.clientConf.Addr, connectTimeout)
	}

	if nil != err {
		fmt.Println("failed to connect to server", this.clientConf.Addr)
		return false
	} else {
		fmt.Println("socket connected")
	}

	this.conn = conn
	return true
}

func (this *LongLiveConn) makeMagicCode() []byte {
	if nil == this {
		return nil
	}

	magic := make([]byte, 12)
	magic[0] = 0x71
	magic[1] = 0x68

	magic[2] = (byte)(((this.clientConf.PV & 0xF) << 4) | ((this.clientConf.CV & 0xF00) >> 8))
	magic[3] = (byte)(this.clientConf.CV & 0xFF)

	magic[4] = (byte)((this.clientConf.AppID & 0xFF00) >> 8)
	magic[5] = (byte)(this.clientConf.AppID & 0xFF)

	return magic
}

func (this *LongLiveConn) makeSecretRan() []byte {
	return logic.MakeSecretRan(this.accountConf.PWD, this.serverRan)
}

func (this *LongLiveConn) makeVerfCode() string {
	return logic.MakeVerfCode(this.accountConf.ID)
}

func (this *LongLiveConn) queuePing() bool {
	if this.clientConf.SendHeartBeat == 0 {
		return true
	}

	packet := packet{true, nil, nil}
	this.sendChan <- &packet
	return true
}

func (this *LongLiveConn) ping() bool {

	this.conn.SetWriteDeadline(time.Now().Add(writeTimeout))

	if count, err := this.conn.Write(pingData); count != len(pingData) {
		fmt.Println("failed to send ping packet", err)
		return false
	}

	//fmt.Println("ping", time.Now())
	return true
}

func (this *LongLiveConn) sendMessage(p *packet) bool {
	if p == nil {
		fmt.Println("sendMessage packet is nil")
		return false
	}

	// add into pending map
	this.pendingMsg[p] = time.Now().Add(packetTimeout)

	if p.isPing {
		return this.ping()
	}

	if p.data == nil || len(p.data) == 0 {
		fmt.Println("sendMessage data is empty")
		return false
	}

	this.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if count, err := this.conn.Write(p.data); count < len(p.data) {
		fmt.Println("conn.Write failed", err)
		return false
	} else {
		id := int(*p.msg.Msgid)
		name := idNameMap[id]
		if this.dumpStream {
			fmt.Println(name, "sent sn", *p.msg.Sn, "data\n", p.data)
		} else {
			fmt.Println(name, "sent sn", *p.msg.Sn)
		}
	}

	return true
}

func (this *LongLiveConn) queueMessage(m *pb.Message) bool {

	ret := false

	for {
		if m == nil {
			break
		}

		pbData, err := proto.Marshal(m)
		if err != nil {
			fmt.Println("proto.Marshl failed", err)
			break
		}

		secret := []byte(this.key)
		if len(secret) != 0 {
			rc, err := rc4.NewCipher(secret)
			if err != nil {
				fmt.Println("NewCipher failed", err)
				break
			}

			rc.XORKeyStream(pbData, pbData)
		}

		totalLen := (uint32)(len(pbData) + 4)
		if this.getState() == stateRunning {
			totalLen += 12
		}

		buf := bytes.NewBuffer(nil)

		// cache magic code
		if this.getState() == stateRunning {
			count, err := buf.Write(this.makeMagicCode())
			if count != 12 || err != nil {
				fmt.Println("failed to cache magic code", err)
				break
			}
		}

		// cache length bytes
		lenBytes := make([]byte, 4)

		binary.BigEndian.PutUint32(lenBytes, totalLen)
		count, err := buf.Write(lenBytes)
		if err != nil || count != 4 {
			fmt.Println("failed to cache len ")
			break
		}

		// cache PB data
		if count, err = buf.Write(pbData); count != len(pbData) {
			fmt.Println("failed to cache pbdata")
			break
		}

		data := buf.Bytes()

		packet := packet{false, m, data}
		this.sendChan <- &packet

		ret = true
		break
	}

	return ret
}

/**
* receive packet from socket
 */
func (this *LongLiveConn) receivePacket() *packet {

	for {
		if this.conn == nil {
			break
		}

		var err error
		var count int
		var rc *rc4.Cipher
		buf := bytes.NewBuffer(nil)

		lenBytes := make([]byte, 4)
		count, err = this.conn.Read(lenBytes)
		if count != 4 {
			fmt.Println("failed to read data", err, ", data read", count)
			time.Sleep(5 * time.Second)
			break
		}

		if this.dumpStream {
			buf.Write(lenBytes)
		}

		if this.getState() == stateRunning && this.clientConf.PV == 1 {
			if lenBytes[0] != 0x71 || lenBytes[1] != 0x68 {
				fmt.Println("incorrect magic code", lenBytes[0], lenBytes[1])
				break
			}
			// read 2 more bytes
			lenBytes[0], lenBytes[1] = lenBytes[2], lenBytes[3]
			count, err = this.conn.Read(lenBytes[2:4])
			if count != 2 {
				fmt.Println("failed to read extra 2 byte", err)
				break
			}
			if this.dumpStream {
				buf.Write(lenBytes[2:4])
			}
		}

		len := int(binary.BigEndian.Uint32(lenBytes))
		if len == 0 {
			//fmt.Println("pong", time.Now())
			return &packet{isPing: true}
		} else if len < 0 || len > 100000000 {
			fmt.Println("len is abnormal ", lenBytes)
			return nil
		}

		len -= 4
		if this.getState() == stateRunning && this.clientConf.PV == 1 {
			len -= 2
		}

		pbData := make([]byte, len)
		count, err = this.conn.Read(pbData)
		if count != len {
			fmt.Println("failed to read pb data", err, ", data read", count)
			break
		}

		if this.dumpStream {
			buf.Write(pbData)
			//fmt.Println("data received:\n", buf.Bytes())
		}

		// decipher data
		var key = this.key
		if this.getState() == stateInitiated {
			key = this.accountConf.PWD
		}

		dest := make([]byte, len)
		if key != "" {
			rc, err = rc4.NewCipher([]byte(key))
			if err != nil {
				fmt.Println("NewCipher failed")
				break
			}

			rc.XORKeyStream(dest, pbData)
		}

		m := &pb.Message{}
		if key != "" {
			err = proto.Unmarshal(dest, m)
		} else {
			err = proto.Unmarshal(pbData, m)
		}
		// the case password is incorrect
		if err != nil && this.getState() == stateInitiated {
			rc, err = rc4.NewCipher([]byte(this.key))
			rc.XORKeyStream(pbData, pbData)
			err = proto.Unmarshal(pbData, m)
		}

		if err != nil {
			fmt.Print("failed to decode pb", err)
			break
		} else {
			//fmt.Println(m)

			//id := int(*m.Msgid)
			//name := idNameMap[id]

			//fmt.Println(name, "packet received sn", *m.Sn)

			return &packet{false, m, pbData}
		}

		break
	}

	return nil
}

func (this *LongLiveConn) getMessageBox(infoType string) *msgSlotInfo {
	if infoType == this.public.infoType {
		return this.public
	} else if infoType == this.peer.infoType {
		return this.peer
	} else if infoType == this.im.infoType {
		return this.im
	}
	return nil
}

func (this *LongLiveConn) shutdown() {

	if nil == this {
		return
	}

	if this.conn != nil {
		this.conn.Close()
		fmt.Println("socket closed by client.")
	} else {
		fmt.Println("connection is not exist")
	}
}

func Sn() uint64 {
	return logic.GetSn()
}

func randString(length int) string {
	return logic.RandString(length)
}

func (this *LongLiveConn) makeInitLogin() *pb.Message {

	var clientRan string = randString(8)
	var sn = Sn()
	this.clientRan = clientRan
	m := &pb.Message{
		Sender: &this.accountConf.ID,
		Sn:     &sn,
		Msgid:  proto.Uint32(pb.INIT_LOGIN_REQ),
		Req: &pb.Request{
			InitLoginReq: &pb.InitLoginReq{
				ClientRam: &clientRan,
			},
		},
	}

	if len(this.accountConf.Sig) > 0 {
		fmt.Println("sig", this.accountConf.Sig)
		m.Req.InitLoginReq.Sig = &this.accountConf.Sig
	}

	return m
}

func (this *LongLiveConn) makeLogin() *pb.Message {

	var sn = Sn()
	var netType uint32 = 3
	var verfCode = this.makeVerfCode()
	var secretRan []byte = this.makeSecretRan()

	m := &pb.Message{
		Sender:     proto.String(this.accountConf.ID),
		SenderType: proto.String(this.accountConf.AccountType),
		Sn:         &sn,
		Msgid:      proto.Uint32(pb.LOGIN_REQ),
		Req: &pb.Request{
			Login: &pb.LoginReq{
				AppId:      proto.Uint32(uint32(this.clientConf.AppID)),
				NetType:    &netType,
				MobileType: proto.String("dev"),
				ServerRam:  proto.String(this.serverRan),
				SecretRam:  secretRan,
				HeartFeq:   proto.Uint32(uint32(this.clientConf.HeartItvl)),
				Deviceid:   proto.String(this.accountConf.DID),
				Platform:   proto.String(this.accountConf.Plf),
				NotEncrypt: proto.Bool(this.accountConf.NE),
			},
		},
	}

	if len(this.accountConf.Sig) == 0 {
		m.Req.Login.VerfCode = &verfCode
	}

	return m
}

//
// Get message
//
func (this *LongLiveConn) getInfo(boxInfo *msgSlotInfo) bool {
	if boxInfo == nil {
		return false
	}
	return this.queueMessage(this.makeGetInfo(boxInfo))
}

//
//
// GetMultiInfo, so far it is used to get chatroom message
//
func (this *LongLiveConn) GetMultiInfo(infoType string, ids []int64, sparameter []byte) bool {
	if len(infoType) == 0 {
		return false
	}
	return this.queueMessage(this.makeGetMulInfo(infoType, ids, sparameter))
}

func (this *LongLiveConn) makeGetInfo(boxInfo *msgSlotInfo) *pb.Message {
	var sn = Sn()
	start := boxInfo.latestID + 1
	size := int32(binary.MaxVarintLen32)
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.GET_INFO_REQ),
		Req: &pb.Request{
			GetInfo: &pb.GetInfoReq{
				InfoType:      &boxInfo.infoType,
				GetInfoId:     &start,
				GetInfoOffset: &size,
			},
		},
	}
	return m
}

func (this *ChatroomService) Logout() {
	var sn = Sn()
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.LOGOUT_REQ),
		Req: &pb.Request{
			Logout: &pb.LogoutReq{
				Reason: proto.String("1"),
			},
		},
	}

	this.llc.queueMessage(m)
}

//
// Send Service Message
//
func (this *LongLiveConn) SendServiceMessage(serviceID int, data []byte) uint64 {

	if this.clientConf.CheckLoginWhenJoin != 0 && this.getState() != stateLoggedIn {
		fmt.Println("SendServiceMessage: LongLiveConn is not logged in yet, state:", this.getState())
		return 0
	}

	p := this.makeServiceMessage(serviceID, data)
	var sn uint64 = *p.Sn
	this.queueMessage(p)
	return sn
}

func (this *LongLiveConn) makeServiceMessage(serviceID int, data []byte) *pb.Message {
	var sn = Sn()
	srvID := uint32(serviceID)
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.SERVICE_REQ),
		Req: &pb.Request{
			ServiceReq: &pb.Service_Req{
				ServiceId: &srvID,
				Request:   data,
			},
		},
	}
	return m
}

func (this *LongLiveConn) makeGetMulInfo(infoType string, ids []int64, sparameter []byte) *pb.Message {
	var sn = Sn()
	m := &pb.Message{
		Sn:    &sn,
		Msgid: proto.Uint32(pb.GET_MULITI_INFOS_REQ),
		Req: &pb.Request{
			GetMultiInfos: &pb.GetMultiInfosReq{
				InfoType:   &infoType,
				GetInfoIds: ids,
				SParameter: sparameter,
			},
		},
	}

	return m
}

//
// Register a service on long live connection
// Corresponding data pakcet would be delivery to the service if matches
// TODO concurrent access unsafe
//
func (this *LongLiveConn) Register(service ServiceHandler) bool {

	if service == nil {
		return false
	}

	appid, err := service.Register(this)
	if err != nil {
		fmt.Println("service Register failed", err)
		return false
	}

	this.services[appid] = service
	return true
}

//
// enable the switch to print data stream
//
func (this *LongLiveConn) SetDumpStream(value bool) {
	this.dumpStream = value
}

//
// start running
//
//
func (this *LongLiveConn) Start(ch chan string) {
	if nil == this {
		fmt.Println("Start: this is nil")
		return
	}

	oldState := this.getState()
	if oldState == stateIdle {
		if !this.setState(oldState, stateRunning) {
			fmt.Println("arleady running 1")
			return
		}
	} else {
		fmt.Println("already running 2")
		return
	}

	go this.mainLoop(ch)
}

func (this *LongLiveConn) readerLoop() {

	for {
		p := this.receivePacket()

		this.readChan <- p

		if nil == p {
			break
		}
	}

	close(this.readChan)
	fmt.Println("reader goroutine exits")
}

func (this *LongLiveConn) mainLoop(ch chan string) {

	// connect socket
	start := time.Now()
	if this.connect() == false {
		return
	}
	fmt.Println("TC: conn socket cost", time.Now().Sub(start))

	defer this.shutdown()

	// queue initlogin message
	this.lastAction = time.Now()
	m := this.makeInitLogin()
	this.queueMessage(m)

	// start reading data
	go this.readerLoop()

	done := false

	heartbeatDur := time.Duration(this.clientConf.HeartItvl) * time.Second
	heartbeatTimer := time.NewTimer(heartbeatDur)
	packetTimer := time.NewTimer(this.getLeastTimeout())

	for !done {

		select {

		case p := <-this.sendChan:
			if this.sendMessage(p) == false {
				done = true
				break
			} else {
				heartbeatTimer.Reset(heartbeatDur)
			}

		case p := <-this.readChan:
			if this.handlePacket(p) == false {
				done = true
				break
			} else {
				packetTimer.Reset(this.getLeastTimeout())
			}
		case <-heartbeatTimer.C:
			if this.queuePing() == false {
				done = true
				break
			}
		case <-packetTimer.C:
			p := this.getLeastTimeoutPacket()
			if p.isPing {
				fmt.Println("pong is not received in time")
			} else {
				fmt.Println("response of packet", *p.msg.Sn, "is not received in time")
			}
			done = true
			break
		}
	}

	if ch != nil {
		time.Sleep(1 * time.Second)
		ch <- "exit"
	}
}

func (this *LongLiveConn) handlePacket(p *packet) bool {

	if p == nil {
		return false
	}

	// remove message from pending queue
	this.removePendingMsg(p)

	if p.isPing {
		return true
	}

	m := p.msg
	sn := p.msg.GetSn()
	resp := p.msg.GetResp()

	if resp != nil && resp.Error != nil {
		err := *resp.Error
		if pb.ERR_USER_INVALID == *err.Id {
			fmt.Println("ERR_USER_INVALID")
			return false
		}

		fmt.Println("err", *err.Id, string(err.Description))
		return false
	}

	switch *m.Msgid {

	case uint32(pb.INIT_LOGIN_RESP):

		fmt.Println("TC: init login resp", time.Now().Sub(this.lastAction))
		clientRan := m.Resp.InitLoginResp.ClientRam

		if *clientRan != this.clientRan {
			fmt.Println("client ran mismatches")
			return false
		}

		this.serverRan = *(m.Resp.InitLoginResp.ServerRam)

		this.setState(stateRunning, stateInitiated)

		if this.clientConf.Autologin != 0 {
			fmt.Println("will auto login after init login")
			if ret := this.queueMessage(this.makeLogin()); !ret {
				fmt.Println("failed to send login packet")
				return false
			}
			this.lastAction = time.Now()
		} else {
			fmt.Println("will not auto login after init login")
		}

	case uint32(pb.LOGIN_RESP):

		this.setState(stateInitiated, stateLoggedIn)
		this.key = *(m.Resp.Login.SessionKey)
		fmt.Println("TC: login resp", time.Now().Sub(this.lastAction))

		this.getInfo(this.peer)
		this.getInfo(this.public)
		this.getInfo(this.im)
		this.lastAction = time.Now()

	case uint32(pb.GET_INFO_RESP):

		getInfo := *m.Resp.GetInfo
		count := len(getInfo.Infos)
		infoType := *getInfo.InfoType

		fmt.Println("GetInfoResp", infoType, "received, count", count, "latest id", *getInfo.LastInfoId)

		for _, info := range getInfo.Infos {

			var msg UserMessage
			for _, pair := range info.PropertyPairs {
				var key string = string(pair.Key)
				if key == "info_id" {
					msg.ID = binary.BigEndian.Uint64(pair.Value)
				} else if key == "chat_body" {
					msg.Content = pair.Value
				}
			}

			DefaultMessageMeter.Check(string(msg.Content))

			fmt.Println("TC: get info resp", infoType, time.Now().Sub(this.lastAction))
		}

		msgBox := this.getMessageBox(infoType)
		if msgBox != nil {
			msgBox.latestID = *getInfo.LastInfoId
		}

	case uint32(pb.GET_MULTI_INFOS_RESP):

		getInfo := *m.Resp.GetMultiInfos
		count := len(getInfo.Infos)
		infoType := *getInfo.InfoType

		fmt.Println("GetMultiInfoResp", infoType, "received, count", count, "latest id", *getInfo.LastInfoId)

		for _, info := range getInfo.Infos {

			msg := UserMessage{
				InfoType: infoType,
				Sn:       m.GetSn(),
				Valid:    1,
			}

			for _, pair := range info.PropertyPairs {

				var key string = string(pair.Key)
				if key == "info_id" {
					msg.ID = binary.BigEndian.Uint64(pair.Value)
				} else if key == "chat_body" {
					msg.Content = pair.Value
				} else if key == "msg_type" {
					msg.MsgType = binary.BigEndian.Uint32(pair.Value)
				} else if key == "msg_valid" {
					msg.Valid = binary.BigEndian.Uint32(pair.Value)
				}
			}

			service, ok := this.services[ChatroomServiceID]
			if ok {
				service.HandleGetMessage(&msg)
			}
		}

	case uint32(pb.NEW_MESSAGE_NOTIFY):
		ntf := m.Notify.NewinfoNtf
		infoType := *ntf.InfoType
		//fmt.Println("NewMessageNotify received", infoType)

		if infoType == "chatroom" {
			service, ok := this.services[ChatroomServiceID]
			if ok {
				service.HandleServiceMessage(ntf.InfoContent, sn, 0)
			}
		} else if infoType == "group" {
			service, ok := this.services[GroupServiceID]
			if ok {
				service.HandleServiceMessage(ntf.InfoContent, sn, 0)
			}
		} else {
			msgBox := this.getMessageBox(infoType)
			if msgBox != nil {
				this.getInfo(msgBox)
			}
		}

	case uint32(pb.SERVICE_RESP):
		if *resp.ServiceResp.ServiceId == ChatroomServiceID {
			service, ok := this.services[ChatroomServiceID]
			if ok {
				service.HandleServiceMessage(resp.ServiceResp.Response, sn, 1)
			}
		} else if *resp.ServiceResp.ServiceId == GroupServiceID {
			service, ok := this.services[GroupServiceID]
			if ok {
				service.HandleServiceMessage(resp.ServiceResp.Response, sn, 1)
			}
		}

	default:
		fmt.Println("handlePacket msgid: ", *m.Msgid, " sn: ", *m.Sn)
	}

	return true
}

func (this *LongLiveConn) getLeastTimeout() time.Duration {

	now := time.Now()
	var least time.Duration = time.Duration(360) * time.Hour
	// pendingMsg  map[uint64]time.Time // used to store pending packet
	for _, t := range this.pendingMsg {
		diff := t.Sub(now)
		if diff < least {
			least = diff
		}
	}

	if least < 0 {
		least = time.Duration(1) * time.Nanosecond
	}

	return least
}

func (this *LongLiveConn) getLeastTimeoutPacket() *packet {
	now := time.Now()
	var p *packet = nil
	var least time.Duration = time.Duration(360) * time.Hour
	// pendingMsg  map[uint64]time.Time // used to store pending packet
	for pkt, t := range this.pendingMsg {
		diff := t.Sub(now)
		if diff < least {
			least = diff
			p = pkt
		}
	}
	return p
}

func (this *LongLiveConn) removePendingMsg(p *packet) {

	if p == nil {
		return
	}

	delete(this.pendingMsg, p)

	for pkt, _ := range this.pendingMsg {
		if p.isPing && pkt.isPing {
			delete(this.pendingMsg, pkt)
		} else if pkt.msg != nil && p.msg != nil {
			if pkt.msg.GetSn() == p.msg.GetSn() {
				delete(this.pendingMsg, pkt)
			}
		}
	}
}

func (this *LongLiveConn) getState() int32 {
	return atomic.LoadInt32(&this.state)
}

func (this *LongLiveConn) setState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&this.state, old, new)
}
