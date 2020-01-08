package com.huajiao.comm.groupchat;

import android.annotation.SuppressLint;
import android.content.Context;
import android.os.Handler;
import android.util.Log;

import com.google.protobuf.micro.InvalidProtocolBufferMicroException;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.Tuple;
import com.huajiao.comm.groupchatresults.GetmsgResult;
import com.huajiao.comm.groupchatresults.GroupNotifyResult;
import com.huajiao.comm.groupchatresults.Result;
import com.huajiao.comm.groupchatresults.SyncResult;
import com.huajiao.comm.im.api.ILongLiveConn;
import com.huajiao.comm.im.api.LongLiveConnFactory;

import com.huajiao.comm.im.packet.NotificationPacket;
import com.huajiao.comm.im.packet.Packet;
import com.huajiao.comm.im.packet.SrvMsgPacket;
import com.huajiao.comm.protobuf.GroupChatProto;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.StringTokenizer;
import java.util.TreeMap;

/**
 * Created by zhangjun-s on 16-7-19.
 */

public class GroupChatHelper implements Serializable {
    private static final String TAG = "GPH";//"ChatroomHelper";

    private ILongLiveConn _llc=null;
    private boolean _has_shutdown = false;
    private FlowService flowService_;
    /**
     * 用于关联sn结果， 对于没有网络或者超时的情况， 无法知道对应的的请求， 所以需要关联一下
     * */
    @SuppressLint("UseSparseArrays")
    private HashMap<Long, Integer> _pending_actions = new HashMap<>();

    public List<Long> reqsn_list;

    public static final Object flow_lock_ = "flow_lock_";
    private static final Object conn_lock_ = "conn_lock";

    private HashMap<Tuple<String,Long,Boolean>, Long> getmsg_reqs_ = new HashMap<>();//groupid,startid,direction->sn
    private long last_sync_all_sn = 0;//sync all req sn
    private TreeMap<Long, GroupChatProto.GroupUpPacket> get_pending_reqs_ = new TreeMap<>();//sn->reqpack
    private TreeMap<Long, GroupChatProto.GroupUpPacket> sync_pending_reqs_ = new TreeMap<>();//sn->reqpack

    @SuppressLint("UseSparseArrays")
    private HashMap<Long, GroupChatProto.GroupUpPacket> unack_reqs_ = new HashMap<>();//sn->reqpack
    @SuppressLint("UseSparseArrays")
    private HashMap<Long, ArrayList<Long>> trans_map_ = new HashMap<>();//sn->list(appreqids...)


    public static final int GROUPCHAT_SRV_ID = 10000001;
    public static final String INFO_TYPE_GROUPCHAT = "group";
    static final long MAX_FLOWCTL_MILLSECONDS = 60000;
    private static final int MAX_GET_GROUPNUM = 10;
    private static final int QUEUE_FULL_NUM = 20;

    /** 长链接已经被关闭了， 无法使用方法， 需要重新new一个实例 */
    private static final long ERR_LLC_HAS_BEEN_SHUTDOWN = -2;
    private static final long ERR_LLC_HASNOT_BEEN_INIT  = -3;

    private static GroupChatHelper sInstance = null;
    private long flow_ctl_abs_millseconds = 0;

    long getFlowCtlAbsMills(){//带校正功能
        synchronized (flow_lock_){
            if ( flow_ctl_abs_millseconds - System.currentTimeMillis() > MAX_FLOWCTL_MILLSECONDS){
                flow_ctl_abs_millseconds = System.currentTimeMillis();
            }

            return flow_ctl_abs_millseconds;
        }
    }

    public void setFlowCtlAbsMills(long mills, boolean force){//带校正和检查功能
        synchronized (flow_lock_){
            if ( flow_ctl_abs_millseconds - System.currentTimeMillis() > MAX_FLOWCTL_MILLSECONDS){
                flow_ctl_abs_millseconds = System.currentTimeMillis();
            }

            if ( mills - System.currentTimeMillis() > MAX_FLOWCTL_MILLSECONDS) {
                return;
            }

            if ( mills > flow_ctl_abs_millseconds ) {
                flow_ctl_abs_millseconds = mills;
            }else if ( force ){
                flow_ctl_abs_millseconds = mills;
            }
        }
    }

