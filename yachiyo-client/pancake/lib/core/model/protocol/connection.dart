import 'package:pancake/core/model/protocol/envelope.dart';

abstract final class Connection {
  static final String register = "register";
  static final String registerSuccess = "register_success";
  static final String registerError = "register_error";
  static final String heartbeat = "heartbeat";
  static final String heartbeatRespond = "heartbeatRespond";
  static final String offline = "offline";
}

class Register implements DataPack {
  final String clientType;
  final String clientName;
  final String clientID;

  const Register({
    required this.clientType,
    required this.clientName,
    required this.clientID,
  });

  factory Register.fromJson(Map<String, dynamic> json) {
    return Register(
      clientType: json["client_type"] as String,
      clientName: json["client_name"] as String,
      clientID: json["client_id"] as String,
    );
  }

  @override
  String get type => Connection.register;

  @override
  Map<String, dynamic> toJson() {
    return {
      "client_type": clientType,
      "client_name": clientName,
      "client_id": clientID,
    };
  }
}

class RegisterSuccess implements DataPack {
  final String runtimeName;
  final String runtimeVersion;

  const RegisterSuccess({
    required this.runtimeName,
    required this.runtimeVersion,
  });

  factory RegisterSuccess.fromJson(Map<String, dynamic> json) {
    return RegisterSuccess(
      runtimeName: json["runtime_name"] as String,
      runtimeVersion: json["runtime_version"] as String,
    );
  }

  @override
  String get type => Connection.registerSuccess;

  @override
  Map<String, dynamic> toJson() {
    return {
      "runtime_name": runtimeName,
      "runtime_version": runtimeVersion,
    };
  }
}

class RegisterError implements DataPack {
  final String errorType;

  const RegisterError({
    required this.errorType,
  });

  factory RegisterError.fromJson(Map<String, dynamic> json) {
    return RegisterError(
      errorType: json["error_type"] as String,
    );
  }

  @override
  String get type => Connection.registerError;

  @override
  Map<String, dynamic> toJson() {
    return {
      "error_type": errorType,
    };
  }
}

class Heartbeat implements DataPack {
  const Heartbeat();

  factory Heartbeat.fromJson(Map<String, dynamic> json) {
    return Heartbeat();
  }

  @override
  String get type => Connection.heartbeat;

  @override
  Map<String, dynamic> toJson() {
    return {};
  }
}

class HeartbeatRespond implements DataPack {
  const HeartbeatRespond();

  factory HeartbeatRespond.fromJson(Map<String, dynamic> json) {
    return HeartbeatRespond();
  }

  @override
  String get type => Connection.heartbeatRespond;

  @override
  Map<String, dynamic> toJson() {
    return {};
  }
}

class Offline implements DataPack {
  const Offline();

  factory Offline.fromJson(Map<String, dynamic> json) {
    return Offline();
  }

  @override
  String get type => Connection.offline;

  @override
  Map<String, dynamic> toJson() {
    return {};
  }
}
