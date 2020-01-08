package com.huajiao.comm.common;

import java.io.Serializable;
import android.text.TextUtils;

/**
 * 账号信息
 * */
public class AccountInfo implements Serializable {
	/**
	 * 
	 */
	private static final long serialVersionUID = -389425043421477569L;

	/**
	 * 账号是手机号
	 * */
	public static final int ACCOUNT_TYPE_PHONE = 1;

	/**
	 * 账号是JID
	 * */
	public static final int ACCOUNT_TYPE_JID = 2;

	/**
	 * 账号是手机号字符串形式
	 * */
	public final static String ACCOUNT_TYPE_PHONE_STRING = "phone";

	/**
	 * 账号是Jid字符串形式
	 * */
	public final static String ACCOUNT_TYPE_JID_STRING = "jid";

	public final static int MAX_STRING_LEN = 512;

	private String _account;
	private String _password;
	private String _device_id;
	private String _signature;
	private int _account_type;

	/**
	 * 构建账号信息
	 * 
	 * @param account
	 *            账号信息
	 * @param password
	 *            对应的密码
	 * @param device_id
	 *            设备唯一号， 可以为空
	 * @param signature
	 *            签名
	 * @param account_type
	 *            账号类型 {@link ACCOUNT_TYPE_PHONE} 或 {@link ACCOUNT_TYPE_JID}
	 * @throws IllegalArgumentException
	 *             account 或 password 为空或null
	 */
	public AccountInfo(String account, String password, String device_id, String signature, int account_type) {
		super();

		if (account == null || account.length() == 0 || password == null || password.length() == 0) {
			throw new IllegalArgumentException();
		}

		if (account.length() > MAX_STRING_LEN || password.length() > MAX_STRING_LEN) {
			throw new IllegalArgumentException("a or j length exceeds limit");
		}

		if (device_id.length() > MAX_STRING_LEN) {
			throw new IllegalArgumentException("d length exceeds limit");
		}

		if (signature != null && signature.length() > MAX_STRING_LEN) {
			throw new IllegalArgumentException("s length exceeds limit");
		}

		this._account = account;
		this._password = password;
		this._device_id = device_id;
		this._account_type = account_type;
		this._signature = signature;
	}

	/**
	 * 构建账号信息
	 * 
	 * @param account
	 *            账号信息
	 * @param password
	 *            对应的密码
	 * @param device_id
	 *            设备唯一号， 可以为空
	 * @param signature
	 *            签名
	 */
	public AccountInfo(String account, String password, String device_id, String signature) {
		this(account, password, device_id, signature, ACCOUNT_TYPE_JID);
	}

	public AccountInfo(String jid, String password) {
		if (TextUtils.isEmpty(jid)) {
			throw new IllegalArgumentException("j cann't be empty!!!");
		}
		if (TextUtils.isEmpty(password)) {
			throw new IllegalArgumentException("p cann't be empty!!!");
		}

		if (jid.length() > MAX_STRING_LEN || password.length() > MAX_STRING_LEN) {
			throw new IllegalArgumentException("p or j length exceeds limit");
		}

		_account = jid;
		_password = password;
	}

	public String get_jid() {
		return _account;
	}

	public String get_account() {
		return _account;
	}

	public String get_password() {
		return _password;
	}

	public String get_device_id() {
		return _device_id;
	}

	public int get_account_type() {
		return _account_type;
	}

	/**
	 * 获取账号字符串形式
	 * 
	 * @return 账号类型的字符形式， 如果不支持， 返回null
	 */
	public String get_account_type_string() {
		return get_account_type_string(_account_type);
	}

	public String get_signature() {
		return _signature;
	}

	/**
	 * 获取账号字符串形式
	 * 
	 * @param account_type
	 *            账号类型
	 * @return 账号类型的字符形式， 如果不支持， 返回null
	 */
	public static String get_account_type_string(int account_type) {
		if (account_type == ACCOUNT_TYPE_JID) {
			return ACCOUNT_TYPE_JID_STRING;
		} else if (account_type == ACCOUNT_TYPE_PHONE) {
			return ACCOUNT_TYPE_PHONE_STRING;
		} else {
			return null;
		}
	}

	boolean compareString(String a, String b) {
		if (a == null && b == null) {
			return true;
		}
		if (a != null && b != null) {
			return a.equals(b);
		}
		return false;
	}

	@Override
	public int hashCode() {

		int account_len = _account == null ? 0 : _account.length();
		int password_len = _password == null ? 0 : _password.length();
		int device_id_len = _device_id == null ? 0 : _device_id.length();
		int sig_len = _signature == null ? 0 : _signature.length();

		int hash = account_len + password_len + device_id_len + sig_len + _account_type;

		return hash;
	}

	@Override
	public boolean equals(Object o) {
		if (null == o) {
			return false;
		}

		if (o instanceof AccountInfo) {

			AccountInfo acc = (AccountInfo) o;
			if (!compareString(_account, acc._account) || !compareString(_password, acc._password) || !compareString(_device_id, acc._device_id)
					|| !compareString(_signature, acc._signature)) {
				return false;
			}

			if (this._account_type != acc._account_type) {
				return false;
			}

			return true;
		}

		return false;
	}
}
