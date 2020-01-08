package com.huajiao.comm.chatroom;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.TreeMap;

import android.annotation.SuppressLint;
import android.content.Context;
import android.util.Log;

import com.google.protobuf.micro.ByteStringMicro;
import com.google.protobuf.micro.InvalidProtocolBufferMicroException;
import com.huajiao.comm.chatroomresults.InComingMessage;
import com.huajiao.comm.chatroomresults.JoinResult;
import com.huajiao.comm.chatroomresults.MemberJoinedinNotification;
import com.huajiao.comm.chatroomresults.MemberQuitNotification;
import com.huajiao.comm.chatroomresults.QueryResult;
import com.huajiao.comm.chatroomresults.QuitResult;
import com.huajiao.comm.chatroomresults.Result;
import com.huajiao.comm.chatroomresults.SubscribeResult;
import com.huajiao.comm.chatroomresults.UserInfo;
import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.Utils;
import com.huajiao.comm.im.api.ILongLiveConn;
import com.huajiao.comm.im.api.LongLiveConnFactory;
import com.huajiao.comm.im.packet.MsgPacket;
import com.huajiao.comm.im.packet.MsgResultPacket;
import com.huajiao.comm.im.packet.NotificationPacket;
import com.huajiao.comm.im.packet.Packet;
import com.huajiao.comm.im.packet.SrvMsgPacket;
import com.huajiao.comm.im.packet.StateChangedPacket;
import com.huajiao.comm.monitor.MessageMonitor;
import com.huajiao.comm.protobuf.ChatRoomProto;
import com.huajiao.comm.protobuf.ChatRoomProto.CRPair;
import com.huajiao.comm.protobuf.ChatRoomProto.CRUser;
import com.huajiao.comm.protobuf.ChatRoomProto.ChatRoom;
import com.huajiao.comm.protobuf.ChatRoomProto.ChatRoomDownToUser;
import com.huajiao.comm.protobuf.ChatRoomProto.ChatRoomMNotify;
import com.huajiao.comm.protobuf.ChatRoomProto.ChatRoomNewMsg;
import com.huajiao.comm.protobuf.ChatRoomProto.ChatRoomPacket;
import com.huajiao.comm.protobuf.ChatRoomProto.MemberJoinChatRoomNotify;
import com.huajiao.comm.protobuf.ChatRoomProto.MemberQuitChatRoomNotify;

public class ChatroomHelper implements IChatroomHelper {

	private static final String TAG = "CRH";//"ChatroomHelper";

	private static final int CHATROOM_SRV_ID = 10000006;

	private static final int PAYLOAD_QUERY = 101;
	private static final int PAYLOAD_JOIN = 102;
	private static final int PAYLOAD_QUIT = 103;
	private static final int PAYLOAD_SUBSCRIBE = 109;

	private static final int PAYLOAD_MEM_JOINED_IN_NOTIFY = 1001;

	private static final int PAYLOAD_MEM_QUIT_NOTIFY = 1002;

	// * 1003 -- multinotify
	private static final int PAYLOAD_COMPRESSED_NOTIFY = 1003;

	/** 是否过滤自己发送的消息 */
	private static boolean FILTER_SELF_SENT_MSG = false;

	private static final String PROPERTY_MEM_CNT = "memcount";
	private static final String INFO_TYPE_CHATROOM = "chatroom";

	/** 长链接已经被关闭了， 无法使用方法， 需要重新new一个实例 */
	public static final long ERR_LLC_HAS_BEEN_SHUTDOWN = -2;

	private boolean _has_shutdown = false;

	private boolean _pull_lost = true;
	
	private volatile static boolean sAllowSaveCrMsgLog = true; // 是否允许写入聊天室消息日志	云控控制
	
	public static void setAllowSaveCrMsgLog(boolean allowSaveCrMsgLog) {
		if(sAllowSaveCrMsgLog != allowSaveCrMsgLog) {
			CRLogger.d(TAG, String.format(Locale.getDefault(), "switch save_cr_msg %d -> %d", sAllowSaveCrMsgLog?1:0, allowSaveCrMsgLog?1:0));
			sAllowSaveCrMsgLog = allowSaveCrMsgLog;
		}
	}
	
	public static boolean isAllowSaveCrMsgLog() {
		return sAllowSaveCrMsgLog;
	}

	/**
	 * 用于关联sn结果， 对于没有网络或者超时的情况， 无法知道对应的的请求， 所以需要关联一下
	 * */
	@SuppressLint("UseSparseArrays")
	private HashMap<Long, Integer> _pending_actions = new HashMap<Long, Integer>();

	private static final int PAYLOAD_NEW_MSG = 1000;

