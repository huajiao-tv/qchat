package network

import (
	"errors"
	"github.com/huajiao-tv/qchat/utility/convert"
	"strings"
	"sync"
	"time"
)

const (
	OpcodePing          = 0
	OpcodePong          = 1
	OpcodeBind          = 2
	OpcodeMsg           = 3
	OpcodeMsgAck        = 4
	OpcodeUnbind        = 5
	OpcodeBindAck       = 6
	OpcodeUnbindAck     = 7
	OpcodeKick          = 8 // 踢掉用户
	OpcodeMigrate       = 9 // 迁移用户到给定room
	OpcodeRebind        = 10
	OpcodeRebindAck     = 11
	OpcodeOfflineNotify = 12 // miop5离线补偿,走execcmd
	OpcodeUploadData    = 13 //上行数据
	OpcodeUploadDataAck = 14 //上行数据ack
	OpcodeHandshake     = 15 //握手协商
)

/**
 * 已定义的 Prop 下标有：
 *
 *   t:300
 *   u:hello@world
 *   k:0 (kick, 缺省是1)
 *   d:0 (durable, 缺省是1)
 *   c:1 (缺省不带该key)
 *   OpCode只在version=5时存在
 */
type MiopBuffer struct {
	Version int16
	OpCode  int16
	PropLen int16
	DataLen int32
	Prop    map[string]string
	Data    []byte
}

const (
	MIOP_MAX_DATA_LENGTH int32 = 1024 * 1024
	MiopBindAckMapMax    int   = 150000 //map最大值，大于此值，返回错误，记录日志
)

func NewMiopBuffer() *MiopBuffer {
	return &MiopBuffer{Prop: make(map[string]string)}
}

func (this *MiopBuffer) Reset() error {
	this.Version = 0
	this.OpCode = 0
	this.PropLen = 0
	this.DataLen = 0
	this.Data = nil
	for k, _ := range this.Prop {
		delete(this.Prop, k)
	}
	return nil
}

func (this *MiopBuffer) ReadFrom(tcpConnection *TcpConnection, timeout time.Duration) error {
	if err := tcpConnection.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	versionBuf := make([]byte, 2)
	err := tcpConnection.ReadStream(versionBuf[0:2], 2)
	if err != nil {
		return err
	}
	this.Version = convert.StreamToInt16(versionBuf[0:2], convert.BigEndian)
	switch this.Version {
	case 5, 6:
		return this.decodeMiop5(tcpConnection)
	case 0, 1, 2, 3, 4:
		return this.decodeMiop(tcpConnection)
	default:
		return errors.New("invalid miop version")
	}
	return nil
}

func (this *MiopBuffer) decodeMiop(tcpConnection *TcpConnection) error {
	header := make([]byte, 6)
	err := tcpConnection.ReadStream(header, 6)
	if err != nil {
		return err
	}
	this.PropLen = convert.StreamToInt16(header[0:2], convert.BigEndian)
	this.DataLen = convert.StreamToInt32(header[2:6], convert.BigEndian)
	if this.PropLen > 0 {
		propStream := make([]byte, this.PropLen)
		err := tcpConnection.ReadStream(propStream, int(this.PropLen))
		if err != nil {
			return err
		}
		propString := string(propStream)
		propLines := strings.Split(propString, "\n")
		for _, propLine := range propLines {
			kv := strings.SplitN(propLine, ":", 2)
			if len(kv) == 2 {
				this.Prop[kv[0]] = kv[1]
			}
		}
	}
	if this.DataLen > 0 {
		if this.DataLen <= MIOP_MAX_DATA_LENGTH {
			this.Data = make([]byte, this.DataLen)
			err := tcpConnection.ReadStream(this.Data, int(this.DataLen))
			if err != nil {
				return err
			}
		} else {
			return errors.New("miop data length exceed MIOP_MAX_DATA_LENGTH")
		}
	}
	return nil
}

