package com.huajiao.comm.im.packet;

public class MsgResultPacket extends Packet{

	/**
	 * 
	 */
	private static final long serialVersionUID = -278896296606461682L;


	public MsgResultPacket(long _sn, int _result) {
		super();
		this._sn = _sn;
		this._result = _result;
	}


	protected long _sn;
	protected int _result;
	 
	
	@Override
	public int getAction() {
		return Packet.ACTION_GOT_MSG_RESULT;
	}


	public long get_sn() {
		return _sn;
	}


	public int get_result() {
		return _result;
	}

}
