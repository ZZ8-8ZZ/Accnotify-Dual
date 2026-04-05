# Add project specific ProGuard rules here.
# You can control the set of applied configuration files using the
# proguardFiles setting in build.gradle.
#
# For more details, see
#   http://developer.android.com/guide/developing/tools/proguard.html

# Keep Room entities and DAOs
-keep class com.trah.accnotify.data.** { *; }
-keep class * extends androidx.room.RoomDatabase
-keep @androidx.room.Entity class *
-keepclassmembers class * extends androidx.room.RoomDatabase {
    abstract *;
}

# OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
-keepnames class okhttp3.internal.publicsuffix.PublicSuffixDatabase

# Gson
-keepattributes Signature
-keepattributes *Annotation*
-keep class com.google.gson.** { *; }

# Coroutines
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}

# Markwon
-keep class io.noties.markwon.** { *; }
-dontwarn org.commonmark.ext.gfm.strikethrough.Strikethrough
-dontwarn org.commonmark.ext.gfm.strikethrough.**
-dontwarn org.commonmark.**

# Lottie
-keep class com.airbnb.lottie.** { *; }

# Crypto
-keep class androidx.security.crypto.** { *; }
