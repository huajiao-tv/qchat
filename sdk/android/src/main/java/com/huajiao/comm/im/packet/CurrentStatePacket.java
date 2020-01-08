package com.huajiao.comm.im.packet;

import com.huajiao.comm.im.ConnectionState;

public class CurrentStatePacket extends Packet {
 
	private static final long serialVersionUID = -1626160525299304962L;

	private ConnectionState _current_state;

	
	@Override
	public int getAction() {	 
		return ACTION_CURRENT_STATE;
	}

	/**
	 * @param _current_state
	 */
	public CurrentStatePacket(ConnectionState current_state) {
		super();
		this._current_state = current_state;
	}
	
	public ConnectionState get_current_state() {
		return _current_state;
	}

}
