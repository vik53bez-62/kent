plugins {
  id("com.android.library")
  kotlin("android")
}

android {
  namespace = "com.kent.app.core.ui"
  compileSdk = 34
  defaultConfig { minSdk = 26 }
  buildFeatures { compose = true }
  composeOptions { kotlinCompilerExtensionVersion = "1.5.14" }
}

dependencies {
  implementation(platform("androidx.compose:compose-bom:2024.10.00"))
  implementation("androidx.compose.material3:material3")
  implementation("androidx.compose.ui:ui")
}
