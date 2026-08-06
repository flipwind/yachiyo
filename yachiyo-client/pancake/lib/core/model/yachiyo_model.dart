import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:pancake/core/network/client.dart';

// TODO: switch to gRPC

class Message {
  String role;
  String message;
  int time;

  Message({required this.role, required this.message, required this.time});

  factory Message.fromJson(Map<String, dynamic> json) {
    return Message(
      role: json['role'],
      message: json['message'],
      time: json['timestamp'],
    );
  }
}

class YachiyoModel extends ChangeNotifier {
  final YachiyoClient client;

  YachiyoModel(this.client) {
    client.messages.listen(_handleMessage);
  }

  String status = "unknown";
  List<Message> messages = [];

  void _handleMessage(String message) {
    // Make a suppose that received a json.

    final Map<String, dynamic> parsedJson = jsonDecode(message);

    final type = parsedJson["type"];
    final data = parsedJson["data"];

    if (type == "status") {
      status = data;
    }

    if (type == "message") {
      messages = [];
      for (var msgData in data) {
        messages.add(Message.fromJson(msgData));
      }
    }

    if (type == "message_delta") {
      messages.add(Message.fromJson(data));
    }

    notifyListeners();
  }

  Future<void> connect() {
    return client.connect();
  }

  void sendMessage(String message) {
    final data = {
      "type": "send_message",
      "data": message
    };

    client.send(jsonEncode(data));

    messages.add(Message(role: "user", message: message, time: DateTime.now().millisecondsSinceEpoch ~/ 1000));
    notifyListeners();
  }
}
