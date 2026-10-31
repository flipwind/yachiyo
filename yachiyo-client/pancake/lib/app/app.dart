import 'package:flutter/material.dart';
import 'package:pancake/application/settings/settings.dart';
import 'package:pancake/application/settings/settings_controller.dart';
import 'package:pancake/features/scaffold/pancake_scaffold.dart';
import 'package:provider/provider.dart';

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
    final settings = context.watch<SettingsController>();
    late ThemeMode themeMode;

    switch (settings.settings.themeMode) {
      case ThemeModeType.dark:
        themeMode = ThemeMode.dark;
      case ThemeModeType.light:
        themeMode = ThemeMode.light;
      case ThemeModeType.system:
        themeMode = ThemeMode.system;
    }

    return ListenableBuilder(
      listenable: settings,
      builder: (context, widget) {
        return MaterialApp(
          title: 'Pancake!',
          theme: createTheme(Brightness.light),
          darkTheme: createTheme(Brightness.dark),
          themeMode: themeMode,
          home: PancakeScaffold(),
        );
      },
    );
  }
}
