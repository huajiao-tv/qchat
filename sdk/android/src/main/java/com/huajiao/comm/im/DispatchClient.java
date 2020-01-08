package com.huajiao.comm.im;

import java.security.KeyFactory;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.PublicKey;
import java.security.spec.X509EncodedKeySpec;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Locale;

import android.util.Base64;
import android.util.Log;

import com.huajiao.comm.common.BuildFlag;
import com.huajiao.comm.common.ClientConfig;
import com.huajiao.comm.common.HttpUtils;

import org.json.JSONObject;

import javax.crypto.Cipher;

import static android.R.attr.key;
import static android.R.id.message;

/** Dispatch 客户端实现 **/
class DispatchClient {

	private final static String TAG = "DISPC";

	private ArrayList<IPAddress> _server_list = null;

	private static DispatchClient _instance = new DispatchClient();

	private int _server_index = 0;

	private GetResult _state = GetResult.INIT;

	private static String pubkey = "";

	enum GetResult{
		INIT,
		SUCCESS,
		FAIL,
		PASS,
		QUERYDONE
	}

	static DispatchClient getInstance() {
		return _instance;
	}

	private DispatchClient() {

	}

	synchronized void reset(){
		_state = GetResult.INIT;
		_server_index = 0;
	}

	/**
	 * @param clientConfig clientConfig
	 * @param dispatch_server dispatch server ip or domain
	 * @return Qchat Server address
	 */
	synchronized GetResult getDispatchResultServer(ClientConfig clientConfig,
														  String dispatch_server,
														  int outindex,
														  String uid,
														  IPAddress resultServer) {

		//预处理
		switch (_state) {
			case INIT:
				boolean ret = queryDispatchServer(clientConfig, dispatch_server, uid);
				if ( !ret ) {
					Logger.e(TAG, "queryDispatchServer fail");
				}
				_state = GetResult.QUERYDONE;
				break;
			case PASS:
				if ( outindex < 2 ){ //重新开始
					_state = GetResult.INIT;
				}
				break;
			default:
				break;
		}

		//正式处理
		switch (_state) {
			case INIT:
				boolean ret = queryDispatchServer(clientConfig, dispatch_server, uid);
				if ( !ret ) {
					Logger.e(TAG, "queryDispatchServer fail");
				}
			case QUERYDONE:
			case SUCCESS:
				if ( _server_list!=null && _server_list.size()>0 ){
					if (outindex > 1){
						++_server_index;
					}

					if ( _server_index >= _server_list.size() ){
						_server_index = 0;
						_state = GetResult.FAIL;
					} else {
						IPAddress server = _server_list.get(_server_index);
						resultServer.set_ip(server.get_ip());
						resultServer.set_port(server.get_port());
						_state = GetResult.SUCCESS;
					}
				}else{
					_state = GetResult.FAIL;
				}
				break;
			case FAIL:
				_state = GetResult.PASS;
				break;
			default:
				break;
		}

		return _state;
	}

	private static String getString(byte[] b){
		StringBuilder md5StrBuff = new StringBuilder();

		for (int i = 0; i < b.length; i++)
		{
			if (Integer.toHexString(0xFF & b[i]).length() == 1)
				md5StrBuff.append("0").append(Integer.toHexString(0xFF & b[i]));
			else
				md5StrBuff.append(Integer.toHexString(0xFF & b[i]));
		}

		return md5StrBuff.toString();
	}

	private static byte[] getMD5(String val) throws NoSuchAlgorithmException {
		MessageDigest md5 = MessageDigest.getInstance("MD5");
		md5.reset();
		md5.update(val.getBytes());
		byte[] m = md5.digest();//加密
		return m; //getString(m);
	}

	private static PublicKey getPublicKey(String key) throws Exception
	{
		byte[] keyBytes = Base64.decode(key,Base64.DEFAULT);

		X509EncodedKeySpec keySpec = new X509EncodedKeySpec(keyBytes);
		KeyFactory keyFactory = KeyFactory.getInstance("RSA");
		PublicKey publicKey = keyFactory.generatePublic(keySpec);
		return publicKey;
	}

	private boolean verify(String data, String sign) {

		if ( data.equals("") ){
			return false;
		}

		if ( sign.equals("") ){
			return true;
		}

		try {
			byte[] md5 = getMD5(data);
			PublicKey publicKey = getPublicKey(pubkey);
			// decrypts the message
			Cipher cipher = Cipher.getInstance("RSA/ECB/PKCS1Padding");
			cipher.init(Cipher.DECRYPT_MODE, publicKey);
			byte[] dectyptedText = cipher.doFinal(Base64.decode(sign, Base64.DEFAULT));

			if ( dectyptedText.length > md5.length) {
				String mymd5 = getString(md5);
				String yourmd5 = getString(Arrays.copyOfRange(dectyptedText, dectyptedText.length - md5.length, dectyptedText.length));
				Logger.i(TAG, "dispatch diff verify:"+mymd5+"|"+yourmd5);

				return Arrays.equals(md5, Arrays.copyOfRange(dectyptedText, dectyptedText.length - md5.length, dectyptedText.length));
			}else{
				String mymd5 = getString(md5);
				String yourmd5 = getString(dectyptedText);
				Logger.i(TAG, "dispatch same verify:"+mymd5+"|"+yourmd5);

				return Arrays.equals(md5, dectyptedText);
			}
		}catch (Exception e){
			return false;
		}
	}

	/** Query dispatch server for MSG address */
	private boolean queryDispatchServer(ClientConfig clientConfig, String dispatch_server, String uid) {

		if (BuildFlag.DEBUG) {
			Logger.i(TAG, "querying dispatch server["+dispatch_server+"]...");
		}

		boolean result = false;
		if (clientConfig == null || dispatch_server == null ) {
			Log.e(TAG, "invalid arguments");
			return false;
		}

		String paras = String.format(Locale.US, "mobiletype=android&uid=%s", uid);
		String request_url = String.format(Locale.US, LLConstant.DispatchServerUrl, dispatch_server, paras);

		do {
			byte[] body = HttpUtils.get(request_url, null, 1000, 1000);
			if (body == null) {
				Logger.e(TAG, "dispatch response is null");
				break;
			}

			try {
				String response = new String(body);
				Logger.i(TAG, "dispatch response: " + response);

				JSONObject jsonObj = new JSONObject(response);
				String sign = jsonObj.optString("sign");
				String data = jsonObj.optString("data");

				if ( verify(data, sign) ) {
					String servers[] = data.split(";");
					_server_list = new ArrayList<IPAddress>();
					for (String server : servers) {
						String IpPort[] = server.split(":");
						if (IpPort.length > 1) {
							IPAddress ipaddr = new IPAddress(IpPort[0], Integer.parseInt(IpPort[1]));
							_server_list.add(ipaddr);
						} else if (IpPort.length > 0) {
							IPAddress ipaddr = new IPAddress(IpPort[0], 443);
							_server_list.add(ipaddr);
						}
					}

					if (_server_list.size() > 0) {
						result = true;
					}
				}else{
					Logger.e(TAG, "dispatch verify fail");
				}
			} catch (Exception e) {
				Logger.e(TAG, "parse dispatch exception: " + e.getMessage());
			}
		} while (false);

		return result;
	}
}
