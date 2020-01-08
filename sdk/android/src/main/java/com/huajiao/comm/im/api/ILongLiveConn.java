package com.huajiao.comm.im.api;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;

public interface ILongLiveConn {

	/**
	 * 获取 sn, 发送消息的sn 从这里获取， 客户端不能随便填sn
	 * 
	 * @return 返回sn
	 * */
	long get_sn();

	/**
	 * 初始化或者切换账号
	 * 
	 * @param account_info
	 * @param client_config
	 */
	void switch_account(AccountInfo account_info, ClientConfig client_config);

	/**
	 * 发送消息
	 * 
	 * @param receiver
	 *            接受者者的账号
	 * @param msgType
	 *            消息类型
	 * @param sn
	 *            消息sn, 不同消息请使用不同sn， 最好调用 {@link get_sn}获取
	 * @param body
	 *            消息体二进制内容
	 * @param timeoutMs
	 *            消息发送超时， 单位毫秒
	 * @param expireationSec
	 *            消息过期时间单位秒
	 * @return 成功返回true, 否则false
	 * */
	boolean send_message(String receiver, int msgType, long sn, byte[] body, int timeoutMs, int expireationSec);

	/**
	 * 发送消息
	 * 
	 * @param receiver
	 *            接收者
	 * @param account_type
	 *            接收者的账号类型AccountInfo.ACCOUNT_TYPE_*
	 * @param msg_type
	 *            消息类型MessageType
	 * @param sn
	 *            消息sn, 不同消息请使用不同sn， 最好调用 {@link get_sn}获取
	 * @param body
	 *            消息内容
	 * @param timeout_ms
	 *            发送超时时间， 单位毫秒
	 * @param expireation_sec
	 *            消息过期时间, 单位秒
	 * @return 成功返回true
	 */
	boolean send_message(String receiver, int account_type, int msg_type, long sn, byte[] body, int timeoutMs, int expireationSec);

	/**
	 * 发送消息
	 * 
	 * @deprecated 请使用带 account_type参数的send_message
	 * 
	 * @param receiver
	 *            接受者者的账号, 默认账号类型是手机号
	 * @param msg_type
	 *            消息类型
	 * @param sn
	 *            消息sn, 不同消息请使用不同sn， 最好调用 {@link get_sn}获取
	 * @param body
	 *            消息体二进制内容
	 * @param timeout_ms
	 *            消息发送超时， 单位毫秒
	 * 
	 * @return 成功返回true, 否则false
	 * */
	boolean send_message(String receiver, int msg_type, long sn, byte[] body, int timeout_ms);

	/**
	 * 发送消息
	 * 
	 * @deprecated 请使用带 account_type参数的send_message
	 * @param receiver
	 *            接受者的账号
	 * @param msgType
	 *            消息类型
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * @param body
	 *            消息体二进制内容
	 * @param timeoutMs
	 *            消息发送超时， 单位毫秒
	 * 
	 * @return 成功返回true, 否则false
	 * */
	boolean send_message(String receiver, int msgType, long sn, byte[] body);

	/**
	 * 发送service消息
	 * 
	 * @param serviceId
	 *            对应的业务Id
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * @param body
	 *            二进制内容
	 * 
	 * @return 成功返回true, 否则返回false
	 * */
	boolean send_service_message(int serviceId, long sn, byte[] body);

	/**
	 * 查询用户在线状态
	 *  @deprecated 请使用带 account_type参数的query_presence
	 * 
	 * @param users
	 *            要查询的用户的手机号码
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * */
	boolean query_presence(String users[], long sn);

	/**
	 * 查询用户在线状态
	 * 	
	 * @param users
	 *            要查询的用户的手机号码
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * @param account_type
	 *            账号类型
	 * */
	boolean query_presence(String users[], long sn, int account_type);

	/**
	 * 设置心跳超时
	 * 
	 * @param heartbeat_timeout
	 *            超时， 单位毫秒, 最小不小于30000 (即30秒)
	 * */
	void set_heartbeat_timeout(int heartbeat_timeout);

	/**
	 * 同步服务器时间
	 * @return  成功返回true
	 * */
	boolean get_server_time();
	
	/**
	 * 拉取消息
	 * @param info_type
	 * @param ids
	 * @param parameters
	 * @return
	 */
	boolean get_message(String info_type, int [] ids, byte[] parameters);
	
	
	boolean shutdown();

	/***
	 * 异步获取长连接状态
	 * */
	boolean get_current_state();

	/**
	 * 8.0以上是否使用启动前台服务方式
	 * @param open
	 */
	void setOpenForeground(boolean open);
}
