package com.huajiao.comm.im;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.UnsupportedEncodingException;
import java.lang.Thread.UncaughtExceptionHandler;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketAddress;
import java.net.SocketException;
import java.net.SocketTimeoutException;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.HashMap;
import java.util.List;
import java.util.Locale;
import java.util.Random;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentSkipListMap;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

import android.annotation.SuppressLint;
import android.app.AlarmManager;
import android.app.PendingIntent;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;
import android.os.AsyncTask;
import android.os.Build;
import android.os.PowerManager;
import android.os.PowerManager.WakeLock;
import android.os.SystemClock;
import android.util.Log;

import com.google.protobuf.micro.ByteStringMicro;
import com.huajiao.comm.common.HttpUtils;
import com.huajiao.comm.protobuf.messages.CommunicationData;
import com.huajiao.comm.protobuf.messages.CommunicationData.ChatReq;
import com.huajiao.comm.protobuf.messages.CommunicationData.ChatResp;
import com.huajiao.comm.protobuf.messages.CommunicationData.Ex1QueryUserStatusReq;
import com.huajiao.comm.protobuf.messages.CommunicationData.Ex1QueryUserStatusResp;
import com.huajiao.comm.protobuf.messages.CommunicationData.GetInfoReq;
import com.huajiao.comm.protobuf.messages.CommunicationData.GetInfoResp;
import com.huajiao.comm.protobuf.messages.CommunicationData.GetMultiInfosReq;
import com.huajiao.comm.protobuf.messages.CommunicationData.GetMultiInfosResp;
import com.huajiao.comm.protobuf.messages.CommunicationData.Info;
import com.huajiao.comm.protobuf.messages.CommunicationData.InitLoginReq;
import com.huajiao.comm.protobuf.messages.CommunicationData.InitLoginResp;
import com.huajiao.comm.protobuf.messages.CommunicationData.LoginReq;
import com.huajiao.comm.protobuf.messages.CommunicationData.LoginResp;
import com.huajiao.comm.protobuf.messages.CommunicationData.Message;
import com.huajiao.comm.protobuf.messages.CommunicationData.NewMessageNotify;
import com.huajiao.comm.protobuf.messages.CommunicationData.Pair;
import com.huajiao.comm.protobuf.messages.CommunicationData.ReqEQ1User;
import com.huajiao.comm.protobuf.messages.CommunicationData.Request;
import com.huajiao.comm.protobuf.messages.CommunicationData.RespEQ1User;
import com.huajiao.comm.protobuf.messages.CommunicationData.Service_Req;
import com.huajiao.comm.protobuf.messages.CommunicationData.Service_Resp;
import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.BuildFlag;
import com.huajiao.comm.common.ClientConfig;
import com.huajiao.comm.common.IUplink;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.LoggerBase;
import com.huajiao.comm.common.RC4;
import com.huajiao.comm.common.Utils;
import com.huajiao.comm.im.api.Constant;
import com.huajiao.comm.im.NetworkProbe;
import com.huajiao.comm.im.api.MessageType;
import com.huajiao.comm.im.packet.CurrentStatePacket;
import com.huajiao.comm.im.packet.MsgPacket;
import com.huajiao.comm.im.packet.MsgResultPacket;
import com.huajiao.comm.im.packet.NotificationPacket;
import com.huajiao.comm.im.packet.PresencePacket;
import com.huajiao.comm.im.packet.SrvMsgPacket;
import com.huajiao.comm.im.packet.StateChangedPacket;
import com.huajiao.comm.im.util.TimeUtil;

/**
 * @author zhouyuanjiang client connection 实现和维护客户和服务器之间的长连接
 */
class ClientConnection implements IConnection, UncaughtExceptionHandler, IUplink {





	/** 重连的时间间隔 单位秒 */
	private short[] reconnect_intervals = new short[] { 1, 2, 2, 2, 2, 6, 8, 12, 18, 24, 32, 48, 96, 120, 192, 240, 300 };

	/** 重连时间间隔的索引, 如果超出范围则使用最后的值 */
	private volatile int _interval_index = 0;

	private int _cur_hour = -1;
	private int _cur_min = -1;
	private int _login_count = 0;

	private enum HandlePacketResult {
		/** 处理成功 */
		Succeeded,

		/** 处理失败， 需要跳转到Disconnected状态 */
		Failed,

		/** 被T， 需要跳转到LoggedInElsewhere状态 */
		ReloggedIn,

		/** server ask for reconnect to another server */
		ReConnect
	}

	private final static String TAG = "Conn_2080";

	/** 定期发送心跳 */
	private final static String ACTION_PING = LLConstant.PROJECT + "_COMM_LLC_ACTION_PING";

	/** 服务器是否过载 */
	private boolean is_overloaded = false;

	/** 心跳相关参数1分钟 */
	private int _curr_heart = 60000;
	private static int _min_heart = 60000;
	private static int _max_heart = 270000;
	private int _report_heart_time = 120;
	private int _success_heart = 60000; //暂时不用，用于动态确定最佳心跳时间

	/** 上次用户主动PING的时间 */
	private long _last_ping_time = System.currentTimeMillis();

	/** 屏幕变黑的时间 */
	private volatile long _screen_off_time = 0;

	/** android application context */
	private Context _context;

	/** 当前网络类型 */
	private int _net_type = 0;

	private int _lvs_index = 0;

	private ClientConfig _clientConfig = null;

	private WakeLock _ping_WL = null;
	private WakeLock _get_msg_WL = null;
	private WakeLock _business_WL = null;

	/** 心跳包内容 */
	private final static byte[] HeartbeatContent = new byte[] { 0, 0, 0, 0 };

	private long _packet_start = 0;
	private final long[] _rtt = new long[2];

	private final MessageEvent _heartbeat_event = new MessageEvent();
	private final Event _got_heartbeat_ack_event = new Event(LLConstant.EVENT_GOT_HEARTBEAT_ACK);

	private BroadcastReceiver _scheduled_task_receiver = new ScheduledTaskReceiver();

	/** 登录成功后android系统已经持续运行的时间长度 */
	private static long _time_base = 0;

	/** 登录成功后得到的服务器UTC时间 */
	private static long _server_time = 0;

	private Object _time_lock = new Object();

	/** 有限状态列表 */
	protected HashMap<ConnectionState, State> _states = new HashMap<ConnectionState, State>();

	/** 当前的状态 */
	protected volatile State _current_state = null;

	/** 事件队列 */
	protected BlockingQueue<Event> _eventQueue = new LinkedBlockingQueue<Event>();

	/**
	 * 已经通过SOCKET发出， 等待服务器回执的消息队列<br>
	 * 用户消息和底层发送出去的消息都保存在这个HT里面， 通过is_user_message区分
	 */
	// protected Hashtable<Long, MessageEvent> _pendingMessages = new Hashtable<Long, MessageEvent>();
	// 保证升序排列
	protected ConcurrentSkipListMap<Long, MessageEvent> _pendingMessages = new ConcurrentSkipListMap<Long, MessageEvent>();

	/** 回调接口 */
	private IMCallback _inotify;

	private volatile boolean _magic_received = false;
	private static Random _random = new Random();
	private String _server_ran;
	private String _client_ran = getRandomString(8);
	private Socket _socket;

	private boolean _inetAvailable = true;

	private HashMap<String, MessageFlag> _msg_flags = new HashMap<String, MessageFlag>();

	private volatile RC4InputStream _socket_in = null;
	private volatile RC4OutputStream _socket_out = null;

	private Object _connectLock = new Object();

	private Receiver _receiver;
	private Sender _sender;

	private PowerManager _pm;
	private AlarmManager _am;
	private Object _alarmLock = new Object();
	private PendingIntent _pi_ping;

	private volatile boolean _quit = false;

	/**
	 * construct magic code
	 * */
	private byte MagicCode[];

	/***
	 * Indication of socket state
	 **/
	private volatile boolean _connected = false;

	private boolean _account_switched = false;

	/**
	 * 登录成功的时间
	 * */
	private long _connect_time = SystemClock.elapsedRealtime();
	private long _last_disconnect_time = System.currentTimeMillis();

	/**
	 * 是否已经收到服务器的心跳回应
	 * */
	private final static long HEARTBEAT_SN = -123454321;

	private boolean _scheduled_task_started = false;

	/**
	 * Indication of user status in
	 */
	private volatile boolean _logged_in = false;
	private boolean _init_logged_in = false;
	private boolean _init_packtet_sent = false;

	private String _sessionKey;

	private AtomicLong _snSeed = new AtomicLong(System.currentTimeMillis());
	private AtomicInteger _snSeedInt = new AtomicInteger();

	private AccountInfo _account_info;
	private String _jid;

	/** 依然有链接， 只是网络环境发生变化了 */
	private static final String SCHEDULE_TASK_EXTRA_ID = "scheduled_time_id";

	private ConnectivityChangedReceiver _connReceiver = null;
      private NetworkProbe mNetworkProbe;
	private boolean _connectivity_registered = false;

	private ArrayList<IPAddress> _reconnect_hosts = null;

	private final static String REPORT_SERVER_HOST = "s.360.cn";
	private final static String DISPATCH_SERVER_HOST = "testdisp.jiaoyantv.com";

	private boolean _drop_peer_msg = true;

	private DispatchClient _dispatchClient = null;

	/** 注册alarm广播 */
	private void registerAlarmBroadcast() {
		if (!_scheduled_task_started) {
			IntentFilter intent_filter = new IntentFilter();
			intent_filter.addAction(ACTION_PING);
			_context.registerReceiver(_scheduled_task_receiver, intent_filter);
			_scheduled_task_started = true;
		}
	}

	/** 注销alarm广播 */
	private void unregisterAlarmBroadcast() {
		if (_scheduled_task_started) {
			cancel_scheduled_task();
			_context.unregisterReceiver(_scheduled_task_receiver);
			_scheduled_task_started = false;
		}
	}

	/** 取消心跳和取消息定期任务 */
	private void cancel_scheduled_task() {

		if (_am == null || !_scheduled_task_started) {
			return;
		}

		synchronized (_alarmLock) {
			if (_pi_ping != null) {
				_am.cancel(_pi_ping);
				_pi_ping = null;
			}
		}
	}

	/** 构建发消息的PindingIntent */
	private PendingIntent makePingPI(final long nextTime) {
		// _next_ping_time = nextTime;
		Intent heartBeatIntent = new Intent(ACTION_PING);
		heartBeatIntent.putExtra(SCHEDULE_TASK_EXTRA_ID, nextTime);
		return PendingIntent.getBroadcast(_context, get_id(), heartBeatIntent, PendingIntent.FLAG_CANCEL_CURRENT);
	}

	private String makeVerfCode(String jid) {
		final String salt = "3" + 6 + "0" + "tan" + "t" + "an" + "@" + 14 + "0" + 8 + "$";
		String tmp = Utils.MD5(jid + salt);
		return tmp.substring(24);
	}

	/** 调整心跳时间 */
	@SuppressLint("NewApi")
	private void schedule_next_ping() {

		if (_am == null || !_scheduled_task_started) {
			return;
		}

		synchronized (_alarmLock) {
			// 先cancel已存在的HeartBeat任务
			if (_pi_ping != null) {
				_am.cancel(_pi_ping);
			}

			if (isDeviceAwake() || !_inetAvailable || (_current_state != null && _current_state.get_state().equals(ConnectionState.AuthFailed))) {
				//设备亮屏时不调度
				return;
			}

			int timeout = _curr_heart;
			boolean is_screen_on = false;
			if (_pm != null) {
				is_screen_on = _pm.isScreenOn();
			}

			long off_time = SystemClock.elapsedRealtime() - _screen_off_time;

			int persit_offline_time = getPersistOfflineTime();
			// Logger.d(TAG, "consistent offline time: " + persit_offline_time);
			// 稳定的假连接, alarm 设为5分钟一次, 大约300秒还是没能连接上， 那么可能是假连接
			if (persit_offline_time >= 300000) {
				timeout = 300000;
			} else if (!is_screen_on && _screen_off_time != 0) {
				// 午夜锁屏超过 1 小时了， 我们不唤醒手机了， 需要省电
				if (off_time >= 3600000 && is_in_midnight()) {
					Logger.v(TAG, "no longer scheduling p.");
					return;
					// 用户已经锁定屏幕超过30分钟了
				} else if (off_time > 1800000) {
					timeout = get_next_heartbeat_time(timeout);
				}
			}

			long nextTimeHeartBeat = SystemClock.elapsedRealtime() + timeout;
			_pi_ping = makePingPI(nextTimeHeartBeat);

			Logger.v(TAG, "schedule next hb alarm p-> a " + nextTimeHeartBeat + "," + timeout);
			if (Build.VERSION.SDK_INT >= 19) {
				try {
					_am.setExact(AlarmManager.ELAPSED_REALTIME_WAKEUP, nextTimeHeartBeat, _pi_ping);
				} catch (Exception e) {
					_am.set(AlarmManager.ELAPSED_REALTIME_WAKEUP, nextTimeHeartBeat, _pi_ping);
				}
			} else {
				_am.set(AlarmManager.ELAPSED_REALTIME_WAKEUP, nextTimeHeartBeat, _pi_ping);
			}
		}
	}

	/**
	 * 获取心跳超时
	 *
	 * @return 心跳超时，单位毫秒
	 */
	@Override
	public int get_heartbeat_timeout() {
		return _curr_heart;
	}

	/**
	 * 设置客户端心跳发送频率
	 *
	 * @param heartbeat_timeout
	 *            心跳超时(单位毫秒) 值不小于30秒
	 */
	@Override
	public void set_heartbeat_timeout(int heartbeat_timeout) {
		if (heartbeat_timeout >= 30000) {
			_curr_heart = heartbeat_timeout;
		} else {
			if (BuildFlag.DEBUG) {
				Logger.w(TAG, "heartbeat timeout is ignore for it is less then 30000 ms.");
			}
		}
	}

