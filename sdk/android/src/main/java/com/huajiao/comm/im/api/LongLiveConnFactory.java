package com.huajiao.comm.im.api;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;

import android.content.Context;

public class LongLiveConnFactory {

	private static ILongLiveConn _connection = null;
	
	private static final Object _lock = "_lock";

	/**
	 * 创建连接实例，如果已经存在则切换账号<br>
	 * 注意：<br>
	 * 使用新链接时， 旧的会被关闭，同时只能一个链接可用 <br>
	 * 
	 * @param context
	 *            Android context
	 * @param account_info
	 *            账号线相关信息
	 * @throws IllegalArgumentException
	 * 
	 * */
	public static ILongLiveConn create(Context context, AccountInfo account_info, ClientConfig client_config) {

		if (context == null || account_info == null || client_config == null) {
			throw new IllegalArgumentException("invalid arguments detected!!!");
		}

		synchronized (_lock) {
			boolean new_instance_created = false;

			if (_connection == null) {
				_connection = new LongLiveConnImpl(context, account_info, client_config);
				new_instance_created = true;
			}

			// 如果连接已经存在直接切换账号
			if (!new_instance_created) {
				_connection.switch_account(account_info, client_config);
			}
		}

		return _connection;
	}

	/**
	 * 获取连接实例，如果已经存在则返回，否则创建<br>
	 * @throws IllegalArgumentException
	 *
	 * */
	public static ILongLiveConn getDefaultConn() {
		return _connection;
	}
}
