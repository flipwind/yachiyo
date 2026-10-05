import 'dart:async';

import 'package:flutter/material.dart';
import 'package:pancake/application/session/session.dart';
import 'package:pancake/core/model/message.dart';
import 'package:pancake/core/network/connection.dart';

sealed class ProviderEvent {}

class EventErrorMessage extends ProviderEvent {
  final String code;
  final String message;
  EventErrorMessage(this.code, this.message);
}

class PancakeProvider extends ChangeNotifier {
  final PancakeSession _session;

  ConnectionStatus _status = ConnectionStatus.disconnected;
  ConnectionStatus get status => _status;
  final ProviderData _data = ProviderData();
  ProviderData get data => _data;
  final _eventController = StreamController<ProviderEvent>.broadcast();
  Stream<ProviderEvent> get events => _eventController.stream;

  PancakeProvider(this._session) {
    _session.events.listen((event) {
      switch (event) {
        case SessionConnected():
          _status = ConnectionStatus.connected;

        case SessionDisconnected():
          _status = ConnectionStatus.disconnected;
          _data.clear();

        case SessionConnecting():
          _status = ConnectionStatus.connecting;

        case SessionRegistered():
          _data.runtimeName = event.runtimeName;
          _data.runtimeVersion = event.runtimeVersion;
          _data.registered = true;
          _session.reloadMessagesFromRuntime();

        case SessionStateReceived():
          _data.runtimeState = event.state;

        case SessionMessageAppend():
          _data.messages.add(event.message);

        case SessionMessageReload():
          _data.messages = event.messages;

        case SessionErrorMessage():
          _eventController.add(EventErrorMessage(event.code, event.message));
      }

      notifyListeners();
    });
  }

  static Future<PancakeProvider> create() async {
    return PancakeProvider(await PancakeSession.create());
  }

  void sendMessage(String message) {
    _session.sendMessage(message);

    notifyListeners();
  }

  void clearMessages() {
    _data.messages.clear();

    notifyListeners();
  }

  void reloadMessagesFromRuntime() {
    _session.reloadMessagesFromRuntime();
  }

  void changeServerAddr(String address) {
    _data.serverAddr = address;
    _session.start(address);

    notifyListeners();
  }
}

class ProviderData {
  List<Message> messages = [];
  String runtimeName;
  String runtimeVersion;
  String runtimeState;
  String serverAddr;

  bool registered;

  ProviderData({
    this.runtimeName = "",
    this.runtimeVersion = "",
    this.runtimeState = "",
    this.serverAddr = "127.0.0.1:16899",
    this.registered = false,
  });

  void clear() {
    messages.clear();
    runtimeName = "";
    runtimeVersion = "";
    runtimeState = "";
    registered = false;
  }
}
