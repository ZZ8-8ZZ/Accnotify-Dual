package com.trah.accnotify.service

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.os.Build
import com.trah.accnotify.AccnotifyApp

class BootReceiver : BroadcastReceiver() {

    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            // Only start if device is registered
            try {
                val app = AccnotifyApp.getInstance()
                if (app.keyManager.isRegistered != true) {
                    return
                }
            } catch (e: Exception) {
                return
            }

            // Start WebSocket service after boot
            val serviceIntent = Intent(context, WebSocketService::class.java).apply {
                action = WebSocketService.ACTION_CONNECT
            }

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(serviceIntent)
            } else {
                context.startService(serviceIntent)
            }
        }
    }
}
