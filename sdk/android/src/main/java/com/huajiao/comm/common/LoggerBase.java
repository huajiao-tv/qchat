package com.huajiao.comm.common;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.util.Calendar;
import java.util.Date;
import java.util.LinkedList;
import java.util.List;
import java.util.Locale;
import java.util.zip.GZIPOutputStream;

import android.annotation.SuppressLint;
import android.os.Build;
import android.os.Environment;
import android.util.Log;

/** 通用日志类* */
public class LoggerBase {

	/** 这类文件前缀将被上传 */
	private static final String LLC_LOG_PREFIX = "BG";

	private static final String CR_LOG_PREFIX = "CR";

	private static final int DAYS_TO_KEEP_LOG = 3;

	private static final int LOG_PACKET_SIZE = 20480;

	private static final int FILE_BUFFER_SIZE = 128;

	/** 单个文件最大大小 */
	private static final int MAX_LOG_SIZE_ALLOWED = 512000;

	private static final String TAG = "Logger";

	private static List<LoggerBase> _loggers = new LinkedList<LoggerBase>();

	private BufferedOutputStream _stream = null;

	private Date _log_date = null;

	private final Object _lock = new Object();

	protected static String _root_folder = Environment.getExternalStorageDirectory().getPath() + "/sdk";

	private String _cur_folder = _root_folder;

	private String _project_name;

	private String _cur_uid = "00000000";

	private boolean _file_name_has_uid = false;

	public void setCurUid(String uid) {
		synchronized (_lock) {
			_cur_uid = uid;
			closeStream();
		}
	}

	private void closeStream() {
		if (_stream != null) {
			try {
				_stream.flush();
				_stream.close();
			} catch (Exception e) {

			}
			_stream = null;
		}
	}

	/**
	 * 设置文件路径
	 *
	 * @param 日志文件根目录
	 * */
	public static void setDebugFilePath(String rootFolder) {

		if (rootFolder != null && !_root_folder.equals(rootFolder)) {
			_root_folder = rootFolder;
		}
	}
	public static String getDebugFilePath(){
		return _root_folder;
	}

	@Override
	protected void finalize() {
		synchronized (_lock) {
			closeStream();
		}
	}

	protected LoggerBase(String projectName) {
		_project_name = projectName;
		_loggers.add(this);
	}

	@SuppressLint({"NewApi", "SetWorldWritable"})
	@SuppressWarnings("deprecation")
	public void log(String tag, String content) {

		if (!FeatureSwitch.isLogOn() || tag == null || content == null) {
			return;
		}

		synchronized (_lock) {

			int y, m, d, h, mi, s, ms;
			Calendar cal = Calendar.getInstance();

			y = cal.get(Calendar.YEAR);
			m = cal.get(Calendar.MONTH) + 1;
			d = cal.get(Calendar.DATE);
			h = cal.get(Calendar.HOUR_OF_DAY);
			mi = cal.get(Calendar.MINUTE);
			s = cal.get(Calendar.SECOND);
			ms = cal.get(Calendar.MILLISECOND);

			Date now = new Date(y - 1900, m - 1, d);

			// create file
			if (!_cur_folder.equals(_root_folder) || _stream == null || now.compareTo(_log_date) != 0) {

				closeStream();

				_log_date = now;

				File dir = new File(_root_folder);
				try {
					if (!dir.exists()) {
						dir.mkdir();
						dir.setReadable(true);
						dir.setWritable(true, false);
					}

					// Remove old files
					File files[] = dir.listFiles();

					if (files != null && files.length > 0) {
						for (File f : files) {
							// 不移除其他项目的日志, 或者其他文件
							if(_file_name_has_uid) {
								// 文件名带uid(uid可能不止8个字符，所以文件名长度不一定等于26)
								if(f.getName().length() < 26 || f.getName().indexOf(_project_name) != 0) {
									continue;
								}
							} else {
								// 文件名不带uid
								if (f.getName().length() != 17 || f.getName().indexOf(_project_name) != 0) {
									continue;
								}
							}

							int len = _project_name.length();
							String s_y = f.getName().substring(len + 1, 7); // 3
							String s_m = f.getName().substring(len + 6, 10); // 8
							String s_d = f.getName().substring(len + 9, 13); // 11

							try {
								Date d1 = new Date(Integer.parseInt(s_y) - 1900, Integer.parseInt(s_m) - 1, Integer.parseInt(s_d));
								// delete old file
								if (now.getTime() - d1.getTime() >= 86400000 * DAYS_TO_KEEP_LOG) {
									f.delete();
								}
							} catch (NumberFormatException ne) {
								continue;
							}
						}
					}
				} catch (Exception e) {
					e.printStackTrace();
				}

			}
			checkFile(y,m,d);
			// write log
			try {
				String t_name = Thread.currentThread().getName();
				if (t_name == null) {
					t_name = "";
				}

				String thread_id = String.format(Locale.US, "%s(%d)", t_name, Thread.currentThread().getId());
				String realContent = String.format(Locale.US, "%02d:%02d:%02d.%03d|%s|%s|%s| %s\n", h, mi, s, ms, _cur_uid, tag, thread_id, content);
				_stream.write(realContent.getBytes());
				_stream.flush();

			} catch (Throwable t) {
				t.printStackTrace();
			}
		}
	}

