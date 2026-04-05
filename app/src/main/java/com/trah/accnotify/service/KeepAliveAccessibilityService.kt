package com.trah.accnotify.service

import android.accessibilityservice.AccessibilityService
import android.content.Intent
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.view.accessibility.AccessibilityEvent
import com.trah.accnotify.AccnotifyApp

/**
 * Minimal Accessibility Service for Background Persistence.
 *
 * Unlike GKD (which only works when screen is on), this service keeps
 * WebSocket connection alive 24/7 for push notifications.
 *
 * Keep-alive strategy:
 * 1. Fixed interval heartbeat (regardless of screen state)
 * 2. Works alongside JobService for redundancy
 * 3. Automatically restarts WebSocket service when it's not running
 */
class KeepAliveAccessibilityService : AccessibilityService() {

    companion object {
        private const val TAG = "KeepAliveService"
        // Fixed heartbeat interval - WebSocket needs to stay connected 24/7
        private const val HEARTBEAT_INTERVAL_MS = 5 * 60 * 1000L // 5 minutes

        @Volatile
        var isServiceRunning = false
            private set
    }

    private val handler = Handler(Looper.getMainLooper())
    private var lastServiceStartTime = 0L

    private val heartbeatRunnable = object : Runnable {
        override fun run() {
            if (isServiceRunning) {
                Log.d(TAG, "Heartbeat: ensuring WebSocket service is alive")
                startWebSocketServiceIfNeeded()
                handler.postDelayed(this, HEARTBEAT_INTERVAL_MS)
            }
        }
    }

    override fun onServiceConnected() {
        Log.i(TAG, "Keep-alive accessibility service connected")
        super.onServiceConnected()
        isServiceRunning = true

        // Start WebSocket service once when accessibility service starts
        startWebSocketServiceIfNeeded()

        // Schedule periodic heartbeat
        handler.postDelayed(heartbeatRunnable, HEARTBEAT_INTERVAL_MS)
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        return super.onStartCommand(intent, flags, startId)
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent?) {
        // Minimal processing - we don't need to react to events
        // The heartbeat mechanism handles service keep-alive
    }

    override fun onInterrupt() {
        Log.w(TAG, "Keep-alive accessibility service interrupted")
    }

    override fun onUnbind(intent: Intent?): Boolean {
        Log.w(TAG, "Keep-alive accessibility service unbound")
        isServiceRunning = false
        handler.removeCallbacks(heartbeatRunnable)
        return super.onUnbind(intent)
    }

    override fun onDestroy() {
        Log.w(TAG, "Keep-alive accessibility service destroyed")
        isServiceRunning = false
        handler.removeCallbacks(heartbeatRunnable)
        super.onDestroy()
    }

    private fun startWebSocketServiceIfNeeded() {
        val now = System.currentTimeMillis()
        // Throttle service starts to at most once per minute
        if (now - lastServiceStartTime < 60_000L) {
            return
        }
        lastServiceStartTime = now

        val intent = Intent(this, WebSocketService::class.java).apply {
            action = WebSocketService.ACTION_CONNECT
        }
        try {
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
                startForegroundService(intent)
            } else {
                startService(intent)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start WebSocketService: ${e.message}")
        }
    }
}

