package com.huajiao.comm.im;

 class State {

	public State(ConnectionState state) {
		_state = state;
	}

	protected ConnectionState _state;

	/**
	 * 状态名字
	 * */
	public ConnectionState get_state() {
		return _state;
	}

	/**
	 * Got called when any event fired
	 * */
	public void OnEeventFired(Event event) {
		// Log.i("EVENT", String.format("Processing event %d", event._event_id));
	}

	/**
	 * 进入该状态时触发
	 * */
	public void OnEnter() {
	}

	/**
	 * 退出该状态时触发
	 * */
	public void OnExit() {
	}
}
