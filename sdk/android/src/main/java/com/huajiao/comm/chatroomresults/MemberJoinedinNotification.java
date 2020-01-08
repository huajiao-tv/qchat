package com.huajiao.comm.chatroomresults;

import java.util.List;

 

/**
 * 成员增加的通知
 * */
public class MemberJoinedinNotification extends DetailedResult {

	private int _total_member_cnt;

	public MemberJoinedinNotification(long sn, int result, byte[] reason,  int total_member_count, String roomid, String roomname, List<UserInfo> members) {
		super(sn, result, reason, Result.PAYLOAD_MEMBER_JOINED_IN, roomid, roomname, members);
		_total_member_cnt = total_member_count;
		
	}

	/**
	 * @return 获取群成员总人数
	 */
	public int get_total_member_cnt() {
		return _total_member_cnt;
	}
}
