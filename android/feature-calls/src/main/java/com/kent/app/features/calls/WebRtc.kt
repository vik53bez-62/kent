package com.kent.app.features.calls

import android.content.Context
import org.webrtc.PeerConnectionFactory

object WebRtc {
  fun createPeerConnectionFactory(context: Context): PeerConnectionFactory {
    val options = PeerConnectionFactory.InitializationOptions.builder(context).createInitializationOptions()
    PeerConnectionFactory.initialize(options)
    return PeerConnectionFactory.builder().createPeerConnectionFactory()
  }
}