	private ILongLiveConn _llc;
	private ClientConfig _clientConfig;
	private AccountInfo _accountInfo;

	private MessageMonitor _mm;

	private void putAction(long sn, int payload) {
		synchronized (_pending_actions) {
			_pending_actions.put(Long.valueOf(sn), Integer.valueOf(payload));
		}
	}

	private int retrieveAction(long sn) {
		synchronized (_pending_actions) {
			Long key = Long.valueOf(sn);
			if (_pending_actions.containsKey(key)) {
				Integer v = _pending_actions.get(key);
				_pending_actions.remove(key);
				return v.intValue();
			}
			return -1;
		}
	}

	private Result convertToEmtypResult(long sn, int result, int payload, String roomid) {
		byte[] reason = null;
		switch (payload) {
		case PAYLOAD_QUERY:
			return new QueryResult(sn, result, reason, 0, roomid, null, null);

		case PAYLOAD_JOIN:
			return new JoinResult(sn, result, reason, reason, 0, roomid, null, null);

		case PAYLOAD_QUIT:
			return new QuitResult(sn, result, reason);

		case PAYLOAD_SUBSCRIBE:
			return new SubscribeResult(sn, result, reason, null, false);
		}

		return null;
	}

	/**
	 * 加入聊天室
	 * 
	 * @param context
	 * @param accountInfo
	 *            账号信息
	 * @param clientConfig
	 *            客户端配置信息
	 * @param callerInfo
	 *            调用者信息，格式：processName_processId_version_..., 例如com.huajiao_1234_4.2.4.1018, 后面可以跟一些
	 *            其他必要的信息，具体由客户端上层调用决定，以区分是客户端还是第三方使用了花椒sdk的app调用的
	 * @throws IllegalArgumentException
	 */
	public ChatroomHelper(Context context, AccountInfo accountInfo, ClientConfig clientConfig, String callerInfo,
						  CRLogger.LoggerWriterCallback logWriteCallback) {
		if (context == null || accountInfo == null || clientConfig == null) {
			throw new IllegalArgumentException();
		}

		CRLogger.d(TAG, String.format("Create ChatRoomHelper %s", accountInfo.get_account()));
		CRLogger.d(TAG, String.format("%s", callerInfo));
		
		CRLogger.setUid(accountInfo.get_account());
		if(logWriteCallback != null) {
			CRLogger.InitWriteCallBack(logWriteCallback);
		}

		_llc = LongLiveConnFactory.create(context, accountInfo, clientConfig);
		_clientConfig = clientConfig;
		_accountInfo = accountInfo;
		_mm = new MessageMonitor(this, accountInfo.get_account());
	}

	public void setOpenForeground(boolean open) {
		if(_llc != null) {
			_llc.setOpenForeground(open);
		}
	}

	/**
	 * subscribe/unsubscribe聊天室
	 *
	 * @param roomId 聊天室id
	 * @param subscribe subscribe or cancel
	 * @return 大于0: 消息sn, 负数： 失败
	 * @throws IllegalArgumentException
	 * */
	public long subscribeChatroom(String roomId, boolean subscribe) {

		if (_has_shutdown) {
			CRLogger.d(TAG, String.format("subscribe room %s llc has shut down", roomId));
			return ERR_LLC_HAS_BEEN_SHUTDOWN;
		}

		if (roomId == null || roomId.length() == 0) {
			throw new IllegalArgumentException();
		}

		long sn = _llc.get_sn();
		ByteStringMicro broomid = ByteStringMicro.copyFromUtf8(roomId);
		ChatRoomProto.SubscribeRequest subreq = new ChatRoomProto.SubscribeRequest();
		subreq.setRoomid(broomid).setSub(subscribe);

		ChatRoomProto.ChatRoomUpToServer up2server = new ChatRoomProto.ChatRoomUpToServer();
		up2server.setSubreq(subreq);
		up2server.setPayloadtype(PAYLOAD_SUBSCRIBE);

		ChatRoomProto.ChatRoomPacket packet = new ChatRoomProto.ChatRoomPacket();
		packet.setAppid(_clientConfig.getAppId());
		packet.setClientSn(_llc.get_sn());
		packet.setToServerData(up2server);
		packet.setRoomid(broomid);

		if (_llc.send_service_message(CHATROOM_SRV_ID, sn, packet.toByteArray())) {
			putAction(sn, PAYLOAD_SUBSCRIBE);
			CRLogger.i(TAG, String.format(Locale.US, "subscribe %s ok, %d, %d", roomId, sn, isAllowSaveCrMsgLog()?1:0));
			return sn;
		} else {
			CRLogger.e(TAG, String.format(Locale.US, "subscribe %s failed", roomId));
			return -1;
		}
	}

