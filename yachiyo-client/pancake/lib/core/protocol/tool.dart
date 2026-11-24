import 'package:pancake/core/protocol/envelope.dart';

abstract final class Tool {
  static const String updateTools = "update_tools";
  static const String updateToolsFeedback = "update_tools_feedback";
  static const String toolCall = "tool_call";
  static const String toolCallResult = "tool_call_result";
}

class ToolItem {
  final String name;
  final String description;
  final Map<String, dynamic> parameters;

  const ToolItem({
    required this.name,
    required this.description,
    required this.parameters,
  });

  factory ToolItem.fromJson(Map<String, dynamic> json) {
    return ToolItem(
      name: json["name"] as String,
      description: json["description"] as String,
      parameters: json["parameters"] as Map<String, dynamic>,
    );
  }

  Map<String, dynamic> toJson() {
    return {"name": name, "description": description, "parameters": parameters};
  }
}

class UpdateTools implements DataPack {
  final List<String> capability;
  final List<ToolItem> tools;

  const UpdateTools({required this.capability, required this.tools});

  factory UpdateTools.fromJson(Map<String, dynamic> json) {
    final capabilityList = json["capability"] as List<dynamic>;
    final toolList = json["tools"] as List<dynamic>;
    return UpdateTools(
      capability: List<String>.from(capabilityList.whereType<String>()),
      tools: toolList
          .map((tool) => ToolItem.fromJson(tool as Map<String, dynamic>))
          .toList(),
    );
  }

  @override
  String get type => Tool.updateTools;

  @override
  Map<String, dynamic> toJson() {
    return {
      "capability": capability,
      "tools": tools.map((tool) => tool.toJson()).toList(),
    };
  }
}

class UpdateToolsFeedback implements DataPack {
  final String status;
  final List<String> rejectCapability;
  final Map<String, String> rejectTools;

  const UpdateToolsFeedback({
    required this.status,
    required this.rejectCapability,
    required this.rejectTools,
  });

  factory UpdateToolsFeedback.fromJson(Map<String, dynamic> json) {
    final rejectCapability = json["reject_capability"] as List<dynamic>;
    final rejectTools = json["reject_tools"] as Map<String, dynamic>;
    return UpdateToolsFeedback(
      status: json["status"] as String,
      rejectCapability: List<String>.from(rejectCapability.whereType<String>()),
      rejectTools: Map<String, String>.from(rejectTools),
    );
  }

  @override
  String get type => Tool.updateToolsFeedback;

  @override
  Map<String, dynamic> toJson() {
    return {
      "status": status,
      "reject_capability": rejectCapability,
      "reject_tools": rejectTools,
    };
  }
}

class ToolCall implements DataPack {
  final String toolName;
  final String requestID;
  final Map<String, dynamic> payload;

  const ToolCall({
    required this.toolName,
    required this.requestID,
    required this.payload,
  });

  factory ToolCall.fromJson(Map<String, dynamic> json) {
    return ToolCall(
      toolName: json["tool_name"] as String,
      requestID: json["request_id"] as String,
      payload: json["payload"] as Map<String, dynamic>,
    );
  }

  @override
  String get type => Tool.toolCall;

  @override
  Map<String, dynamic> toJson() {
    return {"tool_name": toolName, "request_id": requestID, "payload": payload};
  }
}

class ToolCallResult implements DataPack {
  final String requestID;
  final bool ok;
  final String result;

  const ToolCallResult({
    required this.requestID,
    required this.ok,
    required this.result,
  });

  factory ToolCallResult.fromJson(Map<String, dynamic> json) {
    return ToolCallResult(
      requestID: json["request_id"] as String,
      ok: json["ok"] as bool,
      result: json["result"] as String,
    );
  }

  @override
  String get type => Tool.toolCallResult;

  @override
  Map<String, dynamic> toJson() {
    return {"request_id": requestID, "ok": ok, "result": result};
  }
}
