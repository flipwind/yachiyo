import 'package:flutter/material.dart';
import 'package:pancake/features/home/home_page.dart';

class PancakeApp extends StatelessWidget {
  const PancakeApp({super.key});

  ThemeData createTheme(Brightness brightness) {
    final colors = ColorScheme.fromSeed(
      seedColor: const Color(0xFFa793ff),
      brightness: brightness,
    );

    return ThemeData(useMaterial3: true, colorScheme: colors);
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pancake!',
      theme: createTheme(Brightness.light),
      darkTheme: createTheme(Brightness.dark),
      themeMode: ThemeMode.system,
      home: PancakeHomePage(),
    );
  }
}