	/**
	 * 加入聊天室
	 * 
	 * @param roomId  聊天室id
	 * @return 大于0: 消息sn, 负数： 失败
	 * @throws IllegalArgumentException
	 * */
	public long joinChatroom(String roomId) {
		return joinChatroom(roomId, null);
	}

	/**
	 * 加入聊天室
	 *
	 * @param roomId 聊天室id
	 * @param roomProps 聊天室参数
	 * @return 大于0: 消息sn, 负数： 失败
	 * @throws IllegalArgumentException
	 * */
	public long joinChatroom(String roomId, TreeMap<String, String> roomProps) {
		CRLogger.i(TAG,"start joinChatroom roomId："+ roomId);
		if (_has_shutdown) {
			CRLogger.d(TAG, String.format("join room %s llc has shut down", roomId));
			return ERR_LLC_HAS_BEEN_SHUTDOWN;
		}

		if (roomId == null || roomId.length() == 0) {
			throw new IllegalArgumentException();
		}
		
		_mm.onJoinedIn(roomId);

		long sn = _llc.get_sn();
		StringBuilder propsb = new StringBuilder();

		ByteStringMicro broomid = ByteStringMicro.copyFromUtf8(roomId);
		ChatRoomProto.ApplyJoinChatRoomRequest joinreq = new ChatRoomProto.ApplyJoinChatRoomRequest();

		ChatRoomProto.ChatRoom chatroom = new ChatRoomProto.ChatRoom();
		chatroom.setRoomid(broomid);
		if ( roomProps != null ) {
			for (Map.Entry<String, String> prop : roomProps.entrySet()) {
				propsb.append(prop.getKey()).append(":").append(prop.getValue()).append(";");
				ChatRoomProto.CRPair crpair = new ChatRoomProto.CRPair();
				crpair.setKey(prop.getKey());
				crpair.setValue(ByteStringMicro.copyFromUtf8(prop.getValue()));
				chatroom.addProperties(crpair);
			}
		}
		joinreq.setRoomid(broomid).setNoUserlist(true).setRoom(chatroom);

		ChatRoomProto.ChatRoomUpToServer up2server = new ChatRoomProto.ChatRoomUpToServer();
		up2server.setApplyjoinchatroomreq(joinreq);
		up2server.setPayloadtype(PAYLOAD_JOIN);

		ChatRoomProto.ChatRoomPacket packet = new ChatRoomProto.ChatRoomPacket();
		packet.setAppid(_clientConfig.getAppId());
		packet.setClientSn(_llc.get_sn());
		packet.setToServerData(up2server);
		packet.setRoomid(broomid);

		if (_llc.send_service_message(CHATROOM_SRV_ID, sn, packet.toByteArray())) {
			putAction(sn, PAYLOAD_JOIN);
			CRLogger.i(TAG,String.format(Locale.getDefault(), "send_service_message send join packet sn:%d,roomid:%s",sn,roomId));
			CRLogger.i(TAG, String.format(Locale.US, "join %s ok, %d, %d, %s", roomId, sn, isAllowSaveCrMsgLog()?1:0, propsb.toString()));
			return sn;
		} else {
			CRLogger.i(TAG,String.format(Locale.getDefault(), "send_service_message send join failed sn:%d,roomid:%s",sn,roomId));
			CRLogger.e(TAG, String.format(Locale.US, "join %s failed", roomId));
			return -1;
		}
	}

	/**
	 * 退出聊天室
	 * 
	 * @param roomId
	 *            聊天室id
	 * @return 大于0: 消息sn, 负数： 失败
	 * */
	public long quitChatroom(String roomId) {

		if (_has_shutdown) {
			CRLogger.d(TAG, String.format("quit room %s llc has shut down", roomId));
			return ERR_LLC_HAS_BEEN_SHUTDOWN;
		}

		if (roomId == null || roomId.length() == 0) {
			throw new IllegalArgumentException();
		}

		_mm.onQuit(roomId);

		long sn = _llc.get_sn();

		ByteStringMicro broomid = ByteStringMicro.copyFromUtf8(roomId);
		ChatRoomProto.QuitChatRoomRequest quitreq = new ChatRoomProto.QuitChatRoomRequest();
		quitreq.setRoomid(broomid);

		ChatRoomProto.ChatRoomUpToServer up2server = new ChatRoomProto.ChatRoomUpToServer();
		up2server.setQuitchatroomreq(quitreq);
		up2server.setPayloadtype(PAYLOAD_QUIT);

		ChatRoomProto.ChatRoomPacket packet = new ChatRoomProto.ChatRoomPacket();
		packet.setAppid(_clientConfig.getAppId());
		packet.setClientSn(_llc.get_sn());
		packet.setToServerData(up2server);
		packet.setRoomid(broomid);

		if (_llc.send_service_message(CHATROOM_SRV_ID, sn, packet.toByteArray())) {
			putAction(sn, PAYLOAD_QUIT);
			CRLogger.i(TAG, String.format(Locale.US, "quit %s ok, %d", roomId, sn));
			return sn;
		} else {
			CRLogger.i(TAG, String.format(Locale.US, "quit %s failed", roomId));
			return -1;
		}
	}

