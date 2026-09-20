import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:pancake/app/app.dart';
import 'package:pancake/core/model/state/yachiyo_state.dart';
import 'package:pancake/core/network/client.dart';
import 'package:pancake/core/provider/yachiyo_provider.dart';
import 'package:provider/provider.dart';

void main() {

  WidgetsFlutterBinding.ensureInitialized();

  final yachiyoClient = YachiyoClient();
  final yachiyoState = YachiyoState();

  debugPaintSizeEnabled = false;

  if (Platform.isAndroid) {
    SystemChrome.setSystemUIOverlayStyle(
      const SystemUiOverlayStyle(
        systemNavigationBarColor: Colors.transparent,
        systemNavigationBarContrastEnforced: false,
      ),
    );
  }
  
  runApp(
    MultiProvider(
      providers: [ChangeNotifierProvider(create: (_) => YachiyoProvider(client: yachiyoClient, state: yachiyoState))],
      child: PancakeApp(),
    ),
  );
}