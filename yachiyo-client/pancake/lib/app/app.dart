import 'package:flutter/material.dart';
import 'package:pancake/core/network/client.dart';
import 'package:pancake/features/home/home_page.dart';
import 'package:pancake/features/status/status_model.dart';

class PancakeApp extends StatelessWidget {
  PancakeApp({super.key});

  final statusModel = StatusModel();
  final yachiyoClient = YachiyoClient();

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pancake!',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
      ),
      home: PancakeHomePage(statusModel: statusModel, yachiyoClient: yachiyoClient),
    );
  }
}