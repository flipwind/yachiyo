import 'package:flutter/material.dart';
import 'package:pancake/features/chat/chat_widget.dart';
import 'package:pancake/features/status/server_status_widget.dart';
import 'package:pancake/features/status/yachiyo_status_widget.dart';

class PancakeHomePage extends StatefulWidget {
  const PancakeHomePage({super.key});

  @override
  State<PancakeHomePage> createState() => _PancakeHomePageState();
}

class _PancakeHomePageState extends State<PancakeHomePage> {
  String title = "Pancake!";
  String subtitle = "Yachiyo Runtime Viewer";

  @override
  Widget build(BuildContext context) {
    final textStyle = Theme.of(context).textTheme;
    final colorStyle = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: textStyle.titleLarge),
            Text(subtitle, style: textStyle.bodyMedium?.copyWith(
              color: colorStyle.onSurfaceVariant
            )),
          ],
        ),
        leading: Icon(Icons.gesture),
      ),
      body: Column(
        children: [
          ServerStatusWidget(),
          YachiyoStatusWidget(),
          Expanded(child: ChatWidget()),
        ],
      ),
    );
  }
}
