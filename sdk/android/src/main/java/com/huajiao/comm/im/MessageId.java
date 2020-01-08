package com.huajiao.comm.im;

class MessageId {
	public final static int Ack = 400000;
	
	public final static int ChatReq = 100002;
	public final static int ChatResp = 200002;
	
	public final static int GetInfoReq = 100004;
	public final static int GetInfoResp = 200004;
	
	public final static int LoginReq = 100001;
	public final static int LoginResp = 200001;
	
	public final static int LogoutReq = 100005;
	public final static int LogoutResp = 200005;
	
	public final static int NewMessageNotify = 300000;
	public final static int ReLoginNotify = 300001;
	
	/** 服务器故障，或者移机器， 需要等时间长一点再登录 */
	public final static int ReConnectNotify = 300002;
	
	
	public final static int BatchQueryPresenceReq= 100012;
	public final static int BatchQueryPresenceResp = 200012;
	
	
	public final static int InitLoginReq = 100009;
	public final static int InitLoginResp = 200009;
	
	public final static int QueryPeerMsgMaxIdReq = 100013;
	public final static int QueryPeerMsgMaxIdResp = 200013;
	
	public final static int Service_Req = 100011;
	public final static int Service_Resp = 200011;
	
	
	public final static int GetMultipleInfoReq = 100100;
	public final static int GetMultipleInfoResp = 200100;
	
	
}