func (this *MiopBuffer) decodeMiop5(tcpConnection *TcpConnection) error {
	opCodeBuf := make([]byte, 2)
	err := tcpConnection.ReadStream(opCodeBuf[0:2], 2)
	if err != nil {
		return err
	}
	this.OpCode = convert.StreamToInt16(opCodeBuf[0:2], convert.BigEndian)
	switch this.OpCode {
	case OpcodePing, OpcodePong:
		return nil
	case OpcodeBind, OpcodeUnbind, OpcodeMsgAck, OpcodeBindAck, OpcodeUnbindAck, OpcodeRebind, OpcodeRebindAck,
		OpcodeMsg, OpcodeKick, OpcodeMigrate, OpcodeUploadData, OpcodeUploadDataAck, OpcodeHandshake:
		propLengthBuf := make([]byte, 2)
		err := tcpConnection.ReadStream(propLengthBuf[0:2], 2)
		if err != nil {
			return err
		}
		this.PropLen = convert.StreamToInt16(propLengthBuf[0:2], convert.BigEndian)
		if this.PropLen > 0 {
			propStream := make([]byte, this.PropLen)
			err := tcpConnection.ReadStream(propStream, int(this.PropLen))
			if err != nil {
				return err
			}
			propString := string(propStream)
			propLines := strings.Split(propString, "\n")
			for _, propLine := range propLines {
				kv := strings.SplitN(propLine, ":", 2)
				if len(kv) == 2 {
					this.Prop[kv[0]] = kv[1]
				}
			}
		}
		switch this.OpCode {
		case OpcodeMsg, OpcodeUploadData:
			dataLengthBuf := make([]byte, 4)
			err := tcpConnection.ReadStream(dataLengthBuf, 4)
			if err != nil {
				return err
			}
			this.DataLen = convert.StreamToInt32(dataLengthBuf[0:4], convert.BigEndian)
			if this.DataLen > 0 {
				this.Data = make([]byte, this.DataLen)
				err := tcpConnection.ReadStream(this.Data, int(this.DataLen))
				if err != nil {
					return err
				}
			}
		default:
			return nil
		}
	default:
		return errors.New("invalid miop5 protocol")
	}

	return nil
}

func (this *MiopBuffer) WriteTo(tcpConnection *TcpConnection, timeout time.Duration) error {
	buffer := this.Stream()
	if err := tcpConnection.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	_, err := tcpConnection.Write(buffer)
	return err
}

func (this *MiopBuffer) Stream() []byte {
	if this.Version == 5 || this.Version == 6 {
		return this.StreamMiop5()
	}
	return this.StreamMiop()

}

// 编码prop
func (this *MiopBuffer) encodeProp() string {
	var propString string
	propCount := len(this.Prop)
	if propCount > 0 {
		propLines := make([]string, propCount)
		i := 0
		for k, v := range this.Prop {
			propLines[i] = k + ":" + v
			i++
		}
		propString = strings.Join(propLines, "\n")
	}
	return propString
}

// 编码miop0~4版本
func (this *MiopBuffer) StreamMiop() []byte {
	propString := this.encodeProp()
	propLen := len(propString)
	dataLen := len(this.Data)
	buffer := make([]byte, 8+propLen+dataLen)
	convert.Int16ToStreamEx(buffer[0:2], this.Version, convert.BigEndian)
	convert.Int16ToStreamEx(buffer[2:4], int16(propLen), convert.BigEndian)
	convert.Int32ToStreamEx(buffer[4:8], int32(dataLen), convert.BigEndian)
	if propLen > 0 {
		copy(buffer[8:8+propLen], propString)
	}
	if dataLen > 0 {
		copy(buffer[8+propLen:], this.Data)
	}
	return buffer
}

// 编码miop5版本
func (this *MiopBuffer) StreamMiop5() []byte {
	switch this.OpCode {
	case OpcodePong, OpcodePing:
		buffer := make([]byte, 4)
		convert.Int16ToStreamEx(buffer[0:2], this.Version, convert.BigEndian)
		convert.Int16ToStreamEx(buffer[2:4], this.OpCode, convert.BigEndian)
		return buffer
	case OpcodeMsg, OpcodeUploadData:
		propString := this.encodeProp()
		propLen := len(propString)
		dataLen := len(this.Data)
		buffer := make([]byte, 2+2+2+propLen+4+dataLen)
		convert.Int16ToStreamEx(buffer[0:2], this.Version, convert.BigEndian)
		convert.Int16ToStreamEx(buffer[2:4], this.OpCode, convert.BigEndian)
		convert.Int16ToStreamEx(buffer[4:6], int16(propLen), convert.BigEndian)
		if propLen > 0 {
			copy(buffer[6:6+propLen], propString)
		}
		convert.Int32ToStreamEx(buffer[6+propLen:6+propLen+4], int32(dataLen), convert.BigEndian)
		if dataLen > 0 {
			copy(buffer[6+propLen+4:], this.Data)
		}
		return buffer
	case OpcodeMsgAck, OpcodeBind, OpcodeUnbind, OpcodeRebind, OpcodeRebindAck, OpcodeBindAck, OpcodeUnbindAck, OpcodeKick, OpcodeMigrate, OpcodeUploadDataAck, OpcodeHandshake:
		propString := this.encodeProp()
		propLen := len(propString)
		buffer := make([]byte, 2+2+2+propLen)
		convert.Int16ToStreamEx(buffer[0:2], this.Version, convert.BigEndian)
		convert.Int16ToStreamEx(buffer[2:4], this.OpCode, convert.BigEndian)
		convert.Int16ToStreamEx(buffer[4:6], int16(propLen), convert.BigEndian)
		if propLen > 0 {
			copy(buffer[6:6+propLen], propString)
		}
		return buffer

	}
	return []byte{}
}

