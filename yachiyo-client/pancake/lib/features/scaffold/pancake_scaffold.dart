import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:pancake/application/settings/settings_controller.dart';
import 'package:pancake/features/chat/chat_pane.dart';
import 'package:pancake/features/connect/connect_pane.dart';
import 'package:pancake/features/observe/observe_pane.dart';
import 'package:provider/provider.dart';

class PancakeScaffold extends StatefulWidget {
  const PancakeScaffold({super.key});

  @override
  State<PancakeScaffold> createState() => _PancakeLayout();
}

enum PancakePageIndex { space, chat, observe, connect }

const _mobilePage = [
  PancakePageIndex.chat,
  PancakePageIndex.observe,
  PancakePageIndex.connect,
];

const _desktopPage = [
  PancakePageIndex.space,
  PancakePageIndex.chat,
  PancakePageIndex.observe,
  PancakePageIndex.connect,
];

class _PancakeLayout extends State<PancakeScaffold> {
  PancakePageIndex _page = PancakePageIndex.space;

  void _select(PancakePageIndex page) {
    setState(() {
      _page = page;
    });
  }

  @override
  void initState() {
    super.initState();
    final provider = context.read<PancakeProvider>();
    final settings = context.read<SettingsController>();  // This only read once.

    if (settings.settings.enableReconnectOnStart == true) {
      provider.connect();
    }
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: ((context, constraints) {
        final width = constraints.maxWidth;
        final bool isDesktop = width > 800;

        final pageIndex = isDesktop ? _desktopPage : _mobilePage;
        final displayPage = !isDesktop && _page == PancakePageIndex.space
            ? PancakePageIndex.chat
            : _page;

        final index = pageIndex.indexOf(displayPage);

        void onPageChange(int index) {
          _select(pageIndex[index]);
        }

        if (isDesktop) {
          return _CompactDesktop(
            ratio: width < 1000 ? 3 : 2,
            index: index,
            onPageChange: onPageChange,
          );
        }

        return _Mobile(index: index, onPageChange: onPageChange);
      }),
    );
  }
}

class _Mobile extends StatefulWidget {
  final int index;
  final ValueChanged<int> onPageChange;

  const _Mobile({required this.index, required this.onPageChange});
  @override
  State<StatefulWidget> createState() => _MobileState();
}

class _MobileState extends State<_Mobile> {
  late final StreamSubscription<ProviderEvent> _providerSubsctiption;

  @override
  void initState() {
    super.initState();
    final provider = context.read<PancakeProvider>();
    _providerSubsctiption = provider.events.listen((event) {
      switch (event) {
        case EventSnackbarMessage():
          _show(event.message);
      }
    });
  }

  void _switchConnectPage() {
    widget.onPageChange(2);
  }

  void _switchObservePage() {
    widget.onPageChange(1);
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
          Divider(thickness: 0.6, height: 1),
          SizedBox(height: 8),
          Expanded(
            child: _MobileContentLayer(
              page: widget.index,
              pageSelect: widget.onPageChange,
              switchObservePage: _switchObservePage,
            ),
          ),
        ],
      ),
      bottomNavigationBar: _Nav(
        index: widget.index,
        onSelect: widget.onPageChange,
      ),
    );
  }

  @override
  void dispose() {
    _providerSubsctiption.cancel();
    super.dispose();
  }
}

class _CompactDesktop extends StatefulWidget {
  final int ratio;
  final int index;
  final ValueChanged<int> onPageChange;

  const _CompactDesktop({
    required this.ratio,
    required this.index,
    required this.onPageChange,
  });
  @override
  State<StatefulWidget> createState() => _CompactDesktopState();
}

class _CompactDesktopState extends State<_CompactDesktop> {
  late final StreamSubscription<ProviderEvent> _providerSubsctiption;

  @override
  void initState() {
    super.initState();
    final provider = context.read<PancakeProvider>();
    _providerSubsctiption = provider.events.listen((event) {
      switch (event) {
        case EventSnackbarMessage():
          _show(event.message);
      }
    });
  }

  void _switchConnectPage() {
    widget.onPageChange(3);
  }

