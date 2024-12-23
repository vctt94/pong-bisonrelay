import 'package:flutter/material.dart';
import 'package:pongui/components/home/main_content.dart';
import 'package:pongui/components/home/top_status.dart';
import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/models/pong.dart';
import 'package:provider/provider.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final pongModel = Provider.of<PongModel>(context);

    return SharedLayout(
      title: "Pong Game - Home",
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 1) Top area: bet status + error message + current waiting room
          TopStatusCard(pongModel: pongModel),
          const SizedBox(height: 16),

          // 2) Expanded area for the main content
          Expanded(
            child: MainContent(pongModel: pongModel),
          ),
        ],
      ),
    );
  }
}
