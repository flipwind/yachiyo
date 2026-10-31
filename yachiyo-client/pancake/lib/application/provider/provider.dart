import 'dart:async';

import 'package:flutter/material.dart';
import 'package:pancake/application/session/session.dart';
import 'package:pancake/core/identity/identity.dart';
import 'package:pancake/core/model/message.dart';
import 'package:pancake/core/model/state.dart';
import 'package:pancake/core/network/connection.dart';
import 'package:shared_preferences/shared_preferences.dart';

sealed class ProviderEvent {}

class EventSnackbarMessage extends ProviderEvent {
  final String message;
  EventSnackbarMessage(this.message);
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
          data.phase = RuntimePhase.connecting;

        case SessionRegistered():
          _data.runtimeName = event.runtimeName;
          _data.runtimeVersion = event.runtimeVersion;
          _data.registered = true;
          data.phase = RuntimePhase.online;
          data.registerTime = DateTime.now();
          _session.reloadMessagesFromRuntime();

        case SessionStateReceived():
          _data.runtimeState = event.state;

        case SessionMessageAppend():
          _data.messages.add(event.message);

        case SessionMessageReload():
          _data.messages = event.messages;

        case SessionErrorMessage():
          _eventController.add(EventSnackbarMessage(event.message));
        case SessionSelfErrorMessage():
          _pushSelfError(event.message);
      }

      notifyListeners();
    });
  }

  void _pushSelfError(String message) {
    _eventController.add(EventSnackbarMessage("pancake_error: $message"));
  }

  void pushSnackbarMessage(String message) {
    _eventController.add(EventSnackbarMessage(message));
  }

  static Future<PancakeProvider> create() async {
    final session = await PancakeSession.create();
    final data = await ProviderData.create();
    return PancakeProvider(session, data, session.identity);
  }

  Future<void> disconnect() async {
    await _session.disconnect();
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

  Future<void> changeServerAddr(String address) async {
    await _data.setServerAddr(address);
    notifyListeners();
  }

  Future<void> connect() async {
    final address = data.serverAddr;
    await _session.start(address);
  }

  void changeClientID(String clientID) {
    _session.changeClientID(clientID);

    notifyListeners();
  }

  void changeClientName(String name) {
    _session.changeClientName(name);

    notifyListeners();
  }

  Future<void> randomClientID() async {
    await _session.randomClientID();

    notifyListeners();
  }
}

enum RuntimePhase { offline, connecting, online }

class ProviderData {
  final SharedPreferencesAsync _pref = SharedPreferencesAsync();
  List<Message> messages = [];
  String runtimeName;
  String runtimeVersion;
  RuntimeState? runtimeState;
  String serverAddr;

  bool registered;
  DateTime? registerTime;
  RuntimePhase phase;

  String draftText;

  ProviderData({
    this.runtimeName = "",
    this.runtimeVersion = "",
    required this.serverAddr,
    this.registered = false,
    this.phase = RuntimePhase.offline,
    this.draftText = ""
  });

  static Future<ProviderData> create() async {
    final pref = SharedPreferencesAsync();
    String? addr = await pref.getString("server_addr");

    if (addr == null) {
      addr = "127.0.0.1:16899";
      await pref.setString("server_addr", addr);
    }

    return ProviderData(serverAddr: addr);
  }

  Future<void> setServerAddr(String addr) async {
    serverAddr = addr;
    await _pref.setString("server_addr", addr);
  }

  void changeDraftText(String value) {
    draftText = value;
  }

  void clear() {
    messages.clear();
    runtimeName = "";
    runtimeVersion = "";
    runtimeState = null;
    registered = false;
    registerTime = null;
    phase = RuntimePhase.offline;
  }
}
