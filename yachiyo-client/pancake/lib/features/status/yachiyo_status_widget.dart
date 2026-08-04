import 'package:flutter/material.dart';

class YachiyoStatusWidget extends StatefulWidget {
  const YachiyoStatusWidget({super.key});

  @override
  State<YachiyoStatusWidget> createState() => _YachiyoStatusWidgetState();
}

class _YachiyoStatusWidgetState extends State<YachiyoStatusWidget> {
  @override
  Widget build(BuildContext context) {
    return ExpansionTile(
      shape: const Border(),
      collapsedShape: const Border(),
      leading: Icon(Icons.animation_outlined),
      title: Text("Yachiyo Status"),
      children: [Text("Hello Yachiyo")],
    );
  }
}
