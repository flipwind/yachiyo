import 'dart:async';
import 'dart:io';

import 'package:logger/logger.dart';
import 'package:pancake/core/model/state/yachiyo_state.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

class YachiyoClient {
  final logger = Logger();

  WebSocketChannel? _channel;
  final StreamController<String> _messageController =
      StreamController<String>.broadcast();

  late YachiyoState yachiyoState;
  late YachiyoStatus status;
  void setYachiyoState(YachiyoState state){
    yachiyoState = state;
    status = state.status;
  }

  Stream<String> get messages => _messageController.stream;

  Future<void> connect() async {
    await _channel?.sink.close();
    status = YachiyoStatus.disconnected;
    
    try {
      final channel = WebSocketChannel.connect(
        Uri.parse('ws://${yachiyoState.network.serverAddr}/ws/'),
      );

      try {
        await channel.ready;
        status = YachiyoStatus.connected;
      } on SocketException {
        status = YachiyoStatus.disconnected;
        
        rethrow;
      } on WebSocketChannelException {
        status = YachiyoStatus.disconnected;

        rethrow;
      }

      _channel = channel;
      logger.i("Client Connected.");
      status = YachiyoStatus.connected;

      channel.stream.listen(
        (message) {
          _messageController.add(message);
        },
        onDone: () {
          return;
        },
      );

    } catch (e) {
      logger.e("Connection failed: $e");
      rethrow;
    }
    return;
  }

  void send(String message) {
    if (_channel == null) {
      return;
    }
    _channel!.sink.add(message);
  }
}
