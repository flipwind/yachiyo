import 'dart:async';
import 'dart:convert';

import 'package:logger/logger.dart';
import 'package:pancake/core/identity/identity.dart';
import 'package:pancake/core/model/message.dart';
import 'package:pancake/core/network/connection.dart';
import 'package:pancake/core/protocol/connection.dart';
import 'package:pancake/core/protocol/envelope.dart';
import 'package:pancake/core/protocol/interaction.dart';
import 'package:pancake/core/protocol/state.dart';
import 'package:uuid/uuid.dart';

sealed class SessionEvent {}

class SessionDisconnected extends SessionEvent {}

class SessionConnecting extends SessionEvent {}

class SessionConnected extends SessionEvent {}

class SessionRegistered extends SessionEvent {
  final String runtimeName;
  final String runtimeVersion;
  SessionRegistered(this.runtimeName, this.runtimeVersion);
}

class SessionStateReceived extends SessionEvent {
  final String state;
  SessionStateReceived(this.state);
}

class SessionMessageAppend extends SessionEvent {
  final Message message;
  SessionMessageAppend(this.message);
}

class SessionMessageReload extends SessionEvent {
  final List<Message> messages;
  SessionMessageReload(this.messages);
}

class SessionErrorMessage extends SessionEvent {
  final String message;
  SessionErrorMessage(this.message);
}

class SessionSelfErrorMessage extends SessionEvent {
  final String message;
  SessionSelfErrorMessage(this.message);
}

class PancakeSession {
  final PancakeConnection _connection = PancakeConnection();
  final Identity identity;

  final _eventController = StreamController<SessionEvent>.broadcast();
  Stream<SessionEvent> get events => _eventController.stream;

  Timer? _heartbeatTimer;
  Timer? _stateTimer;

  PancakeSession({required this.identity}) {
    _connection.rawMessages.listen(_handleMessage);

    _connection.status.listen((status) {
      switch (status) {
        case ConnectionStatus.connected:
          _pushSessionEvent(SessionConnected());
        case ConnectionStatus.connecting:
          _pushSessionEvent(SessionConnecting());
        case ConnectionStatus.disconnected:
          _sessionDisconnected();
      }
    });

    _connection.errors.listen((errorMessage) {
      _pushSessionEvent(SessionErrorMessage(errorMessage));
    });
  }

  static Future<PancakeSession> create() async {
    final Identity identity = await Identity.create();

    return PancakeSession(identity: identity);
  }

  Future<void> disconnect() async {
    await _connection.disconnect();
  }

  Future<void> changeClientID(String id) async {
    bool isUUID(String value) {
      return RegExp(
        r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$',
      ).hasMatch(value);
    }

    if (!isUUID(id)) {
      _pushSessionEvent(
        SessionSelfErrorMessage("Entered ClientID is not a valid uuid."),
      );
      return;
    }

    await identity.changeID(id);
  }

  Future<void> changeClientName(String name) async {
    await identity.changeName(name);
  }

  Future<void> randomClientID() async {
    await changeClientID(Uuid().v4());
  }

  void _sessionDisconnected() {
    _pushSessionEvent(SessionDisconnected());
    _stopHeartbeat();
    _stopChangeState();
  }

  Future<void> start(String address) async {
    await _connection.connect(address);
    _register();
  }

  void _register() {
    _sendEnvelope(
      Envelope(
        category: "connection",
        type: "register",
        data: Register(
          clientType: identity.clientType,
          clientName: identity.clientName,
          clientID: identity.clientID,
        ),
      ),
    );
  }

  void _pushSessionEvent(SessionEvent event) {
    _eventController.add(event);
  }

  // Message Handler
  void _handleMessage(String rawMessage) {
    final logger = Logger();
    logger.d("Processing json: $rawMessage");

    final Map<String, dynamic> parsedJson = jsonDecode(rawMessage);
    final envelope = Envelope.fromJson(parsedJson);
    _handleEnvelope(envelope);
  }

