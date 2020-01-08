package com.huajiao.comm.im.rpc;

public class ServiceMsgCmd extends Cmd {

	/**
	 * 
	 */
	private static final long serialVersionUID = -5526974199768120623L;
	
	protected int _service_id;
	protected long _sn;
	protected byte[] _body;
	 
	
	/**
	 * @param service_id
	 * @param sn
	 * @param body
	 * @param business_id
	 */
	public ServiceMsgCmd(int service_id, long sn, byte[] body) {
		super(Cmd.CMD_SEND_SRV_MESSAGE);
		this._service_id = service_id;
		this._sn = sn;
		this._body = body;
	}

	public int get_service_id() {
		return _service_id;
	}

	public long get_sn() {
		return _sn;
	}

	public byte[] get_body() {
		return _body;
	}

}
