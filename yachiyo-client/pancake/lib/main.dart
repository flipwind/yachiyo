import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:pancake/app/app.dart';
import 'package:pancake/core/network/client.dart';

void main() {
  final yachiyoClient = YachiyoClient();
  yachiyoClient.connect();

  debugPaintSizeEnabled = false;
  runApp(PancakeApp());
}