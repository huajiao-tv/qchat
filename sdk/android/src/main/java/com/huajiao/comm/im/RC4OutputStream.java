package com.huajiao.comm.im;

import java.io.IOException;
import java.io.OutputStream;

import com.huajiao.comm.common.RC4;
import com.huajiao.comm.common.Utils;

 class RC4OutputStream {

	private OutputStream m_out;
	private RC4 enc = null;

	public RC4OutputStream(String key, OutputStream stream) {
		m_out = stream;

		if (key != null && key.length() > 0) {
			enc = new RC4(key);
		}
	}

	public void write(byte[] buffer) throws IOException {
		if (null == buffer) {
			return;
		}

		if ( enc != null ){
			byte[] data = enc.encry_RC4_byte(buffer);
			m_out.write(data);
		}else {
			m_out.write(buffer);
		}
	}
	
	public RC4 getRC4(){
		return enc;
	}

	public void write(byte[] buffer, int offset, int count) throws IOException {
		byte[] newBuffer = new byte[count];
		System.arraycopy(buffer, offset, newBuffer, 0, count);
		write(newBuffer);
	}

	public void write(int paramInt) throws IOException {
		write(Utils.int_to_bytes(paramInt));
	}

	public OutputStream getOutputStream() {
		return this.m_out;
	}

}
