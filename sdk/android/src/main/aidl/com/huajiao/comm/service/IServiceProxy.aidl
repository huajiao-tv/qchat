package com.huajiao.comm.service;

interface IServiceProxy {

	/**
	 * 获取 sn, 发送消息的sn 从这里获取， 客户端不能随便填sn 
	 * @return 返回sn
	 * */
	long get_sn(int appid);


    /**
    * 获取当前状态
    * @return Connected, Connecting, Disconnected, AuthFailed, LoggedInElsewhere
    */
	boolean get_current_state(int appid);

	/**
	 * 切换账号
	 * 
	 * @param account
	 *            新账号
	 * @param device_id
	 *            新的设备号
	 * */
	void switch_account(int appid, int clientVersion, String server, String defaultKey, String service,  String account, String password, String device_id, String signature);
 
	/**
	 * 发送消息
	 * 
	 * @param account
	 *            接受者的账号
	 * @param msgType
	 *            消息类型
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * @param body
	 *            消息体
	 * @return 成功返回true, 否则false
	 * */
	boolean send_message(int appid, String receiver, int msgType, long sn, in byte[] body);
	
	
	 /**
	 * 发送service消息
	 * 
	 * @param serviceId
	 *            对应的业务Id
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * @param body
	 *            消息内容
	 * @return 成功返回true, 否则返回false
	 * */
	boolean send_service_message(int appid, int serviceId, long sn, in byte[] body);
	
	
	/****
	* 拉取聊天室消息
	*/
	boolean get_message(int appid, String info_type, in int[] ids, in byte[] parameters);
	
	
	/**
	 * 查询用户在线状态
	 * */
	boolean query_presence(int appid, String users, long sn);
	
	/**
	 * 主动发起心跳包
	 * */
	void send_heartbeat(int appid);

	/**
	 * 设置心跳超时
	 * 
	 * @param heartbeat_timeout
	 *            超时， 单位毫秒, 最小不小于30000 (即30秒)
	 * */
	void set_heartbeat_timeout(int appid, int heartbeat_timeout);

	/**
    * 获取本地和服务器时间的差值
    */
	long get_server_time_diff(int appid);
	
	/**
	* 关闭长连接
	*/
	void shutdown(int appid);

    /**
    * 8.0系统以上，是否使用startForegroundservice方式，向上层发送消息
    */
	void setOpenForeground(boolean open);
}