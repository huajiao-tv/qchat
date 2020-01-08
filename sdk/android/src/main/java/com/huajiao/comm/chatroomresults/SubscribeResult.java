package com.huajiao.comm.chatroomresults;

/**
 * Created by zhangjun-s on 16-9-1.
 */

/**
 * subscribe聊天室的结果
 */
public class SubscribeResult extends Result {
    private String _roomid;
    private boolean _issubscribe;

    public SubscribeResult(long sn, int result, byte[] reason, String roomid, boolean issubscribe) {
        super(sn, result, Result.PAYLOAD_SUBSCRIBE_RESULT, reason);
        _roomid = roomid;
        _issubscribe = issubscribe;
    }

    /**
     * @return 获取roomid
     */
    public String get_roomid() {
        return _roomid;
    }

    /**
     * 获取is or not subscribe
     * */
    public boolean get_subscribe() {
        return _issubscribe;
    }
}
