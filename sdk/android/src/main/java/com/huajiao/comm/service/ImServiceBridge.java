package com.huajiao.comm.service;

import java.util.concurrent.atomic.AtomicLong;

import android.annotation.SuppressLint;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.os.Build;
import android.os.IBinder;
import android.os.SystemClock;
import android.text.TextUtils;
import android.util.Log;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.BuildFlag;
import com.huajiao.comm.common.ClientConfig;
import com.huajiao.comm.im.Logger;
import com.huajiao.comm.im.rpc.AccountCmd;
import com.huajiao.comm.im.rpc.Cmd;
import com.huajiao.comm.im.rpc.GetMessageCmd;
import com.huajiao.comm.im.rpc.GetStateCmd;
import com.huajiao.comm.im.rpc.MsgCmd;
import com.huajiao.comm.im.rpc.PresenceCmd;
import com.huajiao.comm.im.rpc.ServiceMsgCmd;
import com.huajiao.comm.im.rpc.ShutdownCmd;
import com.huajiao.comm.im.rpc.SyncTimeCmd;

/**
 * MediaSDK和ImService 之间沟通的桥梁<br>
 * 指令通过该类发送给ImService<br>
 * ImService的消息和响应通过广播发给该类<br>
 * */
public class ImServiceBridge {

	protected static final String TAG = "BGS-BRI";

	private Context _context;

	private ClientConfig _clientConfig = null;

	private AccountInfo _account_info = null;

	private AtomicLong _snSeed = new AtomicLong(System.currentTimeMillis());

	private Object _lock = new Object();

	private int _boundCnt = 0;

	private boolean _shutdown = false;

	private final static int BIND_TIMEOUT = 1000;

	private final static int BIND_MAX_COUNT = 3;

	private final static int BIND_INTERVAL = 15000;

	private long _last_bind_time = 0;

	private volatile IServiceProxy _service_proxy = null;

	private ServiceConnection _conn = null;

	protected final static int MAX_STRING_LEN = 4096;

	protected final static int MAX_BYTES_LEN = 4096;
	private boolean _openForeground = false;

	private void init_service_connection() {

		_conn = new ServiceConnection() {
			@Override
			public void onServiceDisconnected(ComponentName name) {
				Logger.d(TAG, "Service disconnected");
				synchronized (_lock) {
					_service_proxy = null;
				}
			}

			@Override
			public void onServiceConnected(ComponentName className, IBinder service) {

				Logger.d(TAG, "onServiceConnected");

				try {

					IServiceProxy tmp = IServiceProxy.Stub.asInterface(service);

					if (tmp == null) {
						Logger.w(TAG, "IServiceProxy is null");
						return;
					}

					tmp.switch_account(_clientConfig.getAppId(), _clientConfig.getClientVersion(), _clientConfig.getServer(), _clientConfig.getDefaultKey(),
							_clientConfig.get_im_business_service(), _account_info.get_account(), _account_info.get_password(), _account_info.get_device_id(),
							_account_info.get_signature());

					tmp.setOpenForeground(_openForeground);
					synchronized (_lock) {
						_service_proxy = tmp;
						_lock.notifyAll();
					}

					Logger.d(TAG, "Service bound");

				} catch (Exception e) {
					Logger.w(TAG, "onServiceConnected ex\n" + Log.getStackTraceString(e));
				}
			}
		};
	}

	public ImServiceBridge(Context context, AccountInfo accountInfo, ClientConfig clientConfig) {

		if (context == null || accountInfo == null || clientConfig == null) {
			throw new IllegalArgumentException();
		}

		_context = context.getApplicationContext();
		_account_info = accountInfo;
		_clientConfig = clientConfig;
        if(_account_info!=null && !TextUtils.isEmpty(_account_info.get_account())){
            Logger.setUid(_account_info.get_account());
        }

		init_service_connection();

		bindService();

		wait_for_connection(BIND_TIMEOUT);
	}

