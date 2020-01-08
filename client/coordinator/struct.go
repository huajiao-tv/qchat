package coordinator

import "time"

type AdapterStat struct {
	RecordTime     time.Time // 这条记录创建时间
	PushInterval   int       // push间隔(毫秒)
	TotalPushCount int       // 发到这个聊天室的消息量，普通规则之前
	RealPushCount  int       // 真正发到这个聊天室消息，普通规则之后
	RealPushPerS   int       // 化成每秒push多少条
	SendInterval   int       // 发送间隔
	GetCount       int       // 向消息服务里要多少条
	RealGetCount   int       // 实际取到多少条
	RealGetPerS    int       // 实际取到条数化成每秒多少条
	MemCount       int       // 这个机房有多少人
	TotalMemCount  int       // 总人数
	LastAdjust     time.Time // 最近一次调整发消息次数
	FlowCount      uint64    // 当前的流量
	FlowHum        string    // 可读的流量大小
	OverFlow       float32
	Host           string // 当前的聊天室映射到哪一个coordinator上面
}

type DegradeRequest struct {
	AppId   uint16
	RoomId  string
	Degrade bool
}

type LiveNotifyRequest struct {
	AppId  uint16
	RoomId string
	Start  bool
}
