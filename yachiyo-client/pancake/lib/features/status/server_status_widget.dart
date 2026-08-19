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
  final List<IconData> serverStatusIcon = [Icons.cloud_off, Icons.cloud_outlined];

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
    return Card.outlined(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          ListTile(
            leading: (loading == false)? Icon(serverStatusIcon[serverStatus == YachiyoStatus.registered ? 1 : 0]) : CircularProgressIndicator(),
            title: const Text("Server Status"),
            subtitle: Text(serverStatus == YachiyoStatus.registered ? "Registered" : "Unregistered"),
          ),
          Padding(
            padding: EdgeInsets.fromLTRB(8.0, 0, 8.0, 12.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _textEditingController,
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
                IconButton(onPressed: () {
                  if (_textEditingController.text == "") {
                    String defaultServerAddr = "127.0.0.1:16899";
                    _textEditingController.text = defaultServerAddr;
                    networkState.serverAddr = defaultServerAddr;
                  }
                  onServerAddrChange();
                }, icon: Icon(Icons.refresh)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
