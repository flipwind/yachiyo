import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:provider/provider.dart';

class ObservePane extends StatefulWidget {
  const ObservePane({super.key});

  @override
  State<StatefulWidget> createState() => _ObservePaneState();
}

class _ObservePaneState extends State<ObservePane> {
  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;
    final provider = context.watch<PancakeProvider>();
    final state = provider.data.runtimeState;

    if (state == null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              margin: EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: colors.secondaryContainer,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Padding(
                padding: EdgeInsetsGeometry.all(16),
                child: Icon(Icons.query_stats),
              ),
            ),
            Text("Observe", style: texts.titleMedium),
            Text(
              "Connect the runtime, and query her status.",
              style: texts.bodyMedium,
            ),
          ],
        ),
      );
    } else {
      final affect = state.affect;
      return SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.start,
          mainAxisSize: MainAxisSize.max,
          crossAxisAlignment: CrossAxisAlignment.start,
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
                    SizedBox(height: 4),
                  ],
                ),
              ],
            ),

            Divider(thickness: 0.1),
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.vertical(
                  top: Radius.circular(24),
                  bottom: Radius.circular(8),
                ),
              ),
              child: Padding(
                padding: EdgeInsetsGeometry.symmetric(
                  vertical: 16,
                  horizontal: 12,
                ),
                child: Row(
                  spacing: 12,
                  children: [
                    Icon(Icons.sync, size: 24),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text("Last Update:", style: texts.titleMedium),
                        Text(
                          DateFormat(
                            'yyyy-MM-dd HH:mm:ss',
                          ).format(state.timetick),
                          style: texts.bodyMedium,
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              child: Padding(
                padding: EdgeInsetsGeometry.symmetric(
                  vertical: 16,
                  horizontal: 12,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  spacing: 8,
                  children: [
                    Row(
                      spacing: 12,
                      children: [
                        Icon(Icons.code, size: 24),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text("Affect Engine", style: texts.titleMedium),
                            Text(
                              "Category: ${affect.octants}",
                              style: texts.bodyMedium,
                            ),
                          ],
                        ),
                      ],
                    ),
                    Divider(thickness: 0.5),
                    Column(
                      spacing: 16,
                      children: [
                        _AxisCard(
                          icon: Icons.mood,
                          type: "Pleasure",
                          value: affect.pleasure,
                        ),
                        _AxisCard(
                          icon: Icons.trending_up,
                          type: "Arousal",
                          value: affect.arousal,
                        ),
                        _AxisCard(
                          icon: Icons.self_improvement,
                          type: "Dominance",
                          value: affect.dominance,
                        ),
                      ],
                    ),
                    if (affect.actives.isNotEmpty)
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        spacing: 8,
                        children: [
                          Divider(),
                          Row(
                            spacing: 8,
                            children: [
                              Icon(Icons.queue_music, size: 24),
                              Text("Active Moods", style: texts.titleMedium),
                            ],
                          ),
                          Wrap(
                            alignment: WrapAlignment.start,
                            spacing: 8,
                            runSpacing: 4,
                            children: [
                              for (final mood in affect.actives)
                                Chip(
                                  label: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Text(mood.type),
                                      VerticalDivider(),
                                      Text(mood.intensity.toStringAsFixed(2)),
                                    ],
                                  ),
                                ),
                            ],
                          ),
                        ],
                      ),
                  ],
                ),
              ),
            ),
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.vertical(
                  top: Radius.circular(8),
                  bottom: Radius.circular(24),
                ),
              ),
              child: Padding(
                padding: EdgeInsetsGeometry.symmetric(
                  vertical: 16,
                  horizontal: 12,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  spacing: 8,
                  children: [
                    Row(
                      spacing: 8,
                      children: [
                        Icon(Icons.bolt, size: 24),
                        Text("Energy", style: texts.titleMedium),
                      ],
                    ),
                    LayoutBuilder(
                      builder: ((context, constraints) => ConstrainedBox(
                        constraints: BoxConstraints.expand(height: 24),
                        child: Stack(
                          alignment: AlignmentGeometry.centerStart,
                          children: [
                            Container(
                              height: 24,
                              decoration: BoxDecoration(
                                color: colors.surfaceContainerHighest,
                                borderRadius: BorderRadius.circular(12),
                              ),
                            ),
                            Positioned(
                              child: Container(
                                height: 24,
                                width: constraints.maxWidth * state.energy,
                                decoration: BoxDecoration(
                                  color: colors.secondaryContainer,
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: state.energy >= 0.5
                                    ? Align(
                                        alignment: Alignment.centerRight,
                                        child: Padding(
                                          padding: EdgeInsetsGeometry.only(
                                            right: 12,
                                          ),
                                          child: Text(
                                            state.energy.toStringAsFixed(2),
                                            style: texts.labelMedium?.copyWith(
                                              color:
                                                  colors.onSecondaryContainer,
                                            ),
                                          ),
                                        ),
                                      )
                                    : null,
                              ),
                            ),
                            if (state.energy < 0.5)
                              Positioned(
                                left: constraints.maxWidth * state.energy + 8,
                                child: Text(
                                  state.energy.toStringAsFixed(2),
                                  style: texts.labelMedium?.copyWith(
                                    color: colors.onSecondaryContainer,
                                  ),
                                ),
                              ),
                          ],
                        ),
                      )),
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
}

class _AxisCard extends StatelessWidget {
  final IconData icon;
  final String type;
  final double value;

  const _AxisCard({
    required this.icon,
    required this.type,
    required this.value,
  });

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;
    final spaceValue = 0.5 + 0.5 * value;
    return Column(
      spacing: 4,
      children: [
        Row(
          spacing: 6,
          children: [
            Icon(icon, size: 16),
            Text(type, style: texts.titleSmall),
          ],
        ),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text("-1.0", style: texts.bodySmall),
            Text("0.0", style: texts.bodySmall),
            Text("+1.0", style: texts.bodySmall),
          ],
        ),
        LayoutBuilder(
          builder: ((context, constraints) => ConstrainedBox(
            constraints: BoxConstraints.expand(height: 24),
            child: Stack(
              children: [
                Container(
                  height: 24,
                  decoration: BoxDecoration(
                    color: colors.surfaceContainerHighest,
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                Positioned(
                  left: (constraints.maxWidth - 48) * spaceValue,
                  child: Container(
                    height: 24,
                    width: 48,
                    decoration: BoxDecoration(
                      color: colors.secondaryContainer,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Center(
                      child: Text(
                        value.toStringAsFixed(2),
                        style: texts.labelMedium?.copyWith(
                          color: colors.onSecondaryContainer,
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          )),
        ),
      ],
    );
  }
}
