package com.huajiao.comm.im;

import java.util.Locale;

import com.huajiao.comm.common.RC4;
import android.content.Context;
import android.content.SharedPreferences;
import android.content.SharedPreferences.Editor;
import android.util.Log;

/**
 * 客户端最新消息Id
 * */
class MessageFlag {

	private SharedPreferences _prefs;
	private String _account;
	private RC4 rc4 =  null; 
	
	/**
	 * 加载配置
	 */
	protected void load_flag() {
		_last_msg_id = _prefs.getLong(getCryptKeyV2(), -1);
		if (_last_msg_id != -1) {
			return;
		}
	}

	private String getCryptKeyV2() {
		return String.format(Locale.US, "%d_%s_%s", 2080, rc4.encry_RC4_string(_account), _info_type);
	}

	/**
	 * 保存配置
	 * */
	protected void save_flag() {
		try {
			Editor editor = _prefs.edit();
			editor.putLong(getCryptKeyV2(), _last_msg_id);
			if(!editor.commit()){
				//Logger.e(TAG, "commit failed: " + _last_msg_id);
			} else {
				//Logger.d(TAG, "flag saved:" + _last_msg_id);
			}
		} catch (Exception e) {
			Logger.e("MF", Log.getStackTraceString(e));
		}
	}

	/**
	 * 本地的最新的消息id
	 * */
	private long _last_msg_id;

	 
	protected long get_last_msg_id() {
		return _last_msg_id;
	}

	public void set_last_msg_id(long last_msg_id) {
		if (_account != null && _account.equals(_getting_account)) {
			_last_msg_id = last_msg_id;
			save_flag();
		}
	}

	/**
	 * 最后一次收到通知的时间
	 * */
	private long _last_notify_time;

	/**
	 * 按需设置最后一次通知的时间
	 * */
	public void set_last_notify_time(long notify_time) {
		if (notify_time > _last_notify_time) {
			this._last_notify_time = notify_time;
		}
	}

	public long get_last_notify_time() {
		return _last_notify_time;
	}

	/**
	 * 获取消息动作的发起时间
	 * */
	public long _get_msg_time;

	/**
	 * 获取消息时的SN， 因为取消息有可能出错， 需要匹配
	 * */
	public long _get_msg_sn;

	/**
	 * 是否正在取消息
	 * */
	public boolean _getting_msg;

	/**
	 * 获取消息时用到的账号
	 * */
	public String _getting_account;

	/**
	 * 消息类型
	 * */
	private String _info_type;

	public String get_info_type() {
		return _info_type;
	}

	/**
	 * 清零
	 * */
	public void reset() {
		_last_notify_time = _get_msg_time = 0;
		_getting_msg = false;
		_get_msg_sn = 0;
	}

	public void switch_account(String newAccount) {

		save_flag(); // 保存旧的标记

		reset(); // 清零

		_account = newAccount;

		load_flag(); // 读取配置
	}

	public MessageFlag(Context context, String info_type, String account, String defaultKey) {
		_info_type = info_type;
		_account = account;
		_prefs = context.getSharedPreferences(LLConstant.PREF_KEY_ID, Context.MODE_PRIVATE);
		rc4 = new RC4(defaultKey);
		load_flag();
	}
}
