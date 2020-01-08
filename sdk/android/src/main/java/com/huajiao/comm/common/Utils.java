package com.huajiao.comm.common;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.NetworkInterface;
import java.security.MessageDigest;
import java.util.Date;
import java.util.Enumeration;
import java.util.zip.GZIPInputStream;

import android.annotation.SuppressLint;
import android.app.ActivityManager;
import android.app.ActivityManager.RunningAppProcessInfo;
import android.content.Context;
import android.media.AudioManager;
import android.os.Build;
import android.telephony.TelephonyManager;

/**
 * 通用方法
 * */
public class Utils {

	private static String device_uuid = null;
	private static final char hexDigits[] = { '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f' };

	/**
	 * 获取设备唯一id
	 * */
	@SuppressLint("MissingPermission")
	public final static synchronized String get_device_uuid(Context context) {
		
		if (device_uuid == null) {
			
			String tmDevice;
			String androidId;

			try {
				final TelephonyManager tm = (TelephonyManager) context.getSystemService(Context.TELEPHONY_SERVICE);
				tmDevice = "" + tm.getDeviceId();
				androidId = "" + android.provider.Settings.Secure.getString(context.getContentResolver(), android.provider.Settings.Secure.ANDROID_ID);
			} catch (Exception e) {
				tmDevice = "";
				androidId = "";
			}
			
			device_uuid = MD5("android" + tmDevice + androidId);
		}
		
		return device_uuid;
	}

	public final static String MD5(String s) {
		
		try {

			byte[] btInput = s.getBytes();
			// 获得MD5摘要算法的 MessageDigest 对象
			MessageDigest mdInst = MessageDigest.getInstance("MD5");
			// 使用指定的字节更新摘要
			mdInst.update(btInput);
			// 获得密文
			byte[] md = mdInst.digest();
			// 把密文转换成十六进制的字符串形式
			int j = md.length;
			char str[] = new char[j * 2];
			int k = 0;

			for (int i = 0; i < j; i++) {
				byte byte0 = md[i];
				str[k++] = hexDigits[byte0 >>> 4 & 0xf];
				str[k++] = hexDigits[byte0 & 0xf];
			}

			return new String(str);
		} catch (Exception e) {
			return null;
		}
	}
	
	public final static String toHexString(byte[] bytes){
		if(bytes == null || bytes.length == 0){
			return "";
		}
		
		char str[] = new char[bytes.length * 2];
		int k = 0;

		for (int i = 0; i < bytes.length; i++) {
			byte byte0 = bytes[i];
			str[k++] = hexDigits[byte0 >>> 4 & 0xf];
			str[k++] = hexDigits[byte0 & 0xf];
		}
		
		return new String(str);
	}
	
	
	public final static String MD5(byte[] btInput) {
		 
		try {
			// 获得MD5摘要算法的 MessageDigest 对象
			MessageDigest mdInst = MessageDigest.getInstance("MD5");
			// 使用指定的字节更新摘要
			mdInst.update(btInput);
			// 获得密文
			byte[] md = mdInst.digest();
			// 把密文转换成十六进制的字符串形式
			int j = md.length;
			char str[] = new char[j * 2];
			int k = 0;

			for (int i = 0; i < j; i++) {
				byte byte0 = md[i];
				str[k++] = hexDigits[byte0 >>> 4 & 0xf];
				str[k++] = hexDigits[byte0 & 0xf];
			}

			return new String(str);
		} catch (Exception e) {
			return null;
		}
	}

	public static byte[] int_to_bytes(int paramInt) {
		byte buffer[] = new byte[4];
		buffer[0] = (byte) ((paramInt & 0xFF000000) >> 24);
		buffer[1] = (byte) ((paramInt & 0xFF0000) >> 16);
		buffer[2] = (byte) ((paramInt & 0xFF00) >> 8);
		buffer[3] = (byte) ((paramInt & 0xFF));
		return buffer;
	}

