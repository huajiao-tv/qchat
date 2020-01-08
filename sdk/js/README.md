### 引入 
```javascript
import LiveSocket from '@q/hj-livesocket';
```

### 初始化
```javascript
const liveSocket = new LiveSocket({
        // 必传 这些都是每个业务固定的东西
        flag: '',
        protocolVersion: 0,
        clientVersion: 0,
        appId: 0,
        reserved: '',
        defaultKey: '',
        senderType: '',
        wsServer:'',

        // 非必传 如果需要登录用户的话, 传入下面的 uid 及 sign
        uid: '',
        sign: '',
});

liveSocket.connect()
    .then(() => {
        liveSocket.joinRoom("2131231241").then(() => {
            liveSocket.quitRoom();
        });
    });
```

### 方法
#### connect
连接服务
#### joinRoom(RoomId: String)
加入指定房间
#### quitRoom
退出房间

### emit
#### p2message
```javascript
    liveSocket.event.on('PeerMessage', (data) => {
        console.log('PeerMessage', data);
    });
```
#### 事件列表
| 名称 | 参数 |说明 |
|---|---|---|
|HandShake | |握手成功后触发|
|Login | |登录成功后触发|
|JoinRoomFail |String ApplyJoinChatRoomResponse |进入房间失败时触发, 回调为业务对象或 timeout 字符串|
|QuitRoomFail |String QuitChatRoomResponse |退出房间失败时触发, 回调为业务对象或 timeout 字符串|
|JoinRoomSuccess |ApplyJoinChatRoomResponse |进入房间成功时触发, 回调为业务对象|
|QuitRoomSuccess |QuitChatRoomResponse |退出房间成功时触发, 回调为业务对象|
|UnsupportedMsg |Msg |遇到不支持的msgid类型信息触发, 将回调msg信息, 可在此拓展|
|MemberCountUpdate |Number |房间人数更新触发, 回调为人数|
|Message |Object |房间业务消息, 回调为业务对象|
|PeerMessage |String Object |Peer消息, 回调为业务对象|

#### Message 事件中包含业务数据, 在花椒业务中通过 type 属性区分类型操作

type = 3  直播结束
type = 9  普通消息
type = 30 礼物消息

主要消息类型如上, 此类库不限定业务, 因此不与具体业务重合, 不过多介绍