    void addUnackQueue(long sn, GroupChatProto.GroupUpPacket pack) {
        synchronized (flow_lock_) {
            //清除未回应的请求包
            if ( unack_reqs_.size() > QUEUE_FULL_NUM) {
                for (long oldsn : unack_reqs_.keySet()) {
                    trans_map_.remove(oldsn);//先清除sn和reqid的对应表
                }
                unack_reqs_.clear();
            }
            unack_reqs_.put(sn, pack);
        }
    }

    private GroupChatProto.GroupUpPacket getUnackQueue(long sn) {
        synchronized (flow_lock_) {
            return unack_reqs_.get(sn);
        }
    }

    private GroupChatProto.GroupUpPacket removeUnackQueue(long sn) {
        synchronized (flow_lock_) {
            return unack_reqs_.remove(sn);
        }
    }

    private void addPendingQueue(long sn, int payload, GroupChatProto.GroupUpPacket pack) {
        int size=0;
        synchronized (flow_lock_) {
            if ( payload == Result.PAYLOAD_GETGROUPMSGS) {
                if ( get_pending_reqs_.size() > QUEUE_FULL_NUM ) {
                    get_pending_reqs_.remove(get_pending_reqs_.firstKey());
                }
                get_pending_reqs_.put(sn, pack);
                size = get_pending_reqs_.size();
            }else if ( payload == Result.PAYLOAD_GROUPSYNC) {
                if ( sync_pending_reqs_.size() > QUEUE_FULL_NUM ) {
                    sync_pending_reqs_.remove(sync_pending_reqs_.firstKey());
                }
                sync_pending_reqs_.put(sn, pack);
                size = sync_pending_reqs_.size();
            }
        }

        GPLogger.i(TAG, "addPendingQueue, sn="+sn+",size="+size+",payload="+payload);
    }

    GroupChatProto.GroupUpPacket removePendingQueue(long sn, int payload) {
        synchronized (flow_lock_) {
            if ( payload == Result.PAYLOAD_GETGROUPMSGS) {
                return get_pending_reqs_.remove(sn);
            }else if ( payload == Result.PAYLOAD_GROUPSYNC) {
                return sync_pending_reqs_.remove(sn);
            }
            return null;
        }
    }

    private GroupChatProto.GroupUpPacket mergePendingQueue(long sn, int payload, GroupChatProto.GroupUpPacket pack) {
        synchronized (flow_lock_) {
            if (payload == Result.PAYLOAD_GETGROUPMSGS) {
                GroupChatProto.GroupUpPacket oldpack = get_pending_reqs_.get(sn);
                if (oldpack != null && pack.getGetmsgreqCount() > 0) {
                    for (GroupChatProto.GroupMessageReq req : oldpack.getGetmsgreqList()) {
                        if (req.getGroupid().equals(pack.getGetmsgreq(0).getGroupid())
                                && req.getStartid() == pack.getGetmsgreq(0).getStartid()
                                && req.getOffset() * pack.getGetmsgreq(0).getOffset() > 0
                                )
                        {
                            req.setOffset(pack.getGetmsgreq(0).getOffset());
                            String reqids = req.getTraceid() + "," + pack.getGetmsgreq(0).getTraceid();
                            req.setTraceid(reqids);
                        }
                    }
                    get_pending_reqs_.put(sn, oldpack);
                    return oldpack;
                } else {
                    return null;
                }
            }
            return null;
        }
    }

    private long mergePendingQueue(int payload, GroupChatProto.GroupUpPacket pack) {
        synchronized (flow_lock_) {
            if (payload == Result.PAYLOAD_GETGROUPMSGS) {
                Map.Entry<Long, GroupChatProto.GroupUpPacket> entry = get_pending_reqs_.lastEntry();
                if ( entry == null ){
                    return 0;
                }
                long sn = entry.getKey();
                GroupChatProto.GroupUpPacket oldpack = entry.getValue();
                if ( oldpack.getGetmsgreqCount() + pack.getGetmsgreqCount() < MAX_GET_GROUPNUM ){
                    for (GroupChatProto.GroupMessageReq req : pack.getGetmsgreqList()) {
                        oldpack.addGetmsgreq(req);
                    }
                    get_pending_reqs_.put(sn, oldpack);
                    return sn;
                }else{
                    return 0;
                }
            } else if (payload == Result.PAYLOAD_GROUPSYNC) {
                Map.Entry<Long, GroupChatProto.GroupUpPacket> entry = sync_pending_reqs_.lastEntry();
                if ( entry == null ){
                    return 0;
                }
                long sn = entry.getKey();
                GroupChatProto.GroupUpPacket oldpack = entry.getValue();
                if ( oldpack.getSyncreqCount() == 0 ){
                    return sn;
                }else if ( pack.getSyncreqCount() == 0 ) {
                    sync_pending_reqs_.put(sn, pack);
                    return sn;
                }else {
                    for (GroupChatProto.GroupSyncReq req : pack.getSyncreqList()) {
                        oldpack.addSyncreq(req);
                    }
                    sync_pending_reqs_.put(sn, oldpack);
                    return sn;
                }

            }
            return 0;
        }
    }

