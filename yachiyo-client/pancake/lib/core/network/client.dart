import 'dart:async';

import 'package:web_socket_channel/web_socket_channel.dart';

class YachiyoClient {
  late WebSocketChannel _channel;
  final StreamController<String> _messageController =
      StreamController<String>.broadcast();

  Stream<String> get messages => _messageController.stream;

  Future<void> connect() async {
    try {
      final channel = WebSocketChannel.connect(
        Uri.parse('ws://127.0.0.1:16802/ws/'),
      );

      await channel.ready;
      _channel = channel;
      print("Client Connected.");
      channel.stream.listen((message) {
        print("received: $message");
        _messageController.add(message);
      });
    } catch (e) {
      print("Connect failed: $e");
    }
  }

  void send(String message) {
    _channel.sink.add(message);
  }
}
