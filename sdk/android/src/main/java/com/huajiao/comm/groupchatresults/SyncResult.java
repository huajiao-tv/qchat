package com.huajiao.comm.groupchatresults;

import com.huajiao.comm.protobuf.GroupChatProto;

/**
 * Created by zhangjun-s on 16-7-19.
 */

public class SyncResult extends Result {

    private GroupChatProto.GroupInfo info_;

    public SyncResult(long sn, int result, String reason, int payload, GroupChatProto.GroupInfo info) {
        super(sn, result, payload, reason);
        this.info_ = info;
    }


    public String getGroupID() {
        return info_.getGroupid();
    }

    public long getMaxMsgID() {
        return info_.getMaxid();
    }

    public long getVersion() {
        return info_.getVersion();
    }

    public long getStartID() { return info_.getStartid(); }
}
