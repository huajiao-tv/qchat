package com.huajiao.comm.chatroomresults;

import java.util.ArrayList;
import java.util.List;

class DetailedResult extends Result {

	private String _roomid;
	private String _name;
	private List<UserInfo> _members = new ArrayList<UserInfo>();
	
	/**
	 * @param sn
	 * @param result
	 * @param _roomid
	 * @param _name
	 * @param _members
	 */
	public DetailedResult(long sn, int result, byte [] reason, int payload, String _roomid, String _name, List<UserInfo> members) {
		super(sn, result, payload, reason);
		this._roomid = _roomid;
		this._name = _name;
		
		if(members != null){
			for(UserInfo m : members){
				_members.add(m);
			}
		}
	}

	public DetailedResult(long sn, int result, byte [] reason, int payload, String _roomid, String _name, String[] members) {
		super(sn, result, payload, reason);
		this._roomid = _roomid;
		this._name = _name;
		
		if(members != null){
			for(String m : members){
				_members.add(new UserInfo(m));
			}
		}
	}
	
	/**
	 * @return 聊天室id
	 */
	public String get_roomid() {
		return _roomid;
	}
	/**
	 * @return 聊天室名字
	 */
	public String get_name() {
		return _name;
	}
	/**
	 * @return 用户信息列表
	 */
	public List<UserInfo> get_members() {
		return _members;
	}
}