	/**
	 * 查询聊天室
	 * 
	 * @param roomId
	 *            聊天室id
	 * @param start_index
	 *            成员开始id, 从1开始
	 * @param count
	 *            最多返回的成员数
	 * @return 大于0: 消息sn, 负数： 失败
	 */
	public long queryChatroom(String roomId, int start_index, int count) {

		if (_has_shutdown) {
			return ERR_LLC_HAS_BEEN_SHUTDOWN;
		}

		if (roomId == null || roomId.length() == 0) {
			throw new IllegalArgumentException();
		}

		long sn = _llc.get_sn();
		ByteStringMicro broomid = ByteStringMicro.copyFromUtf8(roomId);

		ChatRoomProto.GetChatRoomDetailRequest queryreq = new ChatRoomProto.GetChatRoomDetailRequest();
		queryreq.setIndex(start_index);
		queryreq.setOffset(count);
		queryreq.setRoomid(broomid);

		ChatRoomProto.ChatRoomUpToServer up2server = new ChatRoomProto.ChatRoomUpToServer();
		up2server.setGetchatroominforeq(queryreq);
		up2server.setPayloadtype(PAYLOAD_QUERY);

		ChatRoomProto.ChatRoomPacket packet = new ChatRoomProto.ChatRoomPacket();
		packet.setAppid(_clientConfig.getAppId());
		packet.setClientSn(_llc.get_sn());
		packet.setToServerData(up2server);
		packet.setRoomid(broomid);

		if (_llc.send_service_message(CHATROOM_SRV_ID, sn, packet.toByteArray())) {
			putAction(sn, PAYLOAD_QUERY);
			return sn;
		} else {
			return -1;
		}
	}

	/**
	 * 关闭长连接资源, 将无法收到通知, 同时需要用对应方法时， 需要重新new一个对象
	 * */
	public void shutdownLongLiveConn() {
		_llc.shutdown();
		_has_shutdown = true;
	}

	public List<Result> parsePacket(Packet packet) {
		
		// CRLogger.d(TAG, String.format("parse packet %d", packet.hashCode()));
		
		List<Result> results = null;
		
		try {
			results = parsePacketInner(packet);
		} catch (Exception e) {
			e.printStackTrace();
		}

		/*if (null != results) {
			CRLogger.d(TAG, String.format("results num %d", results.size()));
			for(int i=0;i<results.size();i++) {
				Result re = results.get(i);
				if (re.get_payload_type() == Result.PAYLOAD_JOIN_RESULT) {
					try {
						
						JoinResult jr = (JoinResult) re;
						String partner_str = "null";
						if (jr.get_partnerdata() != null && jr.get_partnerdata().length > 0) {
							partner_str = new String(jr.get_partnerdata());
						}
						CRLogger.i(TAG, String.format(Locale.US, "sn %d, result %d, partner data %s", jr.get_sn(), jr.get_result(), partner_str));
					} catch (Throwable tr) {
						tr.printStackTrace();
					}
				} else if(re.get_payload_type() == Result.PAYLOAD_INCOMING_MESSAGE) {
					CRLogger.d(TAG, String.format("%d: %d", i, re.get_payload_type()));
					InComingMessage msg = (InComingMessage) re;
					String text = CRPacketHelper.parseMsg(new String(msg.get_content()));
					if(text != null) {
						CRLogger.i(TAG, text);
					} 
				} // else if
			}
		} // if
		*/

		return results;
	}

