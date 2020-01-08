package com.huajiao.comm.im;

/**
 * 协议和上层之间的接口
 * */
public interface INotify {

	/**
	 * 状态改变， 用于通知是否登录成功: <br>
	 * Connected:表明已经连接上<br>
	 * AuthFailed:表明鉴权失败客户端需要更新密码
	 * 
	 * @param old_state
	 *            旧状态
	 * @param new_state
	 *            新状态(目前的状态 )
	 * */
	void onStateChanged(int appid, ConnectionState old_state, ConnectionState new_state);

	/**
	 * 消息发送结果回执
	 * 
	 * @param result
	 *            消息结果 请参照 Error.MESSAGE_* 值
	 * @param sn
	 *            消息的SN， 通过SN比对标识对应的消息发送成功，或者失败
	 * @param sent_server_time
	 *            消息发送的服务器时间
	 * */
	void onMessageResult(int appid, int result, long sn, long sent_server_time);

	/**
	 * service消息
	 * 
	 * @param service_id
	 *            业务Id
	 * 
	 * @param result
	 *            消息结果 请参照 Error.MESSAGE_* 值
	 * @param sn
	 *            消息的SN， 通过SN比对标识对应的消息发送成功，或者失败
	 * @param data
	 *            service 数据
	 * */
	void onServiceMessageResult(int appid, int service_id, int result, long sn, byte[] data);

	/**
	 * 收到了新消息
	 * 
	 * @param sender
	 *            发送者的UID
	 * @param info_type
	 *            消息类型， 区分是单聊消息还是推送消息
	 * @param msg_id
	 *            : 消息编号, 按顺序递增， id由服务器确定
	 * @param msg_type
	 *            消息类型, 参照 MessageType类
	 * @param sn
	 *            序列号, 由发送端客户端决定
	 * @param time_sent
	 *            服务器消息入库EPOCH时间
	 * @param body
	 *            消息的二进制内容， 由msg_type标明统一的编解码方式
	 * */
	void onMessage(int appid, final String sender, final String info_type, int msg_type, long msg_id, long sn, long time_sent, final byte[] body, long latestMsgId);

	/**
	 * 查询在线 的响应
	 * 
	 * @param sn
	 * @param result
	 *            0 成功， 其他失败
	 * @param presences
	 *            按如下格式, 没3个一组表示一个用户的状态<br>
	 *            string userid<br>
	 *            string user_type<br>
	 *            string(int) status<br>
	 * <br>
	 *            status 说明<br>
	 *            0: 未注册; <br>
	 *            1: 已注册, offline, not reachable; <br>
	 *            2: registry, offline, reachable; <br>
	 *            3: registry, online, reachable<br>
	 * */
	void onPresenceUpdated(int appid, long sn, int result, String presences[]);

	/**
	 * 收到通知，该类通知不是持久保存在数据库的
	 * 
	 * @param info_type
	 *            通知类型
	 * @param content
	 *            内容
	 * @param id
	 *            消息id
	 * */
	void onNotification(int appid, String info_type, byte[] content, long id);
}
