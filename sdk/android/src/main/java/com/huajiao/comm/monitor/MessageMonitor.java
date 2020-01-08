package com.huajiao.comm.monitor;

import java.util.ArrayList;
import java.util.Hashtable;
import java.util.LinkedList;
import java.util.List;
import java.util.Locale;

import android.os.AsyncTask;
import android.os.SystemClock;
import android.text.TextUtils;
import android.util.Log;

import com.huajiao.comm.chatroom.CRLogger;
import com.huajiao.comm.chatroom.ChatroomHelper;
import com.huajiao.comm.chatroom.IChatroomHelper;
import com.huajiao.comm.common.BuildFlag;
import com.huajiao.comm.common.FeatureSwitch;
import com.huajiao.comm.common.HttpUtils;
import com.huajiao.comm.common.ITimerCallback;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.TimerManager;
import com.huajiao.comm.im.ConnectionState;

/**
 * 监控是否有丢消息的情况， 如果有发起GET请求。
 * */
public class MessageMonitor implements ITimerCallback {

	private final static String TAG = "MSG-MON";

	private final static String INFO_TYPE_CHATROOM = "chatroom";
	

	private final static String REPORT_SERVER_HOST = "";

	// private final static int LOST_ARRAY_MAX_SIZE = 500; // 已改用云控

	private final static int MAX_REPORT_TIMES = 1;

	/** 每次最多取多少条消息 */
	private final static int MAX_GET_LEN = 100;

	/*** 拉取消息超时的底线 */
	private final static int GET_MSG_MIN_TIMEOUT = 5000;

	/** 多长时间缺失的消息开始出发 拉取消息的逻辑 */
	private final static int GET_MSG_MAX_TIMEOUT = 5000;

	/*** 多长时间消息还没有拉回来就认为丢消息了 */
	private final static int REPORT_MSG_MAX_TIMEOUT = 240000;

	private final static int REPORT_MSG_MIN_TIMEOUT = 180000;

	/** 消息丢失超载保护的时间 */
	private final static int MSG_OVERLOAD_TIMEOUT = 2 * 60000;

	/*** lost message ht, key: id, value: expected arrival time */
	private Hashtable<Integer, Long> _lost = new Hashtable<Integer, Long>();

	/*** 当前的聊天室id */
	private volatile String _roomid = null;

	private String _uid;

	private ConnectionState _conn_state = ConnectionState.Connected;

	private int _report_times = 0;

	private IChatroomHelper _chatroom = null;

	/*** get timer id */
	private int _get_tid = -1;

	/** report message lost timer */
	private int _report_tid = -1;

	/*** 过载timer id */
	private int _overload_tid = -1;

	/** 上次已经拉取消息能覆盖的最大值 */
	private int _last_get_id = 0;

	/** 已经push下来的最打的msg_id */
	private int _max_msg_id = 0;

	/** 长连接断开的次数 打点用到 */
	private int _disconnected_count = 0;

	private int _last_report_id = 0;

	private boolean _has_invalid_msg = false;

	/** 是否启用消息拉取补偿丢失的消息 */
	private volatile boolean _enable_compensation = true;

	private Object _lock = new Object();

	private TimerManager _tm = new TimerManager("MSG-TIMER");

	/**
	 * @param chatroom
	 * @param uid
	 */
	public MessageMonitor(IChatroomHelper chatroom, String uid) {
		super();

		_get_tid = _tm.addTimer(this);
		_report_tid = _tm.addTimer(this);
		_overload_tid = _tm.addTimer(this);
		_chatroom = chatroom;
		_uid = uid;

		reset_room(null);
	}

	long now() {
		return SystemClock.elapsedRealtime();
	}