	private void checkFile(int y,int m,int d){

		try {
			File dir = new File(_root_folder);
			String name = "";
			if(_file_name_has_uid) {
				// 文件名带uid
				name = String.format(Locale.getDefault(), "%s_%04d_%02d_%02d_%s.log", _project_name, y, m, d, _cur_uid);
			} else {
				// 文件名不带uid
				name = String.format(Locale.getDefault(), "%s_%04d_%02d_%02d.log", _project_name, y, m, d);
			}
			File file=new File(dir.getPath(), name);
			if(file.createNewFile()){
				closeStream();
				if (!_cur_folder.equals(_root_folder)) {
					_cur_folder = _root_folder;
				}
			}
			if (_stream == null) {
				_stream = new BufferedOutputStream(new FileOutputStream(file, true), FILE_BUFFER_SIZE);
			}



		} catch (Exception e1) {
			e1.printStackTrace();
		}
	}

	/**
	 * 传送SDK日志
	 *
	 * @param receiver
	 *            日志接收者
	 * */
	public static void upload(IUplink uplink, String uid, String receiver, long sn) {

		if (!FeatureSwitch.isLogOn()) {
			return;
		}

		if (receiver == null || receiver.length() == 0 || uplink == null) {
			return;
		}

		File gzipped_file = null;
		FileInputStream fin1 = null;

		try {

			File dir = new File(_root_folder);
			if (!dir.exists()) {
				return;
			}

			byte[] buffer = new byte[4096];
			byte[] body = new byte[LOG_PACKET_SIZE + 8];
			int dataRead = 0;
			int total_size = 0;

			// 创建临时文件
			gzipped_file = File.createTempFile(System.currentTimeMillis() + "", null);
			GZIPOutputStream out_str = new GZIPOutputStream(new FileOutputStream(gzipped_file));

			File files[] = dir.listFiles();

			for (File f : files) {

				// 只上传LLC和CR日志
				if (f.getName().length() != 17 || (f.getName().indexOf(LLC_LOG_PREFIX) != 0 && f.getName().indexOf(CR_LOG_PREFIX) != 0) || f.length() == 0) {
					continue;
				}

				if (f.length() > MAX_LOG_SIZE_ALLOWED) {
					continue;
				}

				if (BuildFlag.DEBUG) {
					Log.d(TAG, "compress file: " + f.getName());
				}

				// Length (4 bytes) == 17 bytes + file length
				int len = (int) (17 + f.length());

				// 写总大小
				out_str.write(Utils.int_to_bytes(len));

				// 写文件名
				out_str.write(f.getName().getBytes());

				// write file content
				FileInputStream fin = new FileInputStream(f);
				while ((dataRead = fin.read(buffer, 0, buffer.length)) > 0) {
					out_str.write(buffer, 0, dataRead);
				}

				fin.close();
			}

			out_str.flush();
			out_str.close();

			total_size = (int) gzipped_file.length();
			if (total_size == 0) {
				return;
			}

			if (BuildFlag.DEBUG) {
				Log.d(TAG, String.format("File size is %d", total_size));
			}

			// 读取文件并发送, 每个包控制在 （50KB） 51200 bytes
			int total_packet = total_size / LOG_PACKET_SIZE;
			int current_packet = 1;
			if ((total_size % LOG_PACKET_SIZE) > 0) {
				++total_packet;
			}

			fin1 = new FileInputStream(gzipped_file);

			while (current_packet <= total_packet) {

				// 最后一个包需要特殊处理
				if (current_packet == total_packet) {
					int remainingSize = total_size % LOG_PACKET_SIZE;
					if (remainingSize > 0) {
						body = new byte[remainingSize + 8];
					}
				}

				int total_read = 0;
				int rem = body.length - 8;

				byte[] btmp = Utils.int_to_bytes(total_packet);
				System.arraycopy(btmp, 0, body, 0, 4);

				btmp = Utils.int_to_bytes(current_packet);
				System.arraycopy(btmp, 0, body, 4, 4);

				while (rem > 0) {
					dataRead = fin1.read(body, total_read + 8, rem);
					if (dataRead <= 0) {
						Log.e(TAG, "Failed to read file log.");
						return;
					}

					total_read += dataRead;
					rem = body.length - 8 - total_read;
				}

				if (uplink.send_data(receiver, body, System.currentTimeMillis())) {
					if (BuildFlag.DEBUG) {
						Log.d(TAG, String.format("uploda log msg sent %d(%d), body lenght is %d ", current_packet, total_packet, body.length));
					}
				} else {
					if (BuildFlag.DEBUG) {
						Log.d(TAG, "send_data failed.");
					}
				}
				current_packet++;
			}

		} catch (Exception e) {
			e.printStackTrace();
		} finally {

			if (fin1 != null) {
				try {
					fin1.close();
				} catch (IOException e) {
					e.printStackTrace();
				}
			}

			if (gzipped_file != null) {
				gzipped_file.delete();
				gzipped_file = null;
			}
		}
	}
}
