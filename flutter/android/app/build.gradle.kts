import java.util.Properties

plugins {
    id("com.android.application")
    id("kotlin-android")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

// Load signing config from key.properties if present (keystore should NOT be committed)
val keystorePropertiesFile = rootProject.file("key.properties")
val keystoreProperties = Properties().apply {
    if (keystorePropertiesFile.exists()) {
        keystorePropertiesFile.inputStream().use { load(it) }
    }
}
val hasReleaseKeystore = keystorePropertiesFile.exists()

android {
    namespace = "com.lfwqsp2641.nettestflutter"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }

    kotlinOptions {
        jvmTarget = JavaVersion.VERSION_11.toString()
    }

    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "com.lfwqsp2641.nettestflutter"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    signingConfigs {
        if (hasReleaseKeystore) {
            create("release") {
                storeFile = file(keystoreProperties.getProperty("storeFile"))
                storePassword = keystoreProperties.getProperty("storePassword")
                keyAlias = keystoreProperties.getProperty("keyAlias")
                keyPassword = keystoreProperties.getProperty("keyPassword")
            }
        }
    }

    buildTypes {
        release {
            // If keystore provided, use release signing; otherwise fall back to debug so `flutter run --release` still works
            signingConfig = if (hasReleaseKeystore) signingConfigs.getByName("release") else signingConfigs.getByName("debug")
        }
    }
}

flutter {
    source = "../.."
}

// Pre-build task: copy native .so from repo lib/ into jniLibs so they are packaged into the APK/AAB
// This avoids committing binaries; run automatically before build.
val syncJniLibs by tasks.registering {
    group = "build"
    description = "Sync Go .so files into app/src/main/jniLibs per ABI"
    doLast {
        // Determine repo root: android/ is at flutter/android, so repo root is parent of parent
        val repoRoot = project.rootDir.parentFile?.parentFile
        val libDir = repoRoot?.resolve("lib")
        if (libDir == null || !libDir.exists()) return@doLast

        val mapping = mapOf(
            "arm64-v8a" to "libandroidnetcore_arm64-v8a.so",
            "armeabi-v7a" to "libandroidnetcore_armeabi-v7a.so",
            "x86" to "libandroidnetcore_x86.so",
            "x86_64" to "libandroidnetcore_x86_64.so",
        )
        mapping.forEach { (abi, soName) ->
            val src = libDir.resolve(soName)
            if (src.exists()) {
                val dstDir = project.file("src/main/jniLibs/$abi")
                dstDir.mkdirs()
                val dst = dstDir.resolve(soName)
                src.copyTo(dst, overwrite = true)
                // Also provide a canonical name for loading as "libnetcore.so" or System.loadLibrary("netcore")
                val alias = dstDir.resolve("libnetcore.so")
                src.copyTo(alias, overwrite = true)
            }
        }
    }
}

tasks.matching { it.name == "preBuild" }.configureEach {
    dependsOn(syncJniLibs)
}
