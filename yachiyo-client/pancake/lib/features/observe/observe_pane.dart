import 'package:flutter/material.dart';

class ObservePane extends StatefulWidget {
  const ObservePane({super.key, required this.state});
  final String state;

  @override
  State<StatefulWidget> createState() => _ObservePaneState();
}

class _ObservePaneState extends State<ObservePane> {
  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    return SafeArea(
      child: Padding(
        padding: EdgeInsetsGeometry.symmetric(horizontal: 16, vertical: 8),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.start,
          mainAxisSize: MainAxisSize.max,
          crossAxisAlignment: CrossAxisAlignment.start,
          spacing: 4,
          children: [
            Row(
              spacing: 12,
              children: [
                Icon(Icons.bedtime_outlined),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "From the former to the future,",
                      style: texts.bodyMedium,
                    ),
                    Text("Heartbeating.", style: texts.titleLarge),
                  ],
                ),
              ],
            ),
            Divider(thickness: 0.1),
            Text(widget.state),
          ],
        ),
      ),
    );
  }
}
