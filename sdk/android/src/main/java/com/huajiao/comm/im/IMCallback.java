package com.huajiao.comm.im;

import com.huajiao.comm.im.packet.CurrentStatePacket;
import com.huajiao.comm.im.packet.MsgPacket;
import com.huajiao.comm.im.packet.MsgResultPacket;
import com.huajiao.comm.im.packet.NotificationPacket;
import com.huajiao.comm.im.packet.PresencePacket;
import com.huajiao.comm.im.packet.SrvMsgPacket;
import com.huajiao.comm.im.packet.StateChangedPacket;

/**
 * 协议和上层之间的接口
 * */
public interface IMCallback {

	 
	void onStateChanged(StateChangedPacket packet);

 
	void onMessageResult(MsgResultPacket packet);

 
	void onServiceMessageResult(SrvMsgPacket packet);

 
	void onMessage(MsgPacket packet);

	 
	void onPresenceUpdated(PresencePacket packet);
 
	
	void onNotification(NotificationPacket packet);
	
	
	void onCurrentState(CurrentStatePacket packet);
}
