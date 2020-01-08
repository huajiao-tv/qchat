package com.huajiao.comm.im;

import java.io.IOException;
import java.io.InputStream;
import com.huajiao.comm.common.RC4;

class RC4InputStream {

	private InputStream m_in;
	private RC4 dec = null;

	public RC4InputStream(String key, InputStream in) {
		m_in = in;
		if (key != null && key.length() > 0) {
			dec = new RC4(key);
		}
	}

	public InputStream getInputStream() {
		return this.m_in;
	}

	/**
	 * 读取未解密的数据
	 * */
	public int read_raw_data(byte[] buffer) throws IOException {
		return read(buffer, 0, buffer.length, true);
	}

	/**
	 * 读取未解密的数据
	 * */
	public int read_raw(byte[] buffer, int offset, int length) throws IOException {
		return read(buffer, offset, length, true);
	}

	/***
	 * 确保读取指定长度的数据
	 * */
	public synchronized int read(byte[] buffer, int offset, int length, boolean read_raw_data) throws IOException {

		int newOffset = 0, bytesRead, remainingLen = length;
		byte newBuffer[] = new byte[length];

		do {

			bytesRead = m_in.read(newBuffer, newOffset, remainingLen);

			if (bytesRead <= 0) {
				// socket has been closed somehow
				break;
			}

			remainingLen -= bytesRead;
			newOffset += bytesRead;

		} while (remainingLen > 0);

		if (bytesRead > 0) {
			if (!read_raw_data) {
				if ( dec != null ) {
					newBuffer = dec.decry_RC4(newBuffer);
				}
			}
			System.arraycopy(newBuffer, 0, buffer, offset, length);
			return length - remainingLen;
		} else {
			return -1; // error indication
		}
	}
}
