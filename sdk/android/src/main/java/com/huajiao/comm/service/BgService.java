package com.huajiao.comm.service;

import android.app.Notification;
import android.app.PendingIntent;
import android.app.Service;
import android.content.ComponentName;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;
import android.os.Process;
import android.os.RemoteException;
import android.text.TextUtils;
import android.util.Log;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.BuildFlag;
import com.huajiao.comm.common.ClientConfig;
import com.huajiao.comm.common.TimeCostCalculator;
import com.huajiao.comm.im.ConnectionFactory;
import com.huajiao.comm.im.ConnectionState;
import com.huajiao.comm.im.IConnection;
import com.huajiao.comm.im.IMCallback;
import com.huajiao.comm.im.Logger;
import com.huajiao.comm.im.packet.CurrentStatePacket;
import com.huajiao.comm.im.packet.MsgPacket;
import com.huajiao.comm.im.packet.MsgResultPacket;
import com.huajiao.comm.im.packet.NotificationPacket;
import com.huajiao.comm.im.packet.Packet;
import com.huajiao.comm.im.packet.PresencePacket;
import com.huajiao.comm.im.packet.SrvMsgPacket;
import com.huajiao.comm.im.packet.StateChangedPacket;
import com.huajiao.comm.im.packet.TimePacket;
import com.huajiao.comm.im.rpc.AccountCmd;
import com.huajiao.comm.im.rpc.Cmd;
import com.huajiao.comm.im.rpc.GetMessageCmd;
import com.huajiao.comm.im.rpc.MsgCmd;
import com.huajiao.comm.im.rpc.PresenceCmd;
import com.huajiao.comm.im.rpc.ServiceMsgCmd;
import com.huajiao.imsdk.BuildConfig;

import java.util.Locale;

/**
 * Back ground service
 * */
public class BgService extends Service implements IMCallback {

	private ClientConfig _clientConfig;

	private AccountInfo _accountInfo;

	private static final String TAG = "BGS";

	/** 登录账号 */
	public static final String KEY_ACCOUNT = "key1";

	/** 登录的密码 */
	public static final String KEY_PWD = "key2";

	/** device id */
	public static final String KEY_ID = "key3";

	public static final String KEY_SERVER = "key4";

	public static final String KEY_APPID = "key5";

	public static final String KEY_DEFAULT_KEY = "key6";

	public static final String KEY_CLIENT_CONFIG = "key7";

	/** 数据包 */
	public static final String KEY_PACKET = "key8";

	/*** Jid */
	public static final String KEY_JID = "key9";

	/** new state key */
	public static final String KEY_NEW_STATE = "key10";

	/** new state key */
	public static final String KEY_ACCOUNT_INFO = "key11";

	/** ImBridge Cmd */
	public static final String KEY_CMD = "key12";

	/** 停止service的动作 */
	public static final String ACTION_SHUTDOWN = "";

	public static final String ACTION_CLOUD_MSG = "";

	private static final int NOTIFY_ID_PESERVICE_NEW = 1600501;

	private boolean _started_as_foreground = false;

	private boolean _has_pending_get_time_req = false;

	/** IM connection */
	private volatile IConnection _conn = null;

	private ConnectionState _last_reported_state = ConnectionState.Disconnected;

	public static final String KEY_LOG_DIR = "key_log_dir";
	private boolean isOpenForeground = false;// 8.0以上系统是否使用

	private synchronized void switch_account(AccountInfo account_info, ClientConfig clientConfig) {
		if (account_info == null || clientConfig == null) {
			Logger.e(TAG, "argument invalid!");
			return;
		}
		if (_conn == null || !_conn.health_check()) {
			_conn = ConnectionFactory.getInstance().getConnection(BgService.this, account_info, clientConfig, BgService.this);
		} else {
			_conn.switch_account(account_info, clientConfig);
		}
	}

