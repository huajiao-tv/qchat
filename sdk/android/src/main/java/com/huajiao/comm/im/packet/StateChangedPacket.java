package com.huajiao.comm.im.packet;
import com.huajiao.comm.im.ConnectionState;
 
public class StateChangedPacket extends Packet {

	/**
	 * 
	 */
	private static final long serialVersionUID = -3977454633592233372L;
	@Override
	public int getAction() {
		return Packet.ACTION_STATE_CHANGED;
	}

	public StateChangedPacket(ConnectionState _oldState, ConnectionState _newState) {
		super();
		this._oldState = _oldState;
		this._newState = _newState;
	}

	protected ConnectionState _oldState;
	protected ConnectionState _newState;
	public ConnectionState get_oldState() {
		return _oldState;
	}

	public ConnectionState get_newState() {
		return _newState;
	}
}
