import 'package:pancake/core/model/protocol/parser.dart';

class Envelope {
  final String category;
  final String type;
  final DataPack data;

  Envelope({required this.category, required this.type, required this.data});

  factory Envelope.fromJson(Map<String, dynamic> json) {
    return Envelope(
      category: json["category"] as String,
      type: json["type"] as String,
      data: DataPackParser.decode(
        json["type"] as String,
        json["data"] as Map<String, dynamic>,
      ),
    );
  }

  Map<String, dynamic> toJson() {
    return {"category": category, "type": type, "data": data.toJson()};
  }
}

abstract interface class DataPack {
  String get type;
  Map<String, dynamic> toJson();
}