	/**
	 * 处理返回的包
	 * 
	 * @param packet
	 *            service 发送过来的包
	 * @return 返回空 如果是不相关的包， 否则返回对应的Result
	 * */
	private List<Result> parsePacketInner(Packet packet) {

		if (packet == null) {
			return null;
		}

		int result = -1;
		long sn = -1;
		int payload;
		List<Result> results = new ArrayList<Result>();
		
		if (packet.getAction() == Packet.ACTION_GOT_SRV_MSG) {

			SrvMsgPacket srvpacket = (SrvMsgPacket) packet;

			sn = srvpacket.get_sn();
			payload = retrieveAction(sn);
			result = srvpacket.get_result();			
			
			if (srvpacket.get_result() != 0) {
				results.add(convertToEmtypResult(sn, result, payload, null));
				return results;
			}

			if (srvpacket.get_service_id() != CHATROOM_SRV_ID) {
				CRLogger.w(TAG, "unsupported service_id: " + srvpacket.get_service_id());
				return null;
			}
			return this.parseChatroomPacket(sn, srvpacket.get_data(), true);

		} else if (packet.getAction() == Packet.ACTION_NOTIFICATION) {

			NotificationPacket npacket = (NotificationPacket) packet;
			if (npacket.get_info_type() != null && npacket.get_info_type().equals(INFO_TYPE_CHATROOM)) {
				SrvMsgPacket srvpacket1 = new SrvMsgPacket(0, CHATROOM_SRV_ID, 0, npacket.get_info_content());
				return parsePacket(srvpacket1);
			}

		} else if (packet.getAction() == Packet.ACTION_GOT_MSG_RESULT) {

			MsgResultPacket msg_result = (MsgResultPacket) packet;
			sn = msg_result.get_sn();
			payload = retrieveAction(sn);
			if (payload != -1) {
				results.add(convertToEmtypResult(sn, msg_result.get_result(), payload, null));
				return results;
			}

		} else if (packet.getAction() == Packet.ACTION_GOT_MSG) {
			MsgPacket msg_packet = (MsgPacket) packet;
			sn = msg_packet.get_sn();
			payload = retrieveAction(sn);
			if (msg_packet.get_info_type() != null && msg_packet.get_info_type().equals(INFO_TYPE_CHATROOM)) {
				InComingMessage msg = parseChatroomMessage(msg_packet);
				if (msg != null) {
					// 还没发现没运行到这里
					// CRLogger.d(TAG, "GotMsg:"+new String(msg.get_content()));
					if (_mm.onMessage(msg.get_id(), msg.get_max_id(), msg.get_sent_time(), msg_packet.is_valid(), true, _pull_lost)) {
						results.add(msg);
					}
					return results;
				}
			}
		} else if (packet.getAction() == Packet.ACTION_STATE_CHANGED) {
			StateChangedPacket sp = (StateChangedPacket) packet;
			if (sp != null && sp.get_newState() != null) {
				_mm.onStateChanged(sp.get_newState());
			}
		}

		return null;
	}

	private InComingMessage parseChatroomMessage(MsgPacket msg_packet) {

		if (msg_packet == null || msg_packet.get_content() == null || msg_packet.get_content().length == 0) {
			return null;
		}

		ChatRoomNewMsg newMsgNotification = null;
		try {
			newMsgNotification = ChatRoomNewMsg.parseFrom(msg_packet.get_content());
		} catch (InvalidProtocolBufferMicroException e) {
			e.printStackTrace();
		}

		if (newMsgNotification != null) {
			return new InComingMessage(msg_packet.get_sn(), msg_packet.is_valid(), newMsgNotification);
		}

		return null;
	}

