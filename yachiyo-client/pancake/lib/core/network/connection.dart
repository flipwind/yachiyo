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

  final StreamController<String> _errorMessage = StreamController<String>.broadcast();
  Stream<String> get errors => _errorMessage.stream;

  void _setStatus(ConnectionStatus status) {
    _status = status;
    _statusController.add(status);
  }

  void sendRaw(String data) {
    _channel?.sink.add(data);
  }

  void _pushError(String message) {
    _errorMessage.add(message);
  }

  Future<void> disconnect() async {
    await _channel?.sink.close();
    if (_channel != null) _setStatus(ConnectionStatus.disconnected);
    _channel = null;
  }

  Future<void> connect(String address) async {
    await disconnect();
    _setStatus(ConnectionStatus.connecting);

    try {
      final channel = WebSocketChannel.connect(Uri.parse(address));

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
            _pushError("Connection error: ${err.toString()}");
            _channel = null;
            _setStatus(ConnectionStatus.disconnected);
          },
        );
      } catch (e) {
        logger.e("Client connection failed: ${e.toString()}");
        _pushError("Client connection failed: ${e.toString()}");
        _setStatus(ConnectionStatus.disconnected);
      }
    } catch (e) {
      logger.e("Client connection failed: ${e.toString()}");
      _pushError("Client connection failed: ${e.toString()}");
      _setStatus(ConnectionStatus.disconnected);
    }
  }

  Future<void> dispose() async {
    await _channel?.sink.close();
    await _messageController.close();
    await _statusController.close();
  }
}
