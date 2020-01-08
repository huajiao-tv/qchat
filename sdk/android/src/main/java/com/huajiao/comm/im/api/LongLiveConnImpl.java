package com.huajiao.comm.im.api;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;
import com.huajiao.comm.service.ImServiceBridge;

import android.content.Context;


/***
 * 长链接实现，LongLiveConnImpl和ImServiceBridge的差别：<br>
 * LongLiveConnImpl发送的消息的回馈会发送给IM service<br>
 * ImServiceBridge可以控制消息反馈发给phone 或者 IM service 
 * */
class LongLiveConnImpl implements ILongLiveConn {

	private ImServiceBridge _im_srv_bridge;
	
	public LongLiveConnImpl(Context context, AccountInfo accountInfo, ClientConfig clientConfig) {
		_im_srv_bridge = new ImServiceBridge(context, accountInfo, clientConfig);
	}

	@Override
	public long get_sn() {
		return _im_srv_bridge.get_sn();
	}

	@Override
	public boolean send_message(String receiver, int msgType, long sn, byte[] body, int timeoutMs) {
		return _im_srv_bridge.send_message(receiver, AccountInfo.ACCOUNT_TYPE_JID, msgType, sn, body, timeoutMs, 0);
	}

	@Override
	public boolean send_service_message(int serviceId, long sn, byte[] body) {
		return _im_srv_bridge.send_service_message(serviceId, sn, body);
	}

	@Override
	public boolean query_presence(String[] users, long sn) {
		return _im_srv_bridge.query_presence(users, sn, AccountInfo.ACCOUNT_TYPE_JID);
	}

	@Override
	public void set_heartbeat_timeout(int heartbeat_timeout) {
		 
	}

	@Override
	public boolean get_server_time() {
		return _im_srv_bridge.sync_time();
	}

	@Override
	public boolean send_message(String receiver, int msgType, long sn, byte[] body) {
		return _im_srv_bridge.send_message(receiver, AccountInfo.ACCOUNT_TYPE_JID, msgType, sn, body);
	}

	@Override
	public boolean send_message(String receiver, int msgType, long sn, byte[] body, int timeoutMs, int expireationSec) {
		return _im_srv_bridge.send_message(receiver, AccountInfo.ACCOUNT_TYPE_JID, msgType, sn, body, timeoutMs, expireationSec);
	}
 
	@Override
	public void switch_account(AccountInfo account_info, ClientConfig client_config) {		 
		_im_srv_bridge.switch_account(account_info, client_config );
	}

	@Override
	public boolean send_message(String receiver, int account_type, int msg_type, long sn, byte[] body, int timeoutMs, int expireationSec) {
		return _im_srv_bridge.send_message(receiver, account_type, msg_type, sn, body,  timeoutMs, expireationSec);
	}

	@Override
	public boolean query_presence(String[] users, long sn, int account_type) {
		return _im_srv_bridge.query_presence(users, sn, account_type);
	}

	@Override
	public boolean get_message(String info_type, int[] ids, byte[] parameters) {
		return _im_srv_bridge.get_message(info_type, ids, parameters);
	}

	@Override
	public boolean shutdown() {
		return _im_srv_bridge.shutdown();
	}

	@Override
	public boolean get_current_state() {
		return _im_srv_bridge.get_current_state();
	}

	@Override
	public void setOpenForeground(boolean open) {
		_im_srv_bridge.setOpenForeground(open);
	}
}
