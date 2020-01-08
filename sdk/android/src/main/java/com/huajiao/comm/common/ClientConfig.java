package com.huajiao.comm.common;

import java.io.Serializable;

/**
 * 客户端配置包含:<br>
 * Application Id: <br>
 * DefaultKey:<br>
 * Server address:<br>
 * */
public class ClientConfig implements Serializable {
	
	

	/**
	 * 
	 */
	private static final long serialVersionUID = 4057067626811028698L;
	
	public final static int MAX_STRING_LEN = 512;
	
	private int _appid;
	private int _clientVersion;
	private String _defaultKey;
	private String _server;
	private String _im_business_service;
	private int _port=80;

	/**
	 * @param appid
	 *            应用ID
	 * @param clientVersion
	 *            SDK协议版本
	 * @param defaultKey
	 *            默认加密密码
	 * @param server
	 *            服务器地址
	 * @param im_business_service
	 *            用于接受推送通知的service的完整名字
	 *            
	 * @throws IllegalArgumentException
	 */

	public ClientConfig(int appid, int clientVersion, String defaultKey, String server, String im_business_service) throws IllegalArgumentException {

		if (defaultKey == null || defaultKey.length() == 0 || server == null || server.length() == 0 || im_business_service == null
				|| im_business_service.length() == 0) {
			throw new IllegalArgumentException("C-Conf invalid arguments.");
		}

		
		if(defaultKey.length() > MAX_STRING_LEN || server.length() > MAX_STRING_LEN || im_business_service.length() > MAX_STRING_LEN){
			throw new IllegalArgumentException("some argumens exceed length limit.");
		}
		
		_clientVersion = clientVersion;
		_appid = appid;
		_defaultKey = defaultKey;
		_server = server;
		_im_business_service = im_business_service;
	}

    public ClientConfig(int appid, int clientVersion, String defaultKey, String server, String im_business_service, int port) throws IllegalArgumentException {
        this(appid,clientVersion,defaultKey,server,im_business_service);
        _port = port;
    }

    public String getDefaultKey() {
		return _defaultKey;
	}

	public String getServer() {
		return _server;
	}

	public int getPort() {
		return _port;
	}

	public int getAppId() {
		return _appid;
	}

	public String get_im_business_service() {
		return _im_business_service;
	}

	/**
	 * @return the _clientVersion
	 */
	public int getClientVersion() {
		return _clientVersion;
	}
}
