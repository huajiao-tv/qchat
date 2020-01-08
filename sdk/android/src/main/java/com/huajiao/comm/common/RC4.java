package com.huajiao.comm.common;

/**
 * RC4编码实现
 * */
public class RC4 {
	final byte[] key;
	final byte[] keyForUse;

	public RC4(String strKey) {
		this.key = initKey(strKey);
		keyForUse = new byte[this.key.length];
	}

	private void copyKeyForUse() {
		System.arraycopy(key, 0, keyForUse, 0, key.length);
	}

	
	public void crypt(byte[] data, int offset, int length){
		decry_RC4(data, offset, length);
	}
	
	public byte[] decry_RC4(byte[] data, int offset, int length) {
		if (data == null) {
			return null;
		}

		return RC4Base(data, offset, length);
	}

	public byte[] decry_RC4(byte[] data) {
		if (data == null) {
			return null;
		}
		return RC4Base(data);
	}

	/**
	 * ����ʮ���������
	 */
	public String decry_RC4(String data) {
		if (data == null) {
			return null;
		}
		return new String(RC4Base(HexString2Bytes(data)));
	}

	 
	
	public byte[] encry_RC4_byte(String data) {
		if (data == null) {
			return null;
		}
		byte b_data[] = data.getBytes();
		return RC4Base(b_data);
	}

	public byte[] encry_RC4_byte(byte[] data) {
		if (data == null) {
			return null;
		}
		byte b_data[] = data;
		return RC4Base(b_data);
	}

	/**
	 * ����Ϊʮ���������
	 */
	public String encry_RC4_string(String data) {
		if (data == null) {
			return null;
		}
		return toHexString(asString(encry_RC4_byte(data)));
	}

	public static String asString(byte[] buf) {
		StringBuffer strbuf = new StringBuffer(buf.length);
		for (int i = 0; i < buf.length; i++) {
			strbuf.append((char) buf[i]);
		}
		return strbuf.toString();
	}

	private byte[] initKey(String aKey) {
		byte[] b_key = aKey.getBytes();
		byte state[] = new byte[256];

		for (int i = 0; i < 256; i++) {
			state[i] = (byte) i;
		}
		int index1 = 0;
		int index2 = 0;
		if (b_key == null || b_key.length == 0) {
			return null;
		}
		for (int i = 0; i < 256; i++) {
			index2 = ((b_key[index1] & 0xff) + (state[i] & 0xff) + index2) & 0xff;
			byte tmp = state[i];
			state[i] = state[index2];
			state[index2] = tmp;
			index1 = (index1 + 1) % b_key.length;
		}
		return state;
	}

	public static String toHexString(String s) {
		String str = "";
		for (int i = 0; i < s.length(); i++) {
			int ch = (int) s.charAt(i);
			String s4 = Integer.toHexString(ch & 0xFF);
			if (s4.length() == 1) {
				s4 = '0' + s4;
			}
			str = str + s4;
		}
		return str;// 0x��ʾʮ�����
	}

	public static String toHexString(byte[] data) {
		String str = "";
		for (int i = 0; i < data.length; i++) {
			int ch = data[i];
			String s4 = Integer.toHexString(ch & 0xFF);
			if (s4.length() == 1) {
				s4 = '0' + s4;
			}
			str = str + s4;
		}
		return str;// 0x��ʾʮ�����
	}

	public static byte[] HexString2Bytes(String src) {
		int size = src.length();
		byte[] ret = new byte[size / 2];
		byte[] tmp = src.getBytes();
		for (int i = 0; i < size / 2; i++) {
			ret[i] = uniteBytes(tmp[i * 2], tmp[i * 2 + 1]);
		}
		return ret;
	}

	private static byte uniteBytes(byte src0, byte src1) {
		char _b0 = (char) Byte.decode("0x" + new String(new byte[] { src0 })).byteValue();
		_b0 = (char) (_b0 << 4);
		char _b1 = (char) Byte.decode("0x" + new String(new byte[] { src1 })).byteValue();
		byte ret = (byte) (_b0 ^ _b1);
		return ret;
	}

	private byte[] RC4Base(byte[] input) {
		return RC4Base(input, 0, input.length);
	}

	/**
	 * Note: input value will be modified !!!
	 * */
	private byte[] RC4Base(byte[] input, int offset, int length) {
		int x = 0;
		int y = 0;
		int xorIndex;
		// byte[] result = new byte[length];
		byte[] result = input;

		copyKeyForUse();
		for (int i = offset; i < offset + length; i++) {
			x = (x + 1) & 0xff;
			y = ((keyForUse[x] & 0xff) + y) & 0xff;
			byte tmp = keyForUse[x];
			keyForUse[x] = keyForUse[y];
			keyForUse[y] = tmp;
			xorIndex = ((keyForUse[x] & 0xff) + (keyForUse[y] & 0xff)) & 0xff;
			result[i] = (byte) (input[i] ^ keyForUse[xorIndex]);
		}
		
		return result;
	}
}
