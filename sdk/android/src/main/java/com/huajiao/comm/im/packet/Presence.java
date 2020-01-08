package com.huajiao.comm.im.packet;

/**
 * 表示用户的状态
 * */
public class Presence {
	
	/**
	 * 在线
	 * */
	public static final int STATUS_ONLINE = 3;
	
	/**
	 * 离线
	 * */
	public static final int STATUS_OFFLINE_2 = 2;
	
	/**
	 * 离线
	 * */
	public static final int STATUS_OFFLINE = 1;
	
	/**
	 * 用户未注册
	 * */
	public static final int STATUS_UNREGISTERED = 0;

	/**
	 * 状态, STATUS_*
	 * */
	public int get_status() {
		return _status;
	}

	/**
	 * 用户id, 对应手机号
	 * */
	public String get_userid() {
		return _userid;
	}

	/**
	 * id,  类型，目前只会返回   phone
	 * */
	public String get_type() {
		return _type;
	}

	/**
	 * Application Id
	 * */
	public int get_appid() {
		return _appid;
	}

	/**
	 * 登录时的EPV
	 * */
	public int get_epv() {
		return _epv;
	}

	/**
	 * 设备类型 android或者ios
	 * */
	public String get_mobile_type() {
		return _mobile_type;
	}

	private int _status;
	private String _userid;
	private String _type;
	private int _appid;
	private int _epv;
	private String _mobile_type;

	public Presence(String _userid, String _type, int _status, String _mobile_type, int _appid, int _epv) {
		super();
		this._status = _status;
		this._userid = _userid;
		this._type = _type;
		this._appid = _appid;
		this._epv = _epv;
		this._mobile_type = _mobile_type;
	}
}