	/***
	 * put message in proper location
	 * 
	 * @param id
	 *            message id
	 * */
	private boolean put(int id, int max_id, boolean via_get) {
		
		if(ChatroomHelper.isAllowSaveCrMsgLog()) {
			CRLogger.d(TAG, String.format(Locale.US, "put %d %d %d", id, max_id, via_get ? 1 : 0));
		}

		Integer key = Integer.valueOf(id);
		long now = now();
		int maxOverloadMissCount = FeatureSwitch.getMaxOverloadMissCount();

		synchronized (_lock) {
			
			if(JhFlag.enableDebug()) {
				CRLogger.d(TAG, "put-start max_msg_id:"+_max_msg_id+", id:"+id+", maxMissCount:"+maxOverloadMissCount);
			}

			if (_roomid == null) {
				return true;
			}

			if (_max_msg_id == 0) {
				_max_msg_id = id;
				return true;
			}

			if (id == _max_msg_id) {
				// duplicated message
				CRLogger.d(TAG, "put id("+id+") == max_msg_id("+_max_msg_id+")");
				return false;

			} else if (id < _max_msg_id) {
				
				
				// lost message arrives
				if (_lost.containsKey(key)) {
					_lost.remove(key);
					if(JhFlag.enableDebug()) {
						CRLogger.d(TAG, "put id("+id+") < max_msg_id("+_max_msg_id+") contains:"+_lost.containsKey(key));
					}
					return true;
				}
				
				CRLogger.d(TAG, "put id("+id+") < maxid("+_max_msg_id+") in:"+_lost.containsKey(key));

				return false;

			} else { // id is greater than max id
				
				if(JhFlag.enableDebug()) {
					CRLogger.d(TAG, "put id("+id+") > max_msg_id("+_max_msg_id+")");
				}

				int lost_count = id - _max_msg_id - 1;
				int i = 0;
				
				if(JhFlag.enableDebug()) {
					CRLogger.d(TAG, "put cur_lost_count:"+lost_count);
				}

				if (lost_count + _lost.size() > maxOverloadMissCount) {
					i = lost_count - (maxOverloadMissCount - _lost.size()) - 1;
				}
				
				List<Integer> cur_lost_ids = new LinkedList<Integer>();

				for (; i < lost_count; i++) {

					Integer lost_key = Integer.valueOf(_max_msg_id + i + 1);

					if (BuildFlag.DEBUG) {
						Log.i(TAG, "add lost msg " + lost_key.intValue());
					}				

					_lost.put(lost_key, now);
					
					cur_lost_ids.add(lost_key);
					
					// lost hashmap is full
					if (_lost.size() >= maxOverloadMissCount) {

						if (BuildFlag.DEBUG) {
							Log.i(TAG, "overloaded");
						}
						
						if(JhFlag.enableDebug()) {
							CRLogger.e(TAG, "put overloaded");
						}

						reset_state();

						_tm.setOnDemandTimeoutIfOff(_overload_tid, MSG_OVERLOAD_TIMEOUT);
						_enable_compensation = false;

						return true;
					}
				}
				
				if(JhFlag.enableDebug()) {
					CRLogger.d(TAG, "put cur_total_lost_count:"+_lost.size());
				}
				
				if(JhFlag.enableDebug()) {
					CRLogger.d(TAG, "put cur_lost_ids:" + cur_lost_ids);
				}

				if (lost_count > 0) {

					if (!_tm.isTimerActive(_get_tid)) {
						_tm.setOnDemandTimeoutIfOff(_get_tid, GET_MSG_MAX_TIMEOUT);
					}

					if (FeatureSwitch.isReportOn() && !_tm.isTimerActive(_report_tid)) {
						_tm.setOnDemandTimeoutIfOff(_report_tid, REPORT_MSG_MAX_TIMEOUT);
					}
				}

				_max_msg_id = id;
				
				if(JhFlag.enableDebug()) {
					CRLogger.d(TAG, "put-end max_msg_id:" + _max_msg_id);
				}

				return true;
			}
		}
	}
	
	public boolean isMyRoomMsg(String roomid) {
		if(TextUtils.isEmpty(roomid)) {
			return false;			
		}
		return roomid.equals(_roomid);
	}
	
	public String getCurRoomid() {
		return _roomid;
	}

	public void onJoinedIn(String roomid) {
		reset_room(roomid);
	}

	public void onQuit(String roomid) {
		reset_room(null);
	}

	public void onStateChanged(ConnectionState state) {

		if (state == null) {
			return;
		}

		_conn_state = state;

		if (state.equals(ConnectionState.AuthFailed)) {
			reset_room(null);
		} else if (_conn_state.equals(ConnectionState.Connected)) {

			// 如果端连接恢复后， 用户依然在聊天室里面， 拉取一次消息
			getMessageWhenConnected();

		} else {
			_disconnected_count++;
		}
	}

	/**
	 * 检查消息
	 * 
	 * @param id
	 *            : 当前消息id
	 * @param maxid
	 *            : 最大消息id
	 * @param send_time
	 *            : 消息发送时间
	 * @param valid
	 *            : 是否有效
	 * @param via_get
	 *            : 是否是拉取来的消息
	 * @param get_msg_on
	 *            拉取消息补偿开关是否打开
	 * 
	 * @return true: 显示消息， false: 不显示
	 */
	public boolean onMessage(int id, int maxid, long send_time, boolean valid, boolean via_get, boolean get_msg_on) {

		if(JhFlag.enableDebug()) {
			CRLogger.d(TAG, "onMsg id:" + id + ", maxid:" + maxid + ", yxbc:" + _enable_compensation + ", via_get:" + via_get+", valid:"+valid+", cur_roomid:"+_roomid);
		}

		//上层不允许消息补偿
		if (!get_msg_on) {
			CRLogger.d(TAG, "onMessage get_msg_on is false");
			return true;
		}

		//逻辑不允许消息补偿
		if (!_enable_compensation) {
			return true;
		}

		// switched off, return message directly
		if (!FeatureSwitch.isPullingOn()) {
			if (_max_msg_id != 0) {
				reset_state();
			}
			return true;
		}

		if (!valid) {
			_has_invalid_msg = true;
		}
		
		if (_roomid == null) {
			return true;
		}

		boolean result = put(id, maxid, via_get);
		
		if(JhFlag.enableDebug()) {
			CRLogger.d(TAG, "onMsg id:" + id + ", return:"+(result && valid));
		}
		
		if (result && valid) {
			return true;
		}

		return false;
	}

