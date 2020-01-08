package com.huajiao.comm.im;

public interface INetworkChanged {

	/**
	 * @param available: 网络是否可用
	 * */
	void onNetworkChanged(boolean available, int net_type, int sub_type);
}
