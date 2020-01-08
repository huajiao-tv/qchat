package com.huajiao.comm.common;

import android.os.SystemClock;

public class TimeCostCalculator {
	
	private long _start = SystemClock.elapsedRealtime();
	
	public TimeCostCalculator(){
		
	}
	
	public long getCost(){
		return  SystemClock.elapsedRealtime() - _start;
	}	
}
