import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:pancake/core/model/protocol/envelope.dart';
import 'package:pancake/core/model/protocol/interaction.dart';
import 'package:pancake/core/model/protocol/state.dart';
import 'package:pancake/core/model/state/yachiyo_state.dart';
import 'package:pancake/core/network/client.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

import '../model/message.dart';
import '../model/protocol/connection.dart';

class YachiyoProvider extends ChangeNotifier {
  late final StreamSubscription<String> _messageSubscription;
  Timer? _heartbeatTimer;
  Timer? _stateTimer;

  final YachiyoClient client;
  final YachiyoState state;

  YachiyoProvider({required this.client, required this.state}) {
    client.setYachiyoState(state);
    _messageSubscription = client.messages.listen(_handleMessage);
  }

  // Receiving

  void _handleMessage(String message) {
    final Map<String, dynamic> parsedJson = jsonDecode(message);
    final envelope = Envelope.fromJson(parsedJson);
    _handlingEnvelope(envelope);

    notifyListeners();
  }

  void _handlingEnvelope(Envelope envelope) {
    switch (envelope.category) {
      case "connection":
        _processConnection(envelope);
      case "interaction":
        _processInteraction(envelope);
      case "state":
        _processState(envelope);
    }
  }

  void _processConnection(Envelope envelope) {
    final envelopeData = envelope.data;
    switch (envelopeData) {
      case RegisterSuccess():
        _setState(YachiyoStatus.registered);
        state.runtime.runtimeName = envelopeData.runtimeName;
        state.runtime.runtimeVersion = envelopeData.runtimeVersion;
        notifyListeners();
      case RegisterError():
        switch (envelopeData.errorType) {
          case "client_conflict": reGenerateClientID();
          default: throw Exception("Unexcepted error when register: ${envelopeData.errorType}");
        }
    }
  }

  void _processInteraction(Envelope envelope) {
    final envelopeData = envelope.data;
    switch (envelopeData) {
      case RuntimeMessage():
        state.runtime.messages.add(
          Message(
            reply: envelopeData.reply,
            role: "assistant",
            message: envelopeData.message,
            time: DateTime.now(),
          ),
        );
      case RelativeMessageHistory():
        state.runtime.messages = envelopeData.messages.map((message) {
          switch (message) {
            case RuntimeMessage():
              return Message(
                reply: message.reply,
                role: "assistant",
                message: message.message,
                time: DateTime.fromMillisecondsSinceEpoch(message.time * 1000),
              );
            case ClientMessage():
              return Message(
                role: "user",
                message: message.message,
                time: DateTime.fromMillisecondsSinceEpoch(message.time * 1000),
              );
            default:
              throw FormatException("Unknown message type: ${message.type}");
          }
        }).toList();
    }
  }

  void _processState(Envelope envelope) {
    final envelopeData = envelope.data;
    switch (envelopeData) {
      case RuntimeState():
        state.runtime.state = envelopeData.state;
        notifyListeners();
    }
  }

  // Sending

  void _sendMessage(Envelope envelope) {
    final data = envelope.toJson();
    client.send(jsonEncode(data));
  }

  // network
  Future<void> reGenerateClientID() async {
    final pref = await SharedPreferences.getInstance();
    var clientID = Uuid().v4();
    await pref.setString("client_id", clientID);

    await start();
  }

  Future<void> start() async {
    final pref = await SharedPreferences.getInstance();
    var clientID = pref.getString("client_id");

    if (clientID == null) {
      clientID = Uuid().v4();
      await pref.setString("client_id", clientID);
    }

    state.client.clientID = clientID;
    _setState(YachiyoStatus.connecting);
    try {
      await client.connect();
      _setState(YachiyoStatus.connected);
      await _register();
    } catch (_) {
      _setState(YachiyoStatus.disconnected);
    }
  }

  Future<void> _register() async {
    _setState(YachiyoStatus.registering);
    _sendMessage(
      Envelope(
        category: "connection",
        type: "register",
        data: Register(
          clientType: state.client.clientType,
          clientName: state.client.clientName,
          clientID: state.client.clientID,
        ),
      ),
    );
  }

  Future<void> connect() {
    return client.connect();
  }

  // heartbeat

  void _startHeartbeat() {
    _heartbeatTimer?.cancel();

    _heartbeatTimer = Timer.periodic(
      const Duration(seconds: 20),
      (_) => _sendMessage(
        Envelope(category: "connection", type: "heartbeat", data: Heartbeat()),
      ),
    );
  }

  void _stopHeartbeat() {
    _heartbeatTimer?.cancel();
  }

  // state change

  void _startChangeState() {
    _stateTimer?.cancel();

    _stateTimer = Timer.periodic(
      const Duration(seconds: 1),
      (_) => _sendMessage(
        Envelope(
          category: "state",
          type: "runtime_state_request",
          data: RuntimeStateRequest(),
        ),
      ),
    );
  }

  void _stopChangeState() {
    _stateTimer?.cancel();
  }

  // message

  void sendMessage(String message) {
    _sendMessage(
      Envelope(
        category: "interaction",
        type: "client_message",
        data: ClientMessage(
          message: message,
          time: DateTime.now().millisecondsSinceEpoch ~/ 1000,
        ),
      ),
    );
  }

  // utils

  void _setState(YachiyoStatus status) {
    state.status = status;

    switch (status) {
      case YachiyoStatus.disconnected:
        _stopHeartbeat();
        _stopChangeState();
      case YachiyoStatus.registered:
        _startHeartbeat();
        _startChangeState();
      default:
        break;
    }
  }

  void clearMessages() {
    state.runtime.messages.clear();
  }

  void refreshMessagesFromRuntime() {
    _sendMessage(
      Envelope(
        category: "interaction",
        type: "get_relative_message_history",
        data: GetRelativeMessageHistory(),
      ),
    );
  }

  @override
  void dispose() {
    _messageSubscription.cancel();
    super.dispose();
  }
}
