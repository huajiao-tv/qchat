package main

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	//"github.com/huajiao-tv/qchat/servers/tester/ben/ben_protobuf/proto"
	"encoding/binary"
	"flag"
	"runtime"
	"sync"
	"time"

	"os"
	"runtime/pprof"

	"github.com/huajiao-tv/qchat/logic/pb"
)

// 包全局变量
var (
	testcase string
	protosn  string
	gonum    int
	num      int
	stage    int

	wg *sync.WaitGroup
)

const (
	MSGIDS     = 10
	MSGBODYLEN = 1024
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&testcase, "tc", "pack", "testcase:pack or unpack")
	flag.StringVar(&protosn, "proto", "getinforesp", "protosn:getinforesp")

	flag.IntVar(&gonum, "gonum", 1, "go routine num")
	flag.IntVar(&num, "num", 1, "test num")
	flag.IntVar(&stage, "stage", 1, "test stage")

	flag.Parse()
}

func getProto() (m *pb.Message) {

	switch protosn {
	case "getinforeq":
		ids := make([]int64, MSGIDS)
		for i := 0; i < MSGIDS; i++ {
			ids = append(ids, int64(1000000+i))
		}

		m = &pb.Message{
			Sn:    proto.Uint64(65535),
			Msgid: proto.Uint32(pb.GET_MULITI_INFOS_REQ),
			Req: &pb.Request{
				GetMultiInfos: &pb.GetMultiInfosReq{
					InfoType:   proto.String("chatroom"),
					SParameter: []byte("1000006"),
					GetInfoIds: ids,
				},
			},
		}
	case "getinforesp":
		infos := make([]*pb.Info, 0, 10)
		buf := make([]byte, MSGIDS*20)

		for i := 0; i < MSGIDS; i++ {
			valid, infoID, timestamp := buf[i*20:i*20+4], buf[i*20+4:i*20+12], buf[i*20+12:i*20+20]
			data := make([]byte, MSGBODYLEN)

			binary.BigEndian.PutUint32(valid, 1)
			binary.BigEndian.PutUint64(timestamp, 1)
			infos = append(infos, &pb.Info{
				PropertyPairs: []*pb.Pair{
					&pb.Pair{
						Key:   []byte("msg_valid"),
						Value: valid,
					},
					&pb.Pair{
						Key:   []byte("info_id"),
						Value: infoID,
					},
					&pb.Pair{
						Key:   []byte("chat_body"),
						Value: data,
					},
					&pb.Pair{
						Key:   []byte("time_sent"),
						Value: timestamp,
					},
				},
			})
		}

		m = &pb.Message{
			Sn:    proto.Uint64(65535),
			Msgid: proto.Uint32(pb.GET_MULTI_INFOS_RESP),
			Resp: &pb.Response{
				GetMultiInfos: &pb.GetMultiInfosResp{
					InfoType:   proto.String("chatroom"),
					Infos:      infos,
					LastInfoId: proto.Int64(12345),
					SParameter: []byte("1000006"),
				},
			},
		}
	}

	return
}

func pack() {

	for i := 0; i < gonum; i++ {
		go func() {
			m := getProto()
			defer wg.Done()
			for i := 0; i < num; i++ {
				_, err := proto.Marshal(m)
				if err != nil {
					fmt.Println("proto.Marshl failed", err)
					break
				}
			}
		}()
	}
}

func unpack() {
	m := getProto()
	packed, err := proto.Marshal(m)
	if err != nil {
		fmt.Println("proto.Marshl failed", err)
		return
	}

	for i := 0; i < gonum; i++ {
		go func() {
			defer wg.Done()

			for i := 0; i < num; i++ {
				unpacked := &pb.Message{} // 返回给客户端的pb
				if err := proto.Unmarshal(packed, unpacked); err != nil {
					fmt.Println("proto.Unmarshl failed", err)
					break
				}
			}
		}()
	}
}

func sample() {
	defer wg.Done()
	m := getProto()
	packed, err := proto.Marshal(m)
	if err != nil {
		fmt.Println("proto.Marshl failed", err)
	} else {
		fmt.Println("proto.Marshl len=", len(packed))
	}
}

func test() {
	wg = new(sync.WaitGroup)
	wg.Add(gonum)

	begin := time.Now()

	fmt.Println("testcase = ", testcase)

	switch testcase {
	case "pack":
		pack()
	case "unpack":
		unpack()
	case "sample":
		sample()
	}

	wg.Wait()

	end := time.Now()

	duration := end.Sub(begin)
	sec := duration.Seconds()
	fmt.Println(float64(gonum*num)/sec, sec)
}

func main() {
	cpufile, _ := os.Create("cpu.out")
	defer cpufile.Close()

	pprof.StartCPUProfile(cpufile)

	for i := 0; i < stage; i++ {
		test()
	}

	pprof.StopCPUProfile()
}