	private boolean handle_cmd(Cmd cmd) {

		boolean result = false;

		if (_conn == null) {
			if (BuildFlag.DEBUG) {
				Log.e(TAG, "handle_cmd _conn is null!!");
			}
			return result;
		}

		switch (cmd.get_cmd_code()) {

		case Cmd.CMD_GET_SERVER_TIME:
			if (_conn != null) {
				onTimeSyncRequest();
				result = true;
			}
			break;

		case Cmd.CMD_QUERY_PRESENCE:
			PresenceCmd presence_cmd = (PresenceCmd) cmd;
			if (null != presence_cmd) {
				String userlist[] = presence_cmd.get_users().split("_");
				result = _conn.query_presence(userlist, presence_cmd.get_sn());
			}
			break;

		case Cmd.CMD_SEND_MESSAGE:
			MsgCmd msg_cmd = (MsgCmd) cmd;
			if (null != msg_cmd) {
				result = _conn.send_message(msg_cmd.get_receiver(), msg_cmd.get_account_type(), msg_cmd.get_msg_type(), msg_cmd.get_sn(), msg_cmd.get_body(),
						msg_cmd.get_timeout_ms(), msg_cmd.get_expiration_sec());
			}

			break;

		case Cmd.CMD_SEND_SRV_MESSAGE:
			ServiceMsgCmd srv_cmd = (ServiceMsgCmd) cmd;
			if (null != srv_cmd) {
				result = _conn.send_service_message(srv_cmd.get_service_id(), srv_cmd.get_sn(), srv_cmd.get_body());
			}

			break;

		case Cmd.CMD_SWITCH_ACCOUNT:
			AccountCmd acc_cmd = (AccountCmd) cmd;
			if (null != acc_cmd) {
				_conn.switch_account(acc_cmd.get_account_info(), acc_cmd.get_client_config());
				result = true;
			}
			break;

		case Cmd.CMD_GET_MESSAGE:
			GetMessageCmd get_msg_cmd = (GetMessageCmd) cmd;
			if (null != get_msg_cmd) {
				result = _conn.get_message(get_msg_cmd.get_info_type(), get_msg_cmd.get_ids(), get_msg_cmd.get_paramters());
			}
			break;

		case Cmd.CMD_GET_LLC_STATE:
			result = _conn.get_current_state();
			break;

		case Cmd.CMD_SHUTDOWN:

			shutdown_llc();
			result = true;

			break;

		default:

			Log.e(TAG, "UNKNOWN CMD " + cmd.get_cmd_code());

			break;
		}

		return result;
	}

	@Override
	public IBinder onBind(Intent intent) {

		init_connection(intent);

		if (BuildFlag.DEBUG) {
			Logger.d(TAG, "onBind called");
		}


		return _binder;
	}

	private void init_connection(Intent intent) {

		if (intent == null) {
			return;
		}

		if (intent.hasExtra(KEY_ACCOUNT_INFO) && intent.hasExtra(KEY_CLIENT_CONFIG)) {
			AccountInfo account_info = (AccountInfo) intent.getSerializableExtra(KEY_ACCOUNT_INFO);
			ClientConfig client_config = (ClientConfig) intent.getSerializableExtra(KEY_CLIENT_CONFIG);
			String log_dir=intent.getStringExtra(KEY_LOG_DIR);
			if(!TextUtils.isEmpty(log_dir)){
				Logger.setDebugFilePath(log_dir);
			}

			if (account_info != null && client_config != null) {
				_accountInfo = account_info;
				_clientConfig = client_config;

				if(!TextUtils.isEmpty(_accountInfo.get_account())){
					Logger.setUid(_accountInfo.get_account());
				}
				switch_account(_accountInfo, _clientConfig);
			} else {
				Logger.w(TAG, "invalid args 1");
			}
		}
	}

	@Override
	public int onStartCommand(Intent intent, int flags, int startId) {
		super.onStartCommand(intent, flags, startId);

		TimeCostCalculator calc = new TimeCostCalculator();

		if (intent == null) {
			return START_STICKY;
		}

		init_connection(intent);

		if (BuildFlag.DEBUG) {
			Logger.d(TAG, String.format(Locale.US, "onStartCommand flags %d, startId %d", flags, startId));
		}

		if (intent.getAction() != null && intent.getAction().equals(ACTION_SHUTDOWN)) {
			shutdown_llc();
			stopSelf();
			return START_NOT_STICKY;
		}

		if (intent.hasExtra(KEY_CMD)) {
			Cmd cmd = (Cmd) intent.getSerializableExtra(KEY_CMD);
			if (cmd != null) {
				try {
					boolean result = handle_cmd(cmd);
					if (!result) {
						Log.e(TAG, "handle_cmd failed: " + cmd.toString());
					}
				} catch (Exception e) {
					Logger.e(TAG, "handle_cmd ex\n" + Log.getStackTraceString(e));
				}
			}
		}

		Logger.d(TAG, "onStartCommand consumes " + calc.getCost());

		return START_STICKY;
	}

