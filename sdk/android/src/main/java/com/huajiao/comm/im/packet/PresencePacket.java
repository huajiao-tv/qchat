package com.huajiao.comm.im.packet;

public class PresencePacket extends Packet {

	/**
	 */
	private static final long serialVersionUID = 6651656532286705583L;

	@Override
	public int getAction() {
		return Packet.ACTION_PRESENCE_UPDATED;
	}

	public PresencePacket(long sn, int result, Object[] presences) {
		super();
		this._sn = sn;
		this._result = result;
		this._presences = presences;
	}

	protected long _sn;
	protected int _result;
	protected Object[] _presences;
	protected transient Presence[] _presences_ex;
	 

	public long get_sn() {
		return _sn;
	}

	public int get_result() {
		return _result;
	}

	/**
	 * <b>返回的状态获取状态</b> 按如下格式, 每6个属性表示一个用户的状态, 第一列代表实际的数据类型， 直接转换即可<br>
	 * <li>string userid: 用户id</li><br>
	 * <li>string user_type: 用户id的类型， phone说明是手机号， jid则表示是jid, 目前是phone</li><br>
	 * <li>Integer status: 表示用户的在线状态， 具体请看 后面的 status 说明</li><br>
	 * <li>string mobile type: 区分移动端和非移动端， 目前客户端不用关心</li><br>
	 * <li>Integer appid：用户可能在多个程序上注册， 表示这些属性是用户哪个appid</li><br>
	 * <li>Integer: end point version, 客户端版本</li> <br>
	 * <br>
	 * status 说明<br>
	 * 0: 未注册; <br>
	 * 1: 已注册, offline, not reachable; <br>
	 * 2: registry, offline, reachable; <br>
	 * 3: registry, online, reachable<br>
	 * <br>
	 * mobile type 说明<br>
	 * android<br>
	 * ios <br>
	 * epv 说明<br>
	 * 客户端版本号， 把4字节直接转换成int32即可
	 * 
	 **/
	public Object[] get_presences() {
		return _presences;
	}

	public Presence[] get_presences_ex() {
		
		if (_presences == null || _presences.length < 6) {
			return _presences_ex;
		}

		if (_presences_ex == null || _presences_ex.length == 0) {
			
			_presences_ex = new Presence[_presences.length / 6];
			int p = 0;
			
			for (int i = 0; i < _presences.length; i += 6) {
				
				int index = i;
				String user_id = (String) _presences[index++];
				String user_type = (String) _presences[index++];
				int status = (Integer) _presences[index++];
				String mobile_type = (String) _presences[index++];
				int appid = (Integer) _presences[index++];
				int epv = (Integer) _presences[index++];
				
				_presences_ex[p++] = new Presence(user_id, user_type, status, mobile_type, appid, epv);
			}
		}

		return _presences_ex;
	}
}
