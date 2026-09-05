class Message {
  bool? reply;
  String role;
  String message;
  DateTime time;

  Message({
    this.reply,
    required this.role,
    required this.message,
    required this.time,
  });
}