	/**
	 * 比较二进制内容是否相等
	 * */
	public static boolean compare_bytes(byte[] b1, byte[] b2, int len) {
		if (len == 0 || b1 == null || b2 == null) {
			return false;
		}

		for (int i = 0; i < len; i++) {
			if (b1[i] != b2[i]) {
				return false;
			}
		}

		return true;
	}

	public static long bytes_to_long(byte[] buffer) {
		if (null == buffer || buffer.length < 8) {
			return -1;
		}

		long p = 0;

		for (int i = 0; i < 8; i++) {
			if (buffer[i] < 0) {
				p += ((256 + (long) buffer[i]) << ((7 - i) * 8));

			} else {
				p += ((long) buffer[i] << ((7 - i) * 8));
			}
		}

		return p;
	}

	public static boolean is_valid_phonenumber(String pn) {
		if (pn == null || pn.length() != 11) {
			return false;
		}

		try {
			long r = Long.parseLong(pn);
			if (r > 10000000000L) {
				return true;
			}
		} catch (Exception ex) {
		}

		return false;
	}

	public static int bytes_to_int(byte[] buffer, int start_index) {
		if (null == buffer || buffer.length < 4 || start_index + 3 >= buffer.length) {
			return -1;
		}
		int p = 0;
		for (int i = start_index; i < start_index + 4; i++) {
			if (buffer[i] < 0) {
				p += ((256 + (int) buffer[i]) << ((3 + start_index - i) * 8));

			} else {
				p += ((int) buffer[i] << ((3 + start_index - i) * 8));
			}
		}
		return p;
	}

	public static byte[] long_to_byte(long data) {
		byte[] bytes = new byte[8];
		bytes[7] = (byte) (data & 0xff);
		bytes[6] = (byte) ((data >> 8) & 0xff);
		bytes[5] = (byte) ((data >> 16) & 0xff);
		bytes[4] = (byte) ((data >> 24) & 0xff);
		bytes[3] = (byte) ((data >> 32) & 0xff);
		bytes[2] = (byte) ((data >> 40) & 0xff);
		bytes[1] = (byte) ((data >> 48) & 0xff);
		bytes[0] = (byte) ((data >> 56) & 0xff);
		return bytes;
	}

	public static byte[] int_to_byte(int data) {
		byte[] bytes = new byte[4];
		bytes[3] = (byte) (data & 0xff);
		bytes[2] = (byte) ((data >> 8) & 0xff);
		bytes[1] = (byte) ((data >> 16) & 0xff);
		bytes[0] = (byte) ((data >> 24) & 0xff);
		return bytes;
	}

	public static byte[] short_to_byte(short data) {
		byte[] bytes = new byte[4];
		bytes[1] = (byte) (data & 0xff);
		bytes[0] = (byte) ((data >> 8) & 0xff);
		return bytes;
	}

	@SuppressWarnings("deprecation")
	public static String getDateString(long stamp) {
		Date date = new Date(stamp);
		return date.toLocaleString();

	}

	public static int compute_rough_level(byte[] data, int len) {

		int level = 0;

		for (int i = 0; i < data.length; i += 2) {
			int sample = (short) (data[i] + (data[i + 1] << 8));
			if (sample < 0) {
				sample = -sample;
			}

			level = (sample + level) >> 2;
		}

		return level;
	}

	/**
	 * @param data
	 *            PCM 数据
	 * @param rate
	 *            比率
	 * */
	public static void adjust_amplitude_pcm16bit(byte[] data, float rate) {
		if (data == null) {
			return;
		}

		for (int i = 0; i < data.length; i += 2) {
			short sample = (short) (data[i] + (data[i + 1] << 8));
			sample *= rate;
			data[i] = (byte) (sample & 0xff);
			data[i + 1] = (byte) ((sample & 0xff00) >> 8);
		}
	}

	/**
	 * get device manufacturer and model
	 * */
	public static String getModel() {
		String ma = Build.MANUFACTURER == null ? "" : Build.MANUFACTURER;
		String mo = Build.MODEL == null ? "" : Build.MODEL;
		return String.format("%s_%s", ma, mo);
	}

