package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/huajiao-tv/qchat/logic"
	"github.com/huajiao-tv/qchat/logic/pb"
	"github.com/huajiao-tv/qchat/utility/network"
)

var (
	Total, Conn         int64
	Ping, Pong          int64
	Msg, MsgUnknow      int64
	SReed, SWrite       int64 // 服务端读写QPS
	EachRead, TotalRead int64 // 每一个时间段读的数据，总共读取的数据

	DefaultKey       []byte        = []byte("894184791415baf5c113f83eaff360f0")
	Appid            uint16        = 1080
	ConnectTimeout   time.Duration = 2 * time.Second
	WriteTimeout     time.Duration = 2 * time.Second
	ReadTimeout      time.Duration = 2 * time.Second
	HeartBeatTimeout time.Duration = 60 * time.Second

	Count   int
	Gateway string
	Room    string
	RoomGen int // 房间号是生成，而非传进去的
	Admin   string
	Detail  bool
	Begin   int  // uid的超始id
	Web     bool //是否是websocket

	DetailMap     map[string]int64
	DetailMapLock sync.Mutex
)

func Info(args ...interface{}) {
	fmt.Print("INFO:")
	fmt.Println(args...)
	fmt.Print("\n")
}
func Succ(args ...interface{}) {
	fmt.Print("Succ:")
	fmt.Println(args...)
	fmt.Print("\n")
}
func Fail(args ...interface{}) {
	fmt.Print("Fail:")
	fmt.Println(args...)
	fmt.Print("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF\n\n")
}
func ExecNTimes(n int, f func()) {
	sleepTime := time.Second / time.Duration(n)
	next := time.Now()
	for {
		next = next.Add(sleepTime)
		go f()
		left := next.Sub(time.Now())
		if left > 0 {
			time.Sleep(left)
		}
	}
}

var lastMsg int64

func ShowInfo() {
	eachRead := EachRead
	atomic.StoreInt64(&EachRead, 0)
	TotalRead += eachRead
	if Detail {
		fmt.Printf("Conn:%d/%d\tQps:%d/%d\tMsg/s:%d\t\tEachRead/ReadTotal:%d/%d MB\t",
			atomic.LoadInt64(&Conn),
			atomic.LoadInt64(&Total),
			atomic.LoadInt64(&SWrite),
			atomic.LoadInt64(&SReed),
			atomic.LoadInt64(&Msg)-lastMsg,
			atomic.LoadInt64(&eachRead)/1024/1024,
			atomic.LoadInt64(&TotalRead)/1024/1024,
		)
		DetailMapLock.Lock()
		ss := make([]string, 0, len(DetailMap))
		var DetailTotal int64
		for k, v := range DetailMap {
			ss = append(ss, fmt.Sprintf("%s:%d", k, v))
			DetailTotal += v
		}
		fmt.Printf("detail:%d\t", DetailTotal)
		sort.Strings(ss)
		for _, v := range ss {
			fmt.Print(v + "\t")

		}
		DetailMap = make(map[string]int64)
		fmt.Printf("\n")
		DetailMapLock.Unlock()
	} else {
		fmt.Printf("Conn:%d/%d\tPing:%d/%d\tMsg:%d/%d\tMsg/s:%d\tQps:%d/%d\tEachRead/ReadTotal:%d/%d MB\n",
			atomic.LoadInt64(&Conn),
			atomic.LoadInt64(&Total),
			atomic.LoadInt64(&Ping),
			atomic.LoadInt64(&Pong),
			atomic.LoadInt64(&Msg),
			atomic.LoadInt64(&MsgUnknow),
			atomic.LoadInt64(&Msg)-lastMsg,
			atomic.LoadInt64(&SWrite),
			atomic.LoadInt64(&SReed),
			atomic.LoadInt64(&eachRead)/1024/1024,
			atomic.LoadInt64(&TotalRead)/1024/1024,
		)
	}
	lastMsg = atomic.LoadInt64(&Msg)
}

func init() {
	flag.StringVar(&Room, "r", "", "room id")
	flag.BoolVar(&Detail, "d", false, "是否打印详细消息条数")
	flag.BoolVar(&Web, "w", false, "是否是websocket")
	flag.StringVar(&Gateway, "g", "127.0.0.1:8080", "gateway addr")
	flag.StringVar(&Admin, "a", "", "gateway admin addr")
	flag.IntVar(&Count, "c", 10, "client num")
	flag.IntVar(&Begin, "b", 0, "begin uid")
	flag.IntVar(&RoomGen, "R", 0, "生成房间数量(覆盖r)，如果小于0，用户随机加入这些房间，其它的话均匀分布, 房间号范围: [0~R)")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	DetailMap = make(map[string]int64)

}

