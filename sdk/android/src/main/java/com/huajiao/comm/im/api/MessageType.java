package com.huajiao.comm.im.api;

/**
 * 消息类型
 * */
public class MessageType {

	/**
	 * 文本消息： 内容是UTF-8编码的文字
	 * */
	public final static int TEXT = 0;

	/**
	 * 语音消息
	 * */
	public final static int VOICE = 1;

	/**
	 * 图片消息，body是UTF-8编码的json，格式为{" thumb_url ": "http://xxx", "url": "http://xxx" }
	 * */
	public final static int PICTURE = 2;

	/**
	 * 应用层消息送达确认 (接收方 发送 给 发生方)<br>
	 * 回执消息，body是UTF-8编码的sn串，sn为收到的消息的sn
	 * */
	public final static int MSG_ACK = 500;

	/**
	 * 电话信令
	 * */
	public final static int PHONE_CMD = 100;
	
	
	/**
	 * 保留的命令最小值, 这些type会被存在peer消息盒子里面
	 * */
	public final static int CMD_MIN = 100;
	
	/**  特别的公共推送消息写在了单聊盒子里面 **/
	public final static int PUBLIC_MESSAGE = 100;
	
	 
	
	/*** 系统保留的消息， 上传日志请求 **/
	public final static int UPLOAD_LOG_REQ = 300;
	
	/*** 系统保留的消息， 上传日志响应包* */
	public final static int UPLOAD_LOG_RES = 301;
	
	/** 强制刷新云端配置 */
	public final static int REFRESH_CLOUD_CONFIG = 302;
	
	/***
	 * 保留的命令最大值， 这些type会被存在peer消息盒子里面
	 * */
	public final static int CMD_MAX = 150;

	/**
	 * 通讯录的好友注册了
	 * */
	public final static int CONTACT_REGISTERED_NOTIFICATION = 200;

	/**
	 * 云名片更新通知
	 */
	public final static int CLOUDCARD_UPDATE_NOTIFICATION = 201;

	/**
	 * 手机号码变化通知
	 */
	public final static int CHANGE_PHONENUMBER_NOTIFICATION = 202;

	/**
	 * 添加阅后即焚之后，原有的消息类型不变，加一个标示位判断阅后即焚消息。<br>
	 * int com.huajiao.comm.im.api.MessageType.MASK_ EPHEMERAL=0x1<<15 <br>
	 * 如果type|MessageType.MASK_ EPHEMERAL != 0 即为阅后即焚消息。 type & (~ MessageType.MASK_ EPHEMERAL) 即为消息类型（文本、语音、图片）。
	 * */
	public final static int MASK_EPHEMERAL = 0x1 << 15;

}
