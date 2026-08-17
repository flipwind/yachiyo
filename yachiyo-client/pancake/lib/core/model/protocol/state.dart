import 'package:pancake/core/model/protocol/envelope.dart';

abstract final class State {
  static final String runtimeStateRequest = "runtime_state_request";
  static final String runtimeState = "runtime_state";
}

class RuntimeStateRequest implements DataPack {
  const RuntimeStateRequest();

  factory RuntimeStateRequest.fromJson(Map<String, dynamic> json) {
    return RuntimeStateRequest();
  }

  @override
  String get type => State.runtimeStateRequest;

  @override
  Map<String, dynamic> toJson() {
    return {};
  }
}


class RuntimeState implements DataPack {
  final String state;

  const RuntimeState({
    required this.state,
  });

  factory RuntimeState.fromJson(Map<String, dynamic> json) {
    return RuntimeState(
      state: json["state"] as String,
    );
  }

  @override
  String get type => State.runtimeState;

  @override
  Map<String, dynamic> toJson() {
    return {
      "state": state,
    };
  }
}