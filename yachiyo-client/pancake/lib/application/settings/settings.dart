enum ThemeModeType { dark, light, system }

class PancakeSettings {
  final ThemeModeType themeMode;
  final bool enableReconnectOnStart;

  PancakeSettings({
    this.themeMode = ThemeModeType.system,
    this.enableReconnectOnStart = false,
  });

  PancakeSettings copyWith({
    ThemeModeType? themeMode,
    bool? enableReconnectOnStart,
  }) {
    return PancakeSettings(
      themeMode: themeMode ?? this.themeMode,
      enableReconnectOnStart:
          enableReconnectOnStart ?? this.enableReconnectOnStart,
    );
  }
}
