import 'package:pancake/core/model/protocol/envelope.dart';

abstract final class Interaction {
  static const String clientMessage = "client_message";
  static const String runtimeMessage = "runtime_message";
  static const String getRelativeMessageHistory =
      "get_relative_message_history";
  static const String relativeMessageHistory = "relative_message_history";
}

abstract interface class MessageDataPack implements DataPack {}

class ClientMessage implements MessageDataPack {
  final String message;
  final int time;

  const ClientMessage({required this.message, required this.time});

  factory ClientMessage.fromJson(Map<String, dynamic> json) {
    return ClientMessage(
      message: json["message"] as String,
      time: json["time"] as int,
    );
  }

  @override
  String get type => Interaction.clientMessage;

  @override
  Map<String, dynamic> toJson() {
    return {"message": message, "time": time};
  }
}

class RuntimeMessage implements MessageDataPack {
  final bool reply;
  final String message;
  final bool isInitiative;
  final int time;

  const RuntimeMessage({
    required this.reply,
    required this.message,
    required this.isInitiative,
    required this.time,
  });

  factory RuntimeMessage.fromJson(Map<String, dynamic> json) {
    return RuntimeMessage(
      reply: json["reply"] as bool,
      message: json["message"] as String,
      isInitiative: json["is_initiative"] as bool,
      time: json["time"] as int,
    );
  }

  @override
  String get type => Interaction.runtimeMessage;

  @override
  Map<String, dynamic> toJson() {
    return {
      "reply": reply,
      "message": message,
      "is_initiative": isInitiative,
      "time": time,
    };
  }
}

class GetRelativeMessageHistory implements DataPack {
  const GetRelativeMessageHistory();

  factory GetRelativeMessageHistory.fromJson(Map<String, dynamic> json) {
    return GetRelativeMessageHistory();
  }

  @override
  String get type => Interaction.getRelativeMessageHistory;

  @override
  Map<String, dynamic> toJson() {
    return {};
  }
}

class RelativeMessageHistory implements DataPack {
  final List<MessageDataPack> messages;

  const RelativeMessageHistory({required this.messages});

  factory RelativeMessageHistory.fromJson(Map<String, dynamic> json) {
    final messages = (json["messages"] as List).map((message) {
      final messageJson = message as Map<String, dynamic>;
      switch (messageJson["type"]) {
        case Interaction.clientMessage:
          return ClientMessage.fromJson(
            messageJson["data"] as Map<String, dynamic>,
          );
        case Interaction.runtimeMessage:
          return RuntimeMessage.fromJson(
            messageJson["data"] as Map<String, dynamic>,
          );
        default:
          throw FormatException("Unknown message type: ${messageJson["type"]}");
      }
    }).toList();

    return RelativeMessageHistory(messages: messages);
  }

  @override
  String get type => Interaction.relativeMessageHistory;

  @override
  Map<String, dynamic> toJson() {
    return {
      "messages": messages
          .map(
            (message) => {
              "category": "interaction",
              "type": message.type,
              "data": message.toJson(),
            },
          )
          .toList(),
    };
  }
}
