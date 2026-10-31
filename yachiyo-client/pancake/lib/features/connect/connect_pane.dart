import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:pancake/application/settings/settings.dart';
import 'package:pancake/application/settings/settings_controller.dart';
import 'package:pancake/core/network/connection.dart';
import 'package:provider/provider.dart';

class ConnectPane extends StatefulWidget {
  const ConnectPane({super.key});

  @override
  State<StatefulWidget> createState() => _ConnectPaneState();
}

class _ConnectPaneState extends State<ConnectPane> {
  final TextEditingController _addressEditingController =
      TextEditingController();
  final TextEditingController _nameEditingController = TextEditingController();
  final TextEditingController _idEditingController = TextEditingController();

  bool loading = false;
  String? _errorText;

  Future<void> onServerAddrChange() async {
    final status = context.read<PancakeProvider>().status;
    if (status == ConnectionStatus.connecting) return;
    setState(() {
      loading = true;
    });

    final provider = context.read<PancakeProvider>();
    final addr = _addressEditingController.text;
    await provider.changeServerAddr(addr);
    await provider.connect();

    setState(() {
      loading = false;
    });
  }

  void setErrorText(String? text) {
    setState(() {
      _errorText = text;
    });
  }

  Future<void> onIdentityChanged() async {
    setErrorText(null);
    final provider = context.read<PancakeProvider>();
    bool isUUID(String value) {
      return RegExp(
        r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$',
      ).hasMatch(value);
    }

    if (isUUID(_idEditingController.text) == false) {
      setErrorText("Client ID should be a UUID.");
      return;
    }

    provider.changeClientName(_nameEditingController.text);
    provider.changeClientID(_idEditingController.text);
    provider.pushSnackbarMessage("Successfully changed identity.");
  }

  Future<void> onClientIDRandom() async {
    final provider = context.read<PancakeProvider>();
    await provider.randomClientID();
    _idEditingController.text = provider.identity.clientID;
  }

