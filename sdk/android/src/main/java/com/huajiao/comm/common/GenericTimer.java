package com.huajiao.comm.common;

import android.util.Log;

public class GenericTimer {

	private int _id = 0;
	private int _interval = 0;
	public long _last = 0;
	private ITimerCallback _callback;
	private boolean _fire_on_demand = false;
	private boolean _event_fired = false;

	private static final String TAG = "T";

	/**
	 * @param id
	 *            unique id
	 * @param interval
	 * @param callback
	 * */
	public GenericTimer(int id, int interval, ITimerCallback callback) {
		this(id, interval, callback, false);
	}

	/**
	 * @param id unique id
	 * @param interval 仅当fire_on_demand为false时才有效
	 * @param callback
	 * @param fire_on_demand false表示repeat计时器。
	 * 
	 * */
	public GenericTimer(int id, int interval, ITimerCallback callback, boolean fire_on_demand) {	
		_id = id;
		_interval = interval;
		_callback = callback;
		_last = System.currentTimeMillis();
		_fire_on_demand = fire_on_demand;
		
		// do not fire when created
		if (_fire_on_demand) {
			_event_fired = true;
		}
	}

	public int getId() {
		return _id;
	}

	public int getInterval() {
		return _interval;
	}

	/**
	 * Fire timeout event
	 * */
	public void onInterval(int id) {
		_callback.onInterval(id);
	}

	public boolean is_fire_on_demand() {
		return _fire_on_demand;
	}

	public boolean is_event_fired() {
		return _event_fired;
	}

	public void set_event_fired(boolean event_fired) {		 
		 _event_fired = event_fired;
	}

	public void set_interval(int interval) {
		if (interval <= 0) {
			if (BuildFlag.DEBUG) {
				Log.e(TAG, "interval is invalid: " + interval);
			}
			return;
		}
		_interval = interval;
	}
}