	/**
	 * if target service is down, bind will be called
	 * */
	@SuppressLint("InlinedApi") 
	private void bindService() {

		synchronized (_lock) {

			if (_service_proxy == null) {

				// bindService is ok, but ServiceProxy is not returned.
				if (_boundCnt > BIND_MAX_COUNT) {
					try {
						Intent stop_intent = new Intent(_context, BgService.class);
						_context.unbindService(_conn);
						_context.stopService(stop_intent);
						_boundCnt = 0;
					} catch (Exception e) {
						Logger.e(TAG, "unbind failed " + Log.getStackTraceString(e));
					}
				}

				Intent service_intent = new Intent(_context, BgService.class);
				service_intent.putExtra(BgService.KEY_ACCOUNT_INFO, _account_info);
				service_intent.putExtra(BgService.KEY_CLIENT_CONFIG, _clientConfig);
				service_intent.putExtra(BgService.KEY_LOG_DIR,Logger.getDebugFilePath());

				int flags = Context.BIND_AUTO_CREATE;
				if (Build.VERSION.SDK_INT >= 14) {
					flags |= Context.BIND_IMPORTANT;
				}

				if (_context.bindService(service_intent, _conn, flags)) {
					_boundCnt++;
					Logger.d(TAG, "bindService returns ok");
				} else {
					_boundCnt = 0;
					Logger.e(TAG, "bindService failed");
				}

				_last_bind_time = SystemClock.elapsedRealtime();
			}
		}
	}

	private boolean wait_for_connection(int time_ms) {

		if (_service_proxy != null) {
			return true;
		}

		synchronized (_lock) {
			if (_service_proxy == null) {
				try {
					_lock.wait(time_ms);
				} catch (InterruptedException e) {
					e.printStackTrace();
				}
			}

			return (_service_proxy != null);
		}
	}

	public synchronized boolean check_service() {

		if (_shutdown) {
			return false;
		}

		if (_service_proxy != null) {
			return true;
		}

		// do not block too frequently
		if (SystemClock.elapsedRealtime() - this._last_bind_time > BIND_INTERVAL) {
			if (BuildFlag.DEBUG) {
				Logger.w(TAG, "BgService is not bound, try to bind");
			}

			bindService();
			return wait_for_connection(BIND_TIMEOUT);
		}

		return (_service_proxy != null);
	}

	private boolean send_cmd(Cmd cmd) {

		boolean result = false;

		try {

			check_service();

			Intent service_intent = new Intent(_context, BgService.class);
			service_intent.putExtra(BgService.KEY_CMD, cmd);
			_context.startService(service_intent);
			result = true;

			if (BuildFlag.DEBUG) {
				Logger.i(TAG, "Service unbound, send_cmd via startService: " + cmd.toString());
			}

		} catch (Exception e) {
			Logger.e(TAG, "send_cmd " + Log.getStackTraceString(e));
		}

		return result;
	}

	public long get_sn() {
		return _snSeed.incrementAndGet();
	}

	public synchronized void switch_account(AccountInfo account_info, ClientConfig client_config) {

		if (_service_proxy != null) {
			try {
				_service_proxy.switch_account(client_config.getAppId(), client_config.getClientVersion(), client_config.getServer(),
						client_config.getDefaultKey(), client_config.get_im_business_service(), account_info.get_account(), account_info.get_password(),
						account_info.get_device_id(), account_info.get_signature());
				return;
			} catch (Exception e) {
				Logger.e(TAG, "s w " + Log.getStackTraceString(e));
			}
		}

		AccountCmd cmd = new AccountCmd(account_info, client_config);
		send_cmd(cmd);
	}

	public boolean send_message(String receiver, int account_type, int msgType, long sn, byte[] body) {
		return send_message(receiver, account_type, msgType, sn, body, 0, 0);
	}