  @override
  void initState() {
    super.initState();
    _addressEditingController.text = "ws://127.0.0.1:16899/ws/";
    _nameEditingController.text = "Pancake!";
    _idEditingController.text = "c17439a5-b121-48b2-b933-de9787becc30";
  }

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final registered = context.watch<PancakeProvider>().data.registered;
    final settingsController = context.watch<SettingsController>();
    final settings = settingsController.settings;
    return SingleChildScrollView(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Padding(
        padding: EdgeInsetsGeometry.symmetric(vertical: 12, horizontal: 20),
        child: Column(
          spacing: 16,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text("Connect to the runtime.", style: texts.titleLarge),
                Text("pursuing your communication.", style: texts.bodyMedium),
              ],
            ),
            TextField(
              decoration: InputDecoration(
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                prefixIcon: Icon(Icons.link),
                labelText: "Runtime Address",
              ),
              controller: _addressEditingController,
              onSubmitted: (value) => onServerAddrChange(),
            ),
            Row(
              spacing: 8,
              children: [
                Expanded(
                  child: FilledButton.icon(
                    onPressed: loading || registered
                        ? null
                        : () => onServerAddrChange(),
                    icon: loading
                        ? SizedBox(
                            width: 12,
                            height: 12,
                            child: CircularProgressIndicator(strokeWidth: 3),
                          )
                        : Icon(registered ? Icons.done : Icons.north_east),
                    label: Text(registered ? "Connected" : "Connect"),
                  ),
                ),
                Consumer<PancakeProvider>(
                  builder: (context, provider, child) {
                    return AnimatedSize(
                      duration: const Duration(milliseconds: 200),
                      curve: Curves.easeInOut,
                      child: SizedBox(
                        width: registered ? null : 0,
                        child: registered
                            ? OutlinedButton(
                                onPressed: () => provider.disconnect(),
                                child: Text("Break"),
                              )
                            : SizedBox.shrink(),
                      ),
                    );
                  },
                ),
              ],
            ),
            Divider(thickness: 0.5),
            ExpansionTile(
              leading: Icon(Icons.fingerprint),
              title: Text("Identity of this client", style: texts.titleMedium),
              shape: Border(),
              tilePadding: EdgeInsets.symmetric(),
              childrenPadding: EdgeInsetsGeometry.symmetric(vertical: 8),
              children: [
                Column(
                  spacing: 12,
                  children: [
                    TextField(
                      enabled: !registered,
                      controller: _nameEditingController,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                        labelText: "Client Name",
                      ),
                    ),
                    TextField(
                      enabled: !registered,
                      controller: _idEditingController,
                      decoration: InputDecoration(
                        errorText: _errorText,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                        labelText: "Client ID/UUID",
                      ),
                    ),
                    Text(
                      "The same Client ID refers to the same buddy for the runtime.",
                      style: texts.bodySmall,
                    ),

                    Row(
                      spacing: 8,
                      children: [
                        FilledButton(
                          onPressed: registered
                              ? null
                              : () => onIdentityChanged(),
                          child: Text("Save"),
                        ),
                        TextButton(
                          onPressed: registered
                              ? null
                              : () => onClientIDRandom(),
                          child: Text("Random ID"),
                        ),
                        IconButton(
                          onPressed: registered
                              ? null
                              : () async {
                                  await Clipboard.setData(
                                    ClipboardData(
                                      text: _idEditingController.text,
                                    ),
                                  );
                                },
                          icon: Icon(Icons.copy),
                          iconSize: 16,
                        ),
                        if (registered)
                          Text(
                            "*Disconnect to change.",
                            style: texts.labelSmall,
                          ),
                      ],
                    ),
                  ],
                ),
              ],
            ),
            _RuntimeIdentityWidget(),
            Divider(),
            Row(
              spacing: 12,
              children: [
                Icon(Icons.settings, size: 24),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text("Settings", style: texts.titleMedium),
                    Text(
                      "About application performance",
                      style: texts.bodyMedium,
                    ),
                  ],
                ),
              ],
            ),
            Column(
              children: [
                ListTile(
                  leading: Icon(Icons.refresh),
                  title: Text("Auto reconnect"),
                  subtitle: Text("Reconnect once when App starts."),
                  trailing: Switch(
                    value: settings.enableReconnectOnStart,
                    onChanged: (choice) {
                      settingsController.setEnableReconnectOnStart(choice);
                    },
                  ),
                ),
                ListTile(
                  leading: Icon(Icons.palette_outlined),
                  title: Text("Auto theme"),
                  subtitle: Text("Follow the system."),
                  trailing: DropdownMenu(
                    dropdownMenuEntries: <DropdownMenuEntry>[
                      DropdownMenuEntry(
                        value: ThemeModeType.system,
                        label: "System",
                        leadingIcon: Icon(Icons.contrast),
                      ),
                      DropdownMenuEntry(
                        value: ThemeModeType.dark,
                        label: "Dark",
                        leadingIcon: Icon(Icons.dark_mode_outlined),
                      ),
                      DropdownMenuEntry(
                        value: ThemeModeType.light,
                        label: "Light",
                        leadingIcon: Icon(Icons.light_mode_outlined),
                      ),
                    ],
                    onSelected: (value) {
                      settingsController.setThemeMode(value);
                    },
                    initialSelection: settings.themeMode,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _RuntimeIdentityWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;
    final data = context.watch<PancakeProvider>().data;

    return Card.filled(
      color: colors.secondaryContainer,
      child: Padding(
        padding: EdgeInsetsGeometry.symmetric(horizontal: 24, vertical: 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          spacing: 4,
          children: [
            Image.asset("assets/pancake.png", height: 64, width: 64),
            Text("Now heartbeating.", style: texts.titleMedium),
            SizedBox(height: 12),
            for (var item in [
              (
                "Runtime",
                data.runtimeName == "" ? "Maybe Yachiyo?" : data.runtimeName,
              ),
              (
                "Version",
                data.runtimeVersion == "" ? "Mystery!" : data.runtimeVersion,
              ),
              (
                "Last Update",
                data.registerTime == null
                    ? "Later"
                    : DateFormat("HH:mm").format(data.registerTime!),
              ),
            ])
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(
                      item.$1,
                      style: texts.bodyMedium!.copyWith(
                        color: colors.onSecondaryContainer,
                      ),
                    ),
                  ),
                  Expanded(child: Text(item.$2)),
                ],
              ),
          ],
        ),
      ),
    );
  }
}
