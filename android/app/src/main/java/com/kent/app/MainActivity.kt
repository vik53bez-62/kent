package com.kent.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.kent.app.ui.theme.KentTheme
import com.kent.app.features.chat.ChatListScreen
import com.kent.app.features.ai.AssistantPanel

class MainActivity : ComponentActivity() {
  override fun onCreate(savedInstanceState: Bundle?) {
    super.onCreate(savedInstanceState)
    setContent {
      KentTheme {
        val nav = rememberNavController()
        NavHost(navController = nav, startDestination = "chats") {
          composable("chats") { ChatListScreen(onOpenAssistant = { nav.navigate("assistant") }) }
          composable("assistant") { AssistantPanel(onBack = { nav.popBackStack() }) }
        }
      }
    }
  }
}