    private void addTransmap(long sn, long reqid) {
        synchronized (flow_lock_) {
            ArrayList<Long> reqlist = trans_map_.get(sn);
            if ( reqlist == null ) {
                reqlist = new ArrayList<>();
                if ( trans_map_.size() > QUEUE_FULL_NUM ){
                    trans_map_.clear();
                }
                trans_map_.put(sn, reqlist);
            }
            reqlist.add(reqid);
        }
    }

    private void addTransmap(long sn, ArrayList<Long> reqids) {
        synchronized (flow_lock_) {
            ArrayList<Long> reqlist = trans_map_.get(sn);
            if ( reqlist == null ) {
                reqlist = new ArrayList<>();
                if ( trans_map_.size() > QUEUE_FULL_NUM ){
                    trans_map_.clear();
                }
                trans_map_.put(sn, reqlist);
            }
            reqlist.addAll(reqids);
        }
    }

    private void removeTransmap(long sn, StringBuffer reqidsinfo, List<Long> reqids){
        synchronized (flow_lock_) {
            ArrayList<Long> reqidlist = trans_map_.get(sn);
            if (reqidlist != null) {
                StringBuilder info = (new StringBuilder()).append("sn:").append(sn).append(";");
                for (long reqid : reqidlist) {
                    info.append("reqid=").append(reqid).append(",");
                }
                trans_map_.remove(sn);
                reqidsinfo.append(info.toString());
                if ( reqids != null ) {
                    reqids.addAll(reqidlist);
                }
            }
        }
    }

    private void addGetQueue(String groupid, long startid, int offset, long sn) {
        synchronized (flow_lock_) {
            if ( getmsg_reqs_.size() > QUEUE_FULL_NUM ){
                getmsg_reqs_.clear();
            }
            getmsg_reqs_.put(new Tuple<>(groupid,startid,offset>0), sn);
        }
    }

    private long getGetQueue(String groupid, long startid, int offset) {
        synchronized (flow_lock_) {
            Long sn = getmsg_reqs_.get(new Tuple<>(groupid,startid,offset>0));
            if ( sn != null ){
                return sn;
            }else{
                return 0;
            }
        }
    }

    private GroupChatHelper() {
        reqsn_list = Collections.synchronizedList(new ArrayList<Long>());
        flowService_ = new FlowService(this);
    }
    
    public static GroupChatHelper getInstance() {
		synchronized (GroupChatHelper.class) {
			if(sInstance == null) {
				sInstance = new GroupChatHelper();
			}
		}

        synchronized (conn_lock_) {
            sInstance._llc = LongLiveConnFactory.getDefaultConn();
        }
    	return sInstance;
    }   

    public boolean is_llc_shutdown() {
        return _has_shutdown;
    }

    public static class GroupMsgReq{
        public GroupMsgReq(String groupid, long startid, int offset, long reqid) {
			this.groupid = groupid;
			this.startid = startid;
			this.offset = offset;
            this.reqid = reqid;
		}

		java.lang.String groupid;
        long startid;
        int offset;
        long reqid;
    }

    private boolean isInvalidConn() {
        synchronized (conn_lock_) {
            if (_llc == null) {
                _llc = LongLiveConnFactory.getDefaultConn();
            }
        }

        return _llc==null;
    }

    public ILongLiveConn getConn(){
        synchronized (conn_lock_) {
            if (_llc == null) {
                _llc = LongLiveConnFactory.getDefaultConn();
            }
        }
        return _llc;
    }