	/**
	 * 重置聊天室房间
	 * */
	private void reset_room(String roomid) {

		synchronized (_lock) {
			boolean reset = false;
			if (_roomid != null && roomid != null) { // 两个都不为空
				if (!_roomid.equals(roomid)) { // 房间发生变化
					reset = true;
					CRLogger.d(TAG, String.format("switch room %s -> %s", _roomid, roomid));
					_roomid = roomid;
					reset_state();
				}
			} else if (roomid != null || _roomid != null) { // 有1个不为空
				reset = true;
				CRLogger.d(TAG, String.format("switch room %s -> %s", _roomid, roomid));
				_roomid = roomid;
				reset_state();
			}
			if(!reset) { // 两个都为空 或者 两个都不为空且两个相同 
				// 日志中添加"->"字符 2016.8.8				
				CRLogger.d(TAG, String.format("switch room %s -> %s", _roomid, roomid));
			}
		}
	}

	/**
	 * 重置状态,但是不重置当前room
	 * */
	private void reset_state() {

		_last_get_id = 0;
		_last_report_id = 0;
		_max_msg_id = 0;
		_disconnected_count = 0;
		_has_invalid_msg = false;

		_lost.clear();

		_tm.cancelOnDemandTimer(_get_tid);
		_tm.cancelOnDemandTimer(_report_tid);
		_tm.cancelOnDemandTimer(_overload_tid);

		if (BuildFlag.DEBUG) {
			if (_roomid == null) {
				Log.d(TAG, "reset_state: cleared");
			} else {
				Log.d(TAG, "reset_state: join " + _roomid);
			}
		}

		_enable_compensation = true;
	}

	/**
	 * timer超时回调
	 * */
	@Override
	public void onInterval(int id) {

		if (id == _get_tid) {
			if (BuildFlag.DEBUG) {
				Log.d(TAG, "do get message ");
			}

			getMessage();

		} else if (id == _report_tid) {
			if (BuildFlag.DEBUG) {
				Log.d(TAG, "do report loss");
			}

			reportLoss();

			_has_invalid_msg = false;

		} else if (id == _overload_tid) {

			if (BuildFlag.DEBUG) {
				Log.d(TAG, "overloaded finished");
			}

			_enable_compensation = true;
		}
	}

	/***
	 * 计算哪些消息超时了， 需要拉取，把id保存到keys里面
	 * 
	 * @param min_timeout
	 *            最小超时时间
	 * @param max_timeout
	 *            最大超时时间
	 * @param ref_id
	 *            消息id需要大于该参照值
	 * @param keys
	 *            用来存储需要拉取的消息id
	 * */
	private MissInfo computeMissInfo(int min_timeout, int max_timeout, int ref_id, List<Integer> keys) {

		if (keys == null) {
			return null;
		}

		keys.clear();
		long now = now();
		int least_timeout = -1;
		int elapsed_time = -1;

		synchronized (_lock) {

			if (_roomid == null || _lost.size() == 0) {
				return null;
			}

			for (Integer key : _lost.keySet()) {
				int diff = (int) (now - _lost.get(key).longValue());

				if (key.intValue() > ref_id) {
					if (diff >= min_timeout) {
						int i = 0;
						while (i < keys.size()) {
							if (keys.get(i).intValue() > key.intValue()) {
								break;
							} else {
								i++;
							}
						}
						keys.add(i, key);
					} else if (diff > elapsed_time) {
						elapsed_time = diff;
					}
				}
			}

			if (keys.size() > 0) {
				if (elapsed_time == -1) {
					least_timeout = 0;
				} else {
					least_timeout = max_timeout - elapsed_time;
				}
				return new MissInfo(keys.get(keys.size() - 1).intValue(), least_timeout, _roomid);
			}
		}

		return null;
	}

