package com.trah.accnotify.util

import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.core.app.NotificationCompat
import com.trah.accnotify.AccnotifyApp
import com.trah.accnotify.R
import com.trah.accnotify.ui.MainActivity
import java.util.concurrent.atomic.AtomicInteger

object NotificationHelper {

    private val notificationIdCounter = AtomicInteger(100)

    fun showNotification(
        context: Context,
        messageId: String,
        title: String,
        body: String,
        group: String? = null,
        url: String? = null
    ) {
        val notificationManager = AccnotifyApp.getInstance().getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        val notificationId = notificationIdCounter.getAndIncrement()

        // 处理消息体 - 简化 JSON 显示
        val displayBody = formatBodyForNotification(body)

        // Content intent - open app to view full message
        val contentIntent = Intent(context, MainActivity::class.java).apply {
            putExtra("message_id", messageId)
            putExtra("open_messages", true)
        }

        val pendingIntent = PendingIntent.getActivity(
            context,
            notificationId,
            contentIntent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        // URL 打开动作（如果有）
        val urlIntent = if (!url.isNullOrEmpty()) {
            PendingIntent.getActivity(
                context,
                notificationId + 10000,
                Intent(Intent.ACTION_VIEW, Uri.parse(url)),
                PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
            )
        } else null

        // Use group parameter if provided, otherwise use messageId as group key
        val groupKey = group ?: "accnotify_$messageId"

        val builder = NotificationCompat.Builder(context, AccnotifyApp.CHANNEL_MESSAGES)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(title)
            .setContentText(displayBody)
            // 使用 BigTextStyle 显示更多内容
            .setStyle(NotificationCompat.BigTextStyle()
                .bigText(displayBody)
                .setBigContentTitle(title))
            .setPriority(NotificationCompat.PRIORITY_MAX)
            .setDefaults(NotificationCompat.DEFAULT_ALL)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .setVibrate(longArrayOf(0, 500, 200, 500))
            .setGroup(groupKey)

        // 添加 URL 打开按钮
        if (urlIntent != null) {
            builder.addAction(
                R.drawable.ic_fluent_link,
                "打开链接",
                urlIntent
            )
        }

        notificationManager.notify(messageId, notificationId, builder.build())
    }

    /**
     * 格式化消息体用于通知显示
     * 对于长 JSON 数据进行简化处理
     */
    private fun formatBodyForNotification(body: String): String {
        val trimmed = body.trim()
        
        // 检查是否包含 "--- 完整数据 ---" 分隔符，只显示摘要部分
        val separatorIndex = trimmed.indexOf("--- 完整数据 ---")
        if (separatorIndex > 0) {
            return trimmed.substring(0, separatorIndex).trim() + "\n(点击查看完整数据)"
        }
        
        // 检查是否是 JSON 格式
        if ((trimmed.startsWith("{") && trimmed.endsWith("}")) ||
            (trimmed.startsWith("[") && trimmed.endsWith("]"))) {
            // JSON 数据 - 尝试提取关键信息或简化显示
            return try {
                val gson = com.google.gson.Gson()
                val json = gson.fromJson(trimmed, com.google.gson.JsonObject::class.java)
                
                // 尝试提取常见的消息字段
                val messageField = json.get("message")?.asString
                    ?: json.get("text")?.asString
                    ?: json.get("content")?.asString
                    ?: json.get("body")?.asString
                
                if (messageField != null) {
                    messageField
                } else {
                    // 如果太长，截断并提示
                    if (trimmed.length > 300) {
                        trimmed.take(300) + "...\n(点击查看完整内容)"
                    } else {
                        trimmed
                    }
                }
            } catch (e: Exception) {
                // JSON 解析失败，直接显示
                if (trimmed.length > 300) {
                    trimmed.take(300) + "...\n(点击查看完整内容)"
                } else {
                    trimmed
                }
            }
        }
        
        // 非 JSON 数据
        return if (body.length > 500) {
            body.take(500) + "...\n(点击查看完整内容)"
        } else {
            body
        }
    }
}