    /**
     * 取群消息
     *
     * @param reqs 多个取群消息的汇总请求
     * @return 大于0: 消息sn, 负数： 失败
     * @throws IllegalArgumentException
     * */
    public long getGroupMessages(List<GroupMsgReq> reqs) {

        if (_has_shutdown) {
            return ERR_LLC_HAS_BEEN_SHUTDOWN;
        }

        if (reqs == null || reqs.size() == 0) {
            throw new IllegalArgumentException();
        }

        if ( isInvalidConn() ) {
            return ERR_LLC_HASNOT_BEEN_INIT;
        }

        GroupChatProto.GroupUpPacket packet = new GroupChatProto.GroupUpPacket();
        packet.setPayload(Result.PAYLOAD_GETGROUPMSGS);

        StringBuilder groupids = new StringBuilder();
        StringBuilder reqids = new StringBuilder();
        long oldsn = 0;
        String groupid=null;
        long startid=0;
        int offset=0;
        ArrayList<Long> reqidlist = new ArrayList<>();

        for(GroupMsgReq req:reqs)
        {
            GroupChatProto.GroupMessageReq getreq = new GroupChatProto.GroupMessageReq();
            groupids.append(req.groupid).append("-");
            groupid = req.groupid;
            startid = req.startid;
            offset = req.offset;
            reqidlist.add(req.reqid);
            reqids.append(req.reqid).append(",");
            getreq.setGroupid(req.groupid);
            getreq.setStartid(req.startid);
            getreq.setOffset(req.offset);
            getreq.setTraceid(String.valueOf(req.reqid));
            oldsn = getGetQueue(req.groupid, req.startid, req.offset);
            packet.addGetmsgreq(getreq);
        }

        if ( reqs.size() == 1 && oldsn != 0 && mergePendingQueue(oldsn, Result.PAYLOAD_GETGROUPMSGS, packet)!=null ){
            addTransmap(oldsn, reqidlist);
            GPLogger.i(TAG, String.format(Locale.US, "getgroupmsgreq %s merge, sn=%d, reqid=%s", groupids.toString(), oldsn, reqids.toString()));
            return oldsn;
        }else{
            long sn = mergePendingQueue(Result.PAYLOAD_GETGROUPMSGS, packet);
            if ( sn > 0 ){
                GPLogger.i(TAG, String.format(Locale.US, "getgroupmsgreq %s merge, sn=%d, reqid=%s", groupids.toString(), sn, reqids.toString()));
            } else {
                sn = _llc.get_sn();
                addPendingQueue(sn, Result.PAYLOAD_GETGROUPMSGS, packet);//加入代发请求列表
                flowService_.enqueueReq(sn, reqids.toString(), Result.PAYLOAD_GETGROUPMSGS);
                GPLogger.i(TAG, String.format(Locale.US, "getgroupmsgreq %s enqueue, sn=%d, reqid=%s", groupids.toString(), sn, reqids.toString()));
            }
            if ( reqs.size() == 1 ) {
                addGetQueue(groupid, startid, offset, sn);//做合并请求用，当前只合并单一请求
            }

            addTransmap(sn, reqidlist);
            return sn;
        }
    }

    /**
     * 取群消息
     * @param groupids group id list
     * @param reqid client reqid
     * @return 大于0: 消息sn, 负数： 失败
     * @throws IllegalArgumentException
     * */

    public long syncGroupInfo(String[] groupids, long reqid) {

        if (_has_shutdown) {
            return ERR_LLC_HAS_BEEN_SHUTDOWN;
        }

        if ( isInvalidConn() ) {
            return ERR_LLC_HASNOT_BEEN_INIT;
        }

        if ( groupids == null ){
            groupids = new String[0];
        }

        GroupChatProto.GroupUpPacket packet = new GroupChatProto.GroupUpPacket();
        packet.setPayload(Result.PAYLOAD_GROUPSYNC);
        StringBuilder sbgroupids = new StringBuilder();

        if ( groupids.length > 0 ) {
            for( String groupid : groupids) {
                GroupChatProto.GroupSyncReq req = new GroupChatProto.GroupSyncReq();
                req.setGroupid(groupid);
                sbgroupids.append(groupid).append("-");
                packet.addSyncreq(req);
            }
        }

        long sn = mergePendingQueue(Result.PAYLOAD_GROUPSYNC, packet);
        if ( sn > 0 ){
            GPLogger.i(TAG, String.format(Locale.US, "syncgroupinfo groupids=null merge, sn=%d, reqid=%d", sn, reqid));
        } else {
            sn = _llc.get_sn();
            addPendingQueue(sn, Result.PAYLOAD_GROUPSYNC, packet);//加入代发请求列表
            flowService_.enqueueReq(sn, String.valueOf(reqid), Result.PAYLOAD_GROUPSYNC);
            GPLogger.i(TAG, String.format(Locale.US, "syncgroupinfo groupids=%s enqueue, sn=%d, reqid=%d", sbgroupids.toString(), sn, reqid));
        }

        addTransmap(sn, reqid);
        return sn;
    }

