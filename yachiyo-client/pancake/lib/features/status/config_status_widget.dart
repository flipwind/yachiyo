import 'package:flutter/material.dart';
import 'package:logger/web.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:pancake/core/network/connection.dart';
import 'package:provider/provider.dart';

class ServerStatusWidget extends StatefulWidget {
  const ServerStatusWidget({super.key});

  @override
  State<ServerStatusWidget> createState() => _ServerStatusWidgetState();
}

class _ServerStatusWidgetState extends State<ServerStatusWidget> {
  final logger = Logger();
  final List<IconData> serverStatusIcon = [
    Icons.cloud_off,
    Icons.cloud_outlined,
  ];

  bool loading = false;

  final TextEditingController _textEditingController = TextEditingController();
  final FocusNode _focusNode = FocusNode();

  Future<void> onServerAddrChange() async {
    final status = context.read<PancakeProvider>().status;
    if (status == ConnectionStatus.connecting) return;
    setState(() {
      loading = true;
    });

    final provider = context.read<PancakeProvider>();
    final addr = _textEditingController.text;
    provider.changeServerAddr(addr);

    setState(() {
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final data = context.watch<PancakeProvider>().data;
    final registered = context.watch<PancakeProvider>().data.registered;
    if (_focusNode.hasFocus == false) {
      _textEditingController.text = data.serverAddr;
    }

    return Card.filled(
      margin: EdgeInsets.all(8.0),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(
            leading: (loading == false)
                ? Icon(serverStatusIcon[registered ? 1 : 0])
                : CircularProgressIndicator(),
            title: const Text("Server Status"),
            subtitle: Text(registered ? "Registered" : "Unregistered"),
          ),
          Padding(
            padding: EdgeInsets.fromLTRB(8.0, 0, 8.0, 12.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    keyboardType: TextInputType.url,
                    controller: _textEditingController,
                    focusNode: _focusNode,
                    autocorrect: false,
                    onSubmitted: (value) {
                      onServerAddrChange();
                    },
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: "Server Address",
                      hintText: "127.0.0.1:16899",
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () {
                    if (_textEditingController.text == "") {
                      String defaultServerAddr = "127.0.0.1:16899";
                      _textEditingController.text = defaultServerAddr;
                    }
                    onServerAddrChange();
                  },
                  icon: Icon(Icons.refresh),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class ClientIDWidget extends StatefulWidget {
  const ClientIDWidget({super.key});

  @override
  State<ClientIDWidget> createState() => _ClientIDWidgetState();
}

class _ClientIDWidgetState extends State<ClientIDWidget> {
  final TextEditingController _textEditingController = TextEditingController();
  final FocusNode _focusNode = FocusNode();

  String? _errorText;

  void setErrorText(String? text) {
    setState(() {
      _errorText = text;
    });
  }

  Future<void> onClientIDChanged() async {
    setErrorText(null);
    final provider = context.read<PancakeProvider>();
    bool isUUID(String value) {
      return RegExp(
        r'^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$',
      ).hasMatch(value);
    }

    if (isUUID(_textEditingController.text) == false){
      setErrorText("Client ID should be a UUID.");
      return;
    }
    provider.changeClientID(_textEditingController.text);
  }

  Future<void> onClientIDRandom() async {
    final provider = context.read<PancakeProvider>();
    await provider.randomClientID();
    _textEditingController.text = provider.identity.clientID;
  }

  @override
  void initState() {
    super.initState();
    final identity = context.read<PancakeProvider>().identity;
    _textEditingController.text = identity.clientID;
  }

  @override
  Widget build(BuildContext context) {
    return Card.filled(
      margin: EdgeInsets.symmetric(horizontal: 8.0),
      child: Padding(
        padding: EdgeInsetsGeometry.all(8.0),
        child: Row(
          children: [
            Padding(
              padding: EdgeInsetsGeometry.symmetric(horizontal: 8.0),
              child: Icon(Icons.fingerprint_rounded),
            ),
            Expanded(
              child: TextField(
                keyboardType: TextInputType.text,
                controller: _textEditingController,
                focusNode: _focusNode,
                autocorrect: false,
                onSubmitted: (value) {
                  onClientIDChanged();
                },
                decoration: InputDecoration(
                  errorText: _errorText,
                  border: OutlineInputBorder(),
                  labelText: "Client ID",
                ),
              ),
            ),
            IconButton(onPressed: () {
              onClientIDRandom();
            }, icon: Icon(Icons.shuffle_rounded)),
            IconButton(
              onPressed: () {
                onClientIDChanged();
              },
              icon: Icon(Icons.check_rounded),
            ),
          ],
        ),
      ),
    );
  }
}

class ConfigStatusBadge extends StatefulWidget {
  const ConfigStatusBadge({super.key});

  @override
  State<ConfigStatusBadge> createState() => _ConfigStatusBadgeState();
}

class _ConfigStatusBadgeState extends State<ConfigStatusBadge> {
  final List<IconData> serverStatusIcon = [
    Icons.cloud_off,
    Icons.cloud_outlined,
  ];

  bool loading = false;

  @override
  Widget build(BuildContext context) {
    final registered = context.watch<PancakeProvider>().data.registered;

    final ThemeData theme = Theme.of(context);
    final ColorScheme colorScheme = theme.colorScheme;

    return TextButton.icon(
      style: TextButton.styleFrom(
        foregroundColor: colorScheme.onPrimary,
        backgroundColor: colorScheme.primary,

        padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        minimumSize: Size.zero,
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      ),
      onPressed: () {
        showModalBottomSheet(
          showDragHandle: true,
          isScrollControlled: true,
          context: context,
          builder: (context) {
            return Padding(
              padding: EdgeInsets.only(
                bottom: MediaQuery.viewInsetsOf(context).bottom,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                spacing: 4.0,
                children: [ClientIDWidget(), ServerStatusWidget()],
              ),
            );
          },
        );
      },
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(registered ? "Registered" : "Unregistered"),
          Icon(Icons.arrow_drop_down_rounded),
        ],
      ),
      icon: (loading == false)
          ? Icon(serverStatusIcon[registered ? 1 : 0])
          : CircularProgressIndicator(),
    );
  }
}
