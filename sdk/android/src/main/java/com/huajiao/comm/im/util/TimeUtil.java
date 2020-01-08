package com.huajiao.comm.im.util;

import java.text.SimpleDateFormat;
import java.util.Date;

public class TimeUtil {
	
	public static String getFullDateTime(long ms) {
		Date date = new Date(ms);
		SimpleDateFormat fmt = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss:SSS");
		return fmt.format(date);
	}
	
	@SuppressWarnings("deprecation")
	public static int getHours(long ms) {
		Date date = new Date(ms);
		return date.getHours();
	}
	
	public static String getFullTime(long ms) {
		Date date = new Date(ms);
		SimpleDateFormat fmt = new SimpleDateFormat("HH:mm:ss:SSS");
		return fmt.format(date);
	}
	
	public static String getDate(long ms) {
		Date date = new Date(ms);
		SimpleDateFormat fmt = new SimpleDateFormat("yyyy-MM-dd");
		return fmt.format(date);
	}
	
	public static String getDateTime(long ms) {
		Date date = new Date(ms);
		SimpleDateFormat fmt = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
		return fmt.format(date);
	}
	
	public static String getTime(long ms) {
		Date date = new Date(ms);
		SimpleDateFormat fmt = new SimpleDateFormat("HH:mm:ss");
		return fmt.format(date);
	}
	
	// ==========================================

	public static String getFullDateTime() {
		Date date = new Date();
		SimpleDateFormat fmt = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss:SSS");
		return fmt.format(date);
	}
	
	public static String getFullTime() {
		Date date = new Date();
		SimpleDateFormat fmt = new SimpleDateFormat("HH:mm:ss:SSS");
		return fmt.format(date);
	}
	
	public static String getDate() {
		Date date = new Date();
		SimpleDateFormat fmt = new SimpleDateFormat("yyyy-MM-dd");
		return fmt.format(date);
	}
	
	public static String getDateTime() {
		Date date = new Date();
		SimpleDateFormat fmt = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");
		return fmt.format(date);
	}
	
	public static String getTime() {
		Date date = new Date();
		SimpleDateFormat fmt = new SimpleDateFormat("HH:mm:ss");
		return fmt.format(date);
	}
}
