import 'dart:async';
import 'dart:io';

import 'package:logger/logger.dart';
import 'package:pancake/core/model/state/yachiyo_state.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import '../model/state/network_state.dart';

class YachiyoClient {
  final logger = Logger();

  WebSocketChannel? _channel;
  final StreamController<String> _messageController =
      StreamController<String>.broadcast();

  late YachiyoState yachiyoState;
  late NetworkState networkState;
  void setYachiyoState(YachiyoState state){
    yachiyoState = state;
    networkState = state.network;
  }

  Stream<String> get messages => _messageController.stream;

  Future<String?> connect() async {
    await _channel?.sink.close();
    networkState.serverConnected = false;
    
    try {
      final channel = WebSocketChannel.connect(
        Uri.parse('ws://${networkState.serverAddr}/ws/'),
      );

      try {
        await channel.ready;
        networkState.serverConnected = true;
      } on SocketException catch (e) {
        networkState.serverConnected = false;
        
        logger.e(e);
        return e.message;
      } on WebSocketChannelException catch (e) {
        networkState.serverConnected = false;

        logger.e(e);
        return "WebSocketChannelException";
      }

      _channel = channel;
      logger.i("Client Connected.");
      networkState.serverConnected = true;

      channel.stream.listen(
        (message) {
          logger.i("received: $message");
          _messageController.add(message);
        },
        onDone: () {
          return;
        },
      );

      if (networkState.serverConnected == false) {
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
    return networkState.serverConnected;
  }

  String getServerAddr() {
    return networkState.serverAddr;
  }

  Future<bool> changeServerAddr(String addr) async {
    networkState.serverAddr = addr;
    var result = await connect();
    if (result == null){
      return true;
    }
    return false;
  }
}
