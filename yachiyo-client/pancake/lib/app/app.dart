import 'package:flutter/material.dart';
import 'package:pancake/features/home/home_page.dart';

class PancakeApp extends StatelessWidget {
  PancakeApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pancake!',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
      ),
      home: PancakeHomePage(),
    );
  }
}