import 'package:flutter/material.dart';
import 'package:pancake/core/provider/yachiyo_provider.dart';
import 'package:provider/provider.dart';

class YachiyoStatusWidget extends StatelessWidget {
  const YachiyoStatusWidget({super.key});

  @override
  Widget build(BuildContext context) {
    return ExpansionTile(
      expandedCrossAxisAlignment: CrossAxisAlignment.stretch,
      childrenPadding: EdgeInsets.symmetric(horizontal: 16.0),
      initiallyExpanded: true,
      shape: const Border(),
      collapsedShape: const Border(),
      leading: Icon(Icons.animation_outlined),
      title: Text("Yachiyo Status"),
      children: [
        Consumer<YachiyoProvider>(
          builder: (_, provider, _) {
            return Text(provider.state.runtime.state??"Unknown state.");
          },
        ),
      ],
    );
  }
}
