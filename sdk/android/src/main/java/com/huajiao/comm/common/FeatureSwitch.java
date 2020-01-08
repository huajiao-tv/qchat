package com.huajiao.comm.common;

import java.util.Locale;

import android.util.Log;

/**
 * feature switches
 * */
public class FeatureSwitch {
	
	
	// TODO
	private static boolean _log_on = true;
	private static boolean _report_on = false;
	private static boolean _pulling_on = false;
	private static int maxOverloadMissCount = 500;

	private static boolean openStartForeground = false;// startService还是使用startForegroundService

	/**
	 * 设置各种云控开关
	 * */
	public static void setSwitch(int value) {
		
		 _log_on = (value & 1) > 0;
		
		 _pulling_on = (value & 2) > 0;
		
		_report_on = (value & 4) > 0;	
		
		if(BuildFlag.DEBUG){
			_log_on = true;
		}
		
		Log.i("HJFS", String.format(Locale.US, "set %d", value));
	}

	public static void setOpenStartForeground(boolean open) {
		Log.e("BGS","setOpenStartForeground = "+open);
		openStartForeground = open;
	}

	public static boolean isOpenStartForeground() {
		return openStartForeground;
	}

	public static void setMaxOverloadMissCount(int maxOverloadMissCount) {
		if(maxOverloadMissCount > 0) {
			FeatureSwitch.maxOverloadMissCount = maxOverloadMissCount;
		}
	}
	
	public static int getMaxOverloadMissCount() {
		return maxOverloadMissCount;
	}

	/**
	 * 是否开启日志
	 * */
	public static boolean isLogOn() {
		return _log_on;
	}
	
	/**
	 * 是否打印
	 * */
	public static boolean isReportOn(){
		return _report_on;
	}
	
	/**
	 * 是否拉取消息
	 * */
	public static boolean isPullingOn(){
		return _pulling_on;
	}
}
