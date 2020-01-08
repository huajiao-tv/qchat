
#import <Foundation/Foundation.h>

/**
 * 定义项目中用到的常量
 *
 */


/**
 * 消息类型
 * */
typedef enum _MessageType
{
    /**
     * 文本消息： 内容是UTF-8编码的文字
     * */
    TEXT = 0,
    
    /**
     * 电话信令
     * */
    PHONE_CMD = 100,
    
    /**
     * 通讯录的好友注册了
     * */
    CONTACT_REGISTERED_NOTIFICATION = 200,
    
    /**
     * 系统保留的消息， 上传日志请求
     * */
     UPLOAD_LOG_REQ = 300,
    
    /**
     * 系统保留的消息， 上传日志响应包
     * */
     UPLOAD_LOG_RES = 301,
    
}MessageType;



/**
 * 定义了用户的状态
 */
enum IMUserState
{
    /**
     * 初始状态，user一开始就是这个状态
     */
    IM_State_Init = 0,
    
	/**
	 * 正在登录服务器
	 **/
	IM_State_Connecting = 1,
	
	/**
	 * 已经登录成功
	 * */
	IM_State_Connected = 2,
    
    /**
	 * 与服务器断开了连接,是因为用户切换系统网络造成的断链
	 * */
	IM_State_Disconnected = 3,
    
    /**
     * 服务器不可达
     */
    IM_State_Not_Reachable = 4,
    
    /**
     * 登录msgrouter失败，即：用户id或密码错误
     */
    IM_State_Login_Fail = 5,
    
    /**
     * msgrouter提示需要重新登录，通常是账号在其他设备登陆导致
     */
    IM_State_Relogin_Need = 6,
    
    /**
     * 未知状态，正常不会有该状态
     * **/
    IM_State_UnKnown = 100
};
typedef enum IMUserState IMUserState;


/**
 * 定义错误代码
 */
enum IMErrorCode
{
    /*
     * 成功，表示没有错误
     */
    IM_Success = 0,
    
    /*
     * 当前没有网络连接
     */
    IM_NoConnection = 1,
    
    /*
     *注册地址不可达
     */
    IM_RegistryServerUnReachable = 2,
    
    /*
     * 上传服务器不可达
     */
    IM_UploadServerUnReachable = 3,
    
    /*
     * 下载服务器不可达
     */
    IM_DownloadServerUnReachable = 4,
    
    /*
     * 接口服务器不可达
     */
    IM_WebServerUnReachable = 5,
    
    /*
     * 登录服务器不可达
     */
    IM_LoginServerUnReachable = 6,
    
    /**
	 * 与服务器断开了， 原因是密码或者用户ID不正确
	 * */
	IM_AuthFailed = 7,
    
    /**
     * 操作超时
     */
    IM_OperateTimeout = 8,
    
    /**
     * 无效参数
     */
    IM_InvalidParam = 9,
    
    /**
     * 操作被取消
     */
    IM_OperateCancel = 10,
    
    /**
     * 发送数据失败
     */
    IM_SendDataFail = 11,

    /**
     * 无效参数
     */
    IM_InvalidNetData = 12,

    /**
     * 未知错误
     */
    IM_UnknowError = 1000
};
typedef enum IMErrorCode IMErrorCode;


/**
 * 定义错误代码
 */
enum IMFeatureCode
{
    /**
     * 控制类型
     */
    IM_Feature_Ctrl = 0,
    
    /**
     * 单聊
     */
    IM_Feature_Peer = 1,
    
    /**
     * 群聊
     */
    IM_Feature_GroupChat = 2,
    
    /**
     * 聊天室
     */
    IM_Feature_ChatRoom = 3,
    
    /**
     * 圈子
     */
    IM_Feature_Circle = 4,
    
    /**
     * 全量push
     */
    IM_Feature_Public = 5,
    
    /**
     * 频道
     */
    IM_Feature_Channel = 7,

    /**
     * 单聊
     */
    IM_Feature_PrivateChat = 8,
    
    IM_Feature_IM = 9,
    
    /**
     * 通知
     */
    IM_Feature_Notify = 100
};
typedef enum IMFeatureCode IMFeatureCode;


/**
 *
 */
enum IMTaskType
{
    /**
     *
     */
    IM_TaskType_Normal = 0,
    
    /**
     * 打开网络通知
     */
    IM_TaskType_Network_Poweron = 1,
    
    /**
     * 系统网路关闭
     */
    IM_TaskType_Network_Poweroff = 2,
    
    /**
     * socket断开
     */
    IM_TaskType_Socket_Disconnect = 3,
    
    /**
     * 心跳包
     */
    IM_TaskType_HeartBeat = 5,
    
};
typedef enum IMTaskType IMTaskType;


