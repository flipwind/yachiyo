import 'package:flutter/material.dart';
import 'package:logger/web.dart';
import 'package:pancake/core/model/yachiyo_model.dart';
import 'package:provider/provider.dart';

class ServerStatusWidget extends StatefulWidget {
  const ServerStatusWidget({super.key});

  @override
  State<ServerStatusWidget> createState() => _ServerStatusWidgetState();
}

class _ServerStatusWidgetState extends State<ServerStatusWidget> {
  final logger = Logger();
  final List<IconData> serverStatusIcon = [Icons.cloud_off, Icons.cloud_outlined];
  int serverConnected = 0;
  String serverAddr = "";

  bool loading = false;

  final TextEditingController _textEditingController = TextEditingController();

  Future<void> onServerAddrChange() async {
    setState(() {
      loading = true;
    });
    var result = await context.read<YachiyoModel>().onServerAddrChange(serverAddr);
    setState(() {
      logger.d("setstate serverconnected $result");
      if (result == false) {
        serverConnected = 0;
      } else {
        serverConnected = 1;
      }
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Card.outlined(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          ListTile(
            leading: (loading == false)? Icon(serverStatusIcon[serverConnected]) : CircularProgressIndicator(),
            title: const Text("Server Status"),
            subtitle: Text(serverConnected == 1 ? "Connected" : "Unconnected"),
          ),
          Padding(
            padding: EdgeInsets.fromLTRB(8.0, 0, 8.0, 12.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _textEditingController,
                    onChanged: (value) {
                      serverAddr = value;
                    },
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
                  if (serverAddr == "") {
                    String defaultServerAddr = "127.0.0.1:16899";
                    _textEditingController.text = defaultServerAddr;
                    serverAddr = defaultServerAddr;
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
