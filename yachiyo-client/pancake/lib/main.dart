import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:pancake/app/app.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:provider/provider.dart';

Future<void> main() async{
  WidgetsFlutterBinding.ensureInitialized();

  final provider = await PancakeProvider.create();

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
      providers: [ChangeNotifierProvider.value(value: provider)],
      child: PancakeApp(),
    ),
  );
}