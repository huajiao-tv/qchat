package com.huajiao.comm.im.packet;

public class TimePacket extends Packet {

	/**
	 * 
	 */
	private static final long serialVersionUID = 4111532561553456977L;

	@Override
	public int getAction() {
		return Packet.ACTION_SYNC_TIME;
	}
	
	public TimePacket(long diff){
		_diff = diff;
	}
	
	protected long _diff;

	/**
	 * 服务器和客户端SystemClock. 的差距
	 * */
	public long get_diff() {
		return _diff;
	}
}