	public static int filterOutLowVoice(byte inOut[], int len, int threshold) {
		int lenProcessed = 0;

		for (int i = 0; i < len / 2; i += 2) {
			int magnitude = (int) (inOut[2 * i + 1]) << 8 + inOut[2 * i];

			if (Math.abs(magnitude) < threshold) {
				inOut[2 * i] = (byte) (inOut[2 * i] / 2);
				inOut[2 * i + 1] = (byte) (inOut[2 * i + 1] / 2);
			}

			lenProcessed += 2;
		}

		return lenProcessed;
	}

	public static int MagnifyOutLowVoice(byte inOut[], int len, double ratio) {
		int lenProcessed = 0;

		/*
		 * for(int i = 0; i < data.length; i+=2){ short sample = (short)(data[i]
		 * + (data[i+1] << 8)); sample *= rate; data[i] = (byte)(sample & 0xff);
		 * data[i +1] = (byte)((sample & 0xff00) >> 8); }
		 */

		for (int i = 0; i < len / 2; i++) {
			int magnitude = (int) (inOut[2 * i] + (inOut[2 * i + 1] << 8)); // there
																			// is
																			// a
																			// trap,
																			// need
																			// to
																			// conver
																			// to
																			// short
																			// at
																			// last

			magnitude = (int) (magnitude * ratio);

			if (magnitude > 32677)
				magnitude = 32677;
			else if (magnitude < -32677)
				magnitude = -32677;

			inOut[2 * i] = (byte) (((short) magnitude) & 0xff);
			inOut[2 * i + 1] = (byte) (((short) magnitude) >> 8);

			lenProcessed += 2;
		}

		return lenProcessed;
	}

	public static boolean is_data_empty(byte[] sourceData, int len) {

		if (sourceData == null || sourceData.length < len) {
			return false;
		}

		for (int i = 0; i < len; i++) {
			if (sourceData[i] != 0) {
				return false;
			}
		}

		return true;
	}

	/**
	 * in place resampling data to 8KHz mono
	 * */
	public static int resample_to_8k_mono(int sourceSampleRateHz, int sourceChannelNum, byte[] sourceData, int len) {
		if (sourceData == null || sourceData.length < len) {
			return -1;
		}

		// no need to resample data
		if (sourceSampleRateHz == 8000 && sourceChannelNum == 1) {
			return len;
		}

		// 16Khz mono
		if (sourceSampleRateHz == 16000 && sourceChannelNum == 1) {

			// incorrect len
			if (len % 4 != 0) {
				return -2;
			}

			int alen = 0;
			for (int i = 0; i < len; i += 4) {
				short s1 = (short) (sourceData[i] + (sourceData[i + 1] << 8));
				// short s2 = (short) (sourceData[i + 2] | (sourceData[i + 3] <<
				// 8));
				// short mixed = (short) ((s1 + s2) >> 1);
				short mixed = s1;

				sourceData[alen++] = (byte) (mixed & 0xff);
				sourceData[alen++] = (byte) ((mixed & 0xff00) >> 8);
			}

			return alen;
		}

		// 16Khz stereo
		if (sourceSampleRateHz == 16000 && sourceChannelNum == 2) {

			// incorrect len
			if (len % 8 != 0) {
				return -2;
			}

			int alen = 0;
			for (int i = 0; i < len; i += 8) {

				short s1l = (short) (sourceData[i] + (sourceData[i + 1] << 8));
				short s1r = (short) (sourceData[i + 2] + (sourceData[i + 3] << 8));
				// short s2l = (short) (sourceData[i + 4] | (sourceData[i + 5]
				// << 8));
				// short s2r = (short) (sourceData[i + 6] | (sourceData[i + 7]
				// << 8));

				// short mixed1 = (short) ((s1l + s1r) >> 1);
				// short mixed2 = (short) ((s2l + s2r) >> 1);

				short mixed = (short) ((s1l + s1r) >> 1);

				sourceData[alen++] = (byte) (mixed & 0xff);
				sourceData[alen++] = (byte) ((mixed & 0xff00) >> 8);
			}

			return alen;
		}

		return -4;
	}