    private void registerConnectivityReceiver() {
        if (!_connectivity_registered) {
            IntentFilter filter = new IntentFilter();
            filter.addAction(ConnectivityManager.CONNECTIVITY_ACTION);
            filter.addAction(Intent.ACTION_SCREEN_ON);
            filter.addAction(Intent.ACTION_SCREEN_OFF);
            filter.addAction(Intent.ACTION_POWER_CONNECTED);
            filter.addAction(Intent.ACTION_POWER_DISCONNECTED);
            filter.addAction(Intent.ACTION_USER_PRESENT);
            filter.addAction(Intent.ACTION_BATTERY_CHANGED);
            filter.setPriority(999);

            IntentFilter mNetworkProbeFilter = new IntentFilter();
            mNetworkProbeFilter.addAction(ConnectivityManager.CONNECTIVITY_ACTION);
            mNetworkProbeFilter.addAction(Intent.ACTION_SCREEN_ON);
            mNetworkProbeFilter.addAction(Intent.ACTION_SCREEN_OFF);
            mNetworkProbeFilter.addAction(Intent.ACTION_POWER_CONNECTED);
            mNetworkProbeFilter.addAction(Intent.ACTION_POWER_DISCONNECTED);
            mNetworkProbeFilter.addAction(Intent.ACTION_USER_PRESENT);
            mNetworkProbeFilter.addAction(Intent.ACTION_BATTERY_CHANGED);
            mNetworkProbeFilter.setPriority(999);

            _context.registerReceiver(mNetworkProbe, mNetworkProbeFilter);
            _context.registerReceiver(_connReceiver, filter);
            _connectivity_registered = true;
        }
    }

	private void unregisterConnectivityReceiver() {
		if (_connectivity_registered) {
			_context.unregisterReceiver(_connReceiver);
			_connectivity_registered = false;
		}
	}

	protected ClientConnection(Context context, AccountInfo account_info, ClientConfig clientConfig, IMCallback notify) {
		init(context, account_info, clientConfig, notify);
	}

	/**
	 * 映射网络类型
	 *
	 * @return net_type<br>
	 *         0:Unknown <br>
	 *         1:2G <br>
	 *         2:3G <br>
	 *         3:Wi-Fi <br>
	 *         4:Ethernet<br>
	 *         5:4G
	 * */
	protected int map_network_type(int type, int sub_type) {
		switch (type) {

		case ConnectivityManager.TYPE_WIFI:
			return 3;

		case ConnectivityManager.TYPE_MOBILE:
			switch (sub_type) {
			case LLConstant.NETWORK_TYPE_EDGE:
			case LLConstant.NETWORK_TYPE_GPRS:
				return 1;

			case LLConstant.NETWORK_TYPE_LTE:
				return 5;

			case LLConstant.NETWORK_TYPE_HSDPA:
			case LLConstant.NETWORK_TYPE_HSPA:
			case LLConstant.NETWORK_TYPE_HSPAP:
			case LLConstant.NETWORK_TYPE_HSUPA:
			case LLConstant.NETWORK_TYPE_UMTS:
			case LLConstant.NETWORK_TYPE_EVDO_0:
			case LLConstant.NETWORK_TYPE_EVDO_A:
			case LLConstant.NETWORK_TYPE_EVDO_B:
				return 2;

			default:
				return 0;
			}

		case ConnectivityManager.TYPE_ETHERNET:
			return 4;

		default:
			return 0;
		}
	}

	private class ConnectivityChangedReceiver extends BroadcastReceiver implements INetworkChanged {

		/** 上次的网络类型 */
		private int last_network_type = -1;

		private boolean last_inet_available;
		private ConnectivityManager mConnMgr;

		public ConnectivityChangedReceiver(Context context) {

			super();

			mConnMgr = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
			if (mConnMgr != null) {
				NetworkInfo aActiveInfo = mConnMgr.getActiveNetworkInfo();
				if (null != aActiveInfo) {
					last_inet_available = true;
					// _last_connected_net_type =
					last_network_type = aActiveInfo.getType();
				}
			}

			NetworkProbe.registerCallback(this);
		}

		private String get_str_nettype(int inettype) {
            switch (inettype) {
                case 1 :
                    return "2G";
                case 2:
                    return "3G";
                case 3:
                    return "Wi-Fi";
                case 4:
                    return "Ethernet";
                case 5:
                    return "4G";
                default:
                    return "Unknown";
            }
        }

		@Override
		public void onNetworkChanged(boolean available, int net_type, int sub_type) {

			if (available) {

				_net_type = map_network_type(net_type, sub_type);

				if (!last_inet_available) {
					Logger.i(TAG, String.format("network is available: " + get_str_nettype(_net_type)));
					notifyNetworkStateChange(true);
					// NIC transition
				} else if (last_network_type != net_type) {
					Logger.i(TAG, String.format(Locale.getDefault(), "network transition(net_type): %d ==>  %d", last_network_type, net_type));
					// 网络类型发现变化， 重连
					_dispatchClient.reset();
					pushEvent(new Event(LLConstant.EVENT_NETWORK_TYPE_CHANAGED, SystemClock.elapsedRealtime()));
				} else if (_current_state != null && _current_state.get_state() != ConnectionState.Connected) {
					// 如果网络有变化，尝试连接
					pushEvent(new Event(LLConstant.EVENT_CONNECT));
					if ( _interval_index > LLConstant.CONTINUOUS_FAILURES_TO_TRY_VIP) {
						_interval_index = 0;
						Logger.d(TAG, "onNetworkChanged: interval_index=0");
					}
				} else {
					pushEvent(_heartbeat_event);
				}

				_inetAvailable = true;
				last_inet_available = true;
				last_network_type = net_type;

			} else {

				Logger.i(TAG, String.format("network is unavailable."));
				_inetAvailable = false;
				notifyNetworkStateChange(false);
				last_inet_available = false;
			}
		}

		@Override
		public void onReceive(Context context_unused, Intent intent) {

			if (intent == null) {
				return;
			}

			String action = intent.getAction();
			if (action == null) {
				return;
			}

			if (action.equals(ConnectivityManager.CONNECTIVITY_ACTION)) {
				if (mConnMgr == null) {
					mConnMgr = (ConnectivityManager) _context.getSystemService(Context.CONNECTIVITY_SERVICE);
				}
				if (mConnMgr != null) {
					NetworkInfo aActiveInfo = mConnMgr.getActiveNetworkInfo();

					if (null != aActiveInfo) {
						onNetworkChanged(true, aActiveInfo.getType(), aActiveInfo.getSubtype());
					} else {
						onNetworkChanged(false, -1, -1);
					}
				}
			} else if (action.equals(Intent.ACTION_SCREEN_ON) || action.equals(Intent.ACTION_USER_PRESENT) || action.equals(Intent.ACTION_POWER_DISCONNECTED)) {
				if (_am != null) {
					synchronized (_alarmLock) {
						if (_am != null && _pi_ping != null) {
							_am.cancel(_pi_ping);
						}
					}
				}
				pushEvent(_heartbeat_event);
				_screen_off_time = 0;

				if (action.equals(Intent.ACTION_SCREEN_ON)) {
					releaseWakeLock(_business_WL);
				}

			} else if (action.equals(Intent.ACTION_SCREEN_OFF)) {
				_screen_off_time = SystemClock.elapsedRealtime();
				schedule_next_ping();
			}
		}
	}

	private boolean isInetAvailabe() {
		ConnectivityManager mConnMgr = (ConnectivityManager) _context.getSystemService(Context.CONNECTIVITY_SERVICE);
		if (mConnMgr != null) {
			NetworkInfo aActiveInfo = mConnMgr.getActiveNetworkInfo(); // 获取活动网络连接信息
			if (aActiveInfo != null) {
				_net_type = map_network_type(aActiveInfo.getType(), aActiveInfo.getSubtype());
			}
			return aActiveInfo != null && aActiveInfo.isAvailable();
		}
		return false;
	}

	class ScheduledTaskReceiver extends BroadcastReceiver {
		@Override
		public void onReceive(Context context, Intent intent) {
            try {
                if (intent == null) {
                    return;
                }

                String action = intent.getAction();
                if (action == null) {
                    return;
                }

                long id = intent.getLongExtra(SCHEDULE_TASK_EXTRA_ID, -1);
                if (id == -1) {
                    Logger.w(TAG, "id is -1, a ignored");
                    return;
                }

                // 需要考虑无网络的场景和连接失败的场景， 即使失败，下一个心跳任然要schedule, 否则CPU一旦休眠再也没有机会检查连接状态
                if (ACTION_PING.equals(action)) {
                    Logger.i(TAG, "ScheduledTaskReceiver: Recv Alarm Wakeup");
                    if (isInetAvailabe() && !isDeviceAwake()) {
                        acquireWakeLock(_ping_WL, TimeConst.WL_TIME_OUT);
                        Logger.v(TAG, "A : " + action);
                        pushEvent(_heartbeat_event);
                        schedule_next_ping();
                    } else {
                        Logger.v(TAG, "Ignore a as Inet is unavailable.");
                    }
                }
            }catch (Exception e){
                Logger.e(TAG, "S  Exception: " + Log.getStackTraceString(e));
            }
		}
	};

	/**
	 * 验证失败的状态
	 * */
	private class AuthFailedState extends State {

		public AuthFailedState() {
			super(ConnectionState.AuthFailed);
		}

		/**
		 * 进入该状态时触发
		 * */
		@Override
		public void OnEnter() {
			close();
			cancel_scheduled_task();
		}

		/**
		 * Got called when any event fired
		 * */
		@Override
		public void OnEeventFired(Event event) {

			super.OnEeventFired(event);

			boolean need_to_login = false;

			switch (event.get_event_id()) {

			case LLConstant.EVENT_CONNECT:
				need_to_login = true;
				break;

			case LLConstant.EVENT_CREDENTIAL_UPDATED: // 用户改密码了
				if (do_update_credential(event._arg)) {
					need_to_login = true;
				}

				break;

			case LLConstant.EVENT_SEND_MSG: {
				Message msg = ((MessageEvent) event).get_message();
				if (_inotify != null) {
					notifyFailedMessage(msg.getSn(), Constant.RESULT_UNAUTHORIZED, msg);
				}
			}
				break;

			case LLConstant.EVENT_INET_AVAILABLE:
			case LLConstant.EVENT_INET_UNAVAILABLE:
			case LLConstant.EVENT_NETWORK_TYPE_CHANAGED:
			case LLConstant.EVENT_SOCK_CLOSED:

				break;

			default:
				Logger.w(TAG, String.format(Locale.getDefault(), "%s : e unhandled: %d", this.getClass().getName(), event.get_event_id()));
				break;
			}

			if (need_to_login) { // 密码更新了， 或者用户要求登录，那么立即登录吧
				if (ClientConnection.this._inetAvailable) {
					set_currentState(ConnectionState.Connecting);
				} else {
					set_currentState(ConnectionState.Disconnected);
				}
			}
		}
	}

	/**
	 * 别处登录的状态
	 * */
	private class LoggedInElsewhereState extends State {
		public LoggedInElsewhereState() {
			super(ConnectionState.LoggedInElsewhere);
		}

		/**
		 * 进入该状态时触发
		 * */
		@Override
		public void OnEnter() {
			ClientConnection.this.pushEvent(new Event(LLConstant.EVENT_CONNECT));
		}

		/**
		 * Got called when any event fired
		 * */
		@Override
		public void OnEeventFired(Event event) {
			super.OnEeventFired(event);
			switch (event._event_id) {
			case LLConstant.EVENT_CONNECT:
				set_currentState(ConnectionState.Disconnected);
				break;
			}
		}
	}

	/**
	 * 连接断开的状态
	 * */
	private class DisconnectedState extends State {
		public DisconnectedState() {
			super(ConnectionState.Disconnected);
		}

		/**
		 * 进入该状态时触发
		 * */
		@Override
		public void OnEnter() {
			close();
			if (_inetAvailable) { // 自动登录逻辑
				pushEvent(new Event(LLConstant.EVENT_CONNECT));
			}
		}

