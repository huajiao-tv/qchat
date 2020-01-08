package com.huajiao.comm.im.packet;

public class NotificationPacket extends Packet {

	private static final long serialVersionUID = 3429746993563634565L;

	
	private String _info_type;
	private byte [] _info_content;
	private long _info_id = -1;
	
	public NotificationPacket(String info_type, byte[] info_content, long info_id) {
		super();
		_info_type = info_type;
		_info_content = info_content;
		_info_id = info_id;
	}

	@Override
	public int getAction() {
		return Packet.ACTION_NOTIFICATION;
	}

	public String get_info_type() {
		return _info_type;
	}

	public byte[] get_info_content() {
		return _info_content;
	}

	public long get_info_id() {
		return _info_id;
	}
}
