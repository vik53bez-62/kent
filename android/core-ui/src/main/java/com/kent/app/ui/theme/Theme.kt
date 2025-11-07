package com.kent.app.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val KentLightColors = lightColorScheme(
  primary = Color(0xFF2563EB),
  onPrimary = Color.White,
  primaryContainer = Color(0xFFDBEAFE),
  onPrimaryContainer = Color(0xFF1E3A8A),
  secondary = Color(0xFF06B6D4),
  onSecondary = Color(0xFF002B31),
  secondaryContainer = Color(0xFFCFFAFE),
  onSecondaryContainer = Color(0xFF164E63),
  tertiary = Color(0xFFA78BFA),
  onTertiary = Color(0xFF2E1065),
  tertiaryContainer = Color(0xFFEDE9FE),
  background = Color(0xFFF8FAFC),
  surface = Color(0xFFFFFFFF),
  surfaceVariant = Color(0xFFE2E8F0),
  outline = Color(0xFF94A3B8),
  error = Color(0xFFEF4444)
)

private val KentDarkColors = darkColorScheme(
  primary = Color(0xFF2563EB),
  onPrimary = Color.White,
  primaryContainer = Color(0xFF1E3A8A),
  onPrimaryContainer = Color(0xFFDBEAFE),
  secondary = Color(0xFF06B6D4),
  onSecondary = Color(0xFF002B31),
  secondaryContainer = Color(0xFF164E63),
  onSecondaryContainer = Color(0xFFCFFAFE),
  tertiary = Color(0xFFA78BFA),
  onTertiary = Color(0xFFEDE9FE),
  background = Color(0xFF0B1220),
  surface = Color(0xFF0F172A),
  surfaceVariant = Color(0xFF1E293B),
  outline = Color(0xFF94A3B8),
  error = Color(0xFFEF4444)
)

@Composable
fun KentTheme(darkTheme: Boolean = false, content: @Composable () -> Unit) {
  MaterialTheme(
    colorScheme = if (darkTheme) KentDarkColors else KentLightColors,
    content = content
  )
}