		/**
		 * Got called when any event fired
		 * */
		@Override
		public void OnEeventFired(Event event) {

			super.OnEeventFired(event);

			boolean need_to_login = false;

			long connect_interval = 1000;

			switch (event.get_event_id()) {

			// 服务器返回的数据包
			case LLConstant.EVENT_GOT_PACKET:

				// 过滤掉上一个账号的信息
				if (event.get_account() == _account_info.get_account()) {
					try {
						HandlePacketResult result = handlePacket((Message) event._arg);
						if (result.equals(HandlePacketResult.Failed)) {
							set_currentState(ConnectionState.Disconnected);
						} else if (result.equals(HandlePacketResult.ReloggedIn)) {
							set_currentState(ConnectionState.LoggedInElsewhere);
						} else if (result.equals(HandlePacketResult.ReConnect)) {
							set_currentState(ConnectionState.LoggedInElsewhere);
						}
					} catch (Exception e2) {
						if (BuildFlag.DEBUG) {
							Logger.e(TAG, "HandlePacket threw: " + e2.getMessage());
						}
						set_currentState(ConnectionState.Disconnected);
					}
				} else {
					Logger.d(TAG, "p is filtered.");
				}

				break;

			case LLConstant.EVENT_CREDENTIAL_UPDATED: // 用户改密码了
				if (do_update_credential(event._arg)) {
					if (_inetAvailable) {
						need_to_login = true; // 用户主动登录
					}
				}
				break;

			case LLConstant.EVENT_SEND_HEARTBEAT:
			case LLConstant.EVENT_GET_MSG:
			case LLConstant.EVENT_CONNECT:
				if (_inetAvailable) {
					need_to_login = true; // 用户主动登录
				}
				break;

			case LLConstant.EVENT_NETWORK_TYPE_CHANAGED:
			case LLConstant.EVENT_INET_AVAILABLE:
				need_to_login = true;
				break;

			case LLConstant.EVENT_SEND_MSG: {
				if (_inetAvailable) {
					need_to_login = true; // 用户主动登录
					_interval_index = 0;
				}
			}
				break;

			case LLConstant.EVENT_INET_UNAVAILABLE:
			case LLConstant.EVENT_SOCK_CLOSED:
			case LLConstant.EVENT_DISCONNECT:
				break;

			default:
				Logger.e(TAG, String.format(Locale.getDefault(), "%s : e unhandled: %d", this.getClass().getName(), event.get_event_id()));
				break;
			}

			if (need_to_login) {

				if (_interval_index < reconnect_intervals.length && _interval_index >= 0) {
					connect_interval = 1000 * reconnect_intervals[_interval_index++];
				} else if (_interval_index >= reconnect_intervals.length) {
					connect_interval = 1000 * reconnect_intervals[reconnect_intervals.length - 1];
				} else {
					_interval_index = 0;
					connect_interval = 1000 * reconnect_intervals[_interval_index++];
				}

				// account is switched, connect immediately
				if (_account_switched) {
					_account_switched = false;
					connect_interval = 1000;
				}

				if (connect_interval > 0) {

					if (is_overloaded) {
						is_overloaded = false;
						Logger.w(TAG, "server is overloaded, set longer connect interval");
						connect_interval = TimeConst.OVERLOADED_LOGIN_INTERVAL + _random.nextInt(TimeConst.OVERLOADED_LOGIN_INTERVAL);
					}

					int minimal_wait_time = ClientConnection.this.getMinimalWaitTime();
					try {
						if (minimal_wait_time > 0) {
							Logger.d(TAG, "minimal w t" + minimal_wait_time);
							Thread.sleep(minimal_wait_time);
						}
					} catch (InterruptedException e1) {

					}

					try {
						synchronized (_connectLock) {
							_connectLock.wait(connect_interval);
						}
					} catch (InterruptedException e) {
						return;
					}
				}

				if (_inetAvailable) {
					set_currentState(ConnectionState.Connecting);
				}
			}
		}
	}

	/**
	 * 已经连接上了
	 * */
	private class ConnectedState extends State {

		public ConnectedState() {
			super(ConnectionState.Connected);
		}

		@Override
		public void OnEnter() {
			if (!send_user_messages()) {
				set_currentState(ConnectionState.Disconnected);
			}
		}

		@Override
		public void OnExit() {
			close();
		}

		@Override
		public void OnEeventFired(Event event) {
			super.OnEeventFired(event);

			switch (event.get_event_id()) {

			// 服务器返回的数据包
			case LLConstant.EVENT_GOT_PACKET:

				// 过滤掉上一个账号的信息
				if (event.get_account() == _account_info.get_account()) {
					try {
						HandlePacketResult result = handlePacket((Message) event._arg);
						if (result.equals(HandlePacketResult.Failed)) {
							set_currentState(ConnectionState.Disconnected);
						} else if (result.equals(HandlePacketResult.ReloggedIn)) {
							set_currentState(ConnectionState.LoggedInElsewhere);
						} else if (result.equals(HandlePacketResult.ReConnect)) {
							set_currentState(ConnectionState.LoggedInElsewhere);
						}
					} catch (Exception e2) {
						if (BuildFlag.DEBUG) {
							Logger.e(TAG, "handlePacket threw: " + e2.getMessage());
						}
						set_currentState(ConnectionState.Disconnected);
					}
				} else {
					Logger.w(TAG, "p is filtered.");
				}

				break;

			case LLConstant.EVENT_SOCK_CLOSED: // 因故SOCKET断开了
			case LLConstant.EVENT_INET_UNAVAILABLE: // 网络断开了
				// 网络迁移 Wi-Fi --> Mobile, Mobile--> Wi-Fi etc..
			case LLConstant.EVENT_NETWORK_TYPE_CHANAGED:

				// Logger.d(TAG, String.format("event %d e_time %d, c_time %d",
				// event.get_event_id(), event.get_time(), _connect_time));

				if (event.get_time() > _connect_time) {
					set_currentState(ConnectionState.Disconnected);
				} else {
					Logger.w(TAG, String.format(Locale.getDefault(), "event dropped for it is out of date %d", event.get_event_id()));
				}

				break;

			case LLConstant.EVENT_GET_MSG:
				GetMsgEvent getMsgEvent = (GetMsgEvent) event;
				if (getMsgEvent != null) {
					if (!(getMessageInner(getMsgEvent.get_info_type(), getMsgEvent.get_ids(), getMsgEvent.get_parameters()))) {
						set_currentState(ConnectionState.Disconnected);
					}
				}
				break;

			case LLConstant.EVENT_CREDENTIAL_UPDATED: // 或者密码被更改了, 我们需要重新连接
				if (do_update_credential(event._arg)) {
					set_currentState(ConnectionState.Disconnected);
				}
				break;

			case LLConstant.EVENT_DISCONNECT: // 用户要求登出
				set_currentState(ConnectionState.Disconnected);
				return;

			case LLConstant.EVENT_SEND_MSG:
				if (!send_user_messages()) {
					set_currentState(ConnectionState.Disconnected);
				}
				break;

			case LLConstant.EVENT_CONNECT:
			case LLConstant.EVENT_INET_AVAILABLE:
				// ?
				break;

			default:
				Logger.e(TAG, String.format(Locale.getDefault(), "%s : e unhandled: %d", this.getClass().getName(), event.get_event_id()));
				break;
			}
		}
	}

	/**
	 * 正在连接服务器
	 * */
	private class ConnectingState extends State {

		/**
		 * 子状态等待初始登录包的响应
		 * */
		private final static int S_WAIT_FOR_INIT_RESP = 0;

		/**
		 * 子状态等待登录包的响应
		 * */
		private final static int S_WAIT_FOR_LOGIN_RESP = 1;

		int _cur_sub_state = S_WAIT_FOR_INIT_RESP;

		// sn 匹配， 过滤掉上次连接的残留包
		long init_sn = 0;
		long login_sn = 0;

		public ConnectingState() {
			super(ConnectionState.Connecting);
		}

		/**
		 * 进入该状态时立即连接
		 * */
		@Override
		public void OnEnter() {
			// 恢复状态
			_cur_sub_state = S_WAIT_FOR_INIT_RESP;

			boolean inetAvailable = _inetAvailable;
			long start = System.currentTimeMillis();

			if (inetAvailable && connect()) {

				Logger.d(TAG, "connected.");

				// 通知接收线程SOCKET已经连接上了, 可以开始接收数据
				_receiver.sendCmd(Receiver.CMD_START);

				// 发送初始登录包
				init_sn = get_sn();
				if (!sendInitLogin(init_sn)) {
					Logger.e(TAG, "Failed to send il req.");
					set_currentState(ConnectionState.Disconnected);
				}
			} else {
				set_currentState(ConnectionState.Disconnected);
			}

			if (inetAvailable) {
				Logger.i(TAG, "connect costs: " + (System.currentTimeMillis() - start));
			}
		}

		/**
		 * 预先检查包是否有错误
		 *
		 * @return 0 没问题<br>
		 *         1 包不正确<br>
		 *         其他： 服务器返回的错误代码
		 * */
		int checkPacket(Message packet) {

			long sn = packet.getSn();

			// 去除消息队列等待的消息， 因为收到服务器的响应了
			Long lsn = Long.valueOf(sn);
			if (_pendingMessages.containsKey(lsn)) {
				_pendingMessages.remove(lsn);
			}

			if (!packet.hasResp()) {
				Logger.e(TAG, "packet has no resp, sub_state is " + _current_state);
				return 1;
			}

			if (packet.getResp().hasError() && packet.getResp().getError() != null) {
				com.huajiao.comm.protobuf.messages.CommunicationData.Error err = packet.getResp().getError();
				int err_code = err.getId();

				// server error treated as overloaded
				if (err_code == Error.DATABASE_IS_TOO_BUSY || err_code == Error.SERVER_OVERLOADED || err_code == Error.SES_REFUSED
						|| err_code == Error.DATABSE_EXCEPTION || err_code == Error.SERVER_LOGIN__FAILED || err_code == Error.SESSION_EXCEPTION) {
					is_overloaded = true;
				}

				return err_code;
			}

			return 0;
		}

		@Override
		public void OnEeventFired(Event event) {
			super.OnEeventFired(event);

			switch (event.get_event_id()) {

			// 服务器返回的数据包
			case LLConstant.EVENT_GOT_PACKET:

				Message packet = (Message) event._arg;

				// 预检查
				int err = checkPacket(packet);

				if (_cur_sub_state == S_WAIT_FOR_INIT_RESP) {

					ConnectionState nextState = ConnectionState.Connecting;

					do {

						if (packet.getSn() != init_sn) {
							Logger.w(TAG, "A drop useless packet: " + packet.getMsgid());
							return;
						}

						if (err != 0) {
							nextState = ConnectionState.Disconnected;
							break;
						}

						if (packet.getMsgid() != MessageId.InitLoginResp || !packet.getResp().hasInitLoginResp()) {
							Logger.e(TAG, "A: resp is not found.");
							nextState = ConnectionState.Disconnected;
							break;
						}

						// compute the first RTT
						_rtt[0] = SystemClock.elapsedRealtime() - _packet_start;

						InitLoginResp init_login_res = packet.getResp().getInitLoginResp();
						_server_ran = init_login_res.getServerRam();
						_init_logged_in = true;
						_socket_in = new RC4InputStream(_account_info.get_password(), _socket_in.getInputStream());

						login_sn = get_sn();
						// 发送登录包
						if (!sendLogin(login_sn)) {
							Logger.e(TAG, "Failed to send B.");
							nextState = ConnectionState.Disconnected;
							break;
						}

						_cur_sub_state = S_WAIT_FOR_LOGIN_RESP;

					} while (false);

					set_currentState(nextState);

				} else if (_cur_sub_state == S_WAIT_FOR_LOGIN_RESP) {

					ConnectionState nextState = ConnectionState.Disconnected;

					do {

						if (packet.getSn() != login_sn) {
							Logger.w(TAG, "B drop useless p: " + packet.getMsgid());
							return;
						}

						if (err == Error.USER_INVALID) {
							Logger.e(TAG, "Get error USER_INVALID when log in");
							nextState = ConnectionState.AuthFailed;
							break;
						} else if (err != 0) {
							break;
						}

						if (packet.getMsgid() != MessageId.LoginResp || !packet.getResp().hasLogin()) {
							Logger.e(TAG, "r is not found.");
							break;
						}

						// compute the second RTT
						_rtt[1] = SystemClock.elapsedRealtime() - _packet_start;

						LoginResp loginres = packet.getResp().getLogin();
						_sessionKey = loginres.getSessionKey();

						if (_sessionKey == null || _sessionKey.length() == 0) {
							Logger.i(TAG, "login: use special sessionkey");
							//break; may be null, mean follow
						}

						Logger.i(TAG, String.format(Locale.getDefault(), "F %d, %d", _rtt[0], _rtt[1]));

						synchronized (_time_lock) {
							// 获取服务器的登录时间, 注意: 只需要赋值一次
							// if (_time_base == 0) {
							long avg_rtt = (_rtt[0] + _rtt[1]) / 2;
							_server_time = ((long) loginres.getTimestamp()) * 1000;
							_server_time -= (avg_rtt / 2);
							_time_base = SystemClock.elapsedRealtime();
							// }
						}

						_socket_in = new RC4InputStream(_sessionKey, _socket_in.getInputStream());
						_socket_out = new RC4OutputStream(_sessionKey, _socket_out.getOutputStream());
						_connect_time = SystemClock.elapsedRealtime();
						_logged_in = true;

						if ( (_reconnect_hosts != null && _reconnect_hosts.size() > 0) || getAllMessage() ) {
							_interval_index = 0;
							_lvs_index = 0;
							nextState = ConnectionState.Connected;
							_reconnect_hosts = null;
						}

					} while (false);
					set_currentState(nextState);
				}

				break;

			case LLConstant.EVENT_SEND_HEARTBEAT:

			case LLConstant.EVENT_SOCK_CLOSED: // 因故SOCKET断开了
			case LLConstant.EVENT_INET_UNAVAILABLE: // 网络断开了
				// 网络迁移 Wi-Fi --> Mobile, Mobile--> Wi-Fi etc..
			case LLConstant.EVENT_NETWORK_TYPE_CHANAGED:

				// Logger.d(TAG, String.format("event %d e_time %d, c_time %d",
				// event.get_event_id(), event.get_time(), _connect_time));
				if (event.get_time() > _connect_time) {
					set_currentState(ConnectionState.Disconnected);
				} else {
					Logger.w(TAG, String.format(Locale.getDefault(), "e dropped for OOD %d", event.get_event_id()));
				}

				break;

			case LLConstant.EVENT_CREDENTIAL_UPDATED: // 或者密码被更改了, 我们需要重新连接
				if (do_update_credential(event._arg)) {
					set_currentState(ConnectionState.Disconnected);
				}
				break;

			case LLConstant.EVENT_DISCONNECT: // 用户要求登出
				set_currentState(ConnectionState.Disconnected);
				return;

			case LLConstant.EVENT_GET_MSG:

			case LLConstant.EVENT_CONNECT:

			case LLConstant.EVENT_SEND_MSG:
				// 这个时候还无法发送消息， 不过别担心， 未发送的消息都保存在消息队列里面， 一旦登录成功会被自动发送
				break;

			default:
				Logger.e(TAG, "Connecting State: unexpected e: " + event.get_event_id());
				break;
			}

		}
	}

	/**
	 * 负责连接socket, 发送数据
	 * */
	private class Sender extends Thread {

