package com.huajiao.comm.im.rpc;

public class ShutdownCmd extends Cmd {

	/**
	 * shutdown
	 */
	private static final long serialVersionUID = 933754201365548612L;

	public ShutdownCmd() {
		super(CMD_SHUTDOWN);
	}
}
