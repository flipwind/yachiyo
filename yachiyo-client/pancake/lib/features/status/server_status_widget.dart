import 'package:flutter/material.dart';
import 'package:logger/web.dart';
import 'package:pancake/core/model/state/yachiyo_state.dart';
import 'package:pancake/core/provider/yachiyo_provider.dart';
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

  Future<void> onServerAddrChange() async {
    setState(() {
      loading = true;
    });

    final provider = context.read<YachiyoProvider>();
    provider.state.network.serverAddr = _textEditingController.text;
    await provider.start();

    setState(() {
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final networkState = context.watch<YachiyoProvider>().state.network;
    final serverStatus = context.watch<YachiyoProvider>().state.status;
    _textEditingController.text = networkState.serverAddr;

    return Card.filled(
      margin: EdgeInsets.all(8.0),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(
            leading: (loading == false)
                ? Icon(
                    serverStatusIcon[serverStatus == YachiyoStatus.registered
                        ? 1
                        : 0],
                  )
                : CircularProgressIndicator(),
            title: const Text("Server Status"),
            subtitle: Text(
              serverStatus == YachiyoStatus.registered
                  ? "Registered"
                  : "Unregistered",
            ),
          ),
          Padding(
            padding: EdgeInsets.fromLTRB(8.0, 0, 8.0, 12.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    keyboardType: TextInputType.url,
                    controller: _textEditingController,
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
                      networkState.serverAddr = defaultServerAddr;
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

class ServerStatusBadge extends StatefulWidget {
  const ServerStatusBadge({super.key});

  @override
  State<ServerStatusBadge> createState() => _ServerStatusBadgeState();
}

class _ServerStatusBadgeState extends State<ServerStatusBadge> {
  final List<IconData> serverStatusIcon = [
    Icons.cloud_off,
    Icons.cloud_outlined,
  ];

  bool loading = false;

  @override
  Widget build(BuildContext context) {
    final serverStatus = context.watch<YachiyoProvider>().state.status;

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
              child: ServerStatusWidget(),
            );
          },
        );
      },
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            serverStatus == YachiyoStatus.registered
                ? "Registered"
                : "Unregistered",
          ),
          Icon(Icons.arrow_drop_down_rounded),
        ],
      ),
      icon: (loading == false)
          ? Icon(
              serverStatusIcon[serverStatus == YachiyoStatus.registered
                  ? 1
                  : 0],
            )
          : CircularProgressIndicator(),
    );
  }
}
