package com.huajiao.comm.im.rpc;

import java.io.Serializable;

public abstract class Cmd implements Serializable{
	
	private static final long serialVersionUID = 8326213556925558869L;
	
	/** 切换账号  */
	public final static int CMD_SWITCH_ACCOUNT = 1;
	
	public final static int CMD_SEND_MESSAGE = 2;
	
	public final static int CMD_SEND_SRV_MESSAGE = 3;
	
	public final static int CMD_QUERY_PRESENCE = 4;
	
	public final static int CMD_REGISTER_FILTER_SERVICE = 5;
	
	public final static int CMD_GET_SERVER_TIME = 6;
	
	public final static int CMD_GET_MESSAGE = 7;
	
	public final static int CMD_GET_LLC_STATE = 8;
	
	public final static int CMD_SHUTDOWN = 9;
	
	private int _cmd_code;
	
	public int get_cmd_code() {
		return _cmd_code;
	}

	public Cmd(int code){
		_cmd_code = code;
	}
}