	public boolean send_message(String receiver, int account_type, int msgType, long sn, byte[] body, int timeout_ms, int expiration_sec) {

		if (_service_proxy != null) {
			try {
				return _service_proxy.send_message(_clientConfig.getAppId(), receiver, msgType, sn, body);
			} catch (Exception e) {
				Logger.e(TAG, "send_message " + Log.getStackTraceString(e));
			}
		}

		MsgCmd cmd = new MsgCmd(receiver, account_type, msgType, sn, body, timeout_ms, expiration_sec);
		return send_cmd(cmd);
	}

	public synchronized boolean send_service_message(int serviceId, long sn, byte[] body) {

		if (body == null || body.length == 0) {
			Logger.e(TAG, "body is empty");
			return false;
		}

		if (body.length > MAX_BYTES_LEN) {
			Logger.e(TAG, "body size exceeds limit");
			return false;
		}

		if (_service_proxy != null) {
			try {
				return _service_proxy.send_service_message(_clientConfig.getAppId(), serviceId, sn, body);
			} catch (Exception e) {
				Logger.e(TAG, "send_service_message " + Log.getStackTraceString(e));
			}
		}

		ServiceMsgCmd cmd = new ServiceMsgCmd(serviceId, sn, body);
		return send_cmd(cmd);
	}

	public boolean query_presence(String[] users, long sn, int account_type) {

		String combined = "";

		for (int i = 0; i < users.length; i++) {
			String s = users[i];
			if (i != users.length - 1) {
				combined += s + "_";
			} else {
				combined += s;
			}
		}

		PresenceCmd cmd = new PresenceCmd(combined, sn, account_type);
		return send_cmd(cmd);
	}

	public boolean sync_time() {

		if (_service_proxy != null) {
			try {
				long diff = _service_proxy.get_server_time_diff(_clientConfig.getAppId());
				if (diff > 0) {
					return true;
				}
			} catch (Exception e) {
				Logger.e(TAG, "sync_time " + Log.getStackTraceString(e));
			}
		}

		return send_cmd(new SyncTimeCmd());
	}

	public synchronized boolean get_message(String info_type, int[] ids, byte[] parameters) {

		if (parameters == null || parameters.length == 0) {
			Log.e(TAG, "parameters is empty");
			return false;
		}

		if (parameters.length > MAX_BYTES_LEN) {
			Log.e(TAG, "parameters size exceeds limit");
			return false;
		}

		if (ids == null || ids.length == 0) {
			Log.e(TAG, "empty ids");
			return false;
		}

		if (ids.length > MAX_BYTES_LEN / 4) {
			Log.e(TAG, "ids size exceeds limit");
			return false;
		}

		if (_service_proxy != null) {
			try {
				return _service_proxy.get_message(_clientConfig.getAppId(), info_type, ids, parameters);
			} catch (Exception e) {
				Logger.e(TAG, "get_message " + Log.getStackTraceString(e));
			}
		}

		GetMessageCmd cmd = new GetMessageCmd(info_type, ids, parameters);
		return send_cmd(cmd);
	}

	public boolean get_current_state() {

		if (_service_proxy != null) {
			try {
				return _service_proxy.get_current_state(_clientConfig.getAppId());
			} catch (Exception e) {
				Logger.e(TAG, "get_current_state " + Log.getStackTraceString(e));
			}
		}

		return send_cmd(new GetStateCmd());
	}

	public boolean shutdown() {

		Logger.i(TAG, "shutdown called");

		_shutdown = true;

		if (_service_proxy != null) {
			try {
				_service_proxy.shutdown(_clientConfig.getAppId());
				return true;
			} catch (Exception e) {
				Logger.e(TAG, "shutdown " + Log.getStackTraceString(e));
			}
		}

		return send_cmd(new ShutdownCmd());
	}

	public void setOpenForeground(boolean open) {
		_openForeground = open;
		if (_service_proxy != null) {
			try {
				_service_proxy.setOpenForeground(open);
			} catch (Exception e) {
				Logger.e(TAG, "setOpenForeground " + Log.getStackTraceString(e));
			}
		}
	}
}