func genRooms() []string {
	rooms := []string{}
	if RoomGen != 0 {
		c := RoomGen
		if c < 0 {
			c = -c
		}
		for i := 0; i < c; i++ {
			rooms = append(rooms, strconv.Itoa(i))
		}
	} else {
		rooms = strings.Split(Room, ",")
		if len(rooms) == 0 {
			rooms = []string{""}
		}
	}
	return rooms
}

func ReadQps() {
	response, err := http.Get("http://" + Admin + "/monitor/data")
	if err != nil {
		Fail("ReadQps fail:", err)
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		Fail("ioutil.ReadAll error", err)
		return
	}
	var f interface{}
	if err := json.Unmarshal(body, &f); err != nil {
		Fail("json.Unmarshal error", err)
	}
	if step1, ok := f.(map[string]interface{}); ok {
		if step2, ok := step1["data"].(map[string]interface{}); ok {
			if r, ok := step2["reading"].(float64); ok {
				atomic.StoreInt64(&SReed, int64(r))
			} else {
				Fail("reading not a float64")
			}
			if w, ok := step2["writing"].(float64); ok {
				atomic.StoreInt64(&SWrite, int64(w))
			} else {
				Fail("writint not a float64")
			}
			return
		}
	}
	Fail("readQps format error")

}

func main() {
	go ExecNTimes(1, ReadQps)
	go ExecNTimes(1, ShowInfo)
	now := time.Now().Unix()

	if Admin == "" {
		Admin = strings.Split(Gateway, ":")[0] + ":16200"
	}
	rooms := genRooms()
	for i := 0; i < Count; i++ {
		if RoomGen < 0 {
			time.Sleep(time.Millisecond * 5)
			if Begin == 0 {
				go RaiseClient(strconv.Itoa(int(now)*100000+i), rooms[rand.Int()%len(rooms)])
			} else {
				go RaiseClient(strconv.Itoa(Begin+i), rooms[rand.Int()%len(rooms)])
			}
		} else {
			time.Sleep(time.Millisecond * 5)
			if Begin == 0 {
				go RaiseClient(strconv.Itoa(int(now)*100000+i), rooms[i%len(rooms)])
			} else {
				go RaiseClient(strconv.Itoa(Begin+i), rooms[i%len(rooms)])
			}
		}
	}
	select {}
}