	public static short compute_max_amplitude(byte[] sourceData, int len) {

		if (sourceData == null || sourceData.length < len) {
			return Short.MAX_VALUE;
		}

		short max = 0;

		for (int i = 0; i < len; i += 2) {
			short s1 = (short) (sourceData[i] + (sourceData[i + 1] << 8));
			if (s1 < 0) {
				s1 = (short) -s1;
			}
			if (s1 > max) {
				max = s1;
			}
		}

		return max;
	}

	public static String AudioModeToString(int mode) {
		switch (mode) {
		case AudioManager.MODE_IN_CALL:
			return "MODE_IN_CALL";
		case AudioManager.MODE_NORMAL:
			return "MODE_IN_NORMAL";
		case AudioManager.MODE_IN_COMMUNICATION:
			return "MODE_IN_COMMUNICATION";
		default:
			return "" + mode;
		}
	}

	public static void xor_encrypt(byte[] data, int len, byte key) {
		if (data == null || data.length < len) {
			return;
		}

		for (int i = 0; i < len; i++) {
			data[i] ^= key;
		}
	}

	public static String getLocalHostIp() {
		String ipaddress = "";
		try {
			Enumeration<NetworkInterface> en = NetworkInterface.getNetworkInterfaces();
			// 遍历所用的网络接口
			while (en.hasMoreElements()) {
				NetworkInterface nif = en.nextElement();// 得到每一个网络接口绑定的所有ip
				Enumeration<InetAddress> inet = nif.getInetAddresses();
				// 遍历每一个接口绑定的所有ip
				while (inet.hasMoreElements()) {
					InetAddress ip = inet.nextElement();
					if (!ip.isLoopbackAddress() && InetAddressUtils.isIPv4Address(ip.getHostAddress())) {
						ipaddress = ip.getHostAddress();
						return ipaddress;
					}
				}
			}
		} catch (Throwable e) {
			e.printStackTrace();
		}
		return ipaddress;

	}

	public static boolean isUdpPortAvailable(short port) {

		try {
			DatagramSocket ds = new DatagramSocket(port);
			ds.close();
			return true;
		} catch (Throwable se) {
			se.printStackTrace();
		}

		return false;
	}
	
	
	public static byte[] ungzip(byte[] input) {
		byte result[] =null;
		if (input == null || input.length < 1) {
			return result;
		}

		byte[] buffer = new byte[1024];
		int dataRead = 0;
		ByteArrayOutputStream outputStream = null;
		GZIPInputStream gzippedStream = null;
		try {
			outputStream = new ByteArrayOutputStream();

			gzippedStream = new GZIPInputStream(new ByteArrayInputStream(input));

			while ((dataRead = gzippedStream.read(buffer, 0, buffer.length)) > 0) {
				outputStream.write(buffer, 0, dataRead);
			}
			result = outputStream.toByteArray();
		} catch (Exception e) {
			e.printStackTrace();
		}finally {
			if(outputStream!=null){
				try {
					outputStream.close();
				} catch (IOException e1) {
					e1.printStackTrace();
				}
			}
			if(gzippedStream!=null){
				try {
					gzippedStream.close();
				} catch (IOException e1) {
					e1.printStackTrace();
				}
			}
		}
		return result;
	}

	public static String getAppNameByPID(Context context, int pid){
	    
		ActivityManager manager = (ActivityManager) context.getSystemService(Context.ACTIVITY_SERVICE);
	    
	    for(RunningAppProcessInfo processInfo : manager.getRunningAppProcesses()){
	        if(processInfo.pid == pid){
	            return processInfo.processName;
	        }
	    }
	    
	    return "";
	}
	
	public static String getStackTrace(StackTraceElement[] eles) {
		StringBuffer sb = new StringBuffer();
		for(StackTraceElement ele : eles) {
			sb.append("\n"+ele);
		}
		return sb.toString();
	}	
}
