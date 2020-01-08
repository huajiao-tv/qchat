package com.huajiao.comm.im;

/*** 时间相关常量  */
class TimeConst {

	/** 取消息， 或者ping的WakeLock超时10秒*/
	public static final int WL_TIME_OUT = 10000;

	/** socket连接超时 */
	public static final int SOCKET_CONNECT_TIMEOUT = 5000;
	
	/** 每15分钟取一次消息 */
	public final static long GET_MSG_INTERVAL = 900000;

	/** 回调超时10秒 */
	public final static int NOTIFY_CB_TIMEOUT = 10000;

	/** 每次最多取多少条消息 */
	public final static int MAX_MSG_COUNT_PER_QUERY = 5;

	/** 消息发送超时  */
	public final static int MSG_SEND_TIMEOUT = 120000;

	/** 心跳ACK超时 10 秒*/
	public final static int HEARTBEAT_ACK_TIMEOUT = 10000;

	/** 普通包超时  10秒*/
	public final static int PACKET_RESP_TIMEOUT = 10000;
	
	/** SERVICE PACKET超时 20秒*/
	public final static int SRV_PACKET_RESP_TIMEOUT = 20000;

	/** 两次登录之间的最少时间间隔, 5秒, 平均2.5秒*/
	public final static int DEFAULT_LOGIN_INTERVAL = 5000;

	/** 超过负载后的登录最短间隔时间 5 分钟 , 超过负载后的登录最长间隔时间 10 分钟 */
	public final static int OVERLOADED_LOGIN_INTERVAL = 300000;
}
