import 'package:pancake/core/protocol/envelope.dart';

abstract final class State {
  static final String runtimeStateRequest = "runtime_state_request";
  static final String runtimeState = "runtime_state";
}

class RuntimeStateRequest implements DataPack {
  const RuntimeStateRequest();

  factory RuntimeStateRequest.fromJson(Map<String, dynamic> json) {
    return RuntimeStateRequest();
  }

  @override
  String get type => State.runtimeStateRequest;

  @override
  Map<String, dynamic> toJson() {
    return {};
  }
}

class RuntimeStateData implements DataPack {
  final int timetick;
  final double energy;
  final RuntimeAffectData affect;

  RuntimeStateData({
    required this.timetick,
    required this.energy,
    required this.affect,
  });

  factory RuntimeStateData.fromJson(Map<String, dynamic> json) {
    return RuntimeStateData(
      timetick: json["timetick"] as int,
      energy: (json["energy"] as num).toDouble(),
      affect: RuntimeAffectData.fromJson(json["affect"] as Map<String, dynamic>),
    );
  }

  @override
  String get type => State.runtimeState;

  @override
  Map<String, dynamic> toJson() {
    return {
      "timetick": timetick,
      "energy": energy,
      "affect": affect.toJson(),
    };
  }
}

class RuntimeAffectData {
  final String octants;
  final double pleasure;
  final double arousal;
  final double dominance;
  final List<OCCEmotionData> actives;

  RuntimeAffectData({
    required this.octants,
    required this.pleasure,
    required this.arousal,
    required this.dominance,
    required this.actives,
  });

  factory RuntimeAffectData.fromJson(Map<String, dynamic> json) {
    return RuntimeAffectData(
      octants: json["octants"] as String,
      pleasure: (json["pleasure"] as num).toDouble(),
      arousal: (json["arousal"] as num).toDouble(),
      dominance: (json["dominance"] as num).toDouble(),
      actives: (json["actives"] as List<dynamic>)
          .map(
            (emotion) => OCCEmotionData.fromJson(emotion as Map<String, dynamic>),
          )
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      "octants": octants,
      "pleasure": pleasure,
      "arousal": arousal,
      "dominance": dominance,
      "actives": actives.map((item) => item.toJson()).toList(),
    };
  }
}

class OCCEmotionData {
  final String type;
  final double intensity;

  OCCEmotionData({required this.type, required this.intensity});

  factory OCCEmotionData.fromJson(Map<String, dynamic> json) {
    return OCCEmotionData(
      type: json["type"] as String,
      intensity: (json["intensity"] as num).toDouble(),
    );
  }

  Map<String, dynamic> toJson() {
    return {"type": type, "intensity": intensity};
  }
}
