/**
 * 
 */
package com.huajiao.comm.chatroomresults;

/**
 * 加入聊天室的结果
 */
public class JoinResult extends DetailedResult {
	
	private int _total_member_cnt;
	private byte _partnerdata[] = null;
	
	public JoinResult(long sn, int result, byte[] reason, byte[] partnerdata,int total_cnt, String _roomid, String _name, String[] members) {
		super(sn, result, reason, Result.PAYLOAD_JOIN_RESULT, _roomid, _name, members);
		_total_member_cnt = total_cnt;
		_partnerdata = partnerdata;
	}
	
	/**
	 * @return 获取群成员总人数
	 */
	public int get_total_member_cnt() {
		return _total_member_cnt;
	}
	
	/**
	 * 获取自定义数据
	 * */
	public byte[] get_partnerdata() {
		return _partnerdata;
	}
}
