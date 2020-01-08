package com.huajiao.comm.common;

public interface IUplink {
	boolean send_data(String receiver, byte[] body, long sn);
}
