package com.huajiao.comm.im.rpc;

import com.huajiao.comm.common.AccountInfo;
import com.huajiao.comm.common.ClientConfig;

/**
 * 切换账号
 * */
public class AccountCmd extends Cmd {

	private static final long serialVersionUID = 3801905735476487276L;

	private AccountInfo _account_info;
	private ClientConfig _client_config;
	
	public AccountCmd(AccountInfo account_info, ClientConfig client_config) {
		super(Cmd.CMD_SWITCH_ACCOUNT);		
		
		if(account_info == null || client_config == null){
			throw new IllegalArgumentException("account_info|client_config");
		}
		
		this._account_info = account_info;
		this._client_config = client_config;
	}


	public AccountInfo get_account_info() {
		return _account_info;
	}

	public ClientConfig get_client_config() {
		return _client_config;
	}
}
