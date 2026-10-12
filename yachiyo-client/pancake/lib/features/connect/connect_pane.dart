import 'package:flutter/material.dart';

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

  @override
  void initState() {
    super.initState();
    _addressEditingController.text = "ws://127.0.0.1/ws/";
    _nameEditingController.text = "Pancake!";
    _idEditingController.text = "c17439a5-b121-48b2-b933-de9787becc30";
  }

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;

    return Padding(
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
          ),
          Row(
            spacing: 8,
            children: [
              Expanded(
                child: FilledButton.icon(
                  onPressed: () {},
                  icon: Icon(Icons.north_east),
                  label: Text("Connect"),
                ),
              ),
              OutlinedButton(onPressed: () {}, child: Text("break")),
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
                    controller: _nameEditingController,
                    decoration: InputDecoration(
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                      labelText: "Client Name",
                    ),
                  ),
                  TextField(
                    style: texts.bodySmall,
                    controller: _idEditingController,
                    decoration: InputDecoration(
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
                      FilledButton(onPressed: () {}, child: Text("Save")),
                      TextButton(onPressed: () {}, child: Text("Random ID")),
                      IconButton(
                        onPressed: () {},
                        icon: Icon(Icons.copy),
                        iconSize: 16,
                      ),
                    ],
                  ),
                ],
              ),
            ],
          ),
          Card.filled(
            color: colors.secondaryContainer,
            child: Padding(
              padding: EdgeInsetsGeometry.symmetric(
                horizontal: 24,
                vertical: 16,
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                spacing: 4,
                children: [
                  Image.asset("assets/pancake.png", height: 64, width: 64),
                  Text("Now heartbeating.", style: texts.titleMedium),
                  SizedBox(height: 12),
                  for (var item in [
                    ("Runtime", "Yachiyo"),
                    ("Version", "v0.1.0"),
                    ("Last Update", "23:55"),
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
          ),
        ],
      ),
    );
  }
}
