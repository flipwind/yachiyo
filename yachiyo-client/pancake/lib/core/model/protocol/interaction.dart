import 'package:pancake/core/model/protocol/envelope.dart';

abstract final class Interaction {
  static final String clientMessage = "client_message";
  static final String runtimeMessage = "runtime_message";
}

class ClientMessage implements DataPack {
  final String message;

  const ClientMessage({
    required this.message,
  });

  factory ClientMessage.fromJson(Map<String, dynamic> json) {
    return ClientMessage(
      message: json["message"] as String,
    );
  }

  @override
  String get type => Interaction.clientMessage;

  @override
  Map<String, dynamic> toJson() {
    return {
      "message": message,
    };
  }
}

class RuntimeMessage implements DataPack {
  final String message;
  final bool isInitiative;

  const RuntimeMessage({
    required this.message,
    required this.isInitiative,
  });

  factory RuntimeMessage.fromJson(Map<String, dynamic> json) {
    return RuntimeMessage(
      message: json["message"] as String,
      isInitiative: json["is_initiative"] as bool,
    );
  }

  @override
  String get type => Interaction.runtimeMessage;

  @override
  Map<String, dynamic> toJson() {
    return {
      "message": message,
      "is_initiative": isInitiative,
    };
  }
}