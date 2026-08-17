import 'package:pancake/core/model/protocol/interaction.dart';
import 'package:pancake/core/model/protocol/state.dart';

import 'connection.dart';
import 'envelope.dart';

typedef DataPackFactory = DataPack Function(Map<String, dynamic> json);

class DataPackParser {
  static final Map<String, DataPackFactory> _parser = {
    Connection.register: Register.fromJson,
    Connection.registerSuccess: RegisterSuccess.fromJson,
    Connection.registerError: RegisterError.fromJson,
    Connection.heartbeat: Heartbeat.fromJson,
    Connection.heartbeatRespond: HeartbeatRespond.fromJson,
    Connection.offline: Offline.fromJson,

    Interaction.clientMessage: ClientMessage.fromJson,
    Interaction.runtimeMessage: RuntimeMessage.fromJson,

    State.runtimeStateRequest: RuntimeStateRequest.fromJson,
    State.runtimeState: RuntimeState.fromJson,
  };

  static DataPack decode(String type, Map<String, dynamic> json) {
    final factory = _parser[type];
    if (factory == null) {
      throw Exception("Unknown DataPack type: $type");
    }

    return factory(json);
  }
}