func RaiseWeb(uid string, room string) {
	atomic.AddInt64(&Total, 1)
	u := url.URL{Scheme: "ws", Host: Gateway, Path: "/"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	conn := &network.WebSocketConnection{
		Conn: c,
	}

	initXimp, err := makeInitLogin(uid)
	if err != nil {
		Fail("makeInitLogin error:", err)
		return
	}
	bs, _ := initXimp.Encode()
	if _, err := conn.WriteBytes(bs, WriteTimeout); err != nil {
		Fail("write initXimp error:", err)
		return
	}

	initResp := network.NewXimpBuffer()
	if err := initResp.ReadFrom(false, conn, ReadTimeout); err != nil {
		Fail("initResp readfrom error:", err)
		return
	}
	serverRam, err := unpackInitResp(initResp)
	if err != nil {
		Fail("unpackInitResp error:", err)
		return
	}
	loginXimp, err := makeLogin(uid, serverRam)
	if err != nil {
		Fail("makeLogin error:", err)
		return
	}
	bs, _ = loginXimp.Encode()
	if _, err := conn.WriteBytes(bs, WriteTimeout); err != nil {
		Fail("write loginXimp error:", err)
		return
	}
	loginResp := network.NewXimpBuffer()
	if err := loginResp.ReadFrom(false, conn, ReadTimeout); err != nil {
		Fail("loginResp readfrom error:", err)
		return
	}

	key, err := unpackLoginResp(uid, loginResp)
	if err != nil {
		Fail("unpackLoginResp error:", err)
		return
	}

	if room != "" {
		joinXimp, err := makeJoin([]byte(key), room)
		if err != nil {
			Fail("makeJoin error:", err)
			return
		}
		bs, _ = joinXimp.Encode()
		if _, err := conn.WriteBytes(bs, WriteTimeout); err != nil {
			Fail("write joinXimp  error:", err)
			return
		}
		joinResp := network.NewXimpBuffer()
		if err := joinResp.ReadFrom(false, conn, ReadTimeout); err != nil {
			Fail("joinResp readfrom error:", err)
			return
		}
		if err := unpackJoinResp([]byte(key), joinResp); err != nil {
			Fail("unpackJoinResp error:", err)
			return
		}
	}

	atomic.AddInt64(&Conn, 1)
	go func() {
		t := time.NewTicker(HeartBeatTimeout)
		for {
			select {
			case <-t.C:
				bs, _ := makePing().Encode()
				if _, err := conn.WriteBytes(bs, WriteTimeout); err != nil {
					Fail("send ping error:", err)
					return
				}
				atomic.AddInt64(&Ping, 1)
			}
		}
	}()
	go func() {
		defer func() {
			conn.Close()
			atomic.AddInt64(&Conn, -1)
		}()
		for {
			msg := network.NewXimpBuffer()
			if err := msg.ReadFrom(false, conn, HeartBeatTimeout*2); err != nil {
				Fail("ReadFrom error:", err)
				return
			}
			unpackMsg([]byte(key), msg)
		}
	}()
}

func RaiseClient(uid string, room string) {
	if Web {
		RaiseWeb(uid, room)
	} else {
		RaiseMobile(uid, room)
	}
}

func RaiseMobile(uid string, room string) {
	atomic.AddInt64(&Total, 1)
	conn, err := network.TcpConnect("", Gateway, ConnectTimeout)
	if err != nil {
		Fail("Ximp:Connect to Gateway error:", err)
		return
	}
	initXimp, err := makeInitLogin(uid)
	if err != nil {
		Fail("makeInitLogin error:", err)
		return
	}
	if err := initXimp.WriteTo(conn, WriteTimeout); err != nil {
		Fail("write initXimp error:", err)
		return
	}

	initResp := network.NewXimpBuffer()
	if err := initResp.ReadFrom(false, conn, ReadTimeout); err != nil {
		Fail("initResp readfrom error:", err)
		return
	}

	serverRam, err := unpackInitResp(initResp)
	if err != nil {
		Fail("unpackInitResp error:", err)
		return
	}
	loginXimp, err := makeLogin(uid, serverRam)
	if err != nil {
		Fail("makeLogin error:", err)
		return
	}
	if err := loginXimp.WriteTo(conn, WriteTimeout); err != nil {
		Fail("write loginXimp error:", err)
		return
	}
	loginResp := network.NewXimpBuffer()
	if err := loginResp.ReadFrom(false, conn, ReadTimeout); err != nil {
		Fail("loginResp readfrom error:", err)
		return
	}

	key, err := unpackLoginResp(uid, loginResp)
	if err != nil {
		Fail("unpackLoginResp error:", err)
		return
	}

	if room != "" {
		joinXimp, err := makeJoin([]byte(key), room)
		if err != nil {
			Fail("makeJoin error:", err)
			return
		}
		if err := joinXimp.WriteTo(conn, WriteTimeout); err != nil {
			Fail("write joinXimp  error:", err)
			return
		}
		joinResp := network.NewXimpBuffer()
		if err := joinResp.ReadFrom(false, conn, ReadTimeout); err != nil {
			Fail("joinResp readfrom error:", err)
			return
		}
		if err := unpackJoinResp([]byte(key), joinResp); err != nil {
			Fail("unpackJoinResp error:", err)
			return
		}
	}

	atomic.AddInt64(&Conn, 1)
	go func() {
		t := time.NewTicker(HeartBeatTimeout)
		for {
			select {
			case <-t.C:
				if err := makePing().WriteTo(conn, WriteTimeout); err != nil {
					Fail("send ping error:", err)
					return
				}
				atomic.AddInt64(&Ping, 1)
			}
		}
	}()
	go func() {
		defer func() {
			conn.Close()
			atomic.AddInt64(&Conn, -1)
		}()
		for {
			msg := network.NewXimpBuffer()
			if err := msg.ReadFrom(false, conn, HeartBeatTimeout*2); err != nil {
				Fail("ReadFrom error:", err)
				return
			}
			unpackMsg([]byte(key), msg)
		}
	}()
}

func unpackJoinResp(key []byte, joinResp *network.XimpBuffer) error {
	m, err := unpackService(key, joinResp)
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

func unpackService(key []byte, resp *network.XimpBuffer) (*pb.Message, error) {
	if err := resp.Decrypt(key); err != nil {
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

func unpackMsg(key []byte, msg *network.XimpBuffer) {
	atomic.AddInt64(&EachRead, int64(len(msg.DataStream)+4))
	if msg.IsHeartbeat {
		atomic.AddInt64(&Pong, 1)
		return
	}
	if len(msg.DataStream) <= 0 {
		atomic.AddInt64(&MsgUnknow, 1)
		return
	}
	if err := msg.Decrypt(key); err != nil {
		atomic.AddInt64(&MsgUnknow, 1)
		return
	}

	m := &pb.Message{}
	if err := proto.Unmarshal(msg.DataStream, m); err != nil {
		atomic.AddInt64(&MsgUnknow, 1)
		return
	}
	if *m.Msgid != uint32(pb.NEW_MESSAGE_NOTIFY) {
		atomic.AddInt64(&MsgUnknow, 1)
		return
	}
	atomic.AddInt64(&Msg, 1)
	if !Detail {
		return
	}

	if *m.Notify.NewinfoNtf.InfoType == "chatroom" {
		var p pb.ChatRoomPacket
		if err := proto.Unmarshal(m.Notify.NewinfoNtf.InfoContent, &p); err != nil {
			Fail("unmarshal chat room notify error", err)
			return
		}

		DetailMapLock.Lock()
		switch p.ToUserData.GetPayloadtype() {
		case pb.CR_PAYLOAD_INCOMING_MSG:
			var j PayloadJson
			if err := json.Unmarshal(p.ToUserData.Newmsgnotify.Msgcontent, &j); err != nil {
				Fail("palyload is not json")
			} else {
				DetailMap[j.ToKey()] += 1
			}
		case pb.CR_PAYLOAD_COMPRESSED:
			for _, n := range p.ToUserData.Multinotify {
				if *n.Type != pb.CR_PAYLOAD_INCOMING_MSG {
					continue
				}
				data := ungzip(n.Data)
				// 解压
				packet := pb.ChatRoomNewMsg{}
				if err := proto.Unmarshal(data, &packet); err != nil {
					Fail("unmarshal new msg error", err)
					continue
				}
				var j PayloadJson
				if err := json.Unmarshal(packet.Msgcontent, &j); err != nil {
					Fail("palyload is not json")
				} else {
					DetailMap[j.ToKey()] += 1
				}

			}
		}
		DetailMapLock.Unlock()
	}
}

type PayloadJson struct {
	Type     int `json:type`
	Priority int `json:priority`
}

func (pj *PayloadJson) ToKey() string {
	return strconv.Itoa(pj.Type) + "-" + strconv.Itoa(pj.Priority)
}

func makePing() *network.XimpBuffer {
	return &network.XimpBuffer{
		IsHeartbeat: true,
	}
}

func unpackLoginResp(uid string, loginResp *network.XimpBuffer) (string, error) {
	if err := loginResp.Decrypt([]byte(uid)); err != nil {
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

func makeJoin(key []byte, roomid string) (*network.XimpBuffer, error) {
	rid := []byte(roomid)
	payload := uint32(pb.CR_PAYLOAD_JOIN)
	appid := uint32(Appid)
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
	return makeService(key, pb.CHATROOM_SERVICE_ID, data)
}

func makeService(key []byte, serviceId uint32, data []byte) (*network.XimpBuffer, error) {
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
		Appid:      Appid,
		DataStream: pbData,
	}
	if err := buf.Encrypt(key); err != nil {
		return nil, err
	}
	return buf, nil
}

func makeLogin(uid, serverRam string) (*network.XimpBuffer, error) {
	sn := logic.GetSn()
	var netType uint32 = 3
	var verfCode = logic.MakeVerfCode(uid)
	var secretRan []byte = logic.MakeSecretRan(uid, serverRam)
	m := &pb.Message{
		Sender:     proto.String(uid),
		SenderType: proto.String("jid"),
		Sn:         &sn,
		Msgid:      proto.Uint32(pb.LOGIN_REQ),
		Req: &pb.Request{
			Login: &pb.LoginReq{
				VerfCode:   &verfCode,
				AppId:      proto.Uint32(uint32(Appid)),
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
		Appid:      Appid,
		DataStream: pbData,
	}
	if err := buf.Encrypt(DefaultKey); err != nil {
		return nil, err
	}

	return buf, err
}

func unpackInitResp(initResp *network.XimpBuffer) (string, error) {
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

func makeInitLogin(uid string) (*network.XimpBuffer, error) {
	clientRan := logic.RandString(8)
	sn := logic.GetSn()
	m := &pb.Message{
		Sender: &uid,
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
		Appid:      Appid,
		DataStream: pbData,
	}
	if err := buf.Encrypt(DefaultKey); err != nil {
		return nil, err
	}

	return buf, err
}

func ungzip(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return nil
	}

	var done bool = true
	bufLen := 512
	buf := make([]byte, bufLen)
	writer := bytes.NewBuffer(nil)
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		fmt.Println("gzip.NewReader failed", err)
		return nil
	}

	for {
		count, err1 := reader.Read(buf)
		if count == bufLen {
			writer.Write(buf)
		} else if count > 0 {
			writer.Write(buf[0:count])
			break
		} else {
			fmt.Println("ungzip failed", err1)
			done = false
			break
		}
	}

	reader.Close()

	if done {
		return writer.Bytes()
	}

	return nil
}