    public List<Result> parsePacket(Packet packet, List<Long> allreqids, List<Long> okreqids) {

        List<Result> results = null;

        try {
            results = parsePacketInner(packet, allreqids, okreqids);
        } catch (Exception e) {
            e.printStackTrace();
        }

        return results;
    }
    
    /**
     * 判断是否群聊消息
     * @param packet
     * @return true是 false否
     */
    public static boolean isGroupChatPacket(Packet packet) {
    	if(packet == null) {
    		return false;
    	}
    	switch(packet.getAction()) {
    	case Packet.ACTION_GOT_SRV_MSG:
    		SrvMsgPacket srvpacket = (SrvMsgPacket) packet;
    		return srvpacket.get_service_id() == GROUPCHAT_SRV_ID;
    	case Packet.ACTION_NOTIFICATION:
    		NotificationPacket npacket = (NotificationPacket) packet;
    		return npacket.get_info_type().equals(INFO_TYPE_GROUPCHAT);
    	}
    	return false;
    }

    private List<Result> parsePacketInner(Packet packet, List<Long> allreqids, List<Long> okreqids) {

        if (packet == null) {
            return null;
        }

        int result = -1;
        long sn = -1;
        int payload=0;
        List<Result> results = new ArrayList<Result>();

        // CRLogger.d(TAG, Utils.getStackTrace(Thread.currentThread().getStackTrace()));

        if (packet.getAction() == Packet.ACTION_GOT_SRV_MSG) {

            SrvMsgPacket srvpacket = (SrvMsgPacket) packet;

            sn = srvpacket.get_sn();

            //以下代码处理响应包的数据回收
            //从保存的unack里面取得payload
            GroupChatProto.GroupUpPacket reqpack = getUnackQueue(sn);
            if ( reqpack != null ){
                payload = reqpack.getPayload();
                removeUnackQueue(sn);
            }
            //通知正在等待的pending req可以发出了
            synchronized (flow_lock_) {
                flow_lock_.notifyAll();
            }
            StringBuffer transinfo = new StringBuffer();
            removeTransmap(sn, transinfo, allreqids);
            GPLogger.i(TAG, "recv resp " + transinfo.toString());

            result = srvpacket.get_result();

            if(JhFlag.enableDebug()) {
                GPLogger.d(TAG, "parsePacketInner got_srv_msg[2], sn:"+sn+", payload:"+payload+", result:"+result+", serviceid:"+srvpacket.get_service_id());
            }

            if (srvpacket.get_result() != 0) {
                GPLogger.w(TAG, "service result error: " + srvpacket.get_result());
                results.add(convertToEmtypResult(sn, result, payload, ""));
                return results;
            }

            if (srvpacket.get_service_id() != GROUPCHAT_SRV_ID) {
                GPLogger.w(TAG, "unsupported service_id: " + srvpacket.get_service_id());
                return null;
            }

            return this.parseGroupchatPacket(sn, srvpacket.get_data(), true, allreqids, okreqids);

        } else if (packet.getAction() == Packet.ACTION_NOTIFICATION) {

            NotificationPacket npacket = (NotificationPacket) packet;
            if (npacket.get_info_type() != null && npacket.get_info_type().equals(INFO_TYPE_GROUPCHAT)) {
                SrvMsgPacket srvpacket1 = new SrvMsgPacket(0, GROUPCHAT_SRV_ID, 0, npacket.get_info_content());
                if(JhFlag.enableDebug()) {
                    GPLogger.d(TAG, "parsePacketInner notification[6], info_type:group");
                }
                return parsePacketInner(srvpacket1, allreqids, okreqids);
            }

        }

        return null;
    }

