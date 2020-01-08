package com.huajiao.comm.im;

/**
 * 常见的错误定义
 * */
public class Error {

	// ------------------------------------------------------------------------
	// 下面的是服务器定义的错误
	// ------------------------------------------------------------------------

	// 严重的错误， 服务器会断掉socket
	public final static int CLIENT_IS_TOO_OLD = 1000;
	public final static int MORE_ARGS_ARE_NEEDED = 1001;
	public final static int SERVER_LOGIN__FAILED = 1017;
	public final static int INVALID_SENDER = 1003;
	public final static int ACT_TOO_FREQUENTLY = 1004;
	public final static int UNKNOW_CHAR_TYPE = 1005;
	public final static int DATABSE_EXCEPTION = 1006;
	public final static int SESSION_EXCEPTION = 1007;
	public final static int USER_INVALID = 1008;
	public final static int PACKET_IS_TOO_LARGE = 1009;
	public final static int INVALID_BODY_ID = 1010;
	public final static int SES_REFUSED = 1011;
	
	/** 数据库超过负荷， 需要等一段时间后再连  */
	public final static int DATABASE_IS_TOO_BUSY = 1012;
	
	
	
	public final static int SP_EXCEPTION = 1013;
	
	public final static int SEVER_ERROR_START = 1000;
	public final static int SEVER_ERROR_END = 1013;
	
	/*** 服务器超过负荷， 需要等一段时间后再连 */
	public final static int SERVER_OVERLOADED = 1015;
	
	/*** 在别的设备上登录了 */
	public final static int REGISTERED_ELSEWHERE = 1016;

	// 普通错误客户端可以继续连接， 也可重新登录
	public final static int DB_INNER = 2000;
	public final static int SES_INNER = 2001;
	public final static int SP_INNER = 2002;
	public final static int MSG_FAILED = 2003;

	// 客户端可以忽略， 参数错误的检查一下参数
	public final static int RECEIVER_IS_NOT_REGISTERED_ = 3000;
	public final static int INVALID_USER_QUERY_PARAM = 3001;
	public final static int UNKNOWN_RECEIVER_TYPE = 3002;
	
}
