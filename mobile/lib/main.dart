import 'package:flutter/material.dart';

void main() {
  runApp(const R3TiFaceAttendApp());
}

class R3TiFaceAttendApp extends StatelessWidget {
  const R3TiFaceAttendApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'R3 TI FaceAttend',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.green),
      ),
      home: const Scaffold(body: Center(child: Text('R3 TI FaceAttend'))),
    );
  }
}