  void _handleEnvelope(Envelope envelope) {
    switch (envelope.category) {
      case "connection":
        _processConnection(envelope.data);
      case "interaction":
        _processInteraction(envelope.data);
      case "state":
        _processState(envelope.data);
    }
  }

  void _processConnection(DataPack data) {
    switch (data) {
      case RegisterSuccess():
        _pushSessionEvent(
          SessionRegistered(data.runtimeName, data.runtimeVersion),
        );
        _startHeartbeat();
        _startChangeState();
      case RegisterError():
        switch (data.errorType) {
          case "client_conflict":
            randomClientID();
          default:
            _pushSessionEvent(
              SessionErrorMessage("register_error: ${data.errorType}"),
            );
            _sessionDisconnected();
        }
    }
  }

  void _processInteraction(DataPack data) async {
    int clamp(int min, max, n) {
      if (n < min) return min;
      if (n > max) return max;
      return n;
    }

    switch (data) {
      case RuntimeMessage():
        for (var i = 0; i < data.message.length; i++) {
          final msg = data.message[i];

          if (i != 0) {
            await Future.delayed(
              Duration(milliseconds: clamp(300, 5000, 200 + 100 * msg.length)),
            );
          }
          _pushSessionEvent(
            SessionMessageAppend(
              AssistantMessage(
                content: msg,
                time: DateTime.fromMillisecondsSinceEpoch(data.time * 1000),
                reply: data.reply,
                initiative: data.isInitiative,
              ),
            ),
          );
        }
      case RelativeMessageHistory():
        List<Message> history = [];
        for (final messageDataPack in data.messages) {
          switch (messageDataPack) {
            case RuntimeMessage():
              for (final datapack in messageDataPack.message) {
                history.add(
                  AssistantMessage(
                    content: datapack,
                    time: DateTime.fromMillisecondsSinceEpoch(
                      messageDataPack.time * 1000,
                    ),
                    reply: messageDataPack.reply,
                    initiative: messageDataPack.isInitiative,
                  ),
                );
              }
            case ClientMessage():
              history.add(
                UserMessage(
                  content: messageDataPack.message,
                  time: DateTime.fromMillisecondsSinceEpoch(
                    messageDataPack.time * 1000,
                  ),
                ),
              );
            default:
              throw FormatException(
                "Unknown message type: ${messageDataPack.type}",
              );
          }
        }
        _pushSessionEvent(SessionMessageReload(history));
      case ErrorMessage():
        _pushSessionEvent(SessionErrorMessage("${data.code}: ${data.message}"));
    }
  }

  void _processState(DataPack data) {
    switch (data) {
      case RuntimeState():
        _pushSessionEvent(SessionStateReceived(data.state));
    }
  }

  // Sending Handler
  void _sendEnvelope(Envelope envelope) {
    final data = envelope.toJson();
    _connection.sendRaw(jsonEncode(data));
  }

  void sendMessage(String message) {
    _sendEnvelope(
      Envelope(
        category: "interaction",
        type: "client_message",
        data: ClientMessage(
          message: message,
          time: DateTime.now().millisecondsSinceEpoch ~/ 1000,
        ),
      ),
    );
    _pushSessionEvent(
      SessionMessageAppend(UserMessage(content: message, time: DateTime.now())),
    );
  }

  void reloadMessagesFromRuntime() {
    _sendEnvelope(
      Envelope(
        category: "interaction",
        type: "get_relative_message_history",
        data: GetRelativeMessageHistory(),
      ),
    );
  }

  // Timer
  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = Timer.periodic(
      const Duration(seconds: 20),
      (_) => _sendEnvelope(
        Envelope(category: "connection", type: "heartbeat", data: Heartbeat()),
      ),
    );
  }

  void _stopHeartbeat() {
    _heartbeatTimer?.cancel();
  }

  void _startChangeState() {
    _stateTimer?.cancel();
    _stateTimer = Timer.periodic(
      const Duration(seconds: 1),
      (_) => _sendEnvelope(
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
}
