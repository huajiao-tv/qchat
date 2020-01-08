package com.huajiao.comm.im.packet;

import java.io.Serializable;

public abstract class Packet implements Serializable {
	
	/**
	 * 
	 */
	private static final long serialVersionUID = 5474403097370528163L;

	/**
	 * 获取到消息
	 * */
	public static final int ACTION_GOT_MSG = 1;

	/**
	 * 获取到业务消息
	 * */
	public static final int ACTION_GOT_SRV_MSG = 2;

	/**
	 * 消息发送结果
	 * */
	public static final int ACTION_GOT_MSG_RESULT = 3;

	/**
	 * 长连接状态发生改变
	 * */
	public static final int ACTION_STATE_CHANGED = 4;

	/***
	 * 状态查询结果
	 * */
	public static final int ACTION_PRESENCE_UPDATED = 5;
	
	
	/***
	 * 收到通知
	 * */
	public static final int ACTION_NOTIFICATION = 6;
	
	
	/***
	 * 同步时间
	 * */
	public static final int ACTION_SYNC_TIME = 7;
	
	
	/***
	 * 获取当前状态
	 * */
	public static final int ACTION_CURRENT_STATE = 8;

	/**
	 * 获取对应的动作
	 * */
	public abstract int getAction();
	
	public String toString() {
		
		switch (getAction()) {
		case ACTION_GOT_MSG:
			return "ACTION_GOT_MSG";
			
		case ACTION_GOT_SRV_MSG:
			return "ACTION_GOT_SRV_MSG";
			
		case ACTION_GOT_MSG_RESULT:
			return "ACTION_GOT_MSG_RESULT";

		case ACTION_STATE_CHANGED:
			return "ACTION_STATE_CHANGED";

		case ACTION_PRESENCE_UPDATED:
			return "ACTION_PRESENCE_UPDATED";
			
		case ACTION_NOTIFICATION:
			return "ACTION_NOTIFICATION";
			
		case ACTION_SYNC_TIME:
			return "ACTION_SYNC_TIME";
		}

		return "INVALID_ACTION";
	}
	
	/**
	 * @return the _appid
	 */
	public int get_appid() {
		return 1080;
	}
}
