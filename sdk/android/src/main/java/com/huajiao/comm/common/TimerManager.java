package com.huajiao.comm.common;

import java.util.LinkedList;
import java.util.List;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import android.os.SystemClock;
import android.util.Log;

/**
 * 通用计时器
 * */
public class TimerManager implements Runnable {

	private final static byte ACTION_REMOVE = 1;

	private final static byte ACTION_ADD = 2;

	private final static byte ACTION_SHUTDOWN = 3;

	/** reset on demand timer */
	private final static byte ACTION_RESET = 4;

	/** reset on demand timer, if it is not set **/
	private final static byte ACTION_RESET_IF_NOT_SET = 5;

	/** cancel on demand timer */
	private final static byte ACTION_CANCEL_ON_DEMAND = 6;

	public String action_to_string(int action) {
		switch (action) {
		case ACTION_REMOVE:
			return "ACTION_REMOVE";
		case ACTION_RESET_IF_NOT_SET:
			return "ACTION_RESET_IF_NOT_SET";
		case ACTION_RESET:
			return "ACTION_RESET";
		case ACTION_CANCEL_ON_DEMAND:
			return "ACTION_CANCEL_ON_DEMAND";
		default:
			return Integer.toString(action);
		}
	}

	class Cmd {
		public byte action;
		public int id;
		public int timeout;
		public GenericTimer timer;
	}

	private Thread _thread;
	private BlockingQueue<Cmd> _cmds = new LinkedBlockingQueue<Cmd>();
	private List<GenericTimer> _timers = new LinkedList<GenericTimer>();
	private List<GenericTimer> _timers_to_remove = new LinkedList<GenericTimer>();
	private boolean _shutdown = false;
	private static String TAG = "TimerManager";
	private AtomicInteger _sidGen = new AtomicInteger(1);

	public TimerManager() {
		this("TimerManager");
	}

	public TimerManager(String name) {
		_thread = new Thread(this);
		_thread.setDaemon(true);
		_thread.setName(name);
		_thread.start();
	}

	private long now() {
		return SystemClock.elapsedRealtime();
	}

	private GenericTimer get_timer(int id) {
		for (GenericTimer t : _timers) {
			if (t.getId() == id) {
				return t;
			}
		}
		return null;
	}

	private int get_least_timeout() {

		int timeout = 3000000;
		long now = now();
		int diff = 0;

		for (GenericTimer t : _timers) {
			if (t.is_fire_on_demand()) {
				if (t.is_event_fired()) {
					continue;
				}
			}

			diff = (int) (t._last + t.getInterval() - now);
			if (diff <= 0) {
				timeout = 0;
			} else if (timeout > diff) {
				timeout = diff;
			}
		}

		int t = (int) (timeout <= 0 ? 0 : timeout);
		t += 5;
		return t;
	}

	/**
	 * @param interval 仅仅当fire_on_demand为false时才有效。
	 * @param callback
	 * @param fire_on_demand false表示repeat计时器。
	 * */
/*	public int addTimer(int interval, ITimerCallback callback, boolean fire_on_demand) {

		int id = _sidGen.incrementAndGet();

		GenericTimer t = new GenericTimer(id, interval, callback, fire_on_demand);
		Cmd cmd = new Cmd();
		cmd.id = id;
		cmd.timer = t;
		cmd.action = ACTION_ADD;
		postCmd(cmd);

		return id;
	}*/
	
	/** 相当于fire_on_demand为true的情况，即触发一次的计时器。 */
	public int addTimer(ITimerCallback callback) {

		int id = _sidGen.incrementAndGet();

		GenericTimer t = new GenericTimer(id, 0, callback, true);
		Cmd cmd = new Cmd();
		cmd.id = id;
		cmd.timer = t;
		cmd.action = ACTION_ADD;
		postCmd(cmd);

		return id;
	}

	/** 相当于fire_on_demand为false的情况，即repeat计时器。 */
	public int addTimer(int interval, ITimerCallback callback) {

		int id = _sidGen.incrementAndGet();
		GenericTimer t = new GenericTimer(id, interval, callback);

		Cmd cmd = new Cmd();
		cmd.id = id;
		cmd.timer = t;
		cmd.action = ACTION_ADD;
		postCmd(cmd);

		return id;
	}

	/**
	 * 设置按需触发计时器的下次超时。新的计时器超时会覆盖旧的计时器超时。
	 * */
	public void setOnDemandTimeout(int timer_id, int timeout) {
 
		Cmd cmd = new Cmd();
		cmd.id = timer_id;
		cmd.timeout = timeout;
		cmd.action = ACTION_RESET;
		postCmd(cmd);
	}

