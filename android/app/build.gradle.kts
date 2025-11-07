plugins {
  id("com.android.application")
  kotlin("android")
  id("com.google.dagger.hilt.android")
  kotlin("kapt")
}

android {
  namespace = "com.kent.app"
  compileSdk = 34

  defaultConfig {
    applicationId = "com.kent.app"
    minSdk = 26
    targetSdk = 34
    versionCode = 1
    versionName = "0.1.0"
    vectorDrawables.useSupportLibrary = true
  }

  buildTypes {
    release {
      isMinifyEnabled = true
      proguardFiles(
        getDefaultProguardFile("proguard-android-optimize.txt"),
        "proguard-rules.pro"
      )
    }

    debug { applicationIdSuffix = ".debug" }
  }

  buildFeatures { compose = true }
  composeOptions { kotlinCompilerExtensionVersion = "1.5.14" }
  kotlinOptions { jvmTarget = "17" }
  packaging { resources.excludes += "/META-INF/{AL2.0,LGPL2.1}" }
}

dependencies {
  // Compose (BOM)
  implementation(platform("androidx.compose:compose-bom:2024.10.00"))
  implementation("androidx.compose.material3:material3")
  implementation("androidx.compose.ui:ui")
  implementation("androidx.compose.ui:ui-tooling-preview")
  debugImplementation("androidx.compose.ui:ui-tooling")
  implementation("androidx.activity:activity-compose:1.9.3")
  implementation("androidx.navigation:navigation-compose:2.8.2")

  // Hilt
  implementation("com.google.dagger:hilt-android:2.51.1")
  kapt("com.google.dagger:hilt-compiler:2.51.1")

  // Networking (Ktor)
  implementation("io.ktor:ktor-client-okhttp:2.3.9")
  implementation("io.ktor:ktor-client-content-negotiation:2.3.9")
  implementation("io.ktor:ktor-serialization-kotlinx-json:2.3.9")

  // Coroutines
  implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.1")

  // WebRTC
  implementation("org.webrtc:google-webrtc:1.0.32006")

  // FCM (подключите google-services по необходимости)
  implementation("com.google.firebase:firebase-messaging-ktx:24.0.0")

  // Images
  implementation("io.coil-kt:coil-compose:2.7.0")
}
