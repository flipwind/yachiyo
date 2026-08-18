import 'package:pancake/core/model/state/client_state.dart';
import 'package:pancake/core/model/state/runtime_state.dart';
import 'package:pancake/core/model/state/network_state.dart';

class YachiyoState {
  final NetworkState network = NetworkState();
  final RuntimeState runtime = RuntimeState();
  final ClientState client = ClientState();
  YachiyoStatus status = YachiyoStatus.disconnected;
}

enum YachiyoStatus {
  disconnected,
  connecting,
  connected,
  registering,
  registered
}