	@SuppressWarnings("deprecation")
	@Override
	public void onCreate() {
		super.onCreate();

		// Logger.d(TAG, "BuildFlag.DEBUG:"+BuildFlag.DEBUG);

		if (BuildFlag.DEBUG) {
			Logger.i(TAG, "OnCreate pid  " + Process.myPid());
		}
		Logger.enableXlog(getApplicationContext(),true);

		if (Build.VERSION.SDK_INT < 18) {
			try {
				Notification notification = new Notification();
				notification.flags |= Notification.FLAG_NO_CLEAR;// 加上FLAG_NO_CLEAR标志,避免用户清除通知栏的时候,把本透明的
																	// notification也清掉
				PendingIntent pendingIntent = PendingIntent.getActivity(this, 0, new Intent(), 0);
//				notification.setLatestEventInfo(this, null, null, pendingIntent);
				startForeground(NOTIFY_ID_PESERVICE_NEW, notification);
				_started_as_foreground = true;

			} catch (Exception ex) {
				if (BuildFlag.DEBUG) {
					ex.printStackTrace();
				}
			}
		}
	}

	private void shutdown_llc() {
		try {
			if (_conn != null) {
				_conn.shutdown();
				_conn = null;
				Logger.d(TAG, "service has been shutdown");
			}
		} catch (Exception e) {
			Logger.e(TAG, "sllc ex\n" + Log.getStackTraceString(e));
		}

	}

	@Override
	public void onDestroy() {
		super.onDestroy();

		TimeCostCalculator calc = new TimeCostCalculator();

		if (BuildFlag.DEBUG) {
			Logger.w(TAG, "onDestroy called, stop running");
		}

		shutdown_llc();

		if (Build.VERSION.SDK_INT < 18 && _started_as_foreground) {
			stopForeground(true);
			_started_as_foreground = false;
		}

		Logger.d(TAG, "onDestroy consumes " + calc.getCost());
	}

	@Override
	public synchronized void onStateChanged(StateChangedPacket packet) {

		if (packet == null) {
			return;
		}

		ConnectionState new_state = packet.get_newState();
		if (new_state == null) {
			return;
		}

		if (!new_state.equals(ConnectionState.Connecting) && !_last_reported_state.equals(new_state)) {
			_last_reported_state = new_state;
			pushPacket(packet, -1);
		}

		if (_has_pending_get_time_req && new_state.equals(ConnectionState.Connected)) {
			onTimeSyncRequest();
		}
	}

	@Override
	public void onMessageResult(MsgResultPacket packet) {
		pushPacket(packet, packet.get_sn());
	}

	@Override
	public void onMessage(MsgPacket packet) {
		pushPacket(packet, packet.get_sn());
	}

	@Override
	public void onNotification(NotificationPacket notificationPacket) {
		pushPacket(notificationPacket, -1);
	}

	@Override
	public void onPresenceUpdated(PresencePacket packet) {
		pushPacket(packet, packet.get_sn());
	}

	@Override
	public void onServiceMessageResult(SrvMsgPacket packet) {
		pushPacket(packet, packet.get_sn());
	}

	@Override
	public void onCurrentState(CurrentStatePacket packet) {
		pushPacket(packet, -1);
	}

