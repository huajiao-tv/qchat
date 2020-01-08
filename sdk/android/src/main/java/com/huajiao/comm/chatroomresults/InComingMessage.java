package com.huajiao.comm.chatroomresults;

import com.huajiao.comm.protobuf.ChatRoomProto.ChatRoomNewMsg;

/**
 * 收到的新消息消息
 * */
public class InComingMessage extends Result {

	private static final byte[] _reason = null;
	private String _roomid;
	private String _senderid = "";
	private int _msg_type;
	private int _id;
	private int _max_id;
	private long _sent_time;
	private byte[] _content = null;
	private int _registered_user_cnt;
	private int _total_member_cnt;
	private boolean _valid = true;

	/**
	 * @param sn
	 * @param roomid
	 * @param senderid
	 * @param msg_type
	 * @param content
	 */
	public InComingMessage(long sn, int id, String roomid, String senderid, int msg_type, byte[] content, int enrolled_user_cnt, int mem_cnt, int max_id,
			long sent_time, boolean valid) {
		super(sn, 0, PAYLOAD_INCOMING_MESSAGE, _reason);
		_roomid = roomid;
		_senderid = senderid;
		if (_senderid == null) {
			_senderid = "";
		}

		_msg_type = msg_type;
		_content = content;
		_registered_user_cnt = enrolled_user_cnt;
		_total_member_cnt = mem_cnt;
		_id = id;
		_max_id = max_id;
		_sent_time = sent_time;
		_valid = valid;
	}

	public InComingMessage(long sn, boolean valid, ChatRoomNewMsg newMsg) {
		super(sn, 0, PAYLOAD_INCOMING_MESSAGE, _reason);

		_valid = valid;

		if (newMsg != null) {

			if(newMsg.hasRoomid() && newMsg.getRoomid() != null){
				_roomid = newMsg.getRoomid().toStringUtf8();
			}
			
			if(newMsg.hasSender() && newMsg.getSender() != null && newMsg.getSender().hasUserid()){
				_senderid = newMsg.getSender().getUserid().toStringUtf8();
			}
			if (_senderid == null) {
				_senderid = "";
			}

			_msg_type = newMsg.getMsgtype();
			
			if(newMsg.hasMsgcontent() && newMsg.getMsgcontent() != null && newMsg.getMsgcontent().size() > 0){
				_content = newMsg.getMsgcontent().toByteArray();
			}
			
			_registered_user_cnt = newMsg.getRegmemcount();
			_total_member_cnt = newMsg.getMemcount();
			_id = newMsg.getMsgid();
			_max_id = newMsg.getMaxid();
			_sent_time = newMsg.getTimestamp();
		}
	}

	/**
	 * @return 聊天室id
	 */
	public String get_roomid() {
		return _roomid;
	}

	/**
	 * @return 发送者id
	 */
	public String get_senderid() {
		return _senderid;
	}

	/**
	 * @return 消息类型
	 */
	public int get_msg_type() {
		return _msg_type;
	}

	/**
	 * @return 消息二进制内容
	 */
	public byte[] get_content() {
		return _content;
	}

	/**
	 * 获取注册用户数, 只包括注册的用户数， 未注册的不算在内
	 * */
	public int get_registered_user_cnt() {
		return _registered_user_cnt;
	}

	/**
	 * 获取所有用户数， 包括注册的和未注册的(游客)
	 * */
	public int get_total_member_cnt() {
		return _total_member_cnt;
	}

	/**
	 * 获取消息id
	 * */
	public int get_id() {
		return _id;
	}

	public int get_max_id() {
		return _max_id;
	}

	public long get_sent_time() {
		return _sent_time;
	}

	public boolean is_valid() {
		return _valid;
	}
}