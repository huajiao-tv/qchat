package com.huajiao.comm.im;

import android.os.SystemClock;

import com.huajiao.comm.protobuf.messages.CommunicationData.Message;

 class MessageEvent extends Event {

	protected Message _message = null;

	protected int _send_count  = 0;
	
	public int get_send_count() {
		return _send_count;
	}

	 

	/** 消息发送的时间 */
	protected long _sent_time = 0;

	protected long _construct_time = SystemClock.elapsedRealtime();

	
	
	public long get_construct_time() {
		return _construct_time;
	}

	/** 消息超时 */
	protected int _timeout = 0;

	/** 是否已经通过socket成功发出 */
	protected boolean _has_been_sent = false;

	/** 是否是用户消息， 用户消息指的是用户发出的消息， 否则指底层用户不可见的消息 */
	protected boolean _is_user_message = false;

	/** 是否是心跳 */
	protected boolean _is_heartbeat = false;

	public int get_timeout() {
		return _timeout;
	}

	public Message get_message() {
		return _message;
	}

	public boolean is_heartbeat() {
		return _is_heartbeat;
	}

	/**
	 * 是否是用户消息， 用户消息指的是用户发出的消息， 否则指底层用户不可见的消息
	 * */
	public boolean is_user_message() {
		return _is_user_message;
	}

	/**
	 * 是否已经通过socket成功发出
	 */
	public boolean has_been_sent() {
		return _has_been_sent;
	}

	/**
	 * 是否已经通过socket成功发出
	 */
	public void set_has_been_sent(boolean has_been_sent) {
		_has_been_sent = has_been_sent;
		if(has_been_sent){
			_send_count ++;
		}
	}
	
	/**
	 * @param _is_heartbeat
	 *            the _is_heartbeat to set
	 */
	public void set_is_heartbeat(boolean is_heartbeat) {
		this._is_heartbeat = is_heartbeat;
	}

	/** 获取消息的发送时间， 发送完给服务器后的时间 */
	public long get_sent_time() {
		return _sent_time;
	}

	/** 把当前时间设置为消息的发送时间 */
	public void set_sent_time() {
		_sent_time = SystemClock.elapsedRealtime();
	}

	/*** 应用层消息 */
	public MessageEvent(Message message, int timeout, boolean is_user_message) {
		super(LLConstant.EVENT_SEND_MSG);
		_message = message;
		_timeout = timeout;
		if (_timeout <= 0) {
			_timeout = TimeConst.MSG_SEND_TIMEOUT;
		}
		_is_user_message = is_user_message;
	}

	/** 心跳事件 */
	public MessageEvent() {
		super(LLConstant.EVENT_SEND_HEARTBEAT);
		_is_heartbeat = true;
		_timeout = TimeConst.HEARTBEAT_ACK_TIMEOUT;
	}
 }