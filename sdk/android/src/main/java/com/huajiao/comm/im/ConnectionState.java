package com.huajiao.comm.im;

public enum ConnectionState {
	
	/**
	 * 正在登录服务器
	 **/
	Connecting, 
	
	/**
	 * 已经登录成功
	 * */
	Connected,
 
	/**
	 * 与服务器断开了连接
	 * */
	Disconnected,
	
	/**
	 * 授权失败， 原因是账号或密码不正确
	 * */
	AuthFailed,
	
	/**
	 * 账号在别处登录了
	 * */
	LoggedInElsewhere,
	
	
	/**
	 * 已经被关闭了
	 * */
	Shutdown
	
}
