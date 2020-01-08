package com.huajiao.comm.im;

/** IP and port */
public class IPAddress {

	private String _ip;
	private int _port;

	public IPAddress(String _ip, int _port) {
		super();
		this._ip = _ip;
		this._port = _port;
	}

	public String get_ip() {
		return _ip;
	}

	public int get_port() {
		return _port;
	}

	public String set_ip(String ip) {
		return _ip = ip;
	}

	public int set_port(int port) {
		return _port = port;
	}
}