    private List<Result> parseGroupchatPacket(long sn, byte[] data, boolean valid_msg, List<Long> allreqids, List<Long> okreqids) {

        if (data == null) {
            return null;
        }

        int result = -1;

        String reason = null;
        int sleep = 0;

        List<Result> results = new ArrayList<Result>();

        GroupChatProto.GroupDownPacket groupdp = null;

        try {
            groupdp = GroupChatProto.GroupDownPacket.parseFrom(data);
        } catch (InvalidProtocolBufferMicroException e) {
            GPLogger.e(TAG, Log.getStackTraceString(e));
        }

        if ( groupdp == null ) {
            return null;
        }

        if (groupdp.hasReason() && groupdp.getReason() != null) {
            reason = groupdp.getReason();
        }

        result = groupdp.getResult();

        if (groupdp.hasSleep()) {
            sleep = groupdp.getSleep();
        }

        if(JhFlag.enableDebug()) {
            GPLogger.d(TAG, "parseGroupchatPacket sn:"+sn+", payloadtype:"+groupdp.getPayload());
        }

        if (result != 0) {
            GPLogger.w(TAG, "group result error: " + result);
            results.add(convertToEmtypResult(sn, result, groupdp.getPayload(), ""));
            return results;
        }

        switch (groupdp.getPayload()) {

            case Result.PAYLOAD_GETGROUPMSGS:
                setFlowCtlAbsMills(System.currentTimeMillis()+sleep*1000, true);

                if(JhFlag.enableDebug()) {
                    GPLogger.d(TAG, "getmsgs result group num:" + groupdp.getGetmsgrespCount());
                }

                if (groupdp.getGetmsgrespCount() <= 0) {
                    return null;
                }

                for ( GroupChatProto.GroupMessageResp resp : groupdp.getGetmsgrespList()) {
                    try {
                        addResult(results, new GetmsgResult(sn, result, reason, Result.PAYLOAD_GETGROUPMSGS, resp));
                        String reqids = resp.getTraceid();
                        String[] parts = reqids.split(",");
                        for ( String reqid : parts ) {
                            okreqids.add(Long.valueOf(reqid));
                        }
                    } catch (Exception e) {
                        e.printStackTrace();
                    }
                }

                return results;

            case Result.PAYLOAD_GROUPSYNC:
                setFlowCtlAbsMills(System.currentTimeMillis()+sleep*1000, true);

                if(JhFlag.enableDebug()) {
                    GPLogger.d(TAG, "sync result group num:" + groupdp.getSyncrespCount());
                }

                if (groupdp.getSyncrespCount() <= 0) {
                    return null;
                }

                for ( GroupChatProto.GroupInfo resp : groupdp.getSyncrespList()) {
                    try {
                        addResult(results, new SyncResult(sn, result, reason, Result.PAYLOAD_GROUPSYNC, resp));
                    } catch (Exception e) {
                        e.printStackTrace();
                    }
                }

                okreqids.addAll(allreqids);
                return results;

            case Result.PAYLOAD_NEWMSG_NOTIFY:
                if (groupdp.getNewmsgnotifyCount() <= 0) {
                    return null;
                }

                for ( GroupChatProto.GroupNotify notify : groupdp.getNewmsgnotifyList()) {

                    try {
                        addResult(results, new GroupNotifyResult(sn, result, reason, Result.PAYLOAD_NEWMSG_NOTIFY, notify));
                    } catch (Exception e) {
                        e.printStackTrace();
                    }
                }

                return results;

            default:
                GPLogger.w(TAG, "unknown data");
                return null;
        }

    }

    private Result convertToEmtypResult(long sn, int result, int payload, String reason) {
        switch (payload) {
            case Result.PAYLOAD_GETGROUPMSGS:
                return new GetmsgResult(sn, result, reason, payload, null);

            case Result.PAYLOAD_GROUPSYNC:
                return new SyncResult(sn, result, reason, payload, null);

            case Result.PAYLOAD_NEWMSG_NOTIFY:
                return new GroupNotifyResult(sn, result, reason, payload, null);
        }

        return new Result(sn, result, payload, reason);
    }

    private void addResult(List<Result> results, Result result) {
        if (result != null) {
            results.add(result);
        }
    }
}
