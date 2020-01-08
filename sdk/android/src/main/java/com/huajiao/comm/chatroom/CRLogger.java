package com.huajiao.comm.chatroom;

import com.huajiao.comm.common.FeatureSwitch;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.LoggerBase;
import android.util.Log;

public class CRLogger extends LoggerBase {

	private static final LoggerBase _instance = new CRLogger();	

	private static LoggerWriterCallback mLogWriteCallback = null;

	public interface LoggerWriterCallback {
		void onWriteLog(String tag,String log);
	}

	public CRLogger() {
		super("CR");
	}

	public static void InitWriteCallBack(final  LoggerWriterCallback callback) {
		mLogWriteCallback = callback;
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

		if(mLogWriteCallback!=null) {
			mLogWriteCallback.onWriteLog(tag,"E: " + msg);
		}else {
			_instance.log(tag, "E: " + msg);
		}

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
		if(mLogWriteCallback!=null) {
			mLogWriteCallback.onWriteLog(tag,"I: " + msg);

		}else {
			_instance.log(tag, "I: " + msg);

		}
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
		if(mLogWriteCallback!=null) {
			mLogWriteCallback.onWriteLog(tag,"D: " + msg);

		}else {
			_instance.log(tag, "D: " + msg);

		}
	}

	public static void v(String tag, String msg) {
		if (!FeatureSwitch.isLogOn()) {
			return;
		}

		if(JhFlag.enableDebug()) {
			Log.v(tag, msg);
		}
		if(mLogWriteCallback!=null) {
			mLogWriteCallback.onWriteLog(tag,"V: " + msg);

		}else {
			_instance.log(tag, "V: " + msg);

		}
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
		if(mLogWriteCallback!=null) {
			mLogWriteCallback.onWriteLog(tag,"W: " + msg);

		}else {
			_instance.log(tag, "W: " + msg);

		}
	}

}
