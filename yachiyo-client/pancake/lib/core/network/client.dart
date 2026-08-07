import 'dart:async';
import 'dart:io';

import 'package:logger/logger.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

class YachiyoClient {
  final logger = Logger();
  WebSocketChannel? _channel;
  final StreamController<String> _messageController =
      StreamController<String>.broadcast();

  String serverAddr = "127.0.0.1:16899";
  bool serverConnected = false;
  String? serverFailedReason;

  Stream<String> get messages => _messageController.stream;

  /// May return a error message.
  Future<String?> connect() async {
    await _channel?.sink.close();
    serverConnected = false;
    serverFailedReason = null;
    
    try {
      final channel = WebSocketChannel.connect(
        Uri.parse('ws://$serverAddr/ws/'),
      );

      try {
        await channel.ready;
        serverConnected = true;
      } on SocketException catch (e) {
        serverConnected = false;
        
        logger.e(e);
        return e.message;
      } on WebSocketChannelException catch (e) {
        serverConnected = false;

        logger.e(e);
        return "WebSocketChannelException";
      }

      _channel = channel;
      logger.i("Client Connected.");
      serverConnected = true;

      channel.stream.listen(
        (message) {
          logger.i("received: $message");
          _messageController.add(message);
        },
        onDone: () {
          return;
        },
      );

      if (serverConnected == false) {
        // onDone
        logger.i("Client Disconnected.");
        return "Client Disconnected.";
      }

    } catch (e) {
      logger.e("Connection failed: $e");
      return "$e";
    }
    return null;
  }

  void send(String message) {
    if (_channel == null) {
      return;
    }
    _channel!.sink.add(message);
  }

  bool getServerConnected() {
    return serverConnected;
  }

  String getServerAddr() {
    return serverAddr;
  }

  String getServerFailedReason() {
    return serverFailedReason!;
  }

  Future<bool> changeServerAddr(String addr) async {
    serverAddr = addr;
    var result = await connect();
    if (result == null){
      return true;
    }
    serverFailedReason = result;
    return false;
  }
}
