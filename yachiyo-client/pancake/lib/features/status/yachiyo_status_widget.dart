import 'package:flutter/material.dart';
import 'package:pancake/features/status/status_model.dart';

class YachiyoStatusWidget extends StatelessWidget {
  final StatusModel statusModel;
  const YachiyoStatusWidget({super.key, required this.statusModel});

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: statusModel,
      builder: (context, child) {
        return ExpansionTile(
          expandedCrossAxisAlignment: CrossAxisAlignment.stretch,
          childrenPadding: EdgeInsets.symmetric(horizontal: 16.0),
          initiallyExpanded: true,
          shape: const Border(),
          collapsedShape: const Border(),
          leading: Icon(Icons.animation_outlined),
          title: Text("Yachiyo Status"),
          children: [Text(statusModel.status)],
        );
      },
    );
  }
}
