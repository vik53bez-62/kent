package com.kent.app.features.ai

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun AssistantPanel(onBack: () -> Unit) {
  var prompt by remember { mutableStateOf("") }

  Scaffold(
    topBar = {
      TopAppBar(
        title = { Text("Kent.AI") },
        navigationIcon = { TextButton(onClick = onBack) { Text("Назад") } }
      )
    }
  ) { padding ->
    Column(Modifier.fillMaxSize().padding(padding).padding(16.dp)) {
      OutlinedTextField(
        value = prompt,
        onValueChange = { prompt = it },
        label = { Text("Спроси что угодно…") },
        modifier = Modifier.fillMaxWidth()
      )
      Spacer(Modifier.height(12.dp))
      Button(onClick = { /* TODO: call AI */ }) { Text("Отправить") }
      Spacer(Modifier.height(16.dp))
      Text("Ответ появится здесь…", color = MaterialTheme.colorScheme.outline)
    }
  }
}
