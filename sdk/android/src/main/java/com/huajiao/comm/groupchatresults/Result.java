package com.huajiao.comm.groupchatresults;

/**
 * Created by zhangjun-s on 16-7-19.
 */

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
    public static final int PAYLOAD_NEWMSG_NOTIFY = 200;

    /**
     * 同步群信息对应的结果
     * */
    public static final int PAYLOAD_GROUPSYNC = 108;

    /**
     * 获取群消息对应的的结果
     * */
    public static final int PAYLOAD_GETGROUPMSGS = 109;



    private long _sn;
    private int _result;
    private int _payload_type;
    private String _reason;

    /**
     * Get reason bytes
     * */
    public String get_reason() {
        return _reason;
    }

    /**
     * @param sn
     * @param result
     * @param payload_type
     */
    public Result(long sn, int result, int payload_type, String reason) {
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