	private List<Result> parseChatroomPacket(long sn, byte[] data, boolean valid_msg) {

		if (data == null) {
			return null;
		}

		int result = -1;
		int total_cnt = 0;

		String roomid = null;
		String roomname = null;

		byte[] reason = null;
		byte[] partnerdata = null;

		List<String> members = new ArrayList<String>();
		List<Result> results = new ArrayList<Result>();

		ChatRoomPacket crpacket = null;

		try {
			crpacket = ChatRoomPacket.parseFrom(data);
		} catch (InvalidProtocolBufferMicroException e) {
			CRLogger.e(TAG, Log.getStackTraceString(e));
		}

		if (crpacket == null || !crpacket.hasToUserData() || crpacket.getToUserData() == null) {
			return null;
		}

		ChatRoomDownToUser touser = crpacket.getToUserData();

		if (touser.hasReason() && touser.getReason() != null) {
			reason = touser.getReason().toByteArray();
		}

		result = touser.getResult();
		if (crpacket.hasRoomid() && crpacket.getRoomid() != null) {
			roomid = crpacket.getRoomid().toStringUtf8();
		}
		
		switch (touser.getPayloadtype()) {

		case PAYLOAD_COMPRESSED_NOTIFY:

			if (touser.getMultinotifyCount() <= 0) {
				return null;
			}

			for (ChatRoomMNotify mnote : touser.getMultinotifyList()) {
				if (!mnote.hasData()) {
					continue;
				}

				int all_user_cnt = mnote.getMemcount();
				int enrolled_user_cnt = mnote.getRegmemcount();

				byte[] original_data = mnote.getData().toByteArray();
				if (original_data == null || original_data.length <= 0) {
					continue;
				}

				byte[] ungzipped_data = Utils.ungzip(original_data);

				try {

					switch (mnote.getType()) {

					case PAYLOAD_NEW_MSG:
						ChatRoomNewMsg newMsgNotification = ChatRoomNewMsg.parseFrom(ungzipped_data);
						addResult(results, parseNewMsgNotification(sn, newMsgNotification, enrolled_user_cnt, all_user_cnt, true, true));
						break;

					case PAYLOAD_MEM_JOINED_IN_NOTIFY:
						MemberJoinChatRoomNotify joinNotification = MemberJoinChatRoomNotify.parseFrom(ungzipped_data);
						addResult(results, parseJoininNotification(sn, result, reason, joinNotification));
						break;

					case PAYLOAD_MEM_QUIT_NOTIFY:
						MemberQuitChatRoomNotify quitNotification = MemberQuitChatRoomNotify.parseFrom(ungzipped_data);
						addResult(results, parseQuitNotification(sn, result, reason, quitNotification));
						break;

					default:
						CRLogger.d(TAG, "unknow mnote type >> " + mnote.getType());
						break;

					}

				} catch (Exception e) {
					e.printStackTrace();
					continue;
				}

			}

			return results;

		case PAYLOAD_JOIN:

			if (touser.hasApplyjoinchatroomresp() && touser.getApplyjoinchatroomresp() != null) {

				if (touser.getApplyjoinchatroomresp().hasRoom() && touser.getApplyjoinchatroomresp().getRoom() != null) {
					ChatRoom room = touser.getApplyjoinchatroomresp().getRoom();
					roomid = room.getRoomid().toStringUtf8();
					roomname = room.getName();
					
					// if(JhFlag.enableDebug()) {
						CRLogger.d(TAG, String.format("join room %s hasPartnerdata [%s]", roomid, room.hasPartnerdata() ? "true" : "false"));
					// }
					
					for (CRUser user : room.getMembersList()) {
						if (user.getUserid() == null) {
							continue;
						}
						String uid = user.getUserid().toStringUtf8();
						if (uid.equals(_accountInfo.get_account())) {
							continue;
						}
						members.add(uid);
					}					

					if (room.hasPartnerdata() && room.getPartnerdata() != null) {
						partnerdata = room.getPartnerdata().toByteArray();
						if(JhFlag.enableDebug()) {
							CRLogger.d(TAG, String.format("join room %s Partnerdata [%s]", roomid, room.getPartnerdata().toStringUtf8()));
						}
					}

					for (CRPair pair : room.getPropertiesList()) {
						if (pair.getKey() != null && pair.getKey().equals(PROPERTY_MEM_CNT)) {
							String v = pair.getValue().toStringUtf8();
							try {
								total_cnt = Integer.parseInt(v);
								break;
							} catch (NumberFormatException e) {
							}
						}
					}
				}

				if (touser.getApplyjoinchatroomresp().hasPullLost()){
					_pull_lost = touser.getApplyjoinchatroomresp().getPullLost();
				}
			} else {
				CRLogger.w(TAG, "Apply join resp incomplete.");
			}

			if (result == 0) {
				CRLogger.d(TAG, String.format(Locale.CHINA,"on join %s %d, pull_lost [%b]", roomid, sn, _pull_lost));
				// _mm.onJoinedIn(roomid);
			}

			addResult(results, new JoinResult(sn, result, reason, partnerdata, total_cnt, roomid, roomname, listToArray(members)));

			return results;

		case PAYLOAD_QUIT:
			if (!touser.hasQuitchatroomresp() || touser.getQuitchatroomresp() == null) {
				CRLogger.w(TAG, "Quit resp incomplete.");
			}

			if (result == 0) {
				CRLogger.d(TAG, String.format(Locale.getDefault(), "on quit %s %d", roomid, sn));
				// _mm.onQuit(roomid);
			}
			addResult(results, new QuitResult(sn, result, reason));
			return results;

		case PAYLOAD_SUBSCRIBE:
				if (!touser.hasSubresp() || touser.getSubresp() == null) {
					CRLogger.w(TAG, "subscribe resp incomplete.");
				}

				if (result == 0) {
					CRLogger.d(TAG, String.format(Locale.getDefault(), "on subscribe %s %d", roomid, sn));
				}

				addResult(results, new SubscribeResult(sn, result, reason, touser.getSubresp().getRoomid().toStringUtf8(), touser.getSubresp().getSub()));
				return results;

		case PAYLOAD_QUERY:

			if (touser.hasGetchatroominforesp() && touser.getGetchatroominforesp() != null) {
				if (touser.getGetchatroominforesp().hasRoom() && touser.getGetchatroominforesp().getRoom() != null) {
					ChatRoom room = touser.getGetchatroominforesp().getRoom();
					roomid = room.getRoomid().toStringUtf8();
					roomname = room.getName();

					for (CRUser user : room.getMembersList()) {
						members.add(user.getUserid().toStringUtf8());
					}

					for (CRPair pair : room.getPropertiesList()) {
						if (pair.getKey() != null && pair.getKey().equals(PROPERTY_MEM_CNT)) {
							String v = pair.getValue().toStringUtf8();
							try {
								total_cnt = Integer.parseInt(v);
								break;
							} catch (NumberFormatException e) {
							}
						}
					}
				}
			} else {
				CRLogger.w(TAG, "query resp incomplete.");
			}

			addResult(results, new QueryResult(sn, result, reason, total_cnt, roomid, roomname, listToArray(members)));
			return results;

		case PAYLOAD_NEW_MSG:

			addResult(results, parseNewMsgNotification(sn, touser.getNewmsgnotify(), 0, 0, false, true));
			return results;

		case PAYLOAD_MEM_JOINED_IN_NOTIFY:

			addResult(results, parseJoininNotification(sn, result, reason, touser.getMemberjoinnotify()));
			return results;

		case PAYLOAD_MEM_QUIT_NOTIFY:

			addResult(results, parseQuitNotification(sn, result, reason, touser.getMemberquitnotify()));
			return results;

		default:
			CRLogger.w(TAG, "unknown data");
			return null;
		}

	}

