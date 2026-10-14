import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:pancake/features/chat/chat_pane.dart';
import 'package:pancake/features/connect/connect_pane.dart';
import 'package:pancake/features/observe/observe_pane.dart';
import 'package:provider/provider.dart';

class PancakeScaffold extends StatefulWidget {
  const PancakeScaffold({super.key});

  @override
  State<PancakeScaffold> createState() => _PancakeScaffoldState();
}

class _PancakeScaffoldState extends State<PancakeScaffold> {
  @override
  void initState() {
    super.initState();
    final provider = context.read<PancakeProvider>();
    provider.events.listen((event) {
      switch (event) {
        case EventSnackbarMessage():
          _show(event.message);
      }
    });
  }

  int _page = 0;
  void select(int page) {
    setState(() {
      _page = page;
    });
  }

  void _switchConnectPage() {
    setState(() {
      _page = 2;
    });
  }

  void _show(String message) {
    final colors = Theme.of(context).colorScheme;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          spacing: 8,
          children: [
            Icon(Icons.info_outline, color: colors.onPrimary),
            Expanded(child: Text(message, softWrap: true)),
          ],
        ),
        behavior: SnackBarBehavior.fixed,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          _Head(switchConnectPage: _switchConnectPage),
          Divider(thickness: 0.6),
          Expanded(
            child: _ContentLayer(page: _page, pageSelect: select),
          ),
        ],
      ),
      bottomNavigationBar: _Nav(index: _page, onSelect: select),
    );
  }
}

class _ContentLayer extends StatelessWidget {
  final int page;
  final ValueChanged<int> pageSelect;
  const _ContentLayer({required this.page, required this.pageSelect});

  @override
  Widget build(BuildContext context) {
    switch (page) {
      case 1:
        return ObservePane();
      case 2:
        return ConnectPane();

      // case 0:
      default:
        return ChatPane(pageChange: pageSelect);
    }
  }
}

class _Head extends StatelessWidget {
  final VoidCallback switchConnectPage;
  const _Head({required this.switchConnectPage});

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final phase = context.watch<PancakeProvider>().data.phase;

    return SafeArea(
      bottom: false,
      child: Padding(
        padding: EdgeInsetsGeometry.all(12.0),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              spacing: 12.0,
              children: [
                Image.asset("assets/pancake.png", width: 32, height: 32),
                Text(
                  "Pancake!",
                  style: GoogleFonts.kiwiMaru(
                    fontSize: texts.titleLarge!.fontSize,
                    fontWeight: FontWeight.w500,
                    letterSpacing: -0.5,
                  ),
                ),
              ],
            ),
            _StatusCapsule(onPress: switchConnectPage, phase: phase),
          ],
        ),
      ),
    );
  }
}

class _Nav extends StatelessWidget {
  final int index;
  final ValueChanged<int> onSelect;

  const _Nav({required this.index, required this.onSelect});

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;

    return Container(
      decoration: BoxDecoration(
        border: Border(top: BorderSide(color: Colors.grey, width: 0.1)),
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: EdgeInsetsGeometry.symmetric(vertical: 8, horizontal: 8),
          child: Row(
            children: [
              for (var item in [
                (0, "Chat", Icons.chat_bubble_outline),
                (1, "Observe", Icons.bubble_chart_outlined),
                (2, "Connect", Icons.tune),
              ])
                Expanded(
                  child: InkWell(
                    borderRadius: BorderRadius.circular(16),
                    onTap: () => onSelect(item.$1),
                    child: Container(
                      padding: EdgeInsets.symmetric(vertical: 8),
                      decoration: BoxDecoration(
                        color: index == item.$1
                            ? colors.secondaryContainer
                            : Colors.transparent,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.center,
                        spacing: 8,
                        children: [
                          Icon(item.$3, size: 18),
                          Text(item.$2, style: texts.labelLarge),
                        ],
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatusCapsule extends StatelessWidget {
  final VoidCallback onPress;
  final RuntimePhase phase;

  String _phaseText(RuntimePhase p) {
    switch (p) {
      case RuntimePhase.offline:
        return "Offline";
      case RuntimePhase.connecting:
        return "Connecting";
      case RuntimePhase.online:
        return "Online";
    }
  }

  Color _phaseColor(RuntimePhase p) {
    switch (p) {
      case RuntimePhase.offline:
        return Colors.grey;
      case RuntimePhase.connecting:
        return Colors.blue;
      case RuntimePhase.online:
        return Colors.lightGreenAccent;
    }
  }

  const _StatusCapsule({required this.onPress, required this.phase});

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    final texts = Theme.of(context).textTheme;

    return Material(
      color: colors.secondaryContainer,
      borderRadius: BorderRadius.circular(30),
      child: InkWell(
        onTap: onPress,
        borderRadius: BorderRadius.circular(30),
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Row(
            spacing: 8.0,
            children: [
              Container(
                width: 6,
                height: 6,
                decoration: BoxDecoration(
                  color: _phaseColor(phase),
                  shape: BoxShape.circle,
                ),
              ),
              Text(_phaseText(phase), style: texts.labelMedium),
              Icon(Icons.expand_more, size: 16.0),
            ],
          ),
        ),
      ),
    );
  }
}
