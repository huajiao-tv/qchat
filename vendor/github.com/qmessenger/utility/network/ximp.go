package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/huajiao-tv/qchat/utility/cryption"
)

const (
	MAX_UPLOAD_SIZE = 1024 * 1024
)

// 客户端与服务端的第一次回包
type XimpBuffer struct {
	IsHeartbeat bool    // 是否是心跳包
	IsDecrypt   bool    // 是否已经解密，false时是加密，
	IsClient    bool    // 是否是客户端发上来的包
	HasHeader   bool    // 是否有magic或者flag
	Version     uint16  // 协议版本, 4个bit
	CVersion    uint16  // 客户端版本，12个bit
	Appid       uint16  // 2字节
	Reserved    [6]byte // 6个字节的保留字段
	DataStream  []byte  // proto数据
	// 业务相关
	TimeStamp int64  // 用于chatroom发送消息时的延迟统计
	TraceId   string // 客户端发上来的跟踪ID
}

func NewXimpBuffer() *XimpBuffer {
	return &XimpBuffer{}
}

func (this *XimpBuffer) String() string {
	result := ""
	if this.IsHeartbeat {
		result += "Hb,"
	}
	if this.IsDecrypt {
		result += "Decry,"
	}
	if this.IsClient {
		result += "Cl,"
	}
	if this.HasHeader {
		result += "Header,"
	}
	if this.Version != 0 {
		result += fmt.Sprintf("V:%d-%d-%d,", this.Version, this.CVersion, this.Appid)
	}
	if len(this.DataStream) != 0 {
		result += "DL:" + strconv.Itoa(len(this.DataStream))
	}
	return result
	//	return fmt.Sprintf("Hb:%t,crypt:%t,IsC:%t,Header:%t,V:%d, CV:%d, A:%d, R:%v, DSLen:%d", this.IsHeartbeat, this.IsDecrypt, this.IsClient, this.HasHeader, this.Version, this.CVersion, this.Appid, this.Reserved, len(this.DataStream))
}

// 读取协议的所有内容
// isClient 表示要读取的包是否是客户端传过来的包
func (this *XimpBuffer) ReadFrom(isClient bool, conn INetwork, timeout time.Duration) error {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	this.IsClient = isClient
	buff := make([]byte, 14)
	if err := conn.ReadStream(buff[:2], 2); err != nil {
		return err
	}
	var dataLen uint32
	var headerLen uint32 = 4 // 头部和size的长度
	if bytes.Compare(buff[:2], []byte{'q', 'h'}) == 0 {
		this.HasHeader = true
		headerLen += 2 // 两个magic
		if this.IsClient {
			headerLen += 10     // version和app等信息长度
			var versions uint16 // version和cversion
			if err := binary.Read(conn, binary.BigEndian, &versions); err != nil {
				return err
			}
			this.Version = versions >> 12
			this.CVersion = versions & 0x0FFF
			if err := binary.Read(conn, binary.BigEndian, &this.Appid); err != nil {
				return err
			}
			if err := conn.ReadStream(this.Reserved[0:6], 6); err != nil {
				return err
			}
		}
		if err := binary.Read(conn, binary.BigEndian, &dataLen); err != nil {
			return err
		}
	} else if bytes.Compare(buff[:2], []byte{'w', 'x'}) == 0 {
		if err := conn.ReadStream(buff[2:10], 8); err != nil {
			return err
		}
		lBuf := buff[2:10]
		if bytes.Compare(buff[2:6], []byte{'h', '1', '2', '8'}) == 0 {
			this.HasHeader = true
			this.CVersion = 102
			this.Appid = 2080
			if err := conn.ReadStream(buff[10:14], 4); err != nil {
				return err
			}
			lBuf = buff[6:14]
		}
		length, err := strconv.Atoi(string(lBuf))
		if err != nil {
			return err
		}
		return this.readFromString(conn, length)
	} else {
		if err := conn.ReadStream(buff[2:4], 2); err != nil {
			return err
		}
		// 心跳包
		if bytes.Compare(buff[0:4], []byte{0, 0, 0, 0}) == 0 {
			this.IsHeartbeat = true
			return nil
		}
		r := bytes.NewReader(buff)
		if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
			return err
		}
	}
	dataLen -= headerLen
	if dataLen <= 0 {
		return nil
	}
	if dataLen > MAX_UPLOAD_SIZE {
		fmt.Println("too big upload data", dataLen)
		return errors.New("too big upload data")
	}
	data := make([]byte, dataLen)
	if err := conn.ReadStream(data, int(dataLen)); err != nil {
		return err
	}
	this.DataStream = data
	return nil
}

func (this *XimpBuffer) readFromString(conn INetwork, length int) error {
	if length == 0 {
		this.IsHeartbeat = true
		return nil
	}
	if length > MAX_UPLOAD_SIZE {
		fmt.Println("too big upload data wx", length)
		return errors.New("too big upload data wx")
	}
	str := make([]byte, length)
	if err := conn.ReadStream(str, length); err != nil {
		return err
	}
	data, err := cryption.Base64Decode(string(str))
	if err != nil {
		return err
	}
	this.DataStream = data
	return nil
}

func (this *XimpBuffer) WriteTo(tcpConnection *TcpConnection, timeout time.Duration) error {
	buffer, err := this.Encode()
	if err != nil {
		return err
	}
	if err := tcpConnection.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	if _, err := tcpConnection.Write(buffer); err != nil {
		return err
	}
	return nil
}

// 编码
func (this *XimpBuffer) Encode() ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	if this.IsHeartbeat {
		buff.Write([]byte{0, 0, 0, 0})
		return buff.Bytes(), nil
	}
	var headerLen int = 4
	if this.HasHeader {
		if this.IsClient || this.Version > 0 {
			if _, err := buff.Write([]byte{'q', 'h'}); err != nil {
				return []byte{}, err
			}
			headerLen += 2
		}
		if this.IsClient {
			headerLen += 10
			versions := uint16((this.CVersion & 0x0FFF) | (this.Version & 0xF << 12))
			if err := binary.Write(buff, binary.BigEndian, versions); err != nil {
				return []byte{}, err
			}
			if err := binary.Write(buff, binary.BigEndian, this.Appid); err != nil {
				return []byte{}, err
			}
			buff.Write(this.Reserved[0:6])
		}
	}
	if err := binary.Write(buff, binary.BigEndian, uint32(len(this.DataStream)+headerLen)); err != nil {
		return []byte{}, err
	}
	buff.Write(this.DataStream)
	return buff.Bytes(), nil
}

func (this *XimpBuffer) Encrypt(rkey []byte) error {
	if len(rkey) == 0 {
		return nil
	}

	var err error
	if len(this.DataStream) > 0 {
		if this.DataStream, err = cryption.Rc4Encrypt(this.DataStream, rkey); err != nil {
			return err
		}
	}
	this.IsDecrypt = false
	return nil
}

func (this *XimpBuffer) Decrypt(rkey []byte) error {
	if len(rkey) == 0 {
		return nil
	}

	var err error
	if len(this.DataStream) > 0 {
		if this.DataStream, err = cryption.Rc4Decrypt(this.DataStream, rkey); err != nil {
			return err
		}
	}
	this.IsDecrypt = true
	return nil
}
