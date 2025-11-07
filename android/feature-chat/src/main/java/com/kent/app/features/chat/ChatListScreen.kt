package com.kent.app.features.chat

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun ChatListScreen(onOpenAssistant: () -> Unit) {
  Scaffold(
    topBar = { TopAppBar(title = { Text("Kent") }) },
    floatingActionButton = {
      ExtendedFloatingActionButton(text = { Text("Kent.AI") }, onClick = onOpenAssistant)
    }
  ) { padding ->
    Column(Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
      Text("Ваши чаты появятся здесь", style = MaterialTheme.typography.titleMedium)
      Spacer(Modifier.height(8.dp))
      Text(
        "Нажмите Kent.AI, чтобы задать вопрос или создать заметку",
        color = MaterialTheme.colorScheme.outline
      )
    }
  }
}
