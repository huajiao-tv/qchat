package com.huajiao.comm.im.rpc;

/***
 * 发送消息 
 * */
public class MsgCmd extends Cmd {

	private static final long serialVersionUID = -2007896494031775953L;
	
	protected String _receiver;
	protected int _account_type;
	protected int _msg_type;
	protected long _sn;
	protected byte[] _body;
	protected int _timeout_ms;
	protected int _expiration_sec;
	
	/**
	 * @param receiver
	 * @param account_type
	 * @param msgType
	 * @param sn
	 * @param body
	 * @param business_id
	 * @param timeout_ms
	 * @param expiration_sec
	 */
	public MsgCmd(String receiver, int account_type, int msgType, long sn, byte[] body,  int timeout_ms, int expiration_sec) {
		super(Cmd.CMD_SEND_MESSAGE);
		if(receiver == null || receiver.length() == 0){
			throw new IllegalArgumentException("receiver");
		}
		
		this._receiver = receiver;
		this._account_type = account_type;
		this._msg_type = msgType;
		this._sn = sn;
		this._body = body;
		this._timeout_ms = timeout_ms;
		this._expiration_sec = expiration_sec;
	}


	public String get_receiver() {
		return _receiver;
	}


	public int get_account_type() {
		return _account_type;
	}


	public int get_msg_type() {
		return _msg_type;
	}


	public long get_sn() {
		return _sn;
	}


	public byte[] get_body() {
		return _body;
	}

	public int get_timeout_ms() {
		return _timeout_ms;
	}


	public int get_expiration_sec() {
		return _expiration_sec;
	}

}
