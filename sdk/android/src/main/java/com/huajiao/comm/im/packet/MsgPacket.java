package com.huajiao.comm.im.packet;

public class MsgPacket extends Packet {
	
	protected long _sn;
	protected String _info_type;
	protected String _from;
	protected String _to;
	protected byte[] _content;
	protected long _id;
	protected int _msg_type;
	protected long _date;
	protected long _latest_id;
	protected long _server_time;
	protected boolean _valid;

	/**
	 * 
	 */
	private static final long serialVersionUID = -6692086683471749546L;

	public MsgPacket(long _sn, String info_type, String _from, String to,  byte[] _content, long _id, int _msg_type, long _date, long _latest_id, long server_time, boolean valid) {
		this._sn = _sn;
		this._from = _from;
		this._content = _content;
		this._id = _id;
		this._msg_type = _msg_type;
		this._date = _date;
		this._latest_id = _latest_id;
		_server_time = server_time;
		_info_type = info_type;
		_to = to;
		_valid = valid;
	}

	
	@Override
	public int getAction() {
		return Packet.ACTION_GOT_MSG;
	}

	/**
	 * 消息SN
	 * */
	public long get_sn() {
		return _sn;
	}

	/**
	 * 消息发送者
	 * */
	public String get_from() {
		return _from;
	}

	
	/**
	 * 消息内容
	 * */
	public byte[] get_content() {
		return _content;
	}

	/**
	 * 消息id
	 * */
	public long get_id() {
		return _id;
	}

	/**
	 * 消息业务类型
	 * */
	public int get_msg_type() {
		return _msg_type;
	}

	/**
	 * 消息入库时间EPOCH时间, 单位毫秒
	 * */
	public long get_date() {
		return _date;
	}

	public long get_latest_id() {
		return _latest_id;
	}

	public long get_server_time() {
		return _server_time;
	}

	/**
	 * 消息盒子类型
	 * */
	public String get_info_type() {
		return _info_type;
	}	
	
	public boolean is_valid() {
		return _valid;
	}


	public String get_to() {
		return _to;
	}
}