  void _switchObservePage() {
    widget.onPageChange(2);
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
    final ratio = widget.ratio;
    return Scaffold(
      body: Row(
        children: [
          _Rail(
            index: widget.index,
            onSelect: widget.onPageChange,
            switchConnectPage: _switchConnectPage,
          ),
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.max,
              children: [
                _Head(switchConnectPage: _switchConnectPage, showIcon: false),
                Expanded(
                  child: ConstrainedBox(
                    constraints: BoxConstraints(
                      maxWidth: widget.index == 0 ? 1200 : 800,
                    ),
                    child: _CompactContentLayer(
                      page: widget.index,
                      pageSelect: widget.onPageChange,
                      switchObservePage: _switchObservePage,
                      ratio: ratio,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _providerSubsctiption.cancel();
    super.dispose();
  }
}

class _MobileContentLayer extends StatelessWidget {
  final int page;
  final ValueChanged<int> pageSelect;
  final VoidCallback switchObservePage;
  const _MobileContentLayer({
    required this.page,
    required this.pageSelect,
    required this.switchObservePage,
  });

  @override
  Widget build(BuildContext context) {
    switch (page) {
      case 1:
        return ObservePane();
      case 2:
        return ConnectPane();

      // case 0:
      default:
        return ChatPane(switchObservePage: switchObservePage);
    }
  }
}

class _CompactContentLayer extends StatelessWidget {
  final int page;
  final ValueChanged<int> pageSelect;
  final VoidCallback switchObservePage;
  final int ratio;
  const _CompactContentLayer({
    required this.page,
    required this.pageSelect,
    required this.ratio,
    required this.switchObservePage,
  });

  @override
  Widget build(BuildContext context) {
    switch (page) {
      case 1:
        return ChatPane(switchObservePage: switchObservePage);
      case 2:
        return ObservePane();
      case 3:
        return ConnectPane();

      // case 0
      default:
        return Row(
          mainAxisSize: MainAxisSize.max,
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Expanded(
              flex: ratio,
              child: Card(
                margin: EdgeInsets.only(left: 12, right: 12, bottom: 12),
                child: Padding(
                  padding: EdgeInsetsGeometry.symmetric(
                    vertical: 8,
                    horizontal: 4,
                  ),
                  child: ChatPane(switchObservePage: switchObservePage),
                ),
              ),
            ),
            Expanded(
              flex: 2,
              child: Column(
                mainAxisSize: MainAxisSize.max,
                children: [
                  Expanded(
                    child: Padding(
                      padding: EdgeInsetsGeometry.only(
                        right: 12,
                        top: 12,
                        bottom: 12,
                      ),
                      child: ObservePane(),
                    ),
                  ),
                ],
              ),
            ),
          ],
        );
    }
  }
}

class _Head extends StatelessWidget {
  final VoidCallback switchConnectPage;
  final bool showIcon;
  const _Head({required this.switchConnectPage, this.showIcon = true});

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final phase = context.watch<PancakeProvider>().data.phase;

    return SafeArea(
      bottom: false,
      child: Padding(
        padding: EdgeInsetsGeometry.symmetric(
          horizontal: showIcon ? 12 : 24,
          vertical: showIcon ? 12 : 24,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              spacing: 12.0,
              children: [
                if (showIcon)
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

class _Rail extends StatelessWidget {
  final int index;
  final ValueChanged<int> onSelect;
  final VoidCallback switchConnectPage;

  const _Rail({
    required this.index,
    required this.onSelect,
    required this.switchConnectPage,
  });

  @override
  Widget build(BuildContext context) {
    final texts = Theme.of(context).textTheme;
    final colors = Theme.of(context).colorScheme;

    return Container(
      width: 80,
      decoration: BoxDecoration(
        border: Border(right: BorderSide(color: Colors.grey, width: 0.1)),
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: EdgeInsetsGeometry.symmetric(vertical: 8),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.start,
            crossAxisAlignment: CrossAxisAlignment.center,
            mainAxisSize: MainAxisSize.max,
            children: [
              SizedBox(
                child: Padding(
                  padding: EdgeInsetsGeometry.only(top: 16, bottom: 24),
                  child: Image.asset(
                    "assets/pancake.png",
                    width: 32,
                    height: 32,
                  ),
                ),
              ),
              Column(
                mainAxisAlignment: MainAxisAlignment.start,
                mainAxisSize: MainAxisSize.max,
                spacing: 4,
                children: [
                  for (var item in [
                    (0, "Space", Icons.motion_photos_on_outlined),
                    (1, "Chat", Icons.chat_bubble_outline),
                    (2, "Observe", Icons.bubble_chart_outlined),
                  ])
                    SizedBox(
                      height: 64,
                      width: 64,
                      child: InkWell(
                        borderRadius: BorderRadius.circular(16),
                        onTap: () => onSelect(item.$1),
                        child: Container(
                          decoration: BoxDecoration(
                            color: index == item.$1
                                ? colors.secondaryContainer
                                : Colors.transparent,
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            mainAxisAlignment: MainAxisAlignment.center,
                            crossAxisAlignment: CrossAxisAlignment.center,
                            spacing: 4,
                            children: [
                              Icon(item.$3, size: 20),
                              Text(item.$2, style: texts.labelSmall),
                            ],
                          ),
                        ),
                      ),
                    ),
                ],
              ),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    for (var item in [(3, "Connect", Icons.tune)])
                      SizedBox(
                        height: 64,
                        width: 64,
                        child: InkWell(
                          borderRadius: BorderRadius.circular(16),
                          onTap: () => onSelect(item.$1),
                          child: Container(
                            decoration: BoxDecoration(
                              color: index == item.$1
                                  ? colors.secondaryContainer
                                  : Colors.transparent,
                              borderRadius: BorderRadius.circular(16),
                            ),
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              mainAxisAlignment: MainAxisAlignment.center,
                              crossAxisAlignment: CrossAxisAlignment.center,
                              spacing: 4,
                              children: [
                                Icon(item.$3, size: 20),
                                Text(item.$2, style: texts.labelSmall),
                              ],
                            ),
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
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
