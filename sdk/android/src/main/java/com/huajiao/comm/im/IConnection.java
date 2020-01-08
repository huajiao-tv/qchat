package com.huajiao.comm.im;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;

/**
 * 长连接接口
 * */
public interface IConnection {

	/**
	 * 获取当前状态
	 * */
	boolean get_current_state();

	/**
	 * 获取 sn, 发送消息的sn 从这里获取， 客户端不能随便填sn
	 * 
	 * @return 返回sn
	 * */
	long get_sn();

	/**
	 * @param account_info
	 * @param client_config
	 */
	void switch_account(AccountInfo account_info, ClientConfig client_config);

	/**
	 * 获取账号
	 * 
	 * @return 目前正在使用的账号
	 * */
	String get_account();

	/**
	 * 获取当前用户的JID， 如果登录过
	 * */
	String get_jid();

	/**
	 * 发送消息
	 * 
	 * @param account
	 *            接受者的账号
	 * @param account_type
	 *            账号类型
	 * @param msg_type
	 *            消息类型
	 * @param sn
	 *            不能随便填，需要通过调用 {@link get_sn}获取
	 * @param body
	 *            消息体
	 * @param business_id
	 * 
	 * @param timeout_ms
	 *            发送超时时间，单位毫秒
	 * @param expiration_sec
	 *            过期过期时间， 单位秒
	 * 
	 * @return 成功返回true, 否则false
	 * */
	boolean send_message(String receiver, int account_type, int msg_type, long sn, byte[] body, int timeout_ms, int expiration_sec);

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
	boolean send_service_message(int serviceId, long sn, byte[] body);

	/**
	 * 查询用户在线状态
	 * 
	 * @param users
	 * @param sn
	 * */
	boolean query_presence(String users[], long sn);

	/**
	 * 主动发起心跳包
	 * */
	void send_heartbeat();

	/**
	 * 设置心跳超时
	 * 
	 * @param heartbeat_timeout
	 *            超时， 单位毫秒, 最小不小于30000 (即30秒)
	 * */
	void set_heartbeat_timeout(int heartbeat_timeout);

	/**
	 * 获取心跳超时时间设置
	 * 
	 * @return 返回心跳时间，单位毫秒
	 * */
	int get_heartbeat_timeout();

	/**
	 * 健康检查，检查线程是否在运行
	 * 
	 * @return 如果正在运行返回true, 否则返回false, 可以重新调用工厂类生成对象
	 * */
	boolean health_check();

	/**
	 * 停止
	 * */
	void shutdown();

	/**
	 * 是否已经停止
	 * */
	boolean has_shutdown();

	/**
	 * 获取系统已经运行时间和服务器时间的差值, 计算实际服务器时间时可以用以下方法<br>
	 * SystemClock.elapsedRealtime() + 该差值
	 * 
	 * @return 返回该差值， 如果是-1表示获取时间失败
	 * */
	long get_server_time_diff();

 

	boolean get_message(String info_type, int[] ids, byte[] parameters);

}
