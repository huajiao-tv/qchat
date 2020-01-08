package com.huajiao.comm.im.packet;

public class SrvMsgPacket extends Packet {

	/**
	 * 
	 */
	private static final long serialVersionUID = -279179281147844256L;

	@Override
	public int getAction() {
		return Packet.ACTION_GOT_SRV_MSG;
	}
	
	public SrvMsgPacket(long _sn, int _service_id, int _result, byte[] _data) {
		super();
		this._sn = _sn;
		this._service_id = _service_id;
		this._result = _result;
		this._data = _data;
	}

	protected long _sn;
	protected int _service_id;
	protected int _result;
	protected byte[] _data;
	
	public long get_sn() {
		return _sn;
	}

	public int get_service_id() {
		return _service_id;
	}

	public int get_result() {
		return _result;
	}

	public byte[] get_data() {
		return _data;
	}
}
