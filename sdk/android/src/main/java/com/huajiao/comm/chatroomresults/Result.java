package com.huajiao.comm.chatroomresults;

/**
 * 通用结果基类
 * */
public class Result {

	public static final int ERR_SUCCESS = 0;

	/**
	 * 发送消息失败
	 * */
	public static final int ERR_FAILED_TO_SEND_MSG = 1;

	/**
	 * 消息发送超时
	 * */
	public static final int ERR_MSG_TIMEOUT = 2;

	/**
	 * 新消息
	 * */
	public static final int PAYLOAD_INCOMING_MESSAGE = 100;

	/**
	 * 退出动作对应的的结果
	 * */
	public static final int PAYLOAD_QUIT_RESULT = 101;

	/**
	 * 查询动作对应的结果
	 * */
	public static final int PAYLOAD_QUERY_RESULT = 102;

	/**
	 * 加入对应的结果
	 * */
	public static final int PAYLOAD_JOIN_RESULT = 103;

	/**
	 * subscribe对应的结果
	 * */
	public static final int PAYLOAD_SUBSCRIBE_RESULT = 109;

	/**
	 * 有人加入
	 * */
	public static final int PAYLOAD_MEMBER_JOINED_IN = 201;
	
	/**
	 * 有人退出 
	 * */
	public static final int PAYLOAD_MEMBER_QUIT = 202;
	

	private long _sn;
	private int _result;
	private int _payload_type;
	private byte[] _reason;

	/**
	 * Get reason bytes
	 * */
	public byte[] get_reason() {
		return _reason;
	}

	/**
	 * @param sn
	 * @param result
	 * @param payload_type
	 */
	public Result(long sn, int result, int payload_type, byte[] reason) {
		_sn = sn;
		_result = result;
		_payload_type = payload_type;
		_reason = reason;
	}

	/**
	 * @return 获取sn
	 */
	public long get_sn() {
		return _sn;
	}

	/**
	 * @return 获取结果, 参照 {@link ERR_SUCCESS}, {@link ERR_FAILED_TO_SEND_MSG},
	 *         {@link ERR_MSG_TIMEOUT}
	 * 
	 */
	public int get_result() {
		return _result;
	}

	/**
	 * @return 结果类型， 用于区分是否是JoinResult, QueryResult, QuitResult或者 是 InCommingMessage
	 */
	public int get_payload_type() {
		return _payload_type;
	}
}
