package com.huajiao.comm.common;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.URL;
import java.net.URLConnection;

import android.util.Log;

public class HttpUtils {
	
	
	private static boolean isDEBUG=false;
	

	public static boolean isDEBUG() {
		return isDEBUG;
	}

	public static void setDEBUG(boolean isDEBUG) {
		HttpUtils.isDEBUG = isDEBUG;
	}

	
	//-------------------------------

	/** HTTP connect timeout */
	private static final int CONNECT_TIMEOUT = 15000;

	/** HTTP READ timeout */
	private static final int READ_TIMEOUT = 15000;

	/** Possible response body size */
	private static final int MAX_RESP_SIZE = 10240;
	
	private static String TAG = "HttpUtils";

	
	public static byte[] get(String request_url, byte[] request_data) {
		return get(request_url, request_data, CONNECT_TIMEOUT, READ_TIMEOUT);
	}
	
	
	/**
	 * 访问HTTP服务器
	 * @param request_url  URL
	 * @param request_data Request Body
	 * @return response body
	 */
	public static byte[] get(String request_url, byte[] request_data, int connect_timeout, int read_timeout) {

		byte[] body = null;
		boolean abnormal_size = false;
		
		if (request_url == null || request_url.length() == 0) {
			return body;
		}

		long start = System.currentTimeMillis();
		
		try {

			URL url = new URL(request_url);
			URLConnection urlConnection = url.openConnection();
		 
			urlConnection.setConnectTimeout(connect_timeout);
			urlConnection.setReadTimeout(read_timeout);
			urlConnection.setRequestProperty("content-type", "text/xml");

			if (request_data != null && request_data.length > 0) {
				urlConnection.setDoOutput(true);
				OutputStream out = urlConnection.getOutputStream();
				out.write(request_data);
				out.flush();
				out.close();
			}

			InputStream inputStream = urlConnection.getInputStream();
			ByteArrayOutputStream out_stream = new ByteArrayOutputStream();
			int dataRead = 0;
			int totalDataRead = 0;
			byte[] buffer = new byte[1024];

			while ((dataRead = inputStream.read(buffer)) > 0) {
				out_stream.write(buffer, 0, dataRead);
				totalDataRead += dataRead;
				if (totalDataRead > MAX_RESP_SIZE) {
					inputStream.close();
					abnormal_size = true;
					break;
				}
			}

			if(abnormal_size){
				Log.e(TAG, "payload size is abnormal !!!");
			} else {
				body = out_stream.toByteArray();
			}
			
			out_stream.close();
			inputStream.close();

		} catch (Throwable tr) {
			tr.printStackTrace();
		}
 	
		long consumption = System.currentTimeMillis() - start;
		Log.i(TAG, "get costs (ms): " + consumption);
		
		return body;
	}


	public static boolean touch(String request_url) {
		return touch(request_url, CONNECT_TIMEOUT, READ_TIMEOUT);
	}

		/**
         * Only to create access log
         * @param request_url
         * @param connect_timeout
         * @param read_timeout
         * @return
         */
	public static boolean touch(String request_url, int connect_timeout, int read_timeout) {

		boolean result = false;
		
		if (request_url == null || request_url.length() == 0) {
			return false;
		}

		long start = System.currentTimeMillis();
		
		try {

			URL url = new URL(request_url);
			URLConnection urlConnection = url.openConnection();
		 
			urlConnection.setConnectTimeout(connect_timeout);
			urlConnection.setReadTimeout(read_timeout);
			urlConnection.setRequestProperty("content-type", "text/xml");
			
			InputStream inputStream = urlConnection.getInputStream();			 
			int dataRead = 0;
			
			byte[] buffer = new byte[512];
			dataRead = inputStream.read(buffer);
			if(dataRead > 0){
				result = true;
			}

			inputStream.close();

		} catch (Throwable tr) {
			tr.printStackTrace();
		}
 	
		long consumption = System.currentTimeMillis() - start;
		Log.i(TAG, "touch costs (ms): " + consumption);
		
		return result;
	}
	
	
	

}
