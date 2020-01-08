package com.huajiao.comm.groupchat;

import com.huajiao.comm.common.FeatureSwitch;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.LoggerBase;
import android.util.Log;

public class GPLogger extends LoggerBase {

	private static final LoggerBase _instance = new GPLogger();
	
	public GPLogger() {
		super("GP");
	}
	
	public static void setUid(String uid) {
		_instance.setCurUid(uid);
	}

	/**
	 * Write error log
	 * */
	public static void e(String tag, String msg) {
		if (!FeatureSwitch.isLogOn()) {
			return;
		}
		
		if(JhFlag.enableDebug()) {
			Log.e(tag, msg);
		}

		_instance.log(tag, "E: " + msg);
	}

	/**
	 * Write Information log
	 * */
	public static void i(String tag, String msg) {
		if (!FeatureSwitch.isLogOn()) {
			return;
		}

		if(JhFlag.enableDebug()) {
			Log.i(tag, msg);
		}
		_instance.log(tag, "I: " + msg);
	}

	/**
	 * Write debug log
	 * */
	public static void d(String tag, String msg) {
		if (!FeatureSwitch.isLogOn()) {
			return;
		}

		if(JhFlag.enableDebug()) {
			Log.d(tag, msg);
		}
		_instance.log(tag, "D: " + msg);
	}

	public static void v(String tag, String msg) {
		if (!FeatureSwitch.isLogOn()) {
			return;
		}

		if(JhFlag.enableDebug()) {
			Log.v(tag, msg);
		}
		_instance.log(tag, "V: " + msg);
	}

	/**
	 * Write warning log
	 * */
	public static void w(String tag, String msg) {
		if (!FeatureSwitch.isLogOn()) {
			return;
		}

		if(JhFlag.enableDebug()) {
			Log.w(tag, msg);
		}
		_instance.log(tag, "W: " + msg);
	}

}
