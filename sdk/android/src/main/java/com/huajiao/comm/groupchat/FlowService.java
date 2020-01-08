package com.huajiao.comm.groupchat;

import android.app.IntentService;
import android.os.Handler;
import android.os.HandlerThread;
import android.util.Log;

import com.huajiao.comm.im.api.ILongLiveConn;
import com.huajiao.comm.protobuf.GroupChatProto;

import java.util.Locale;

/**
 * An {@link IntentService} subclass for handling asynchronous task requests in
 * a service on a separate handler thread.
 * <p/>
 * TODO: Customize class - update intent actions, extra parameters and static
 * helper methods.
 */
final class FlowService {
    private Handler handler_=null;
    private HandlerThread handlerThread_;

    private	static final String	TAG	= "FlowService";
    private GroupChatHelper instance_;

    FlowService(GroupChatHelper instance) {
        instance_ = instance;
        handlerThread_ = new HandlerThread("FlowService");
        handlerThread_.start();
        handler_ = new Handler(handlerThread_.getLooper());
    }

    // This method is allowed to be called from any thread
    synchronized void requestStop() {
        // using the handler, post a Runnable that will quit()
        // the Looper attached to our DownloadThread
        // obviously, all previously queued tasks will be executed
        // before the loop gets the quit Runnable
        handler_.post(new Runnable() {
            @Override
            public void run() {
                // This is guaranteed to run on the DownloadThread
                // so we can use myLooper() to get its looper
                Log.i(TAG, "DownloadThread loop quitting by request");
                handlerThread_.quit();
            }
        });
    }

    synchronized void enqueueReq(final long sn, final String reqids, final int payload) {
        // Wrap DownloadTask into another Runnable to track the statistics
        handler_.post(new Runnable() {
            @Override
            public void run() {
                try {
                    long waitmills = instance_.getFlowCtlAbsMills() - System.currentTimeMillis();
                    if ( waitmills > 0 && waitmills < GroupChatHelper.MAX_FLOWCTL_MILLSECONDS ) {
                        GPLogger.i(TAG, (new StringBuilder()).append("send req wait for ").append(waitmills).append("ms").toString());
                        synchronized (GroupChatHelper.flow_lock_) {
                            GroupChatHelper.flow_lock_.wait(waitmills);
                        }
                    }

                    GroupChatProto.GroupUpPacket packet = instance_.removePendingQueue(sn, payload);//用remove而不是get，防止get后被merge
                    ILongLiveConn conn = instance_.getConn();

                    if (conn.send_service_message(GroupChatHelper.GROUPCHAT_SRV_ID, sn, packet.toByteArray())) {
                        instance_.addUnackQueue(sn, packet);
                        instance_.setFlowCtlAbsMills(System.currentTimeMillis()+5000, false);
                        GPLogger.i(TAG, String.format(Locale.US, "send req payload=%d reqid=%s, sn=%d", packet.getPayload(), reqids, sn));
                    } else {
                        GPLogger.e(TAG, String.format(Locale.US, "send req payload=%d reqid=%s,sn=%d failed", packet.getPayload(), reqids, sn));
                    }
                } catch (Exception e) {
                    GPLogger.i(TAG, String.format(Locale.US, "send req reqid=%s, sn=%d exception", reqids, sn));
                }
            }
        });
    }

}
