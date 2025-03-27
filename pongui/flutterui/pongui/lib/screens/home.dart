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
  @override
  Widget build(BuildContext context) {
    final pongModel = Provider.of<PongModel>(context);
    final bool gameInProgress = pongModel.gameStarted;

    return SharedLayout(
      title: "Pong Game - Home",
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          // Only show status elements when game is not in progress
          if (!gameInProgress) ...[
            // 1) Top area: bet status
            Center(
              child: Container(
                width: MediaQuery.of(context).size.width * 0.85,
                margin: const EdgeInsets.only(top: 16.0),
                child: Card(
                  color: const Color(0xFF1B1E2C), // Dark card background
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          "Bet: ${pongModel.betAmt / 1e11}",
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        if (pongModel.betAmt > 0 && pongModel.currentWR == null)
                          ElevatedButton(
                            onPressed: pongModel.createWaitingRoom,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.blueAccent,
                            ),
                            child: const Text("Create Waiting Room"),
                          ),
                      ],
                    ),
                  ),
                ),
              ),
            ),

            // 2) Current waiting room info
            Center(
              child: Container(
                width: MediaQuery.of(context).size.width * 0.85,
                margin: const EdgeInsets.only(top: 16.0),
                child: Card(
                  color: const Color(0xFF1B1E2C), // Dark card background
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          "Current Waiting Room",
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              "Room ID: ${pongModel.currentWR?.id ?? ""}",
                              style: const TextStyle(
                                color: Colors.white,
                              ),
                            ),
                            Text(
                              pongModel.isReady ? "Ready" : "Not Ready",
                              style: TextStyle(
                                color: pongModel.isReady
                                    ? Colors.green
                                    : Colors.white,
                                fontWeight: pongModel.isReady
                                    ? FontWeight.bold
                                    : FontWeight.normal,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          "Players: ${pongModel.currentWR?.players?.length ?? 0} / 2",
                          style: const TextStyle(
                            color: Colors.white,
                          ),
                        ),

                        // Add ready/leave buttons if in a room
                        if (pongModel.currentWR != null) ...[
                          const SizedBox(height: 16),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              ElevatedButton(
                                onPressed: pongModel.toggleReady,
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: pongModel.isReady
                                      ? Colors.orange
                                      : Colors.green,
                                ),
                                child: Text(pongModel.isReady
                                    ? "Cancel Ready"
                                    : "Ready"),
                              ),
                              const SizedBox(width: 8),
                              ElevatedButton(
                                onPressed: () => pongModel.leaveWaitingRoom(),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: Colors.redAccent,
                                ),
                                child: const Text("Leave Room"),
                              ),
                            ],
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
              ),
            ),

            // 3) Error message if exists
            if (pongModel.errorMessage.isNotEmpty)
              Center(
                child: Container(
                  width: MediaQuery.of(context).size.width * 0.85,
                  margin: const EdgeInsets.only(top: 16.0),
                  child: Card(
                    color: Colors.red.shade800,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Padding(
                      padding: const EdgeInsets.all(12.0),
                      child: Row(
                        children: [
                          const Icon(Icons.error, color: Colors.white),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              pongModel.errorMessage,
                              style: const TextStyle(
                                color: Colors.white,
                              ),
                            ),
                          ),
                          Material(
                            color: Colors.transparent,
                            child: InkWell(
                              onTap: () {
                                pongModel.clearErrorMessage();
                              },
                              borderRadius: BorderRadius.circular(20),
                              child: const Padding(
                                padding: EdgeInsets.all(8.0),
                                child: Icon(Icons.close,
                                    color: Colors.white, size: 20),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
          ],

          // 4) Expanded area for the main content
          Expanded(
            child: MainContent(pongModel: pongModel),
          ),
        ],
      ),
    );
  }
}
