import 'package:flutter/material.dart';
import 'package:pongui/components/home/main_content.dart';
import 'package:pongui/components/home/top_status.dart';
import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/models/pong.dart';
import 'package:provider/provider.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  bool _showTopStatus = true;

  @override
  Widget build(BuildContext context) {
    final pongModel = Provider.of<PongModel>(context);

    return SharedLayout(
      title: "Pong Game - Home",
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Toggle button for top status visibility
          Padding(
            padding: const EdgeInsets.only(top: 8.0, right: 16.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton.icon(
                  onPressed: () {
                    setState(() {
                      _showTopStatus = !_showTopStatus;
                    });
                  },
                  icon: Icon(
                      _showTopStatus ? Icons.visibility_off : Icons.visibility),
                  label: Text(_showTopStatus ? "Hide Status" : "Show Status"),
                ),
              ],
            ),
          ),

          // 1) Top area: bet status + error message + current waiting room
          if (_showTopStatus)
            TopStatusCard(
              pongModel: pongModel,
              onErrorDismissed: () {
                pongModel.clearErrorMessage();
              },
            ),
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
