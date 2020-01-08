package com.huajiao.comm.chatroom;

public interface IChatroomHelper {
	boolean getMessage(String info_type, int[] ids, byte[] parameters);
	long getServerTime();
}
