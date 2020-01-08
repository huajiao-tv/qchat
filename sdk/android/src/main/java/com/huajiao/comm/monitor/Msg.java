package com.huajiao.comm.monitor;

import android.os.SystemClock;

public class Msg {
	
	public int id = -1;
	public int expected_recv_ts = 0;
	
	public void set_id(int id){
		this.id = id;
		if(id != -1){
			expected_recv_ts = (int) (SystemClock.elapsedRealtime() / 1000);
		}
	}
	public int get_duration_sec(int now){
		return  now - expected_recv_ts;
	}
	public int get_duration_ms(int now){
		return  (now - expected_recv_ts) * 1000;
	}
}