typedef enum _MsgId {
    // 消息请求类型
    MSG_ID_REQ_EMPTY_HEART_BEAT = 0,
    
    MSG_ID_REQ_LOGIN = 100001,
    MSG_ID_REQ_CHAT = 100002,
    MSG_ID_REQ_QUERY_INFO = 100003,
    MSG_ID_REQ_GET_INFO = 100004,
    MSG_ID_REQ_LOGOUT = 100005,
    MSG_ID_REQ_HEART_BEAT = 100006,
    MSG_ID_REQ_QUERY_USER_STATUS = 100007,
    MSG_ID_REQ_QUERY_USER_REG = 100008,
    MSG_ID_REQ_INIT_LOGIN = 100009,
    
    MSG_ID_REQ_SERVICE_CONTROL = 100011,
    MSG_ID_REQ_EX1_QUERY_USER_STATUS = 100012,

    MSG_ID_REQ_QUERY_QUOTA = 100015,
    MSG_ID_REQ_UPDATE_CONVSUMMARY = 100016,
    MSG_ID_GET_MULTI_INFOS_REQ = 100100,
    
    // 消息响应类型
    MSG_ID_RESP_LOGIN = 200001,
    MSG_ID_RESP_CHAT = 200002,
    MSG_ID_RESP_GET_INFO = 200004,
    MSG_ID_RESP_LOGOUT = 200005,
    // MSG_ID_RESP_HEART_BEAT = 200006,
    MSG_ID_RESP_QUERY_USER_STATUS = 200007,
    MSG_ID_RESP_QUERY_USER_REG = 200008,
    MSG_ID_RESP_INIT_LOGIN = 200009,
    MSG_ID_RESP_EX_QUERY_USER_STATUS = 200010,
    MSG_ID_RESP_SERVICE_CONTROL = 200011,
    MSG_ID_RESP_EX1_QUERY_USER_STATUS = 200012,
    MSG_ID_RESP_QUERY_CONVSUMMARY = 200014,
    MSG_ID_RESP_QUERY_QUOTA = 200015,
    MSG_ID_RESP_UPDATE_CONVSUMMARY = 200016,
    
    MSG_ID_RESP_GET_MULTI_INFOS = 200100,
    
    MSG_ID_NTF_NEW_MESSAGE = 300000,
    MSG_ID_NTF_RELOGIN = 300001,
    MSG_ID_NTF_RECONNECT = 300002
    
}MsgId;


typedef enum _Payload{
    PAYLOAD_REQ_CREATE_CHANNEL = 100000,
    PAYLOAD_REQ_CHECK_CHANNEL = 100001,
    PAYLOAD_REQ_RESTORE_CHANNEL = 100002,
    
    PAYLOAD_RESP_CREATE_CHANNEL = 200000,
    PAYLOAD_RESP_CHECK_CHANNEL = 200001,
    PAYLOAD_RESP_RESTORE_CHANNEL = 200002,
    
    PAYLOAD_NEW_CHANNEL_NOTIFY = 300000,
    
}PayLoad;


typedef enum _service_id{
    SERVICE_ID_MSGROUTER                                                        = 10000000,
    SERVICE_ID_GROUP                                                                  = 10000001,
    SERVICE_ID_DISTRIBUTE                                                          = 10000002,
    SERVICE_ID_CIRCLE                                                                   = 10000003,
    SERVICE_ID_RELATION                                                              = 10000004,
    SERVICE_ID_APNS                                                                      = 10000005,
    SERVICE_ID_CHATROOM                                                          = 10000006,
    SERVICE_ID_VCP                                                                         = 10000007,
    SERVICE_ID_WISH                                                                      = 10000009,
}Service_ID;

typedef enum _Chatroom_Payload{
    PAYLOAD_QUERY_CHATROOM = 101,
    PAYLOAD_JOIN_CHATROOM = 102,
    PAYLOAD_QUIT_CHATROOM = 103,
    PAYLOAD_SUBSCRIBE_CHATROOM = 109,
    PAYLOAD_MESSAGE_CHATROOM = 300,
    
    // chatroom message Notification
    PAYLOAD_NEW_MSG_NTF_CHATROOM = 1000,
    PAYLOAD_JOIN_NTF_CHATROOM = 1001,
    PAYLOAD_QUIT_NTF_CHATROOM = 1002,
    PAYLOAD_MIX_NTF_CHATROOM = 1003,
    
}ChatroomPayload;

typedef enum _Group_Payload{
    // section for client request payload type
    PAYLOAD_REQ_SYNC_GROUP_LIST = 108,
    PAYLOAD_REQ_GET_GROUP_MSGS = 109,

    // section for server response payload type
    PAYLOAD_RESP_SYNC_GROUP_LIST = 108,
    PAYLOAD_RESP_GET_GROUP_MSGS = 109,
    PAYLOAD_RESP_NEW_GROUP_MSG_NOTIFY = 200,

}GroupPayLoad;

typedef enum _Group_Request_type{
    REQ_SYNC_GROUP              = 0,
    REQ_GET_GROUP_MSGS  = 1,
}GroupRequestType;
