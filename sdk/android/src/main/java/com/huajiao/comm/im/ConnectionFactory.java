package com.huajiao.comm.im;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;

import android.content.Context;

/**
 * 连接工厂, 调用getConnection获取连接实例
 * */
public class ConnectionFactory {

	private static final ConnectionFactory _instance = new ConnectionFactory();
	private static IConnection _connection = null;
	private static final Object _lock = new Object();

	/**
	 * 获取工厂单实例
	 * */
	public static ConnectionFactory getInstance() {
		return _instance;
	}

	private ConnectionFactory() {

	}

	/**
	 * 获取连接实例<br>
	 * 注意：<br>
	 * 使用新链接时， 旧的会被关闭，同时只能一个链接可用 <br>
	 * 
	 * @param context
	 *            Android context
	 * @param account
	 *            账号
	 * @param password
	 *            密码
	 * @param device_id
	 *            设备唯一号, 如果为null, SDK生成
	 * @param clientConfig
	 *            客户端配置
	 * 
	 * @param notify
	 *            通知接口实现
	 * 
	 * @throws IllegalArgumentException
	 * 
	 * */
	public IConnection getConnection(Context context, AccountInfo account_info, ClientConfig clientConfig, IMCallback notify) {

		if (context == null) {
			throw new IllegalArgumentException("1st arg");
		}

		if (account_info == null ) {
			throw new IllegalArgumentException("2nd arg");
		}
 
		if (clientConfig == null) {
			throw new IllegalArgumentException("3rd arg");
		}
		
		if(notify == null){
			throw new IllegalArgumentException("4th arg");
		}

		synchronized (_lock) {

			boolean new_instance_created = false;

			if (_connection == null || !_connection.health_check()) {

				if (_connection != null) { // 关闭旧连接, 允许重复关闭
					_connection.shutdown();
				}

				_connection = new ClientConnection(context, account_info, clientConfig, notify);
				new_instance_created = true;
			}

			// 如果连接已经存在直接切换账号
			if (!new_instance_created) {
				_connection.switch_account(account_info, clientConfig);
			}
		}

		return _connection;
	}
}
