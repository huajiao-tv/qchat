package com.huajiao.comm.im.rpc;

import com.huajiao.comm.common.AccountInfo;

/**
 * 查询状态
 * */
public class PresenceCmd extends Cmd {

	private static final long serialVersionUID = -3204802883351271831L;
	protected String _users;
	protected long _sn;
	protected int _account_type;
	protected int _business_id;

	/**
	 * @param users
	 * @param sn
	 * @param business_id
	 * @param account_type: AccountInfo.ACCOUNT_TYPE*
	 * @throws IllegalArgumentException
	 */
	public PresenceCmd(String users,  long sn,  int account_type) {
		super(Cmd.CMD_QUERY_PRESENCE);
		
		if (users == null || users.length() == 0) {
			throw new IllegalArgumentException();
		}
		
		if(account_type != AccountInfo.ACCOUNT_TYPE_JID && account_type != AccountInfo.ACCOUNT_TYPE_PHONE){
			throw new IllegalArgumentException("invalid account type");
		}
		
		this._users = users;
		this._sn = sn;
		 
		this._account_type = account_type;
	}

	public String get_users() {
		return _users;
	}

	public long get_sn() {
		return _sn;
	}
}
