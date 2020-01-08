package com.huajiao.comm.chatroomresults;

public class QuitResult extends Result {
	public QuitResult(long sn, int result, byte [] reason) {
		super(sn, result, Result.PAYLOAD_QUIT_RESULT, reason);
	}
}