	/**
	 * @return 获取IMSDK长连接接口
	 */
	public ILongLiveConn get_llc() {
		return _llc;
	}

	/**
	 * @return 是否已经关闭了长连接, 如果已经关闭， 实例的其他方法将无法使用
	 */
	public boolean is_llc_shutdown() {
		return _has_shutdown;
	}
	
	// 特殊消息roomid不用过滤 当作白名单走绿色通道
	private boolean isWhiteRoomid(String roomid) {
		String worldGiftRoomid = String.valueOf(_clientConfig.getAppId()); // 世界礼物
		if(worldGiftRoomid.equals(roomid)) {
			return true;
		}
		return false;
	}

	private InComingMessage parseNewMsgNotification(long sn, ChatRoomNewMsg newMsg, int enrolled_user_cnt, int mem_cnt, boolean overrideMemCnt, boolean valid) {

		if (newMsg == null) {
			CRLogger.d(TAG, "new msg notification is null!!!");
			return null;
		}
		
		String roomid = "";
		String senderid = "";
		int id = 0;
		byte[] content = null;
		boolean report_msg = true;

		if (newMsg.getRoomid() != null) {
			roomid = newMsg.getRoomid().toStringUtf8();
		}

		if (newMsg.hasMsgid()) {
			id = newMsg.getMsgid();
		}
		
		// 非特殊消息需要过滤roomid
		if(!isWhiteRoomid(roomid)) {
			// 如果不是当前房间的消息：抛弃不处理		
			if(!_mm.isMyRoomMsg(roomid)) {
				if(JhFlag.enableDebug()) {
					// CRLogger.d(TAG, String.format("invalid msg %s %d %d", roomid, newMsg.getMsgid(), newMsg.getMaxid()));
				}
				return null;
			}
		}

		if (newMsg.getSender() != null && newMsg.getSender().getUserid() != null) {
			senderid = newMsg.getSender().getUserid().toStringUtf8();
			if (!overrideMemCnt) {
				mem_cnt = newMsg.getMemcount();
				enrolled_user_cnt = newMsg.getRegmemcount();
			}

			if (FILTER_SELF_SENT_MSG && senderid.equals(_accountInfo.get_account())) {
				CRLogger.d(TAG, "filter message sent by self.");
				return null;
			}
		}

		if (newMsg.hasMsgcontent() && newMsg.getMsgcontent().size() > 0) {
			content = newMsg.getMsgcontent().toByteArray();
			if(JhFlag.enableDebug()) {
				CRLogger.d(TAG, String.format(Locale.getDefault(), "parseNewMsgNotification >> roomid %s msgid %d maxmsgid %s content %s",
						roomid, newMsg.getMsgid(), newMsg.getMaxid(), newMsg.getMsgcontent().toStringUtf8()));
			}					
		} else {
			CRLogger.d(TAG, "new msg has no content");
		}

		// join in and other trivial message has no id
		if (newMsg.hasMsgid()) {			
			report_msg = _mm.onMessage(id, newMsg.getMaxid(), newMsg.getTimestamp(), valid, false, _pull_lost);
			String text = CRPacketHelper.parseMsg(newMsg.getMsgcontent().toStringUtf8());
			if(text != null) {
				if(isAllowSaveCrMsgLog()) {
					CRLogger.d(TAG, String.format("%s %s", report_msg ? "" : "*", text));
				}
			}
		} else {			
			// CRLogger.d(TAG, "new msg has no msgid");
		}

		if (!report_msg) {
			CRLogger.d(TAG, "ignore dup or invalid message.");
			return null;
		}

		return new InComingMessage(sn, id, roomid, senderid, newMsg.getMsgtype(), content, enrolled_user_cnt, mem_cnt, newMsg.getMaxid(),
				newMsg.getTimestamp(), true);
	}

