plugins {
  id("com.android.library")
  kotlin("android")
}

android {
  namespace = "com.kent.app.features.calls"
  compileSdk = 34
  defaultConfig { minSdk = 26 }
}

dependencies {
  implementation("org.webrtc:google-webrtc:1.0.32006")
}
