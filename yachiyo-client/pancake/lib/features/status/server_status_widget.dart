import 'package:flutter/material.dart';

class ServerStatusWidget extends StatefulWidget {
  const ServerStatusWidget({super.key});

  @override
  State<ServerStatusWidget> createState() => _ServerStatusWidgetState();
}

class _ServerStatusWidgetState extends State<ServerStatusWidget> {
  final List<IconData> serverStatusIcon = [Icons.cloud_off, Icons.cloud];
  int serverConnected = 0;

  @override
  Widget build(BuildContext context) {
    return Card.outlined(
      margin: EdgeInsets.all(8.0),
      child: Column(
        children: [
          ListTile(
            leading: Icon(serverStatusIcon[serverConnected]),
            title: const Text("Server Status"),
            subtitle: Text(serverConnected == 1 ? "Connected" : "Unconnected"),
          ),
          Padding(
            padding: EdgeInsets.fromLTRB(8.0, 0, 8.0, 12.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: "Server Address",
                      hintText: "192.168.0.1:16802",
                    ),
                  ),
                ),
                IconButton(onPressed: () => {}, icon: Icon(Icons.refresh)),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
