import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:pancake/features/chat/chat_pane.dart';
import 'package:pancake/features/connect/connect_pane.dart';
import 'package:pancake/features/observe/observe_pane.dart';

class PancakeScaffold extends StatefulWidget {
  const PancakeScaffold({super.key});

  @override
  State<PancakeScaffold> createState() => _PancakeScaffoldState();
}

class _PancakeScaffoldState extends State<PancakeScaffold> {
  int _page = 0;
  void _select(int page) {
    setState(() {
      _page = page;
    });
  }

  void _switchConnectPage() {
    setState(() {
      _page = 2;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          _Head(switchConnectPage: _switchConnectPage),
          Divider(thickness: 0.6),
          Expanded(child: _ContentLayer(page: _page)),
        ],
      ),
      bottomNavigationBar: _Nav(index: _page, onSelect: _select),
    );
  }
}

class _ContentLayer extends StatelessWidget {
  final int page;
  const _ContentLayer({required this.page});

  @override
  Widget build(BuildContext context) {
    switch (page) {
      case 1:
        return ObservePane(state: "hello this is state. How are you today?");
      case 2:
        return ConnectPane();

      // case 0:
      default:
        return ChatPane();
    }
  }
}

class _Head extends StatelessWidget {
  final VoidCallback switchConnectPage;
  const _Head({required this.switchConnectPage});

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;

    return Padding(
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
          Material(
            color: colors.secondaryContainer,
            borderRadius: BorderRadius.circular(30),
            child: InkWell(
              onTap: switchConnectPage,
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
                        color: Colors.grey,
                        shape: BoxShape.circle,
                      ),
                    ),
                    Text("Offline", style: texts.labelMedium),
                    Icon(Icons.expand_more, size: 16.0),
                  ],
                ),
              ),
            ),
          ),
        ],
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
