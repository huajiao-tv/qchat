package com.huajiao.comm.chatroom;

import org.json.JSONObject;

import java.util.Locale;

public class CRPacketHelper {
	
	public static final String TAG = "CRPH";
	
	/*private static String formatTime(long ms) {
		return DateFormat.format("HH:mm:ss", ms).toString();
	}*/
	
	/*public static String parseGiftMsg(String data) {
		try {
			JSONObject jodata = new JSONObject(data);
			if(jodata.getInt("type") != 30) return null;
			// JSONObject jo = new JSONObject();
			StringBuffer sb = new StringBuffer();
			// jo.put("t", "gift");
			// jo.put("rmid", jodata.getInt("roomid"));
			sb.append("{type:30, rm_id:"+jodata.getInt("roomid"));
			if(jodata.has("extends")) {
				JSONObject joEx = jodata.getJSONObject("extends");
				if(joEx.has("receiver")) {
					JSONObject joRecv = joEx.getJSONObject("receiver");
					// jo.put("rcvuid", joRecv.getInt("uid"));
					// jo.put("rcvname", joRecv.getString("nickname"));
					sb.append(", rcv_uid:"+joRecv.getInt("uid"));
				}
				if(joEx.has("sender")) {
					JSONObject joSend = joEx.getJSONObject("sender");
					// jo.put("snduid", joSend.getInt("uid"));
					// jo.put("sndname", joSend.getString("nickname"));
					sb.append(", snd_uid:"+joSend.getInt("uid"));
				}
				if(joEx.has("giftinfo")) {
					JSONObject joGift = joEx.getJSONObject("giftinfo");
					// jo.put("gftid", joGift.getString("giftid"));
					// jo.put("gftname", joGift.getString("giftname"));
					// jo.put("gftamount", joGift.getString("amount"));
					sb.append(", gift_id:"+joGift.getString("giftid"));
					sb.append(", gift_amt:"+joGift.getString("amount"));
					sb.append(", gift_name:"+joGift.getString("giftname"));
				}
			}
			sb.append("}");
			// return jo.toString();
			return sb.toString();
		} catch (Throwable tr) {
			if(JhFlag.enableDebug()) {
				CRLogger.e(TAG, Log.getStackTraceString(tr));
			}
			return null;
		}
	}*/
	public static String parseMsg(String data) {
		// CRLogger.d(TAG, data);
		try {
			JSONObject jodata = new JSONObject(data);
			int type = jodata.getInt("type");
			int roomid = jodata.getInt("roomid");
			if(type == 30) {
				JSONObject joEx = jodata.getJSONObject("extends");
				int senduid = joEx.getJSONObject("sender").getInt("uid");
				String giftid = joEx.getJSONObject("giftinfo").getString("giftid");
				String amount = joEx.getJSONObject("giftinfo").getString("amount");
				return String.format(Locale.getDefault(), "%d %d %d GIVE %s %s", type, roomid, senduid, giftid, amount);
			} else if(type == 9) {
				String text = jodata.getString("text");
				String senduid = jodata.getJSONObject("extends").getString("userid");
				return String.format(Locale.getDefault(), "%d %d %s SAY %s", type, roomid, senduid, text);
			} else {
				return String.format(Locale.getDefault(), "%d %d", type, roomid);
			}
		} catch (Throwable tr) {
			/*if(JhFlag.enableDebug()) {
				CRLogger.e(TAG, Log.getStackTraceString(tr));
			}*/			
		}
		return null;
	}
}
