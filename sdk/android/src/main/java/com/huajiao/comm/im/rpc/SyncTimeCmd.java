package com.huajiao.comm.im.rpc;

public class SyncTimeCmd extends Cmd {

	/**
	 * 
	 */
	private static final long serialVersionUID = 2405281372936101235L;

	public SyncTimeCmd() {
		super(Cmd.CMD_GET_SERVER_TIME);
	}

}
