package com.huajiao.comm.im;

class Event implements Comparable<Event> {

	public static final int PRIORITY_LOWEST = 4;
	public static final int PRIORITY_LOW = 3;
	public static final int PRIORITY_NORMAL = 2;
	public static final int PRIORITY_HIGH = 1;
	public static final int PRIORITY_HIGHEST = 0;

	public Event(byte event_id, long time) {
		this(event_id);
		_time = time;
	}

	public Event(byte event_id, int priority) {
		this(event_id);
		_priority = priority;
	}

	public Event(byte event_id) {
		_event_id = event_id;
	}

	/**
	 * Event with argument
	 * */
	public Event(byte event_id, Object arg) {
		this(event_id);
		this._arg = arg;
	}

	protected byte _event_id;
	protected long _time;
	protected Object _arg;
	protected String _account;
	/**
	 * @return the _account
	 */
	public String get_account() {
		return _account;
	}

	/**
	 * @param _account the _account to set
	 */
	public void set_account(String _account) {
		this._account = _account;
	}

	/**
	 * lower value indicates higher priority
	 * */
	protected int _priority = 5;

	/**
	 * @return the _priority
	 */
	public int get_priority() {
		return _priority;
	}

	/**
	 * @return the event time
	 */
	public long get_time() {
		return _time;
	}

	public byte get_event_id() {
		return _event_id;
	}

	@Override
	public int compareTo(Event r) {
		if (r == null) {
			return -1;
		}

		if (get_priority() == r.get_priority()) {
			return 0;
		}

		return get_priority() < r.get_priority() ? 1 : -1;
	}
}