		@Override
		public void run() {

			Event event = null;

			// 心跳超时
			long timeout = 0;
			// 马上要超时的消息或心跳的剩余超时时间
			long p_timeout = 0;

			// 设置初始状态
			set_currentState(ConnectionState.Disconnected);

			while (!_quit) {

				try {

					timeout = _curr_heart;

					// 看最近是否有消息快要超时
					p_timeout = getLeastTimeout();
					if (_pendingMessages.size() > 0 && p_timeout < timeout) {
						timeout = p_timeout;
					}

					event = _eventQueue.poll(timeout, TimeUnit.MILLISECONDS);

					// 处理事件
					if (event != null) {

						switch (event.get_event_id()) {

						case LLConstant.EVENT_GET_STATE:
							reportState();
							break;

						case LLConstant.EVENT_SEND_HEARTBEAT:
							do_send_heartbeat(true);
							if (get_state() != ConnectionState.Connected && get_state() != ConnectionState.AuthFailed) {
								pushEvent(new Event(LLConstant.EVENT_CONNECT));
							}
							break;

						case LLConstant.EVENT_GOT_HEARTBEAT_ACK:

							// 收到了心跳回复
							Long lsn = Long.valueOf(HEARTBEAT_SN);
							if (_pendingMessages.containsKey(lsn)) {
								_pendingMessages.remove(lsn);
							}

							schedule_next_ping();
							releaseWakeLock(_ping_WL);

							break;

						case LLConstant.EVENT_SEND_MSG:

							// put message in pending map
							MessageEvent me = (MessageEvent) event;
							Long key = Long.valueOf(me.get_message().getSn());
							_pendingMessages.put(key, me);
							// try to send the message, if it is connected
							if (_current_state.get_state() == ConnectionState.Connected) {
								_current_state.OnEeventFired(event);
							}

							break;

						case LLConstant.EVENT_GOT_PACKET:

							// 过滤掉上一个账号的信息
							if (event.get_account() != null && !event.get_account().equals(_account_info.get_account())) {
								Logger.w(TAG, "packet of previous account is filtered.");
								break;
							}

							if (_current_state.get_state() == ConnectionState.Connected) {
								schedule_next_ping();
							}

							_current_state.OnEeventFired(event);
							break;

						default:
							_current_state.OnEeventFired(event);
							break;
						}
					}

					// 如果任何消息超时， 说明很可能跟服务器的连接已经断开
					if (updatePendingMessageStatus() == PendingMessageStatus.TimeoutOccurred
							&& (_current_state != null && !_current_state.get_state().equals(ConnectionState.AuthFailed))) {
						set_currentState(ConnectionState.Disconnected);
					}

					// 每次处理任意事件触发发送心跳包。
					do_send_heartbeat(false);

				} catch (InterruptedException e) {
					continue;
				} catch (Exception e1) {
					Logger.e(TAG, "S  Exception: " + Log.getStackTraceString(e1));
				}
			}

			if (_quit) {
				Logger.d(TAG, "S exits.");
			} else {
				Logger.e(TAG, "S exits abnormally, probably vm is quiting!");
			}

			_quit = true;
		}

		/**
		 * 如果 send_immediately为 true, 频道最小控制在5秒， 否则控制在心跳频率
		 */
		private void do_send_heartbeat(boolean send_immediately) {
			if (get_state().equals(ConnectionState.Connected)) {
				long cur_ping_time = System.currentTimeMillis();
				long diff = cur_ping_time - _last_ping_time;
				if ((send_immediately && diff > 5000) || diff >= _curr_heart) {
					acquireWakeLock(_ping_WL, TimeConst.WL_TIME_OUT);
					_last_ping_time = cur_ping_time;
					if (!send_heartbeat_packet()) {
						set_currentState(ConnectionState.Disconnected);
						releaseWakeLock(_ping_WL);
					}
				}
			}
		}
	}

	/**
	 * 负责接送数据
	 * */
	private class Receiver extends Thread {

		private BlockingQueue<Integer> _threadMessageQueue = new LinkedBlockingQueue<Integer>();

		// 已经登录成功可以开始读取数据
		public static final int CMD_START = 0;

		// 执行退出动作
		public static final int CMD_STOP = 1;

		// 主动关闭socket
		public static final int CMD_CLOSE_SOCKET = 2;

		/**
		 * 给接收线程发送命令<br>
		 * */
		public void sendCmd(int code) {
			_threadMessageQueue.offer(Integer.valueOf(code));
		}

		@Override
		public void run() {

			Integer cmd = Integer.valueOf(0);
			Message message = null;
			String currAcc = null;

			while (!_quit) {

				try {
					cmd = null;
					cmd = _threadMessageQueue.poll(300, TimeUnit.SECONDS);
				} catch (InterruptedException e) {

				}

				if (cmd == null) {
					continue;
				}

				if (cmd.intValue() == CMD_START) {

					Logger.i(TAG, "reading");

					while (!_quit) {
						currAcc = _account_info.get_account();
						message = readPacket();
						if (message == null) { // socket is closed probably
							Logger.i(TAG, "reading failed!!! ");
							Integer temp = _threadMessageQueue.peek();
							if (temp != null && temp.intValue() == CMD_CLOSE_SOCKET) {
								temp = _threadMessageQueue.poll();
							} else {
								pushEvent(new Event(LLConstant.EVENT_SOCK_CLOSED, SystemClock.elapsedRealtime()));
							}
							break;
						} else {
							Event event = new Event(LLConstant.EVENT_GOT_PACKET, message);
							event.set_account(currAcc);
							pushEvent(event);
						}

					} // end of inner while loop

					Logger.i(TAG, "done-reading");

				} else if (cmd.intValue() == CMD_STOP) {
					break;
				} else if (cmd.intValue() == CMD_CLOSE_SOCKET) {
					// 忽略， 可能由于多次调用close造成
				}
			}

			if (_quit) {
				Logger.d(TAG, "Receiver thread exits normally!");
			} else {
				Logger.e(TAG, "Receiver thread exits abnormally, probably vm is quiting!");
			}

			// the critical thread
			_quit = true;
		}
	}


	/***
	 * upload report
	 * */
	private boolean connectfail_report(String uid, String domain, String ip, String did, String reason) {

		if (uid == null || domain == null || ip == null || did == null || reason == null) {
			return false;
		}

		if (REPORT_SERVER_HOST.isEmpty()){
			return false;
		}

		try {
			String url = String.format(Locale.US, "http://%s/huajiao/linkerr.html?ip=%s&rip=%s&net=%d&uid=%s&did=%s&plf=android&r=%s",
					REPORT_SERVER_HOST,
					domain,
					ip,
					_net_type,
					uid,
					java.net.URLEncoder.encode(did, "utf-8"),
					java.net.URLEncoder.encode(reason, "utf-8")
			);

			Logger.d(TAG, "connectfail_report url=" + url);

			ReportTask reportTask = new ReportTask();
			reportTask.execute(url);
		}catch (Exception ex)
		{
			Logger.e(TAG, "connectfail_report fail"+ex.getMessage());
		}

		return true;
	}

	/***
	 * 异步打点connect fail
	 * */
	private static class ReportTask extends AsyncTask<String, Void, Boolean> {

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

