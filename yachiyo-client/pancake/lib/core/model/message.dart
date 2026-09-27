abstract class Message {
  String content;
  DateTime time;

  Message({
    required this.content,
    required this.time,
  });
}

class UserMessage extends Message {
  UserMessage({
    required super.content,
    required super.time
  });
}

class AssistantMessage extends Message {
  bool reply;
  bool initiative;
  AssistantMessage({
    required super.content,
    required super.time,
    required this.reply,
    required this.initiative,
  });
}