func (this *MiopBuffer) HasProp(key string) bool {
	if _, ok := this.Prop[key]; ok {
		return true
	}
	return false
}

func (this *MiopBuffer) IsEmpty() bool {
	return this.DataLen == 0
}

// 判断是否是pong包
func (this *MiopBuffer) IsPongMessage() bool {
	return this.OpCode == OpcodePong
}

// 判断是否是ack包
func (this *MiopBuffer) IsAckMessage() bool {
	return this.OpCode == OpcodeMsgAck
}

// 判断是否是kick包
func (this *MiopBuffer) IsKickMessage() bool {
	return this.OpCode == OpcodeKick
}

// 判断是否是迁移包
func (this *MiopBuffer) IsMigrateMessage() bool {
	return this.OpCode == OpcodeMigrate
}

func (this *MiopBuffer) IsMessage() bool {
	return this.OpCode == OpcodeMsg
}

// 获取上行数据的data字段，返回data,本次解析到的offset
func (this *MiopBuffer) ReadUpdateData(offset int) (string, int, error) {
	if offset == len(this.Data) {
		return "", 0, nil
	}
	if len(this.Data) < offset+4 {
		return "", 0, errors.New("invalid bodyLen")
	}
	bodyLen := convert.StreamToInt32(this.Data[offset:offset+4], convert.BigEndian)
	if bodyLen < 0 {
		return "", 0, errors.New("bodyLen < 0")
	}
	offset = offset + 4
	if len(this.Data) < offset+int(bodyLen) {
		return "", 0, errors.New("invalid body")
	}
	if bodyLen == 0 {
		return "", offset, nil
	}

	data := string(this.Data[offset : offset+int(bodyLen)])
	offset = offset + int(bodyLen)
	return data, offset, nil
}

var MiopBindAckMap map[string][]byte
var MiopBindAckMapLock sync.RWMutex

/* 此处bind和unbind的ACK公用一个map，不过要做区分*/
func MiopBindAckStream(appId, r string, version int16, opCode int16) ([]byte, error) {
	if MiopBindAckMap == nil {
		MiopBindAckMap = make(map[string][]byte, 1)
	}
	var key string
	if opCode == OpcodeBindAck {
		key = appId + "-bind" + r
	} else if opCode == OpcodeUnbindAck {
		key = appId + "-unbind" + r
	} else if opCode == OpcodeRebindAck {
		key = appId + "-rebind" + r
	} else {
		return nil, errors.New("bad param")
	}
	MiopBindAckMapLock.RLock()
	if BindStream, OK := MiopBindAckMap[key]; OK {
		MiopBindAckMapLock.RUnlock()
		return BindStream, nil
	}

	miopBuf := NewMiopBuffer()
	miopBuf.Version = version
	miopBuf.OpCode = opCode
	miopBuf.Prop["id"] = appId
	miopBuf.Prop["r"] = r

	if len(MiopBindAckMap) > MiopBindAckMapMax {
		MiopBindAckMapLock.RUnlock()
		println("Error: MiopBindAckMapMax")
		return miopBuf.Stream(), nil
	}
	MiopBindAckMapLock.RUnlock()
	MiopBindAckMapLock.Lock()
	MiopBindAckMap[key] = miopBuf.Stream()
	MiopBindAckMapLock.Unlock()
	return MiopBindAckMap[key], nil
}

var Miop5PongStream = func() []byte {
	miopBuf := NewMiopBuffer()
	miopBuf.Version = 5
	miopBuf.OpCode = OpcodePong
	return miopBuf.Stream()
}()

var Miop6PongStream = func() []byte {
	miopBuf := NewMiopBuffer()
	miopBuf.Version = 6
	miopBuf.OpCode = OpcodePong
	return miopBuf.Stream()
}()
