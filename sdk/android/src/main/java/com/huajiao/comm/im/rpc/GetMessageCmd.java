package com.huajiao.comm.im.rpc;

public class GetMessageCmd extends Cmd {

	private static final long serialVersionUID = 8950307482247090220L;
	
	private String _info_type;
	private int[] _ids;
	private byte[] _paramters;

	/**
	 * @param code
	 * @param _info_type
	 * @param _ids
	 * @param _paramters
	 */
	public GetMessageCmd(String _info_type, int[] _ids, byte[] _paramters) {
		super(CMD_GET_MESSAGE);
		this._info_type = _info_type;
		this._ids = _ids;
		this._paramters = _paramters;
	}

	public String get_info_type() {
		return _info_type;
	}

	public int[] get_ids() {
		return _ids;
	}

	public byte[] get_paramters() {
		return _paramters;
	}

}
