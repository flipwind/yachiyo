import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
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
  bool isYachiyoStatusShown = true;

  void _toggleYachiyoStatusShown() {
    setState(() {
      isYachiyoStatusShown = !isYachiyoStatusShown;
    });
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
            ServerStatusBadge(),
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
