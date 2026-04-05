package com.trah.accnotify.crypto

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import java.security.KeyPair
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.PrivateKey
import java.security.PublicKey
import java.security.SecureRandom

class KeyManager(private val context: Context) {

    private val keyStore: KeyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply { load(null) }
    private val prefs by lazy { createEncryptedPrefs() }

    private fun createEncryptedPrefs() = EncryptedSharedPreferences.create(
        context,
        PREFS_NAME,
        MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build(),
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    /**
     * Ensure RSA key pair and device key exist
     */
    fun ensureKeysExist() {
        if (!keyStore.containsAlias(RSA_KEY_ALIAS)) {
            generateRSAKeyPair()
        }
        if (getDeviceKey() == null) {
            generateDeviceKey()
        }
    }

    /**
     * Generate RSA-2048 key pair stored in Android Keystore
     */
    private fun generateRSAKeyPair() {
        val keyPairGenerator = KeyPairGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_RSA,
            ANDROID_KEYSTORE
        )

        val spec = KeyGenParameterSpec.Builder(
            RSA_KEY_ALIAS,
            KeyProperties.PURPOSE_DECRYPT or KeyProperties.PURPOSE_ENCRYPT
        )
            .setKeySize(2048)
            .setDigests(KeyProperties.DIGEST_SHA256)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_RSA_OAEP)
            .build()

        keyPairGenerator.initialize(spec)
        keyPairGenerator.generateKeyPair()
    }

    /**
     * Get the RSA key pair
     */
    fun getKeyPair(): KeyPair? {
        return try {
            val privateKey = keyStore.getKey(RSA_KEY_ALIAS, null) as? PrivateKey ?: return null
            val publicKey = keyStore.getCertificate(RSA_KEY_ALIAS)?.publicKey ?: return null
            KeyPair(publicKey, privateKey)
        } catch (e: Exception) {
            null
        }
    }

    /**
     * Get the private key for decryption
     */
    fun getPrivateKey(): PrivateKey? {
        return keyStore.getKey(RSA_KEY_ALIAS, null) as? PrivateKey
    }

    /**
     * Get the public key
     */
    fun getPublicKey(): PublicKey? {
        return keyStore.getCertificate(RSA_KEY_ALIAS)?.publicKey
    }

    /**
     * Export public key in PEM format
     */
    fun exportPublicKeyPEM(): String? {
        val publicKey = getPublicKey() ?: return null
        val encoded = Base64.encodeToString(publicKey.encoded, Base64.NO_WRAP)
        return "-----BEGIN PUBLIC KEY-----\n" +
                encoded.chunked(64).joinToString("\n") +
                "\n-----END PUBLIC KEY-----"
    }

    /**
     * Generate a random 32-character device key
     */
    private fun generateDeviceKey(): String {
        val bytes = ByteArray(24) // 24 bytes = 32 base64 characters
        SecureRandom().nextBytes(bytes)
        val deviceKey = Base64.encodeToString(bytes, Base64.URL_SAFE or Base64.NO_WRAP or Base64.NO_PADDING)
        prefs.edit().putString(PREF_DEVICE_KEY, deviceKey).apply()
        return deviceKey
    }

    /**
     * Get the device key
     */
    fun getDeviceKey(): String? {
        return prefs.getString(PREF_DEVICE_KEY, null)
    }

    /**
     * Regenerate device key only (e.g., if compromised)
     */
    fun regenerateDeviceKey(): String {
        isRegistered = false // Need to re-register
        return generateDeviceKey()
    }

    /**
     * Regenerate both device key and RSA key pair (full reset)
     */
    fun regenerateAllKeys(): String {
        // Delete old RSA key
        if (keyStore.containsAlias(RSA_KEY_ALIAS)) {
            keyStore.deleteEntry(RSA_KEY_ALIAS)
        }
        // Generate new RSA key pair
        generateRSAKeyPair()
        // Generate new device key
        val newDeviceKey = generateDeviceKey()
        // Reset registration status
        isRegistered = false
        return newDeviceKey
    }

    /**
     * Get or set server URL
     */
    var serverUrl: String
        get() = prefs.getString(PREF_SERVER_URL, DEFAULT_SERVER_URL) ?: DEFAULT_SERVER_URL
        set(value) {
            prefs.edit().putString(PREF_SERVER_URL, value).apply()
            addServer(value) // Automatically add to history list
        }

    fun getServers(): Set<String> {
        return prefs.getStringSet(PREF_SERVER_LIST, setOf(DEFAULT_SERVER_URL)) ?: setOf(DEFAULT_SERVER_URL)
    }

    fun addServer(url: String) {
        val current = getServers().toMutableSet()
        current.add(url)
        prefs.edit().putStringSet(PREF_SERVER_LIST, current).apply()
    }

    fun removeServer(url: String) {
        val current = getServers().toMutableSet()
        if (current.contains(url)) {
            current.remove(url)
            prefs.edit().putStringSet(PREF_SERVER_LIST, current).apply()
            
            // If we removed the active one, fallback to default or another one
            if (serverUrl == url) {
                serverUrl = current.firstOrNull() ?: DEFAULT_SERVER_URL
            }
        }
    }

    /**
     * Check if device is registered with server
     */
    var isRegistered: Boolean
        get() = prefs.getBoolean(PREF_IS_REGISTERED, false)
        set(value) = prefs.edit().putBoolean(PREF_IS_REGISTERED, value).apply()

    /**
     * Control whether to show foreground notification for keep-alive
     * Default is true (show notification) for better reliability
     * When disabled, the app relies on AccessibilityService + JobService for keep-alive
     */
    var showForegroundNotification: Boolean
        get() = prefs.getBoolean(PREF_SHOW_FOREGROUND_NOTIFICATION, true)
        set(value) = prefs.edit().putBoolean(PREF_SHOW_FOREGROUND_NOTIFICATION, value).apply()

    companion object {
        private const val ANDROID_KEYSTORE = "AndroidKeyStore"
        private const val RSA_KEY_ALIAS = "accnotify_rsa_key"
        private const val PREFS_NAME = "accnotify_secure_prefs"
        private const val PREF_DEVICE_KEY = "device_key"
        private const val PREF_SERVER_URL = "server_url"
        private const val PREF_SERVER_LIST = "server_list"
        private const val PREF_IS_REGISTERED = "is_registered"
        private const val PREF_SHOW_FOREGROUND_NOTIFICATION = "show_foreground_notification"
        private const val DEFAULT_SERVER_URL = "https://an.trah.cn"
    }

}
