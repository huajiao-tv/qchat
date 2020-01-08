package com.huajiao.comm.im;

/** 消息类型 **/
public class MessageType {

	/**  文本消息： 内容是UTF-8编码的文字 **/
	public final static int TEXT = 0;
	
	/**  特别的公共推送消息写在了单聊盒子里面 **/
	public final static int PUBLIC_MESSAGE = 100;
	
	/*** 电话信令 **/
	public final static int PHONE_CMD = 100;
		
	/*** 通讯录的好友注册了 **/
	public final static int CONTACT_REGISTERED_NOTIFICATION = 200;
	
	/*** 系统保留的消息， 上传日志请求 **/
	public final static int UPLOAD_LOG_REQ = 300;
	
	/*** 系统保留的消息， 上传日志响应包* */
	public final static int UPLOAD_LOG_RES = 301;
	
	/** 强制刷新云端配置 */
	public final static int REFRESH_CLOUD_CONFIG = 302;
	
}