	void pushPacket(Packet packet, long sn) {

		try {

			Intent intent = new Intent();
			intent.putExtra(KEY_PACKET, packet);
			intent.setAction(ACTION_CLOUD_MSG);

			if (_clientConfig != null && _clientConfig.get_im_business_service() != null) {
				intent.setComponent(new ComponentName(getApplicationContext(), _clientConfig.get_im_business_service()));
			} else {
				Logger.e(TAG, "null argument4");
			}

			if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && isOpenForeground) {
				// debug版本启动前台服务
				if(BuildConfig.DEBUG) {
					Log.e(TAG, "startForegroundService");
				}
				startForegroundService(intent);
			} else {
				if(BuildConfig.DEBUG) {
					Log.e(TAG, "startService");
				}
				startService(intent);
			}

			if (BuildFlag.DEBUG) {
				// Logger.d(TAG, "packet delivered: " + sn);
			} else {
				// Log.d(TAG, "packet delivered: " + sn);
			}

		} catch (Exception e) {
			Logger.e(TAG, "packet delivered failed:" + sn + ", ex: " + e.getMessage());
		}
	}

	void onTimeSyncRequest() {

		long time_diff = -1;

		if (_conn != null) {
			time_diff = _conn.get_server_time_diff();
		}

		if (time_diff > 0) {
			TimePacket time_packet = new TimePacket(_conn.get_server_time_diff());
			pushPacket(time_packet, -1);
			_has_pending_get_time_req = false;
		} else {
			Logger.d(TAG, "LLC is not connected, wait till connected.");
			_has_pending_get_time_req = true;
		}
	}

	private final IServiceProxy.Stub _binder = new IServiceProxy.Stub() {
		@Override
		public void send_heartbeat(int appid) throws RemoteException {
			try {
				if (_conn != null) {
					_conn.send_heartbeat();
				}
			} catch (Exception e) {
				e.printStackTrace();
			}
		}

		@Override
		public boolean send_service_message(int appid, int serviceId, long sn, byte[] body) throws RemoteException {
			try {
				if (_conn != null) {
					return _conn.send_service_message(serviceId, sn, body);
				}
			} catch (Exception e) {
				Logger.e(TAG, "1 ex\n" + Log.getStackTraceString(e));
			}

			return false;
		}

		@Override
		public long get_sn(int appid) throws RemoteException {
			try {
				if (_conn != null) {
					return _conn.get_sn();
				}
			} catch (Exception e) {
				Logger.e(TAG, "2 ex\n" + Log.getStackTraceString(e));
			}
			return -1;
		}

		@Override
		public boolean send_message(int appid, String receiver, int msgType, long sn, byte[] body) throws RemoteException {
			try {
				if (_conn == null) {
					return false;
				}

				return _conn.send_message(receiver, AccountInfo.ACCOUNT_TYPE_JID, msgType, sn, body, 0, 0);
			} catch (Exception e) {
				Logger.e(TAG, "3 ex\n" + Log.getStackTraceString(e));
			}

			return false;
		}

		@Override
		public boolean query_presence(int appid, String users, long sn) throws RemoteException {
			try {
				if (_conn == null) {
					return false;
				}

				String userlist[] = users.split("_");
				return _conn.query_presence(userlist, sn);
			} catch (Exception e) {
				Logger.e(TAG, "4 ex\n" + Log.getStackTraceString(e));
			}
			return false;
		}

		@Override
		public void set_heartbeat_timeout(int appid, int heartbeat_timeout) throws RemoteException {

			try {
				if (_conn != null) {
					_conn.set_heartbeat_timeout(heartbeat_timeout);
				}
			} catch (Exception e) {
				Logger.e(TAG, "5 ex\n" + Log.getStackTraceString(e));
			}
		}

		@Override
		public long get_server_time_diff(int appid) throws RemoteException {

			if (_conn == null) {
				return -1;
			}
			try {
				onTimeSyncRequest();

			} catch (Exception e) {
				Logger.e(TAG, "6 ex\n" + Log.getStackTraceString(e));
			}

			try {

				return _conn.get_server_time_diff();
			} catch (Exception e) {
				Logger.e(TAG, "7 ex\n" + Log.getStackTraceString(e));
			}

			return -1;
		}

		@Override
		public boolean get_current_state(int appid) throws RemoteException {
			try {
				if (_conn == null) {
					return false;
				}
				return _conn.get_current_state();
			} catch (Exception e) {
				Logger.e(TAG, "8 ex\n" + Log.getStackTraceString(e));
			}

			return false;
		}

		@Override
		public synchronized void shutdown(int appid) throws RemoteException {
			try {
				shutdown_llc();
			} catch (Exception e) {
				Logger.e(TAG, "9 ex\n" + Log.getStackTraceString(e));
			}
		}

		@Override
		public void setOpenForeground(boolean open) throws RemoteException {
			isOpenForeground = open;
		}

		@Override
		public synchronized boolean get_message(int appid, String info_type, int[] ids, byte[] parameters) throws RemoteException {
			try {
				if (_conn == null) {
					return false;
				}
				return _conn.get_message(info_type, ids, parameters);
			} catch (Exception e) {
				Logger.e(TAG, "10 ex\n" + Log.getStackTraceString(e));
			}
			return false;
		}

		@Override
		public void switch_account(int appid, int clientVersion, String server, String defaultKey, String service, String account, String password,
				String device_id, String signature) throws RemoteException {

			try {
				AccountInfo tmpAccInfo = new AccountInfo(account, password, device_id, signature);
				ClientConfig tmpCC = new ClientConfig(appid, clientVersion, server, defaultKey, service);
				BgService.this.switch_account(tmpAccInfo, tmpCC);
			} catch (Exception e) {
				Logger.e(TAG, "11 ex\n" + Log.getStackTraceString(e));
			}
		}
	};
}
