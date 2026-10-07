import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:pancake/application/provider/provider.dart';
import 'package:provider/provider.dart';
import 'package:pancake/features/chat/chat_widget.dart';
import 'package:pancake/features/status/config_status_widget.dart';
import 'package:pancake/features/status/yachiyo_status_widget.dart';

class PancakeHomePage extends StatefulWidget {
  const PancakeHomePage({super.key});

  @override
  State<PancakeHomePage> createState() => _PancakeHomePageState();
}

class _PancakeHomePageState extends State<PancakeHomePage> {
  String title = "Pancake!";
  String subtitle = "Yachiyo Runtime Viewer";
  bool isYachiyoStatusShown = true;

  @override
  void initState() {
    super.initState();

    final provider = context.read<PancakeProvider>();
    provider.events.listen((event) {
      switch (event) {
        case EventErrorMessage():
          _show("(${event.code}) ${event.message}");
      }
      ;
    });
  }

  void _toggleYachiyoStatusShown() {
    setState(() {
      isYachiyoStatusShown = !isYachiyoStatusShown;
    });
  }

  void _show(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Column(
          children: [
            Row(
              children: [
                Padding(
                  padding: EdgeInsetsGeometry.symmetric(horizontal: 8.0),
                  child: Icon(Icons.warning_amber_rounded, color: Colors.white),
                ),
                Text("Warning", style: TextStyle(fontWeight: FontWeight.bold)),
              ],
            ),
            Padding(
              padding: EdgeInsetsGeometry.symmetric(horizontal: 8.0),
              child: Text(message, softWrap: true),
            ),
          ],
        ),
        width: 280.0, // Width of the SnackBar.
        padding: EdgeInsetsGeometry.symmetric(horizontal: 4.0, vertical: 8.0),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(10.0),
        ),
        action: SnackBarAction(label: "OK", onPressed: () {}),
        persist: true,
      ),
      snackBarAnimationStyle: const AnimationStyle(
        duration: Duration(milliseconds: 100),
        reverseDuration: Duration(milliseconds: 800),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: Icon(Icons.gesture),

        centerTitle: true,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            Text(
              title,
              style: GoogleFonts.kiwiMaru(
                fontWeight: FontWeight.w500,
                letterSpacing: -0.5,
              ),
            ),
            ConfigStatusBadge(),
          ],
        ),

        actions: [
          IconButton(
            onPressed: () => _toggleYachiyoStatusShown(),
            icon: Icon(
              isYachiyoStatusShown == true
                  ? Icons.face_retouching_natural_rounded
                  : Icons.face_retouching_off_rounded,
            ),
          ),
        ],
      ),
      body: Padding(
        padding: EdgeInsetsGeometry.only(
          bottom: MediaQuery.paddingOf(context).bottom,
        ),
        child: Column(
          children: [
            if (isYachiyoStatusShown == true) YachiyoStatusWidget(),
            Expanded(child: ChatWidget()),
          ],
        ),
      ),
    );
  }
}
