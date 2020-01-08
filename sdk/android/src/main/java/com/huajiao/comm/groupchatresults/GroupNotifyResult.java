package com.huajiao.comm.groupchatresults;

import com.huajiao.comm.protobuf.GroupChatProto;

/**
 * Created by zhangjun-s on 16-7-20.
 */

public class GroupNotifyResult extends Result {
    private GroupChatProto.GroupNotify notify_;

    public GroupNotifyResult(long sn, int result, String reason, int payload, GroupChatProto.GroupNotify notify) {
        super(sn, result, payload, reason);
        this.notify_ = notify;
    }

    public String getGroupID() {
        return notify_.getGroupid();
    }

    public String getSummary() {
        return notify_.getSummary();
    }

    public long getMsgID() {
        return notify_.getMsgid();
    }
}