	private MemberJoinedinNotification parseJoininNotification(long sn, int result, byte[] reason, MemberJoinChatRoomNotify notification) {

		String roomid = "";
		String roomname = "";
		int count = 0;

		List<UserInfo> members = new ArrayList<UserInfo>();

		if (null == notification) {
			CRLogger.w(TAG, "member join notify is null");
			return new MemberJoinedinNotification(sn, result, reason, count, roomid, roomname, members);
		}

		if (notification.getRoom() != null) {

			ChatRoom room = notification.getRoom();
			roomid = room.getRoomid().toStringUtf8();
			roomname = room.getName();

			for (CRUser user : room.getMembersList()) {
				byte[] user_data = null;
				if (user.getUserdata() != null) {
					user_data = user.getUserdata().toByteArray();
				}
				members.add(new UserInfo(user.getUserid().toStringUtf8(), user_data));
			}

			count = room.getMembersCount();
			if (room.getPropertiesCount() > 0) {
				for (CRPair pair : room.getPropertiesList()) {
					if (pair.getKey() != null && pair.getKey().equals(PROPERTY_MEM_CNT)) {
						String v = pair.getValue().toStringUtf8();
						try {
							count = Integer.parseInt(v);
							break;
						} catch (NumberFormatException e) {
						}
					}
				}
			}

			// avoid empty join notification?
			if (members.size() == 0) {
				return null;
			}
		} else {
			CRLogger.w(TAG, "member join notify is incomplete");
		}

		return new MemberJoinedinNotification(sn, result, reason, count, roomid, roomname, members);
	}

	private MemberQuitNotification parseQuitNotification(long sn, int result, byte[] reason, MemberQuitChatRoomNotify notification) {

		String roomid = "";
		String roomname = "";
		int count = 0;
		List<String> members = new ArrayList<String>();

		if (null == notification) {
			CRLogger.w(TAG, "member quit notify is null");
			return new MemberQuitNotification(sn, result, reason, count, roomid, roomname, listToArray(members));
		}

		if (notification.getRoom() != null) {

			ChatRoom room = notification.getRoom();
			roomid = room.getRoomid().toStringUtf8();
			roomname = room.getName();

			for (CRUser user : room.getMembersList()) {

				if (user.getUserid() == null) {
					continue;
				}

				String uid = user.getUserid().toStringUtf8();

				if (uid.equals(_accountInfo.get_account())) {
					continue;
				}
				members.add(uid);
			}
			count = room.getMembersCount();
			if (room.getPropertiesCount() > 0) {
				for (CRPair pair : room.getPropertiesList()) {
					if (pair.getKey() != null && pair.getKey().equals(PROPERTY_MEM_CNT)) {
						String v = pair.getValue().toStringUtf8();
						try {
							count = Integer.parseInt(v);
							break;
						} catch (NumberFormatException e) {

						}
					}
				}
			}

		} else {
			CRLogger.w(TAG, "member quit notify is incomplete");
		}

		return new MemberQuitNotification(sn, result, reason, count, roomid, roomname, listToArray(members));
	}

	private void addResult(List<Result> results, Result result) {
		if (result != null) {
			results.add(result);
		}
	}

	private String[] listToArray(List<String> list) {

		if (list == null || list.size() == 0) {
			return null;
		}

		String[] array = new String[list.size()];

		for (int i = 0; i < list.size(); i++) {
			array[i] = list.get(i);
		}

		return array;

	}

	@Override
	public boolean getMessage(String info_type, int[] ids, byte[] parameters) {

		boolean result = false;

		if (ids == null || ids.length == 0) {
			return false;
		}

		if (info_type == null || info_type.length() == 0) {
			return false;
		}

		try {
			if (_llc != null) {
				result = _llc.get_message(info_type, ids, parameters);
			}
		} catch (Throwable tr) {
			tr.printStackTrace();
		}

		return result;
	}

	@Override
	public long getServerTime() {
		long serverTime = -1;
		try {
			if (_llc != null) {
				// serverTime = _llc.get_server_time();
			}
		} catch (Throwable tr) {
			tr.printStackTrace();
		}

		return serverTime;
	}
}
