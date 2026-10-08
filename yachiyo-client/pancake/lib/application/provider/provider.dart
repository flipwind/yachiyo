import 'dart:async';

import 'package:flutter/material.dart';
import 'package:pancake/application/session/session.dart';
import 'package:pancake/core/identity/identity.dart';
import 'package:pancake/core/model/message.dart';
import 'package:pancake/core/network/connection.dart';
import 'package:shared_preferences/shared_preferences.dart';

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
  final ProviderData _data;
  ProviderData get data => _data;
  final _eventController = StreamController<ProviderEvent>.broadcast();
  Stream<ProviderEvent> get events => _eventController.stream;

  final Identity _identity;
  Identity get identity => _identity;

  PancakeProvider(this._session, this._data, this._identity) {
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
        case SessionSelfErrorMessage():
          _pushSelfError(event.message);
      }

      notifyListeners();
    });
  }

  void _pushSelfError(String message) {
    _eventController.add(EventErrorMessage("pancake_error", message));
  }

  static Future<PancakeProvider> create() async {
    final session = await PancakeSession.create();
    final data = await ProviderData.create();
    return PancakeProvider(session, data, session.identity);
  }

  bool sendMessage(String message) {
    if (data.registered != true || status != ConnectionStatus.connected) {
      _pushSelfError("Client not connected");
      return false;
    }

    _session.sendMessage(message);
    notifyListeners();
    return true;
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
    _data.setServerAddr(address);

    notifyListeners();
  }

  void changeClientID(String clientID) {
    _session.changeClientID(clientID);

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
    required this.serverAddr,
    this.registered = false,
  });

  static Future<ProviderData> create() async {
    final pref = await SharedPreferences.getInstance();
    var addr = pref.getString("server_addr");

    if (addr == null) {
      addr = "127.0.0.1:16899";
      await pref.setString("server_addr", addr);
    }

    return ProviderData(serverAddr: addr);
  }

  Future<void> setServerAddr(String addr) async {
    final pref = await SharedPreferences.getInstance();
    await pref.setString("server_addr", addr);
  }

  void clear() {
    messages.clear();
    runtimeName = "";
    runtimeVersion = "";
    runtimeState = "";
    registered = false;
  }
}
