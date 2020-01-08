/**
 * 
 */
package com.huajiao.comm.chatroomresults;


/**
 * 查询聊天室信息的结果
 */
public class QueryResult extends DetailedResult {
	
	private int _total_member_cnt;
	
	public QueryResult(long sn, int result, byte [] reason, int total_cnt, String roomid, String name, String[] members) {
		super(sn, result, reason, Result.PAYLOAD_QUERY_RESULT, roomid, name, members);
		_total_member_cnt = total_cnt;
	}
	
	/**
	 * @return 获取群成员总人数
	 */
	public int get_total_member_cnt() {
		return _total_member_cnt;
	}
}
