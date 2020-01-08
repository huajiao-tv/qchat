package com.huajiao.comm.monitor;

import java.util.Locale;

class MissInfo {

	private int max_id;
	private int least_timeout;
	private String _roomid;

	/**
	 * @param max_id
	 * @param least_timeout
	 */
	public MissInfo(int max_id, int least_timeout, String roomid) {
		super();
		this.max_id = max_id;
		this.least_timeout = least_timeout;
		_roomid = roomid;
	}

	public int getMax_id() {
		return max_id;
	}

	/***
	 * 最小超时的消息的超时时间， 如果0表示没有
	 * */
	public int getLeast_timeout() {
		return least_timeout;
	}
	
	public String getRoomId(){
		return _roomid;
	}

	@Override
	public String toString() {
		return String.format(Locale.US, "max_id %d, least_timeout %d", max_id, least_timeout);
	}
}