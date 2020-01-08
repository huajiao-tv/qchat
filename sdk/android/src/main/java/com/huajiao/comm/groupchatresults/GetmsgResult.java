package com.huajiao.comm.groupchatresults;

import com.huajiao.comm.protobuf.GroupChatProto;

import java.util.List;

/**
 * Created by zhangjun-s on 16-7-19.
 */

public class GetmsgResult extends Result {

    private GroupChatProto.GroupMessageResp resp_;

    public GetmsgResult(long sn, int result, String reason, int payload, GroupChatProto.GroupMessageResp resp) {
        super(sn, result, payload, reason);
        this.resp_ = resp;
    }


    public String getGroupID() {
        return resp_.getGroupid();
    }

    public long getMaxMsgID() {
        return resp_.getMaxid();
    }

    public long getVersion() {
        return resp_.getVersion();
    }

    public long getMsgCount() {
        return resp_.getMsglistCount();
    }

    public List<GroupChatProto.GroupMessage> getMsgList() {
        return resp_.getMsglistList();
    }
}
