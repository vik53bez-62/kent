plugins {
  id("com.android.library")
  kotlin("android")
  id("com.google.dagger.hilt.android")
  kotlin("kapt")
}

android {
  namespace = "com.kent.app.core.network"
  compileSdk = 34
  defaultConfig { minSdk = 26 }
}

dependencies {
  implementation("io.ktor:ktor-client-okhttp:2.3.9")
  implementation("io.ktor:ktor-client-content-negotiation:2.3.9")
  implementation("io.ktor:ktor-serialization-kotlinx-json:2.3.9")
  implementation("com.google.dagger:hilt-android:2.51.1")
  kapt("com.google.dagger:hilt-compiler:2.51.1")
}