			return result;
		}
	}

	/**
	 * 需要确认的消息的处理结果
	 * **/
	private enum PendingMessageStatus {
		/**
		 * 队列已经为空， 没有要处理的消息
		 * */
		QueueIsEmpty,

		/**
		 * 有消息在指定时间内没有得到服务器的确认, 这里需要切换到端口状态了
		 * */
		TimeoutOccurred,

		/**
		 * 有消息需要继续等待服务器的回执
		 * **/
		Continue
	}

	/**
	 * 增加命令到发送队列
	 * */
	private void pushEvent(Event cmd) {

		if (cmd == null) {
			Logger.e(TAG, "cmd is null!");
			return;
		}

		if (!_eventQueue.offer(cmd)) {
			Logger.e(TAG, "event queue offer failed!!!");
		}
	}

	public void reportState() {
		if (_current_state == null) {
			_inotify.onCurrentState(new CurrentStatePacket(ConnectionState.Disconnected));
		} else {
			_inotify.onCurrentState(new CurrentStatePacket(_current_state.get_state()));
		}
	}

	/**
	 * 状态迁移， 注意该函数只能被Sender线程调用， 否则后续结果无法预见
	 * */
	void set_currentState(ConnectionState eState) {

		State state = _states.get(eState);

		if (state == null) {
			return;
		}

		// Ignore, no need to do transition
		if (_current_state != null && state.equals(_current_state)) {
			return;
		}

		State oldState = _current_state;
		State newState = state;

		// 退出状态, 同一个状态变化， 不触发动作
		if (_current_state != null && oldState != newState) {
			_current_state.OnExit();
		}

		_current_state = state;

		if (_inotify != null && oldState != newState) {
			ConnectionState newStateEnum = (newState == null ? ConnectionState.Disconnected : newState.get_state());
			ConnectionState oldStateEnum = (oldState == null ? ConnectionState.Disconnected : oldState.get_state());
			Logger.i(TAG, String.format("s %s ==> %s", oldStateEnum.toString(), newStateEnum.toString()));
			_inotify.onStateChanged(new StateChangedPacket(oldStateEnum, newStateEnum));
		}

		// 进入状态
		if (oldState != newState) {
			state.OnEnter();
		}

		if (oldState != null && oldState.get_state().equals(ConnectionState.Connected) && newState.get_state().equals(ConnectionState.Disconnected)) {
			_last_disconnect_time = System.currentTimeMillis();
		}
	}

	/***
	 * 初始状态机
	 * */
	protected void initState() {
		_states.put(ConnectionState.Disconnected, new DisconnectedState());
		_states.put(ConnectionState.Connecting, new ConnectingState());
		_states.put(ConnectionState.Connected, new ConnectedState());
		_states.put(ConnectionState.AuthFailed, new AuthFailedState());
		_states.put(ConnectionState.LoggedInElsewhere, new LoggedInElsewhereState());
	}

	static String getRandomString(int length) {
		String r = "";
		for (int i = 0; i < length; i++) {
			r += (char) (32 + _random.nextInt(94));
		}
		return r;
	}

	/**
	 * 获取唯一消息SN, 发送消息的SN可以由此接口获取
	 * */
	@Override
	public long get_sn() {
		return _snSeed.incrementAndGet();
	}

	private int get_id() {
		return _snSeedInt.incrementAndGet();
	}

	/**
	 * 传送消息给应用层
	 * */
	private boolean delivery_message(String sender, String receiver, String info_type, int msg_type, long msg_id, long sn, long time_sent, byte[] body,
			long lastest_msg_id, boolean valid) {

		boolean result = false;

		try {

			acquireWakeLock(_business_WL, LLConstant.BUSINESS_WL_TIMEOUT);
			// long _sn, String info_type, String _from, String to, byte[]
			// _content, long _id, int _msg_type, long _date, long _latest_id,
			// long server_time, boolean valid
			MsgPacket packet = new MsgPacket(sn, info_type, sender, receiver, body, msg_id, msg_type, time_sent, lastest_msg_id, getCurrentServerTime(), valid);
			_inotify.onMessage(packet);
			result = true;

		} catch (Exception e) {
			Logger.w(TAG, Log.getStackTraceString(e));
		}
		return result;
	}

	/**
	 * 更新账号, 需要更新账号或密码后才会尝试重新登录, 注意仅当账号发送改变的时候last_msg_id才会生效
	 **/
	@Override
	public void switch_account(AccountInfo account_info, ClientConfig client_config) {
		if(account_info != null && _account_info != null) {
			Logger.setUid(account_info.get_account());
			Logger.i(TAG, "switch acc "+_account_info.get_account()+" -> "+account_info.get_account());
		}
		if (null != account_info && !account_info.equals(_account_info)) {
			_jid = null;
			pushEvent(new Event(LLConstant.EVENT_CREDENTIAL_UPDATED, new Object[] { account_info, client_config }));
		} else {
			Logger.i(TAG, "switch acc but the old_account and new_acount equals,don't nothings");
		}
	}

	/**
	 * @param arguments
	 *            AccountInfo 和 ClientConfig 对象
	 * */
	private boolean do_update_credential(Object arguments) {

		Object[] objects = (Object[]) arguments;
		if (objects == null || objects.length < 2) {
			if (BuildFlag.DEBUG) {
				Log.w(TAG, "do_update_credential incorrect argument nubmer");
			}
			return false;
		}

		AccountInfo account_info = (AccountInfo) objects[0];
		if (account_info == null) {
			return false;
		}

		if (!account_info.equals(_account_info)) {

			for (MessageFlag flag : _msg_flags.values()) {
				flag.switch_account(account_info.get_account());
			}

			_interval_index = _lvs_index = 0;
			_account_info = account_info;

			_account_switched = true;
			// 清除掉所有的消息， 他们可能是上一个账号的我们已经不关心了。
			_pendingMessages.clear();
			return true;
		}

		return false;
	}

	protected boolean sendPacket(Message packet) {
		return sendPacket(packet, false);
	}

	/***
	 * send packet to server
	 *
	 * @param packet
	 * @return
	 */
	protected boolean sendPacket(Message packet, boolean expect_response) {

		boolean result = false;

		do {
			if (null == packet) {
				Logger.e(TAG, "p is null!");
				break;
			}

			if (!this._connected) {
				Logger.e(TAG, String.format(Locale.getDefault(), "msgId:%d, _connected is failed! send packet failed",packet.getSn()));
				break;
			}

			int protobufSize = packet.getSerializedSize();
			int totalLen = protobufSize + 4;
			byte[] pbData = packet.toByteArray();

			try {

				if (!_init_logged_in) {
					totalLen += 12;
				}

				byte[] buffer = new byte[totalLen];
				int index = 0;

				if (!_init_logged_in) {
					System.arraycopy(MagicCode, 0, buffer, 0, MagicCode.length);
					index += MagicCode.length;
				}

				System.arraycopy(Utils.int_to_bytes(totalLen), 0, buffer, index, 4);
				index += 4;

				RC4 enc = _socket_out.getRC4();
				if ( enc != null) { //only enc valid to encry
					pbData = enc.encry_RC4_byte(pbData);
				}
				System.arraycopy(pbData, 0, buffer, index, pbData.length);

				_socket_out.getOutputStream().write(buffer);
				_socket_out.getOutputStream().flush();

				result = true;

				// 加入等待队列
				if (expect_response) {

					MessageEvent me = null;
					Long key = Long.valueOf(packet.getSn());

					if (_pendingMessages.containsKey(key)) {
						me = _pendingMessages.get(key);
					} else {
						me = new MessageEvent(packet, TimeConst.PACKET_RESP_TIMEOUT, false);
						_pendingMessages.put(packet.getSn(), me);
					}

					if (me.get_send_count() > 1) {
						Logger.d(TAG, "resend : " + me.get_message().getSn());
					}

					me.set_sent_time();
					me.set_has_been_sent(true);
				}

				_last_ping_time = System.currentTimeMillis();

			} catch (Exception e) {
				Logger.e(TAG, "msgId:"+ packet.getSn() +",sp  failed: " + e.getMessage());
			}

		} while (false);

		return result;
	}

	/**
	 * 发送初始登录包
	 * */
	private boolean sendInitLogin(long sn) {

		if (_init_packtet_sent) {
			return true;
		}

		_client_ran = getRandomString(8);

		InitLoginReq req1 = new InitLoginReq();
		req1.setClientRam(_client_ran);

		Request req = new Request();
		req.setInitLoginReq(req1);

		if (_account_info.get_signature() != null && _account_info.get_signature().length() > 0) {
			req1.setSig(_account_info.get_signature());
		}

		Message msg = new Message();
		msg.setMsgid(MessageId.InitLoginReq);
		msg.setSn(sn);
		msg.setSender(_account_info.get_account());
		msg.setReq(req);
		_init_packtet_sent = sendPacket(msg, true);

		_packet_start = SystemClock.elapsedRealtime();

		return _init_packtet_sent;
	}

	/**
	 * 发送登录包
	 * */
	private boolean sendLogin(long sn) {

		if (this._logged_in) {
			return true;
		}

		String secSrc = _server_ran + getRandomString(8);
		RC4 rc4 = new RC4(_account_info.get_password());
		byte[] secrets = rc4.encry_RC4_byte(secSrc.getBytes());

		String device_id = (_account_info.get_device_id() == null || _account_info.get_device_id().length() == 0) ? "empty" : _account_info.get_device_id();

		// 心跳设为两分钟， 因为有些机型无法按指定时间唤醒而是会对齐为5分钟， 典型的手机:小米, 如果服务器心跳超时为3分钟 1X3，
		// 那么等手机唤醒时， 服务器已经断开连接了， 如果网络还稳定， 这是不合理的。
		// add not_encrypt field, MUST be true
		LoginReq req1 = new LoginReq().setNetType(_net_type).setMobileType(LLConstant.MOBILE_TYPE).setServerRam(_server_ran).setDeviceid(device_id)
				.setAppId(_clientConfig.getAppId()).setSecretRam(ByteStringMicro.copyFrom(secrets)).setHeartFeq(_report_heart_time).setNotEncrypt(true);

		// if signature is not specified, this is a guest user
		if (_account_info.get_signature() == null || _account_info.get_signature().length() == 0) {
			String verfCode = makeVerfCode(_account_info.get_account());
			req1.setVerfCode(verfCode);
		}

		Request req = new Request();
		req.setLogin(req1);

		Message msg = new Message();
		msg.setSender(_account_info.get_account()).setSn(sn).setReq(req).setMsgid(MessageId.LoginReq).setSenderType("jid");

		boolean result = sendPacket(msg, true);

		_packet_start = SystemClock.elapsedRealtime();

		return result;
	}

	private IPAddress get_server() {

		//reconnect_host
		if ( _reconnect_hosts != null && _reconnect_hosts.size() > 0 ){
			if (_interval_index >= 0 && _interval_index <= _reconnect_hosts.size() ) {
				int index = _interval_index>0 ? _interval_index - 1:_interval_index;
				return _reconnect_hosts.get(index);
			}else {
				_interval_index = 0;
				_reconnect_hosts = null; //try all reconnect_host, won't reagain
			}
		}

		// test server
		if (!_clientConfig.getServer().toLowerCase(Locale.US).equals(LLConstant.OFFICIAL_SERVER)) {
			return new IPAddress(_clientConfig.getServer(), _clientConfig.getPort());
		}

		IPAddress dispatchResultServer = new IPAddress(null, 0);
		DispatchClient.GetResult result = _dispatchClient.getDispatchResultServer(_clientConfig, DISPATCH_SERVER_HOST, _interval_index, _account_info.get_account(), dispatchResultServer);
		if ( result == DispatchClient.GetResult.SUCCESS ){
			return dispatchResultServer;
		}else if ( result == DispatchClient.GetResult.FAIL ){
			_interval_index = 0;
		}

		if (_interval_index > LLConstant.CONTINUOUS_FAILURES_TO_TRY_VIP && LLConstant.LVS_IP.length > 0) {

			IPAddress server = null;

			if (_lvs_index < LLConstant.LVS_IP.length) { // try port 80
				server = new IPAddress(LLConstant.LVS_IP[_lvs_index], LLConstant.PORT[0]);
			} else if (_lvs_index >= LLConstant.LVS_IP.length) { // try port 443
				server = new IPAddress(LLConstant.LVS_IP[_lvs_index % LLConstant.LVS_IP.length], LLConstant.PORT[1]);
			}

			_lvs_index++;
			if (_lvs_index >= LLConstant.LVS_IP.length * 2) {
				_lvs_index = 0;
			}

			if (server != null) {
				return server;
			}
		}

		// by default use server configured.
		return new IPAddress(_clientConfig.getServer(), LLConstant.PORT[_interval_index%2]);
	}

	/**
	 * Connect to server synchronously
	 *
	 * @return
	 * */
	protected boolean connect() {

		if (!this._inetAvailable) {
			return false;
		}

		if (this._connected) {
			Logger.e(TAG, "already connected, ignore!");
			return true;
		}

		try {
			IPAddress server = get_server();
			try {
				this._connected = false;
				this._logged_in = false;
				this._init_packtet_sent = false;

				if (_socket != null && _socket.isConnected()) {
					_socket.close();
				}

				SocketAddress sockaddr = new InetSocketAddress(server.get_ip(), server.get_port());

				_socket = new Socket();
				try {
					_socket.setKeepAlive(true);
				} catch (SocketException se) {
					// Logger.w(TAG, "method 1 failed");
				}

				try {
					_socket.setSoLinger(false, 0);
				} catch (SocketException se) {
					// Logger.w(TAG, "method 2 failed");
				}

				try {
					_socket.setTcpNoDelay(true);
				} catch (SocketException se) {
					// Logger.w(TAG, "method 3 failed");
				}

				Logger.i(TAG, String.format(Locale.US, "connecting to %s:%d", server.get_ip(), server.get_port()));
				_socket.connect(sockaddr, TimeConst.SOCKET_CONNECT_TIMEOUT);

				Logger.i(TAG, "addr: " + _socket.getRemoteSocketAddress().toString());

				InputStream in = _socket.getInputStream();
				OutputStream out = _socket.getOutputStream();

				_socket_in = new RC4InputStream(_clientConfig.getDefaultKey(), in);
				_socket_out = new RC4OutputStream(_clientConfig.getDefaultKey(), out);
				_connected = true;

			} catch (SocketTimeoutException e) {
				String ip;
				try {
					ip = _socket.getRemoteSocketAddress().toString();
				}catch (Exception ex) {
					ip = "null";
				}
				connectfail_report(_account_info.get_account(), server.get_ip(), ip, _account_info.get_device_id(), "timeout");

				Logger.e(TAG, "S timeout");
				_connected = false;
			} catch (Exception ex) {
				// Logger.e(TAG, "C Error: " + ex.getMessage());
				String ip;
				try {
					ip = _socket.getRemoteSocketAddress().toString();
				}catch (Exception ipex) {
					ip = "null";
				}
				connectfail_report(_account_info.get_account(), server.get_ip(), ip, _account_info.get_device_id(), ex.getMessage());

				if(JhFlag.enableDebug()){
					Logger.e(TAG, Log.getStackTraceString(ex));
				} else {
					Logger.e(TAG, ex.getMessage());
				}
				_connected = false;
			}

		}catch (Exception ex) {
			if(JhFlag.enableDebug()){
				Logger.e(TAG, Log.getStackTraceString(ex));
			} else {
				Logger.e(TAG, ex.getMessage());
			}
			_connected = false;
		}

		return _connected;
	}

	protected Message parsePacket(final byte buffer[]) {

		Message msg = null;
		RC4 rc4 = null;
		String decryptKey = null;

		if (buffer == null || buffer.length == 0) {
			return msg;
		}

		try {
			decryptKey = _logged_in ? _sessionKey : _account_info.get_password(); // determine
			if (!_init_logged_in) {
				decryptKey = _clientConfig.getDefaultKey();
			}

			if (decryptKey != null && decryptKey.length() > 0) {
				rc4 = new RC4(decryptKey);
				rc4.decry_RC4(buffer);
			}
			msg = Message.parseFrom(buffer);

		} catch (Exception e) {
			Logger.e(TAG,"parsePacket failed！！ "+e.getLocalizedMessage());
			// Logger.logEx(TAG, e);
			// 如果触发这个问题， 可以尝试使用DK解密， 可能是因为密码不正确

			// 还原数据
			if ( rc4 != null ) {
				rc4.encry_RC4_byte(buffer);
			}
			RC4 dk_rc4 = new RC4(_clientConfig.getDefaultKey());
			dk_rc4.decry_RC4(buffer);

			try {
				msg = Message.parseFrom(buffer);
			} catch (Exception e1) {
				// Logger.logEx(TAG, e);
				Logger.e(TAG,"parsePacket parseFrom failed！！ "+e1.getLocalizedMessage());
			}
		}

		return msg;
	}

	/**
	 * Make magic code flag = "qh" magic = flag(2bytes) +
	 * protocol_version(4bits) + client_version(12bits) + appid(2bytes) +
	 * reserved(6bytes)
	 * */
	protected static byte[] makeMagicCode(int protocolVersion, int clientVersion, int appId) {

		byte[] magic = new byte[12];
		magic[0] = 0x71; // 'q'
		magic[1] = 0x68; // 'h'
		magic[2] = (byte) (((protocolVersion & 0xF) << 4) | ((clientVersion & 0xF00) >> 8));
		magic[3] = (byte) (clientVersion & 0xFF);

		magic[4] = (byte) ((appId & 0xFF00) >> 8);
		magic[5] = (byte) (appId & 0xFF);

		return magic;
	}

	/**
	 * 获取服务器当前时间
	 * */
	private long getCurrentServerTime() {

		/**
		 * 没有登录成功， 重来没有获取到服务器时间
		 * */
		if (_time_base == 0) {
			return System.currentTimeMillis();
		}

		long server_sent_time = _server_time + SystemClock.elapsedRealtime() - _time_base;

		return server_sent_time;
	}

	/**
	 * 目前是否午夜时间
	 * */
	private boolean is_in_midnight() {
		if (getCurrentServerTime() != -1) {
			int hours = TimeUtil.getHours(getCurrentServerTime());
			boolean t = (6 >= hours && hours >= 0);
			// Log.d(TAG, "is_in_midnight: " + t);
			return t;
		}
		return false;
	}

	private boolean isDeviceAwake() {
		if (_pm == null) {
			return false;
		}

		return _pm.isScreenOn();
	}

	/**
	 * 底层处理收到的Message包
	 *
	 * @param packet
	 *            需要处理的包
	 * @return<br> 0: 处理成功<br>
	 *             1: 处理失败， 需要跳转到Disconnected状态 <br>
	 *             2: 被T， 需要跳转到LoggedInElsewhere状态<br>
	 * */
	HandlePacketResult handlePacket(Message packet) {

		HandlePacketResult result = HandlePacketResult.Succeeded;

		long sn = packet.getSn();
		int msg_proto_id = packet.getMsgid();

		// 去除消息队列等待的消息， 因为收到服务器的响应了
		Long lsn = Long.valueOf(sn);
		if (_pendingMessages.containsKey(lsn)) {
			_pendingMessages.remove(lsn);
			if (msg_proto_id != MessageId.Service_Resp && msg_proto_id != MessageId.ChatResp && msg_proto_id != MessageId.GetInfoResp) {
				Logger.i(TAG, String.format(Locale.getDefault(), "ack'd: %d", lsn));
			}
		}

		// 通知比较特殊， 不走Response
		if (msg_proto_id == MessageId.NewMessageNotify && packet.hasNotify() && packet.getNotify().hasNewinfoNtf()) {

			NewMessageNotify notify = packet.getNotify().getNewinfoNtf();
			String info_type = notify.getInfoType();
			if (info_type == null) {
				return result;
			}

			long cur = SystemClock.elapsedRealtime();

			// Logger.i(TAG, String.format("noti: %s", info_type));

			MessageFlag msgFlag = null;

			if (_msg_flags.containsKey(info_type)) {
				msgFlag = _msg_flags.get(info_type);
			}

			if(JhFlag.enableDebug()) {
				Logger.i(TAG, "recv " + info_type + " notify");
			}

			if (null == msgFlag) { // 不认识的info_type 作成通知
				byte[] content = notify.hasInfoContent() && notify.getInfoContent() != null ? notify.getInfoContent().toByteArray() : null;
				long info_id = notify.hasInfoId() ? notify.getInfoId() : -1;
				_inotify.onNotification(new NotificationPacket(info_type, content, info_id));
			} else {
				// 有新消息的时候请不让CPU休眠
				acquireWakeLock(_get_msg_WL, TimeConst.WL_TIME_OUT);
				msgFlag.set_last_notify_time(cur);

				if (!msgFlag._getting_msg) { // 如果正在取消息, 等消息返回后再取, 防止取重复的消息
					if (!getMessage(msgFlag)) {
						return HandlePacketResult.Failed;
					}
				}
			}

			return result;

		} else if (msg_proto_id == MessageId.ReLoginNotify) {
			Logger.i(TAG, "R-L-N");

			if (BuildFlag.DEBUG) {
				Logger.e(TAG, "!!! Relogged in !!!");
			}

			// 服务器请求客户端重新登录
			// 目前就让他们互T
			return HandlePacketResult.ReloggedIn;
		} else if (msg_proto_id == MessageId.ReConnectNotify && packet.hasNotify() && packet.getNotify().hasReconnectNtf() ) {
			Logger.i(TAG, "R-C-N");

			if (BuildFlag.DEBUG) {
				Logger.e(TAG, "!!! ReConnect !!!");
			}

			CommunicationData.ReConnectNotify notify = packet.getNotify().getReconnectNtf();
			_reconnect_hosts = new ArrayList<IPAddress>();

			if ( notify.hasPort() ) {
				if (notify.hasIp()) {
					_reconnect_hosts.add(new IPAddress(notify.getIp(), notify.getPort()));
				}

				for (String ip : notify.getMoreIpsList()) {
					_reconnect_hosts.add(new IPAddress(ip, notify.getPort()));
				}
			}

			return HandlePacketResult.ReConnect;
		}

		if (!packet.hasResp()) {
			// 忽略， 但是不重连
			// Logger.e(TAG, String.format("response is null, msgid = %d",
			// msg_proto_id));
			return HandlePacketResult.Succeeded;
		}

		if (packet.hasResp() && packet.getResp().hasError() && packet.getResp().getError() != null) {
			com.huajiao.comm.protobuf.messages.CommunicationData.Error err = packet.getResp().getError();
			int err_code = err.getId();
			// 出错的情况下， 重设标志的状态

			for (MessageFlag flag : _msg_flags.values()) {
				if (flag._get_msg_sn == sn) {
					flag._getting_msg = false;
				}
			}

			// 严重错误， 需要重新建立链接
			if (Error.SEVER_ERROR_START <= err_code && err_code <= Error.SEVER_ERROR_END) {
				return HandlePacketResult.Failed;
			}

			return HandlePacketResult.Succeeded;
		}

		switch (msg_proto_id) {

		case MessageId.BatchQueryPresenceResp:

			Logger.d(TAG, "p-r, sn = " + packet.getSn());

			Ex1QueryUserStatusResp e1_query_user = packet.getResp().getE1QueryUser();
			Object presences[] = null;

			if (e1_query_user.getUserListCount() > 0) {
				presences = new Object[e1_query_user.getUserListCount() * 6];
				int p_index = 0;
				for (RespEQ1User eq_user : e1_query_user.getUserListList()) {
					if (!eq_user.hasUserid() || !eq_user.hasUserType() || !eq_user.hasStatus()) {
						continue;
					}
					presences[p_index++] = (Object) eq_user.getUserid();
					presences[p_index++] = (Object) eq_user.getUserType();
					presences[p_index++] = (Object) eq_user.getStatus();
					presences[p_index++] = (Object) eq_user.getMobileType();
					presences[p_index++] = (Object) eq_user.getAppId();
					presences[p_index++] = (Object) eq_user.getClientVer();
				}
			}

			_inotify.onPresenceUpdated(new PresencePacket(sn, 0, presences));

			break;

		case MessageId.LogoutResp:

			// LogoutResp logoutres = packet.getResp().getLogout();
			// int logoutResult = logoutres.getResult();
			// Logger.i(TAG,
			// String.format("Logged out successfully!, result = %d",
			// logoutResult));

			break;

		case MessageId.ChatResp:

			ChatResp chat_res = packet.getResp().getChat();
			int chat_result = chat_res.getResult();
			Logger.i(TAG, String.format(Locale.CHINA, "m a-ed %d", packet.getSn()));
			_inotify.onMessageResult(new MsgResultPacket(packet.getSn(), chat_result));

			break;

		case MessageId.Service_Resp:

			Logger.i(TAG, String.format(Locale.CHINA, "srv m-ed handleResponse sn:%d", packet.getSn()));
			Service_Resp service_resp = packet.getResp().getServiceResp();
			if (this._inotify != null) {
				// // long _sn, int _service_id, int _result, byte[] _data
				_inotify.onServiceMessageResult(new SrvMsgPacket(sn, service_resp.getServiceId(), 0, service_resp.getResponse().toByteArray()));
			}

			break;

		case MessageId.GetMultipleInfoResp:
		case MessageId.GetInfoResp: {

			long latest_msg_id = 0;
			MessageFlag msgFlag = null;
			boolean valid = true;
			String msg_info_type = null;
			List<Info> messages = null;

			if (msg_proto_id == MessageId.GetInfoResp) {

				if (!packet.getResp().hasGetInfo()) {
					break;
				}

				GetInfoResp get_info_res = packet.getResp().getGetInfo();
				latest_msg_id = get_info_res.getLastInfoId();
				msg_info_type = get_info_res.getInfoType();
				messages = get_info_res.getInfosList();
				if(JhFlag.enableDebug()) {
					Log.d(TAG, "handlePacket GetInfoResp last_msg_id:"+latest_msg_id+",msg_info_type:"+msg_info_type+",messages.size:"+messages.size());
				}
			} else if (msg_proto_id == MessageId.GetMultipleInfoResp) {
				if (!packet.getResp().hasGetMultiInfos()) {
					break;
				}

				GetMultiInfosResp get_info_res = packet.getResp().getGetMultiInfos();
				latest_msg_id = get_info_res.getLastInfoId();
				msg_info_type = get_info_res.getInfoType();
				messages = get_info_res.getInfosList();
			}

			msgFlag = _msg_flags.get(msg_info_type);

			// chatroom 特殊，找falg是找不到的
			if(msgFlag == null && !LLConstant.INFO_TYPE_CHATROOM.equals(msg_info_type)) break;

			if (messages != null) {

				int index = 0;
				for (Info info : messages) {
					List<Pair> pairs = info.getPropertyPairsList();

					boolean is_dup_msg = false;
					long msg_id = -1;
					long sent_time = 0;
					long msg_sn = 0;
					byte[] body = null;
					StringBuffer infosb = new StringBuffer();

					int msg_type = 0;
					String sender = "";
					String receiver = "";

					for (Pair p : pairs) {

						String key = "";

						try {
							key = p.getKey().toString("utf-8");
						} catch (UnsupportedEncodingException e) {
							// Logger.logEx(TAG, e);
						}

						if (key == null || key.length() == 0) {
							continue;
						}

						byte[] bytes = p.hasValue() && p.getValue() == null ? null : p.getValue().toByteArray();

						if (bytes == null) {
							continue;
						}

						if (key.compareTo("info_id") == 0) {
							msg_id = Utils.bytes_to_long(bytes);
							append(infosb, "\"info_id\":"+msg_id);
							if (msgFlag != null) {
								if (msg_id <= msgFlag.get_last_msg_id()) {
									is_dup_msg = true;
									Logger.i(TAG, String.format(Locale.getDefault(), "G dup %s m, i = %d", msg_info_type, msg_id));
								} else {
									Logger.i(TAG, String.format(Locale.getDefault(), "G %s m, i = %d", msg_info_type, msg_id));
								}
							} else {
								Logger.i(TAG, String.format(Locale.getDefault(), "G %s m, i = %d", msg_info_type, msg_id));
							}
						} else if (key.compareTo("chat_body") == 0) {
							body = bytes;
							append(infosb, "\"chat_body\":"+new String(body));
						} else if (key.compareTo("time_sent") == 0) {
							sent_time = Utils.bytes_to_long(bytes);
							append(infosb, "\"time_sent\":"+sent_time);
						} else if (key.compareTo("msg_type") == 0) {
							msg_type = Utils.bytes_to_int(bytes, 0);
							append(infosb, "\"msg_type\":"+msg_type);
						} else if (key.compareTo("msg_sn") == 0) {
							msg_sn = Utils.bytes_to_long(bytes);
							append(infosb, "\"msg_sn\":"+msg_sn);
						} else if (key.compareTo("msg_valid") == 0) {
							int int_valid = Utils.bytes_to_int(bytes, 0);
							valid = (int_valid == 1);
							append(infosb, "\"msg_valid\":"+valid);
						} else if(key.compareTo("expire_time") == 0) {
							long expire_time = Utils.bytes_to_long(bytes);
							append(infosb, "\"expire_time\":"+expire_time);
						} else if(key.compareTo("msg_box") == 0) {
							int msg_box = Utils.bytes_to_int(bytes, 0);
							append(infosb, "\"msg_box\":"+msg_box);
						}
					}

					if(JhFlag.enableDebug()) {
						// Log.i(TAG, "msgs["+(index++)+"]:{"+infosb.toString()+"}");
						Logger.i(TAG, infosb.toString());
					}

					// 解body
					if (LLConstant.INFO_TYPE_PEER.equals(msg_info_type) && body != null) {
						// 100 是特殊消息
						if (msg_type != MessageType.PUBLIC_MESSAGE) {
							Message packed_msg;
							try {
								packed_msg = Message.parseFrom(body);
								msg_sn = packed_msg.getSn();

								String sender_long_str = packed_msg.getSenderJid();

								int i = sender_long_str.indexOf("#");
								if (i != -1) {
									sender = sender_long_str.substring(0, i);
								} else {
									sender = sender_long_str;
								}

								if (packed_msg.hasReq() && packed_msg.getReq().hasChat()) {
									body = packed_msg.getReq().getChat().getBody().toByteArray();
									msg_type = packed_msg.getReq().getChat().getBodyType();
								}
							} catch (Exception e) {
								e.printStackTrace();
							}
						}
					} else if (LLConstant.INFO_TYPE_CHATROOM.equals(msg_info_type)) {
						byte[] tmp = Utils.ungzip(body);
						if (tmp != null) {
							body = tmp;
						}
					}

					if(JhFlag.enableDebug()) {
						Log.d(TAG, "is_dup_msg:"+is_dup_msg);
					}

					// Skip duplicated message
					if (!is_dup_msg) {
						if (delivery_message(sender, receiver, msg_info_type, msg_type, msg_id, msg_sn, sent_time, body, latest_msg_id, valid)) {
							// 更新消息id
							if (msgFlag != null) {
								if (msg_id > msgFlag.get_last_msg_id()) {
									msgFlag.set_last_msg_id(msg_id);
								}
							}

							if (msg_type == MessageType.UPLOAD_LOG_REQ && msg_info_type.equals(LLConstant.INFO_TYPE_PEER)) {
								LoggerBase.upload(this, _account_info.get_account(), sender, msg_sn);
							}

						} else {
							// 上层处理消息失败， 断开连接重试
							if (!_quit) {
								result = HandlePacketResult.Failed; // 需要重连
							}
						}
					}
				} // end of for loop
			}

			if (messages.size() == 0) {
				Logger.d(TAG, String.format("No %s m.", msg_info_type));
			}

			// 如果最新消息ID小于本地的用服务器的（容错）
			if (msgFlag != null) {
				if (latest_msg_id < msgFlag.get_last_msg_id()) {
					msgFlag.set_last_msg_id(latest_msg_id);
				}else if( _drop_peer_msg && ( latest_msg_id - msgFlag.get_last_msg_id() > 1000 ) && LLConstant.INFO_TYPE_PEER.equals(msg_info_type) ){
					//peer atmost get 1000
					msgFlag.set_last_msg_id(latest_msg_id-1000);
					_drop_peer_msg = false;
				}
			}

			// 如果在获取消息的期间收到新的通知, 那么需要再取一次| 如果这一批消息是5条， 那么意味者服务可能还有新消息， 也需要再取一次
			// 如果取到了消息， 那么要接着取直到没有消息
			// 这里不需要考虑使用锁 （正在判断条件时， 收到了通知的情况）， 因为收通知， 和这里的处理是同一个线程
			if (msgFlag != null) {
				if (msgFlag._get_msg_time < msgFlag.get_last_notify_time() || messages.size() >= TimeConst.MAX_MSG_COUNT_PER_QUERY) {
					result = getMessage(msgFlag) ? HandlePacketResult.Succeeded : HandlePacketResult.Failed; // 防止发送失败
				} else {
					msgFlag._getting_msg = false;

					boolean all_done = true;
					for (MessageFlag flag : _msg_flags.values()) {
						if (flag._getting_msg) {
							all_done = false;
							break;
						}
					}

					// 当单聊消息都返回后再释放取消息锁
					if (all_done) {
						releaseWakeLock(_get_msg_WL);
						releaseWakeLock(_ping_WL);
					}
				}
			}
		}
			break;

		default:
			if (BuildFlag.DEBUG) {
				Logger.w(TAG, String.format(Locale.getDefault(), "%s unknown message id %d", this.get_state().toString(), packet.getMsgid()));
			}
			break;
		}

		return result;
	}

	private void append(StringBuffer sb, String text) {
		if(sb != null) {
			if(sb.length() == 0) {
				sb.append(text);
			} else {
				sb.append(","+text);
			}
		}
	}

	/**
	 * 发送心跳包， 4个byte都是0
	 * */
	private boolean send_heartbeat_packet() {

		boolean result = false;

		try {

			_heartbeat_event.set_is_heartbeat(true);
			_heartbeat_event.set_sent_time();
			_heartbeat_event.set_has_been_sent(true);
			_pendingMessages.put(Long.valueOf(HEARTBEAT_SN), _heartbeat_event);

			_socket.getOutputStream().write(HeartbeatContent);
			_socket.getOutputStream().flush();

			Logger.i(TAG, "p->");

			result = true;

		} catch (IOException e) { // socket is probably closed
			Logger.i(TAG, "p-> failed");
		}

		return result;
	}

	/**
	 * 发送登录请求, 默认情况下（自动登录）不需要调用该方法
	 * */
	protected void login() {
		if (_inetAvailable) {
			ClientConnection.this.pushEvent(new Event(LLConstant.EVENT_CONNECT));
		}
	}

	private boolean getAllMessage() {
		for (MessageFlag flag : _msg_flags.values()) {
			if (!getMessage(flag)) {
				return false;
			}
		}
		return true;
	}

	private boolean getMessageInner(String info_type, int[] ids, byte[] parameters) {

		if (info_type == null || info_type.length() == 0 || ids == null || ids.length == 0) {
			return false;
		}

		long get_msg_sn = get_sn();

		Logger.i(TAG, String.format("G ids %s ", info_type));

		Message msg = new Message().setMsgid(MessageId.GetMultipleInfoReq).setSn(get_msg_sn);

		GetMultiInfosReq get_req = new GetMultiInfosReq().setInfoType(info_type);
		if (parameters != null && parameters.length > 0) {
			get_req.setSParameter(ByteStringMicro.copyFrom(parameters));
		}

		for (int id : ids) {
			get_req.addGetInfoIds(id);
		}

		Request req = new Request();
		req.setGetMultiInfos(get_req);
		msg.setReq(req);

		return sendPacket(msg, true);
	}

	/**
	 * 从服务器取消息
	 *
	 * @param messageFlag
	 *            对应的消息标志, 包含info_type类型， 起始消息id
	 * */
	private boolean getMessage(MessageFlag messageFlag) {

		Message msg = null;

		long get_msg_sn = get_sn();
		long start = messageFlag.get_last_msg_id() <= 0 ? 0 : messageFlag.get_last_msg_id() + 1;

		Logger.i(TAG, String.format(Locale.getDefault(), "G %s m s-i = %d", messageFlag.get_info_type(), start));

		msg = new Message().setMsgid(MessageId.GetInfoReq).setSn(get_msg_sn);

		GetInfoReq get_req = new GetInfoReq().setGetInfoId(start)
				.setGetInfoOffset(TimeConst.MAX_MSG_COUNT_PER_QUERY).setInfoType(messageFlag.get_info_type());

		Request req = new Request();
		req.setGetInfo(get_req);
		msg.setReq(req);

		long cur = SystemClock.elapsedRealtime();

		messageFlag._getting_msg = true;
		messageFlag._get_msg_time = cur;
		messageFlag._get_msg_sn = get_msg_sn;
		messageFlag._getting_account = _account_info.get_account();

		return sendPacket(msg, true);
	}

	/**
	 * 移除非用户消息, 已经发送的消息标注的失败
	 * */
	private void removeNonUserMessages() {

		List<Long> keys2Remove = new ArrayList<Long>();

		for (Long k : _pendingMessages.keySet()) {
			MessageEvent me = _pendingMessages.get(k);
			// 所有非用户消息被移除
			if (!me.is_user_message()) {
				keys2Remove.add(k);
			} else if (me.get_send_count() >= LLConstant.MAX_SEND_COUNT) {
				keys2Remove.add(k);
				notifyFailedMessage(k.longValue(), Constant.RESULT_EXCEEDS_RESEND_LIMIT, me.get_message());
			} else if (me.has_been_sent()) {
				// 已经发送但是还没有收到ACK的reset状态
				// keys2Remove.add(k);
				Logger.i(TAG, String.format(Locale.getDefault(), "reset m, sn = %d", me.get_message().getSn()));
				me.set_has_been_sent(false);
			}
		}

		for (Long k : keys2Remove) {
			_pendingMessages.remove(k);
		}
	}

	private int getPersistOfflineTime() {
		if (_interval_index > 1) {
			return (int) (System.currentTimeMillis() - _last_disconnect_time);
		}
		return 0;
	}

	/**
	 * 跟新已经发送的消息或心跳的状态， 对于超过超时设置没有得到ACK的消息标识为失败， 并通知应用层
	 *
	 * @return: 枚举值: PendingMessageStatus
	 * */
	private PendingMessageStatus updatePendingMessageStatus() {

		// 把所有没有得到服务器回执的消息标注为失败
		List<Long> keys2Remove = new ArrayList<Long>();
		long cur = SystemClock.elapsedRealtime();
		boolean timeout_occurred = false;

		for (Long k : _pendingMessages.keySet()) {

			MessageEvent me = _pendingMessages.get(k);
			long diff = cur - me.get_sent_time();
			long timeout = me.get_timeout();
			long business_timeout = cur - me.get_construct_time();

			if (me.is_heartbeat()) {
				if (diff > timeout) {
					keys2Remove.add(k);
					Logger.i(TAG, String.format(Locale.getDefault(), "p t-out: %d", diff));
				}
			} else if (me.is_user_message()) {

				// 检查业务是否已经超时
				if (business_timeout > me.get_timeout()) {
					keys2Remove.add(k);
					Logger.i(TAG, String.format(Locale.getDefault(), "m t-out: %d, s: %d, id %d", diff, me.get_message().getSn(), me.get_message().getMsgid()));
				}

				// 检查连接层是否已经超时
				if (me.has_been_sent()) {
					if (diff > TimeConst.PACKET_RESP_TIMEOUT) {
						timeout_occurred = true;
						if (!keys2Remove.contains(k)) {
							Logger.d(TAG,
									String.format(Locale.getDefault(), "m t-out: %d, s: %d, id %d, retry in future", diff, me.get_message().getSn(), me.get_message().getMsgid()));
						}
					}
				}
			} else {
				if (diff > timeout) {
					keys2Remove.add(k);
					Logger.i(TAG, String.format(Locale.getDefault(), "p t-out: %d, s: %d, mi %d", diff, me.get_message().getSn(), me.get_message().getMsgid()));
				}
			}
		}

		// 删除超时的消息
		for (Long k : keys2Remove) {

			MessageEvent me = _pendingMessages.get(k);
			if (me.is_user_message()) {
				if (_inotify != null) {
					if (me.get_message().getMsgid() == MessageId.ChatReq) {
						_inotify.onMessageResult(new MsgResultPacket(me.get_message().getSn(), Constant.RESULT_TIMEOUT));
					} else if (me.get_message().getMsgid() == MessageId.Service_Req) {
						int service_id = 0;
						try {
							Message msg = me.get_message();
							service_id = msg.getReq().getServiceReq().getServiceId();
						} catch (Exception e) {

						}
						_inotify.onServiceMessageResult(new SrvMsgPacket(me.get_message().getSn(), service_id, Constant.RESULT_TIMEOUT, null));
					} else if (me.get_message().getMsgid() == MessageId.BatchQueryPresenceReq) {
						_inotify.onPresenceUpdated(new PresencePacket(me.get_message().getSn(), Constant.RESULT_TIMEOUT, null));
					}
				}
			}

			// 没有发出去的包超时， 不应该影响当前的连接
			if (me.has_been_sent()) {
				timeout_occurred = true;
			}

			_pendingMessages.remove(k);
		}

		if (timeout_occurred) {
			return PendingMessageStatus.TimeoutOccurred;
		}

		return _pendingMessages.size() > 0 ? PendingMessageStatus.Continue : PendingMessageStatus.QueueIsEmpty;
	}

	/**
	 * 获取接下来最有可能马上超时的消息的剩余超时时间
	 *
	 * @return 需要等待的时间
	 * */
	private long getLeastTimeout() {

		long cur = SystemClock.elapsedRealtime();
		long min_diff = TimeConst.MSG_SEND_TIMEOUT;

		for (Long k : _pendingMessages.keySet()) {

			MessageEvent me = _pendingMessages.get(k);
			long timeout = me.get_timeout();
			long diff = (timeout + me.get_sent_time()) - cur;

			if (me.is_user_message()) {
				diff = cur - me.get_construct_time();
				if (me.has_been_sent()) {
					long conn_diff = cur - me.get_sent_time();
					if (conn_diff < diff) {
						diff = conn_diff;
					}
				}
			}

			if (diff < min_diff) {
				min_diff = diff;
			}
		}

		return min_diff <= 0 ? 1 : (min_diff + 500);
	}

	/**
	 * 关闭连接
	 * */
	protected void close() {

		schedule_next_ping();
		removeNonUserMessages();

		if (_socket != null && _socket.isConnected()) {
			_receiver.sendCmd(Receiver.CMD_CLOSE_SOCKET);
			try {
				_socket.close();
			} catch (Exception e) {

			} finally {

			}
		}

		for (MessageFlag flag : _msg_flags.values()) {
			flag.reset();
		}

		_magic_received = _connected = _init_logged_in = _logged_in = _init_packtet_sent = false;
	}

	/**
	 * send user message
	 * */
	protected boolean send_user_messages() {

		for (Long k : _pendingMessages.keySet()) {
			if (!_pendingMessages.get(k).has_been_sent() && _pendingMessages.get(k).is_user_message()) {
				if (!sendPacket(_pendingMessages.get(k).get_message(), true)) {
					return false;
				} else {
					Logger.i(TAG, String.format("m %s sent", _pendingMessages.get(k).get_message().getSn()));
				}
			}
		}

		return true;
	}

	/**
	 * 异步发送消息， 如果返回成功消息结果将由INotify接口告知
	 *
	 * @param receiver
	 *            接收者的账号
	 * @param msg_type
	 *            消息类型
	 * @param sn
	 *            消息sn
	 * @param body
	 *            内容体
	 * **/
	@Override
	public boolean send_message(String receiver, int account_type, int msg_type, long sn, byte[] body, int timeoutMs, int expirationSec) {

		if (null == body || receiver == null || receiver.length() == 0) {
			Logger.e(TAG, "s-m: invalid arguments!!!");
			return false;
		}

		ByteStringMicro bstr = ByteStringMicro.copyFrom(body);
		ChatReq chat = new ChatReq().setBody(bstr).setBodyType(msg_type);

		Request req = new Request();
		req.setChat(chat);

		String receiver_type = AccountInfo.get_account_type_string(account_type);
		if (receiver_type == null) {
			Logger.e(TAG, "account type is not supported");
			return false;
		}

		Message msg = new Message().setReceiver(receiver).setReceiverType(receiver_type).setMsgid(MessageId.ChatReq).setSn(sn);
		msg.setReq(req);

		MessageEvent me = new MessageEvent(msg, timeoutMs, true);
		me.set_sent_time();

		// 通知发送消息
		_eventQueue.offer(me);

		// 用户要发消息， 我们马上登录吧
		if (_current_state.get_state().equals(ConnectionState.Disconnected)) {
			synchronized (_connectLock) {
				_connectLock.notifyAll();
			}
		}

		return true;
	}

	@Override
	public boolean send_service_message(int serviceId, long sn, byte[] body) {
		Logger.i(TAG, "CC send_service_message sn:" + sn);
		if (body == null || body.length == 0) {
			Logger.e(TAG, "s_s_m: invalid arguments!!!");
			return false;
		}

		ByteStringMicro bstr = ByteStringMicro.copyFrom(body);

		Service_Req sreq = new Service_Req().setServiceId(serviceId).setRequest(bstr);
		Request req = new Request().setServiceReq(sreq);
		Message msg = new Message().setReceiverType("null").setMsgid(MessageId.Service_Req).setSn(sn).setReq(req);

		MessageEvent me = new MessageEvent(msg, TimeConst.SRV_PACKET_RESP_TIMEOUT, true);
		me.set_sent_time();

		// 通知发送消息
		_eventQueue.offer(me);

		// 用户要发消息， 我们马上登录吧
		if (_current_state.get_state().equals(ConnectionState.Disconnected)) {
			Logger.i(TAG, "CC send_service_message sn: state == Disconnected" + sn);
			synchronized (_connectLock) {
				_connectLock.notifyAll();
			}
		}

		return true;
	}

	/**
	 * 读取底层包，并解析
	 *
	 * @return 正常包， 如果是null表示连接已经中断， 或出现其他无法解决的问题
	 * */
	protected Message readPacket() {

		Message msg = null;

		try {

			while (true) {

				if (!_connected) {
					return msg;
				}

				int totalLen = 0;
				int remainingDataLen = 0;
				byte pbBuffer[] = null;
				int len_of_len = 4;

				if (!_magic_received) {
					len_of_len += 2;
				}

				byte lengthBytes[] = new byte[len_of_len];
				int index = 0, bytesRead = 0;

				while (index < lengthBytes.length && ((bytesRead = _socket_in.read_raw(lengthBytes, index, lengthBytes.length - index)) > 0)) {
					index += bytesRead;
				}

				if (bytesRead < lengthBytes.length) {
					if (bytesRead > 0) {
						Logger.w(TAG, "r insufficient data.");
					} else {
						Logger.w(TAG, "s has been closed.");
					}
					break;
				}

				totalLen = Utils.bytes_to_int(lengthBytes, lengthBytes.length - 4);

				// 服务器发送的心跳包4个字节都是0, 连接后才会发送心跳
				if (_logged_in && totalLen == 0) {
					Logger.i(TAG, "p<-");
					_eventQueue.offer(_got_heartbeat_ack_event);
					continue;
				}

				remainingDataLen = totalLen - 4;
				if (!_magic_received) { // 校验magic, 解决网络被劫持的问题

					if (lengthBytes[0] != 0x71 || lengthBytes[1] != 0x68) {
						Logger.w(TAG, String.format(Locale.US, "hijacked %s", Utils.toHexString(lengthBytes)));
						break;
					}

					remainingDataLen -= 2; // 去除magic_code的长度
					_magic_received = true;
				}

				if (totalLen <= 4 || totalLen > 512000 || remainingDataLen <= 0 || remainingDataLen > 512000) {
					Logger.w(TAG, String.format(Locale.getDefault(), "L is abnormal: %d", totalLen));
					break;
				}

				pbBuffer = new byte[remainingDataLen];
				int dataRead = _socket_in.read_raw_data(pbBuffer);
				if (-1 == dataRead) {
					break;
				}

				msg = parsePacket(pbBuffer);
				// Logger.i(TAG, "msg<-");

				break;
			}
		} catch (SocketException ex) {
			// Logger.w(TAG, "SocketException: " + ex.toString());
			if(JhFlag.enableDebug()) {
				Logger.w(TAG, "read packet throw " + ex.toString());
			} else {
				Logger.w(TAG, ex.getMessage());
			}
		} catch (Exception e) {
			Logger.w(TAG, String.format("r-P threw %s", Log.getStackTraceString(e)));
		}

		return msg;
	}

	/**
	 * 永久关闭连接, 回收资源
	 * */
	@Override
	public synchronized void shutdown() {

		if (_quit) {
			return;
		}

		Logger.i(TAG, "s-do...");

		unregisterAlarmBroadcast();

		unregisterConnectivityReceiver();

		_quit = true;

		if (_receiver != null) {
			_receiver.sendCmd(Receiver.CMD_STOP);
		}

		close();

		if (this._sender != null) {
			this._sender.interrupt();
			try {
				// 最终退出也没关系
				this._sender.join(1000);
			} catch (InterruptedException e) {
				// Logger.logEx(TAG, e);
			}

			_sender = null;
		}

		if (this._receiver != null) {
			this._receiver.interrupt();
			try {
				// 最终退出也没关系
				this._receiver.join(1000);
			} catch (InterruptedException e) {
				// Logger.logEx(TAG, e);
			}

			_receiver = null;
		}

		releaseAllWakeLock();

		if (BuildFlag.DEBUG) {
			Logger.i(TAG, "connection has shut down.");
		}
	}

	void releaseAllWakeLock() {
		releaseWakeLock(_ping_WL);
		releaseWakeLock(_get_msg_WL);
		releaseWakeLock(_business_WL);
	}

	boolean acquireWakeLock(WakeLock wl, long timeout) {

		if (_pm == null || wl == null) {
			Logger.e(TAG, "_args is null!!!");
			return false;
		}

		if (_pm.isScreenOn()) {
			return true;
		}

		try {
			synchronized (wl) {
				if (!wl.isHeld()) {
					String wlName = getWlName(wl);
					wl.acquire(timeout);
					Logger.i(TAG, wlName + " acq'd.");
				}
			}
			return true;
		} catch (Exception e) {
			Logger.e(TAG, "acqWL  exception" + e.getMessage());
		}

		return false;
	}

	String getWlName(WakeLock wl) {

		if (wl == null) {
			return "wl_null";
		}

		if (wl.equals(_business_WL)) {
			return "wl_b";
		}

		if (wl.equals(_get_msg_WL)) {
			return "wl_g";
		}

		if (wl.equals(_ping_WL)) {
			return "wl_p";
		}

		return "wl_u";
	}

	void releaseWakeLock(WakeLock wl) {

		if (wl == null) {
			Logger.e(TAG, "WL is null!!!");
			return;
		}

		try {
			synchronized (wl) {
				if (wl.isHeld()) {
					wl.release();
					String wlName = getWlName(wl);
					Logger.i(TAG, wlName + " released.");
				}
			}
		} catch (Exception e) {
			Logger.e(TAG, "releaseWL  exception" + e.getMessage());
		}
	}

	/**
	 * 获取当前的状态
	 **/
	public ConnectionState get_state() {
		State s = this._current_state;
		if (s != null) {
			return s.get_state();
		}
		return ConnectionState.Disconnected;
	}

	/**
	 * 如果网络状态发送变化，可以提前主动切换状态， 这样状态更及时
	 **/
	protected void notifyNetworkStateChange(boolean is_inet_available) {

		_interval_index = 0;
		_dispatchClient.reset();

		_inetAvailable = is_inet_available;
		byte event_id = is_inet_available ? LLConstant.EVENT_INET_AVAILABLE : LLConstant.EVENT_INET_UNAVAILABLE;
		pushEvent(new Event(event_id, SystemClock.elapsedRealtime()));

		if (is_inet_available) {
			try {
				synchronized (_connectLock) {
					_connectLock.notifyAll();
				}
			} catch (Exception e) {
				return;
			}
		}
	}

	/**
	 * 程序前后台切换, 会影响心跳和重连频率<br>
	 * 注意：目前不支持
	 *
	 * @param is_on_foreground
	 *            程序是否在前台
	 * */
	public void notify_app_state_change(boolean is_on_foreground) {
		// 如果切换到前台， 可以立即登录
		if (is_on_foreground) {
			synchronized (_connectLock) {
				_connectLock.notify();
			}
		}
	}

	private int get_heartbeat_time(String uid, String uiver) {
		if ( uid.equals("20000116") ){
			return _min_heart;
		}

		if ( uiver.toLowerCase().contains("360ui") ){
			_report_heart_time = _max_heart/1000;
			return _max_heart;
		}

		return _min_heart;
	}

	private int get_next_heartbeat_time(int curr) {
		int timeout = (int)(curr * 1.5);

		return timeout<_max_heart?timeout:_max_heart;
	}

	private static String getSystemProperty(String propName){
		String line = "null";
		BufferedReader input = null;
		try
		{
			Process p = Runtime.getRuntime().exec("getprop " + propName);
			input = new BufferedReader(new InputStreamReader(p.getInputStream()), 1024);
			line = input.readLine();
			if ( line==null||line.equals("") ){
				line = "null";
			}
			input.close();
		}
		catch (IOException ex)
		{
			Log.e(TAG, "Unable to read sysprop " + propName, ex);
		}
		finally
		{
			if(input != null)
			{
				try
				{
					input.close();
				}
				catch (IOException e)
				{
					Log.e(TAG, "Exception while closing InputStream", e);
				}
			}
		}

		return line;
	}

	/**
	 * 开始运行
	 * */
	void init(Context context, AccountInfo account_info, ClientConfig clientConfig, IMCallback notify) {

		if (context == null || account_info == null || notify == null || clientConfig == null) {
			throw new IllegalArgumentException();
		}

		_clientConfig = clientConfig;

		Logger.setUid(account_info.get_account());
		String ui = getSystemProperty("ro.build.uiversion");
		_curr_heart = get_heartbeat_time(account_info.get_account(), ui);
		Logger.i(TAG, "conn init for " + ui + "," + Build.BRAND + ", heartbeat time:" + _curr_heart + " ms, report heartbeat time:" + _report_heart_time );

		Logger.i(TAG, String.format("Ver %s, %S", LLConstant.SDK_VER, Utils.getModel()));

		MagicCode = makeMagicCode(LLConstant.PROTOCOL_VERSION, _clientConfig.getClientVersion(), _clientConfig.getAppId());

		_context = context.getApplicationContext();
		_am = (AlarmManager) _context.getSystemService(Context.ALARM_SERVICE);

		_inotify = notify;

		_account_info = account_info;
		MessageFlag peer_msg_flag = new MessageFlag(_context, LLConstant.INFO_TYPE_PEER, _account_info.get_account(), _clientConfig.getDefaultKey());
		_msg_flags.put(LLConstant.INFO_TYPE_PEER, peer_msg_flag);

		MessageFlag public_msg_flag = new MessageFlag(_context, LLConstant.INFO_TYPE_PUBLIC, _account_info.get_account(), _clientConfig.getDefaultKey());
		_msg_flags.put(LLConstant.INFO_TYPE_PUBLIC, public_msg_flag);

		// chatroom标记
//		MessageFlag chatroom_msg_flag = new MessageFlag(_context, LLConstant.INFO_TYPE_CHATROOM, _account_info.get_account(), _clientConfig.getDefaultKey());
//		_msg_flags.put(LLConstant.INFO_TYPE_CHATROOM, chatroom_msg_flag);

		// im标记
		MessageFlag im_msg_flag = new MessageFlag(_context, LLConstant.INFO_TYPE_IM, _account_info.get_account(), _clientConfig.getDefaultKey());
		_msg_flags.put(LLConstant.INFO_TYPE_IM, im_msg_flag);

		_connReceiver = new ConnectivityChangedReceiver(context);
		registerConnectivityReceiver();
		_inetAvailable = isInetAvailabe();

		//dispache client
		_dispatchClient = DispatchClient.getInstance();

		initState();

		_pm = (PowerManager) _context.getSystemService(Context.POWER_SERVICE);

		if (_pm != null) {
			if (!_pm.isScreenOn()) {
				_screen_off_time = SystemClock.elapsedRealtime();
			}

			_ping_WL = _pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "p");
			_ping_WL.setReferenceCounted(false);

			_get_msg_WL = _pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "g");
			_get_msg_WL.setReferenceCounted(false);

			_business_WL = _pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "b");
			_business_WL.setReferenceCounted(false);
		}

		registerAlarmBroadcast();

		_receiver = new Receiver();
		_receiver.setUncaughtExceptionHandler(this);
		_receiver.setDaemon(true);
		_receiver.setName("CN-RECV");

		_receiver.start();

		_sender = new Sender();
		_sender.setUncaughtExceptionHandler(this);
		_sender.setDaemon(true);
		_sender.setName("CN-SEND");

		_sender.start();

	}

	@Override
	public boolean health_check() {

		if (this._quit) {
			return false;
		}

		if (this._receiver == null || !this._receiver.isAlive() || this._sender == null || !this._sender.isAlive()) {
			return false;
		}

		return true;
	}

	@Override
	public boolean has_shutdown() {
		return this._quit;
	}

	@Override
	public String get_account() {
		return _account_info.get_account();
	}

	@Override
	public void send_heartbeat() {
		pushEvent(_heartbeat_event);
	}

	@Override
	public boolean get_current_state() {
		pushEvent(new Event(LLConstant.EVENT_GET_STATE));
		return true;
	}

	@Override
	public boolean query_presence(String[] users, long sn) {
		if (users == null || users.length == 0) {
			Logger.e(TAG, "q_pre: invalid arguments!!!");
			return false;
		}

		Ex1QueryUserStatusReq query_req_builder = new Ex1QueryUserStatusReq();
		for (String user : users) {

			// TODO: do not hard code user type
			ReqEQ1User req1 = new ReqEQ1User().setUserid(user).setUserType("phone").setAppId(_clientConfig.getAppId());
			query_req_builder.addUserList(req1);
		}

		Request req = new Request().setE1QueryUser(query_req_builder);
		Message msg = new Message().setMsgid(MessageId.BatchQueryPresenceReq).setSn(sn).setReq(req);

		MessageEvent me = new MessageEvent(msg, TimeConst.PACKET_RESP_TIMEOUT, true);
		me.set_sent_time();

		_eventQueue.offer(me);

		// 用户要发消息， 我们马上登录吧
		if (_current_state.get_state().equals(ConnectionState.Disconnected)) {
			synchronized (_connectLock) {
				_connectLock.notifyAll();
			}
		}

		return true;
	}

	@Override
	public long get_server_time_diff() {

		if (_current_state.get_state().equals(ConnectionState.Disconnected)) {
			synchronized (_connectLock) {
				_connectLock.notifyAll();
			}
		}

		synchronized (_time_lock) {
			if (_time_base == 0) {
				return -1;
			}
			return _server_time - _time_base;
		}
	}

	@Override
	public String get_jid() {
		return _jid;
	}

	/** 是否是WAP apn需要走dispatch */
	@SuppressLint("DefaultLocale")
	private boolean is_wap_apn() {

		// dispatch is only available for online servers
		if (!_clientConfig.getServer().equals(LLConstant.OFFICIAL_SERVER)) {
			return false;
		}

		final ConnectivityManager connMgr = (ConnectivityManager) _context.getSystemService(Context.CONNECTIVITY_SERVICE);
		if (connMgr == null) {
			return false;
		}

		NetworkInfo networkInfo = connMgr.getActiveNetworkInfo();
		if (networkInfo == null) {
			return false;
		}

		int net_type = networkInfo.getType();
		if (net_type == ConnectivityManager.TYPE_MOBILE) {
			String apn = networkInfo.getExtraInfo();
			if (apn != null && apn.length() > 0) {
				apn = apn.toLowerCase();
				if (apn.indexOf("cmwap") != -1 || apn.indexOf("3gwap") != -1 || apn.indexOf("uniwap") != -1 || apn.indexOf("ctwap") != -1) {
					return true;
				}
			}
		}

		return false;
	}

	@Override
	public void uncaughtException(Thread thread, Throwable ex) {
		String output = "T-C!!!: " + thread.getName() + ":\n";
		Logger.e(TAG, output + Log.getStackTraceString(ex));
	}

	/**
	 * do login frequency check and return minal wait time.
	 * */
	private int getMinimalWaitTime() {

		Calendar calendar = Calendar.getInstance();
		int currentMin = calendar.get(Calendar.MINUTE);
		int currentHour = calendar.get(Calendar.HOUR_OF_DAY);

		if (currentMin == _cur_min && currentHour == _cur_hour) {
			_login_count++;
		} else {
			_login_count = 0;
			_cur_hour = currentHour;
			_cur_min = currentMin;
		}

		if (_login_count >= LLConstant.LOGIN_FREQ_LIMIT) {
			int wait_sec = 60 - calendar.get(Calendar.SECOND);
			return wait_sec * 1000;
		}

		return 0;
	}

	void notifyFailedMessage(long sn, int result, Message message) {

		if (message == null) {
			return;
		}

		try {
			if (message.getMsgid() == MessageId.Service_Req) {
				int srv_id = message.getReq().getServiceReq().getServiceId();
				_inotify.onServiceMessageResult(new SrvMsgPacket(sn, srv_id, result, null));
			} else {
				_inotify.onMessageResult(new MsgResultPacket(sn, result));
			}
		} catch (Exception e) {
			e.printStackTrace();
		}
	}

	// upload log
	@Override
	public boolean send_data(String receiver, byte[] body, long sn) {
		return this.send_message(receiver, AccountInfo.ACCOUNT_TYPE_JID, MessageType.UPLOAD_LOG_RES, sn, body, 0, 0);
	}

	@Override
	public boolean get_message(String info_type, int[] ids, byte[] parameters) {
		GetMsgEvent event = new GetMsgEvent(info_type, ids, parameters);
		pushEvent(event);
		return true;
	}

}
