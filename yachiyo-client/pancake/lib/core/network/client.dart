import 'package:web_socket_channel/web_socket_channel.dart';

class YachiyoClient {
  Future<void> connect() async {
    try {
      final channel = WebSocketChannel.connect(
        Uri.parse('ws://127.0.0.1:16802/ws/'),
      );

      await channel.ready;
      print("Client Connected.");
      channel.stream.listen(
        (message) {
          print("received: $message");
        }
      );
    } catch (e) {
      print("Connect failed: $e");
    }
  }
}
