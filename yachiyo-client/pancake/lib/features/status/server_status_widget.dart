import 'package:flutter/material.dart';
import 'package:logger/web.dart';
import 'package:pancake/application/provider/provider.dart';
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
              child: ServerStatusWidget(),
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
