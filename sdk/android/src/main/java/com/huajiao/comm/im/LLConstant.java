package com.huajiao.comm.im;

class LLConstant {

	/** SDK的版本 */
	public final static String SDK_VER = "20160518-1722";

	public final static String ACCOUNT_TYPE_JID = "jid";

	public final static String ACCOUNT_TYPE_PHONE = "phone";

	public final static int BUSINESS_WL_TIMEOUT = 50000;

	public static final int CONTINUOUS_FAILURES_TO_TRY_VIP = 2;

	public static final byte EVENT_AUTH_FAILED = 1;

	public static final byte EVENT_CONNECT = 2;

	public static final byte EVENT_CREDENTIAL_UPDATED = 3;

	public static final byte EVENT_DISCONNECT = 4;

	public static final byte EVENT_GET_MSG = 5;

	public static final byte EVENT_GOT_HEARTBEAT_ACK = 6;

	public static final byte EVENT_GOT_PACKET = 7;

	public static final byte EVENT_INET_AVAILABLE = 8;

	public static final byte EVENT_INET_UNAVAILABLE = 9;

	public static final byte EVENT_NETWORK_TYPE_CHANAGED = 10;

	public static final byte EVENT_SEND_HEARTBEAT = 11;

	public static final byte EVENT_SEND_MSG = 12;

	public static final byte EVENT_SOCK_CLOSED = 13;

	public static final byte EVENT_GET_STATE = 14;


	/*** 其他业务* */
	public static final int IM_GENERIC_BUSINESS_ID = 2;

	/** 电话业务 * */
	public static final int IM_PHONE_BUSINESS_ID = 1;

	/*** 聊天室消息 */
	public final static String INFO_TYPE_CHATROOM = "chatroom";

	/**
	 * IM消息
	 * */
	public final static String INFO_TYPE_IM = "im";

	/**
	 * 信令
	 * */
	public final static String INFO_TYPE_PEER = "peer";

	public final static String INFO_TYPE_PUBLIC = "public";

	/** 每分钟最多登录次数 */
	public final static int LOGIN_FREQ_LIMIT = 8;

	/** extra message server addresses if DNS request has been hijacked. */
	public static final String LVS_IP[] = new String[] {  };

	/** 用于重发频度控制， 单条消息最多重发一次 */
	public static final int MAX_SEND_COUNT = 2;

	/** 平台 */
	public final static String MOBILE_TYPE = "android";

	/** Current network is 1xRTT */
	public static final int NETWORK_TYPE_1xRTT = 7;
	/** Current network is CDMA: Either IS95A or IS95B */
	public static final int NETWORK_TYPE_CDMA = 4;
	/** Current network is EDGE */
	public static final int NETWORK_TYPE_EDGE = 2;
	/** Current network is eHRPD */
	public static final int NETWORK_TYPE_EHRPD = 14;
	/** Current network is EVDO revision 0 */
	public static final int NETWORK_TYPE_EVDO_0 = 5;
	/** Current network is EVDO revision A */
	public static final int NETWORK_TYPE_EVDO_A = 6;
	/** Current network is EVDO revision B */
	public static final int NETWORK_TYPE_EVDO_B = 12;
	/** Current network is GPRS */
	public static final int NETWORK_TYPE_GPRS = 1;
	/** Current network is HSDPA */
	public static final int NETWORK_TYPE_HSDPA = 8;
	/** Current network is HSPA */
	public static final int NETWORK_TYPE_HSPA = 10;
	/** Current network is HSPA+ */
	public static final int NETWORK_TYPE_HSPAP = 15;
	/** Current network is HSUPA */
	public static final int NETWORK_TYPE_HSUPA = 9;
	/** Current network is iDen */
	public static final int NETWORK_TYPE_IDEN = 11;
	/** Current network is LTE */
	public static final int NETWORK_TYPE_LTE = 13;
	/** Current network is UMTS */
	public static final int NETWORK_TYPE_UMTS = 3;
	/** Network type is unknown */
	public static final int NETWORK_TYPE_UNKNOWN = 0;

	/** 服务器地址 */
	public static final String OFFICIAL_SERVER = "qchat.server.com";

	public final static String DispatchServerUrl = "http://%s/get?%s";

	/** 端口 */
	public final static int PORT[] = new int[] { 80, 443 };

	/** 设置里面的KEY* */
	public final static String PREF_KEY = "ph_llc";

	/** 设置里面的KEY* */
	public final static String PREF_KEY_ID = "ph_llc";

	/**
	 * 支持的项目<br>
	 * 必须是两个英文字母
	 * */
	public final static String PROJECT = "AB";

	/** 长链接协议版本号* */
	public final static int PROTOCOL_VERSION = 1;

}
