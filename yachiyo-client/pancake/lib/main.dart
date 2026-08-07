import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:pancake/app/app.dart';
import 'package:pancake/core/model/yachiyo_model.dart';
import 'package:pancake/core/network/client.dart';
import 'package:provider/provider.dart';

void main() {
  final yachiyoClient = YachiyoClient();

  debugPaintSizeEnabled = false;
  runApp(
    MultiProvider(
      providers: [ChangeNotifierProvider(create: (_) => YachiyoModel(yachiyoClient))],
      child: PancakeApp(),
    ),
  );
}