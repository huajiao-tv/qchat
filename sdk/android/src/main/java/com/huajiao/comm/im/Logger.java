package com.huajiao.comm.im;

import android.annotation.SuppressLint;
import android.content.Context;
import android.os.Environment;
import android.text.TextUtils;
import android.util.Log;

import com.huajiao.comm.common.FeatureSwitch;
import com.huajiao.comm.common.JhFlag;
import com.huajiao.comm.common.LoggerBase;
//import com.huajiao.log.LogReport;

import java.io.File;

public class Logger extends LoggerBase {

	private static final LoggerBase _instance = new Logger();

	private static volatile boolean isEnableXlog = false;
	private static String BASE_ROOT_PATH = null;

//	private LogReport mLogReport;
	public Logger() {
		super("BG");
	}

	public static void enableXlog(Context context,boolean isEnable) {
		isEnableXlog = isEnable;
		if(isEnableXlog) {
			initXlog(context);
		}
	}

	
	public static void setUid(String uid) {
		_instance.setCurUid(uid);
	}

	public static void initXlog(Context context) {
//		System.loadLibrary("c++_shared");
//		System.loadLibrary("marsxlog");
		final String logPath = getRootPath(context) + "logcache";
//		final String cachePath = context.getFilesDir() + "/push_xlog";
//		LogReport.getInstance().setLogPath(logPath).setLogFileName("push").setLogNamePrefix("push_log").init();
	}

	public static String getRootPath(Context context){
		if(BASE_ROOT_PATH == null){
			initRootPath(context);
		}
		return BASE_ROOT_PATH;
	}

	@SuppressLint("SdCardPath")
	private static void initRootPath(Context context) {
		if (BASE_ROOT_PATH == null) {
			File file = null;
			try {
				file = Environment.getExternalStorageDirectory();
				if (file.exists() && file.canRead() && file.canWrite()) {
					//如果可读写，则使用此目录
					String path = file.getAbsolutePath();
					if (path.endsWith("/")) {
						BASE_ROOT_PATH = file.getAbsolutePath() + "huajiaoliving/";
					} else {
						BASE_ROOT_PATH = file.getAbsolutePath() + "/huajiaoliving/";
					}
				}
			} catch (Exception e) {

			}
			if (BASE_ROOT_PATH == null) {
				//如果走到这里，说明外置sd卡不可用
				if (context != null) {
					file = context.getFilesDir();
					String path = file.getAbsolutePath();
					if (path.endsWith("/")) {
						BASE_ROOT_PATH = file.getAbsolutePath() + "huajiaoliving/";
					} else {
						BASE_ROOT_PATH = file.getAbsolutePath() + "/huajiaoliving/";
					}
				} else {
					BASE_ROOT_PATH = "/sdcard/huajiaoliving/";
				}
			}
		}
		File file = new File(BASE_ROOT_PATH);
		if (!file.exists()) {
			file.mkdirs();
		}else if(!file.isDirectory()){
			file.delete();
			file.mkdirs();
		}
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

		if(isEnableXlog) {
			collectEventLog("","",0 ,"BG","tag:"+msg);

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
		if(isEnableXlog) {
			collectEventLog("","",0 ,"BG","tag:"+msg);

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
		if(isEnableXlog) {
			collectEventLog("","",0 ,"BG","tag:"+msg);

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
		if(isEnableXlog) {
			collectEventLog("","",0 ,"BG","tag:"+msg);

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
		if(isEnableXlog) {
			collectEventLog("","",0 ,"BG","tag:"+msg);
		}else {
			_instance.log(tag, "W: " + msg);

		}
	}

	public static void logEx(String tag, Exception e) {
		StringBuffer sb = new StringBuffer("--" + e.getMessage() + "--");
		StackTraceElement[] traces = e.getStackTrace();
		for (StackTraceElement trace : traces) {
			sb.append("\n" + trace.getClassName() + "." + trace.getMethodName() + ":" + trace.getLineNumber());
		}
		e(tag, sb.toString());
	}

	/**
	 * Event 打点
	 *
	 * @param log
	 */
	public static void collectEventLog(String classname, String funcname, int line, String tag, String log) {
		if (TextUtils.isEmpty(log)/*||!WRITE*/) {
			return;
		}
//		LogReport.getInstance().collectLog(tag, log);
	}
}
