import 'dart:async';
import 'package:logger/logger.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

enum ConnectionStatus { disconnected, connecting, connected }

class PancakeConnection {
  final logger = Logger();

  WebSocketChannel? _channel;

  final StreamController<String> _messageController =
      StreamController<String>.broadcast();
  Stream<String> get rawMessages => _messageController.stream;

  final StreamController<ConnectionStatus> _statusController =
      StreamController<ConnectionStatus>.broadcast();
  Stream<ConnectionStatus> get status => _statusController.stream;

  ConnectionStatus _status = ConnectionStatus.disconnected;
  ConnectionStatus get currentStatus => _status;

  void _setStatus(ConnectionStatus status) {
    _status = status;
    _statusController.add(status);
  }

  void sendRaw(String data) {
    _channel?.sink.add(data);
  }

  Future<void> connect(String address) async {
    await _channel?.sink.close();
    _setStatus(ConnectionStatus.disconnected);
    _setStatus(ConnectionStatus.connecting);

    try {
      final channel = WebSocketChannel.connect(Uri.parse("ws://$address/ws/"));

      try {
        await channel.ready;
        _channel = channel;
        _setStatus(ConnectionStatus.connected);

        logger.i("Client Connected.");

        channel.stream.listen(
          (message) {
            if (_channel != channel) return;
            _messageController.add(message);
          },

          onDone: () {
            if (_channel != channel) return;
            _channel = null;
            _setStatus(ConnectionStatus.disconnected);
          },

          onError: (err) {
            if (_channel != channel) return;
            logger.e("Connection error: ${err.toString()}");
            _channel = null;
            _setStatus(ConnectionStatus.disconnected);
          },
        );
      } catch (e) {
        logger.e("Client connection failed: ${e.toString()}");
        _setStatus(ConnectionStatus.disconnected);
        rethrow;
      }
    } catch (e) {
      logger.e("Client connection failed: ${e.toString()}");
      _setStatus(ConnectionStatus.disconnected);
      rethrow;
    }
  }

  Future<void> dispose() async {
    await _channel?.sink.close();
    await _messageController.close();
    await _statusController.close();
  }
}
