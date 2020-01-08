package com.huajiao.comm.im;

import com.huajiao.comm.common.BuildFlag;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.net.ConnectivityManager;
import android.net.NetworkInfo;

public class NetworkProbe extends BroadcastReceiver {

	private static INetworkChanged _inetworkChanged = null;
	private static Object _lock = new Object();

	public static void registerCallback(INetworkChanged networkChanged) {
		if (null != networkChanged) {
			synchronized (_lock) {
				_inetworkChanged = networkChanged;
			}
		}
	}

	@Override
	public void onReceive(Context context, Intent intent) {
		if (BuildFlag.DEBUG){
			Logger.i("HJ-NetworkProbe", "onReceive");
		}
		
		checkNetworkConnected(context);
	}

	public void checkNetworkConnected(Context context) {

		int sub_type = -1;
		int net_type = -1;

		if (context == null) {
			return;
		}

		ConnectivityManager mConnectivityManager = (ConnectivityManager) context.getSystemService(Context.CONNECTIVITY_SERVICE);
		NetworkInfo mNetworkInfo = mConnectivityManager.getActiveNetworkInfo();
		
		if (mNetworkInfo != null) {
			
			sub_type = mNetworkInfo.getSubtype();
			net_type = mNetworkInfo.getType();
			synchronized (_lock) {
				if (_inetworkChanged != null) {
					_inetworkChanged.onNetworkChanged(true, net_type, sub_type);
				}
			}

		} else {
			synchronized (_lock) {
				if (_inetworkChanged != null) {
					_inetworkChanged.onNetworkChanged(false, net_type, sub_type);
				}
			}
		}
	}
}
