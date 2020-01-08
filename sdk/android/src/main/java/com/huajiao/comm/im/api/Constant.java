package com.huajiao.comm.im.api;

/**
 * 常量
 * */
public class Constant {
		
	/**
	 * Message has been sent successfully
	 * */
	public static final int RESULT_SUCCEEDED = 0;
	
	/**
	 * Indication of exception
	 * */
	public static final int RESULT_FAILED = 1;

	/**
	 * execution timeout
	 * */
	public static final int RESULT_TIMEOUT = 2;
	
	
	/***  尚未登陆成功， 需要重新登陆 */
	public static final int RESULT_UNAUTHORIZED = 3;
	
	
	/***  重试次数超过频度限制 */
	public static final int RESULT_EXCEEDS_RESEND_LIMIT = 4;

}
