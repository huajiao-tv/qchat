package com.huajiao.comm.chatroomresults;

/**
 * 用户信息
 * */
public class UserInfo {

	private String _uid;
	private byte[] _data;
	
	public String get_uid() {
		return _uid;
	}

	public byte[] get_user_data() {
		return _data;
	}

	
	public UserInfo(String uid, byte[] data) {
		super();
		this._uid = uid;
		this._data = data;
	}
	
	public UserInfo(String uid) {
		super();
		this._uid = uid;
		this._data = null;
	}
}
