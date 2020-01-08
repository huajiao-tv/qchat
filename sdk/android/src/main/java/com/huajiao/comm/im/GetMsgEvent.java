package com.huajiao.comm.im;

public class GetMsgEvent extends Event {

	
	private String _info_type;
	private int[] _ids; 
	private byte[] _parameters;
	
	
	/**
	 * @param _info_type
	 * @param _start
	 * @param _length
	 * @param _parameters
	 */
	public GetMsgEvent(String _info_type, int[] ids, byte[] parameters) {
		super(LLConstant.EVENT_GET_MSG);
		this._info_type = _info_type;
		this._parameters = parameters;
		_ids = ids;
	}

	public String get_info_type() {
		return _info_type;
	}

	public int[] get_ids() {
		return _ids;
	}

	public byte[] get_parameters() {
		return _parameters;
	}

}