	/**
	 * 长链接恢复后的消息拉取动作，会拉取所有的消息<br>
	 * 因为：保存在_lost里面的消息必然没有收到， 拉取一次作为补偿
	 * 
	 */
	private synchronized void getMessageWhenConnected() {

		if (!_enable_compensation || !_conn_state.equals(ConnectionState.Connected) || _roomid == null) {
			return;
		}

		// 置为0， 这样之前拉过的消息可以再次去拉， 防止丢消息
		_last_get_id = 0;

		getMessage();
	}

	/**
	 * 拉取超时的消息， 最多拉取一次， 一次拉取MAX_GET_LEN条
	 */
	private synchronized void getMessage() {

		if (!_enable_compensation || !_conn_state.equals(ConnectionState.Connected) || _roomid == null) {
			return;
		}

		List<Integer> keys = new ArrayList<Integer>();
		
		MissInfo info = computeMissInfo(GET_MSG_MIN_TIMEOUT, GET_MSG_MAX_TIMEOUT, _last_get_id, keys);
		
		if(JhFlag.enableDebug()) {
			CRLogger.i(TAG, "computeMissInfo result keys_size:"+keys.size());
		}

		if (info == null) {
			Log.e(TAG, "info is null");
			return;
		}

		if (keys.size() > 0) {

			int size = keys.size() > MAX_GET_LEN ? MAX_GET_LEN : keys.size();
			int[] id = new int[size];

			for (int j = 0; j < size; j++) {
				id[j] = keys.get(j);
			}
			
			if(JhFlag.enableDebug()) {
				CRLogger.i(TAG, "getMessage room_id:" + _roomid + ", get_size:" + size + ", ids:" + keys);
			}
			
			_chatroom.getMessage(INFO_TYPE_CHATROOM, id, _roomid.getBytes());
			
			if (BuildFlag.DEBUG) {
				Log.i(TAG, "getMessage: room " + _roomid + ": " + combine(id));
			}

			_last_get_id = id[size - 1];
		}

		if (info.getLeast_timeout() > 0) {

			if (BuildFlag.DEBUG) {
				Log.i(TAG, "we have untimed-out gap, scheduling timer: " + info.getLeast_timeout());
			}

			_tm.setOnDemandTimeout(_get_tid, GET_MSG_MIN_TIMEOUT);
		}
	}

	/*** 检查是否有消息丢失， 并打点报告 */
	private boolean reportLoss() {

		if (!FeatureSwitch.isReportOn()) {
			return true;
		}

		if (_report_times >= MAX_REPORT_TIMES || !_conn_state.equals(ConnectionState.Connected)) {
			return false;
		}

		List<Integer> keys = new ArrayList<Integer>();

		MissInfo info = computeMissInfo(REPORT_MSG_MIN_TIMEOUT, REPORT_MSG_MAX_TIMEOUT, _last_report_id, keys);
		if (info == null) {
			if (BuildFlag.DEBUG) {
				Log.d(TAG, "no msg lost");
			}
			return false;
		}

		if (keys.size() > 0) {
			if (upload_report(_uid, info.getRoomId(), keys.size())) {
				_report_times++;
				_last_report_id = info.getMax_id();
			}
		}

		if (info.getLeast_timeout() > 0 && _report_times < MAX_REPORT_TIMES) {
			if (BuildFlag.DEBUG) {
				Log.i(TAG, "need to re-run report, scheduling timer: " + info.getLeast_timeout());
			}
			_tm.setOnDemandTimeout(_report_tid, info.getLeast_timeout());
		}

		return true;
	}

	/***
	 * upload report
	 * */
	private boolean upload_report(String uid, String roomid, int count) {

		if (uid == null || roomid == null) {
			return false;
		}

		String reason = _has_invalid_msg ? "invalid" : "timeout";

		String url = String.format(Locale.US, "http://%s/message/loss?v=2&plf=android&reason=%s&uid=%s&roomid=%s&c=%d&dc=%d", REPORT_SERVER_HOST, reason, uid,
				roomid, count, _disconnected_count);

		if (System.currentTimeMillis() % 100 == 0) {
			new ReportTask().execute(url);
		}

		return true;
	}

	/***
	 * 异步打点丢失消息
	 * */
	static class ReportTask extends AsyncTask<String, Void, Boolean> {

		@Override
		protected Boolean doInBackground(String... params) {

			if (params == null || params.length < 0) {
				return Boolean.FALSE;
			}

			String url = params[0];
			if (url == null || url.length() < 7) {
				return Boolean.FALSE;
			}

			boolean result = HttpUtils.touch(url, 15000, 10000);

			if (BuildFlag.DEBUG) {
				Log.i(TAG, "r message loss: " + Boolean.toString(result));
			}

			return Boolean.valueOf(result);
		}
	}

	String combine(int[] ids) {
		StringBuilder sb = new StringBuilder();
		for (int i : ids) {
			sb.append(i);
			sb.append(",");
		}
		return sb.toString();
	}
}
