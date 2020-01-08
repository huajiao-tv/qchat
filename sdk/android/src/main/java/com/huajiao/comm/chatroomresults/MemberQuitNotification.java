package com.huajiao.comm.chatroomresults;



/**
 * 成员退出的通知
 * */
public class MemberQuitNotification extends DetailedResult {

	private int _total_member_cnt;
	
	public MemberQuitNotification(long sn, int result, byte[] reason, int total_member_count, String roomid,  String roomname, String[] members) {
		super(sn, result, reason,  Result.PAYLOAD_MEMBER_QUIT, roomid, roomname, members);
		_total_member_cnt = total_member_count;
	}

	/**
	 * @return 获取群成员数
	 */
	public int get_total_member_cnt() {
		return _total_member_cnt;
	}
}
