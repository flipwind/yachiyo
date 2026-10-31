import 'package:flutter/material.dart';
import 'package:pancake/application/settings/settings.dart';
import 'package:shared_preferences/shared_preferences.dart';

abstract final class _Types {
  static final String thememode = "settings.thememode";
  static final String reconnectOnStart = "settings.reconnectOnStart";
}

class SettingsStorage {
  final SharedPreferencesAsync _pref = SharedPreferencesAsync();

  Future<PancakeSettings> load() async {
    final theme = await _pref.getInt(_Types.thememode);
    final reconnect = await _pref.getBool(_Types.reconnectOnStart);

    final themeMode = ThemeModeType.values[theme ?? 2];
    return PancakeSettings(
      themeMode: themeMode,
      enableReconnectOnStart: reconnect ?? false,
    );
  }

  Future<void> save(PancakeSettings settings) async {
    await _pref.setInt(_Types.thememode, settings.themeMode.index);
    await _pref.setBool(
      _Types.reconnectOnStart,
      settings.enableReconnectOnStart,
    );
  }
}

class SettingsController extends ChangeNotifier {
  final SettingsStorage storage;
  PancakeSettings _settings = PancakeSettings();

  SettingsController(this.storage);

  PancakeSettings get settings => _settings;

  static Future<SettingsController> create() async {
    final storage = SettingsStorage();
    return SettingsController(storage);
  }

  Future<void> load() async {
    _settings = await storage.load();
    notifyListeners();
  }

  Future<void> _pending = Future<void>.value();

  Future<void> _pendingProcess(
    PancakeSettings Function(PancakeSettings current) change,
  ) {
    final operation = _pending.then<void>((_) async {
      final updated = change(_settings);
      await storage.save(updated);
      _settings = updated;
      notifyListeners();
    });

    _pending = operation.then<void>((_) {}, onError: (Object err) {});

    return operation;
  }

  Future<void> setThemeMode(ThemeModeType mode) async {
    return _pendingProcess((current) => current.copyWith(themeMode: mode));
  }

  Future<void> setEnableReconnectOnStart(bool choice) async {
    return _pendingProcess(
      (current) => current.copyWith(enableReconnectOnStart: choice),
    );
  }
}