	/**
	 * 如果没有设置，则设置超时。新的计时器超时不会覆盖旧的计时器超时。
	 * */
	public void setOnDemandTimeoutIfOff(int timer_id, int timeout) {
		
		Cmd cmd = new Cmd();
		cmd.id = timer_id;
		cmd.timeout = timeout;
		cmd.action = ACTION_RESET_IF_NOT_SET;
		postCmd(cmd);
	}

	/**
	 * 检查该计时器是否活动状态
	 * 
	 * @param timer_id
	 * @return 计时器是否活跃
	 */
	public boolean isTimerActive(int timer_id) {

		for (GenericTimer t : _timers) {

			if (t.getId() != timer_id) {
				continue;
			}

			// if it is on-demand timer
			if (t.is_fire_on_demand()) {
				if (!t.is_event_fired()) {
					return true;
				}
			} else if (!t.is_fire_on_demand()) {
				// if it is interval timer
				return true;
			}
		}

		return false;
	}

	/**
	 * Cancel on demang timer, not remove it
	 * */
	public void cancelOnDemandTimer(int timer_id) {
		Cmd cmd = new Cmd();
		cmd.id = timer_id;
		cmd.action = ACTION_CANCEL_ON_DEMAND;
		postCmd(cmd);
	}

	public void shutdown() {
		Cmd cmd = new Cmd();
		cmd.action = ACTION_SHUTDOWN;
		postCmd(cmd);
	}

	public void removeTimer(int id) {
		Cmd cmd = new Cmd();
		cmd.id = id;
		cmd.action = ACTION_REMOVE;
		postCmd(cmd);
	}

	private void postCmd(Cmd cmd) {
		_cmds.offer(cmd);
	}

	private void check_timers() {

		long now = now();
		_timers_to_remove.clear();

		for (GenericTimer t : _timers) {
			if (t._last + t.getInterval() <= now) {
				// for on demand timer
				if (t.is_fire_on_demand() && !t.is_event_fired()) {
					t._last = now;
					execute(t);

				 	t.set_event_fired(true);
				 
					// for interval timer
				} else if (!t.is_fire_on_demand()) {
					t._last = now;
					execute(t);
				}
			}
		}
	}

	@Override
	public void run() {

		Cmd cmd = null;

		while (!_shutdown) {

			int timeout = get_least_timeout();

			try {
				cmd = _cmds.poll(timeout, TimeUnit.MILLISECONDS);
			} catch (InterruptedException e) {
				e.printStackTrace();
			}

			if (cmd != null) {

				GenericTimer timer = get_timer(cmd.id);
				if (cmd.action == ACTION_SHUTDOWN) {
					_shutdown = true;
				} else if (cmd.action == ACTION_ADD) {
					_timers.add(cmd.timer);
					if (BuildFlag.DEBUG) {
						Log.d(TAG, "timer added: " + cmd.timer.getId());
					}
				} else if (cmd.action == ACTION_REMOVE) {
					if (timer != null) {
						_timers.remove(timer);
						if (BuildFlag.DEBUG) {
							Log.d(TAG, "timer removed: " + cmd.id);
						}
					}
				} else if (cmd.action == ACTION_RESET) {
					if (timer != null) {
						timer.set_event_fired(false);
						timer._last = now(); // 计时器的起始时间
						timer.set_interval(cmd.timeout);					 
					}
				} else if (cmd.action == ACTION_RESET_IF_NOT_SET) {
					if (timer != null && timer.is_event_fired()) {
						timer.set_event_fired(false);
						timer._last = now(); // 计时器的起始时间
						timer.set_interval(cmd.timeout);
					} else {
						// 当前计时器非法或者还没触发，什么也不做。
					}
				} else if (cmd.action == ACTION_CANCEL_ON_DEMAND) {
					if (timer != null) {
						//Log.w(TAG, "CANCEL_ON_DEMAND: " + timer.getId());
						timer.set_event_fired(true);
					}
				} else {
					Log.w(TAG, "unknown action: " + cmd.action);
				}
			}

			if (_shutdown) {
				if (BuildFlag.DEBUG) {
					Log.i(TAG, "shutdown");
				}
				break;
			} else {
				check_timers();
			}
		}
	}

	private void execute(GenericTimer timer) {
		if (timer == null) {
			return;
		}

		try {		 
			timer.onInterval(timer.getId());		 
		} catch (Exception e) {
			e.printStackTrace();
		}
	}
}
