import 'package:flutter/material.dart';
import 'package:pongui/components/waiting_rooms.dart';
import 'package:pongui/models/pong.dart';

class MainContent extends StatelessWidget {
  final PongModel pongModel;

  const MainContent({
    Key? key,
    required this.pongModel,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    // GAME STARTED
    if (pongModel.gameStarted) {
      return Center(
        child: pongModel.pongGame.buildWidget(
          pongModel.gameState,
          FocusNode(), // or pass the correct FocusNode
        ),
      );
    }

    // READY, BUT WAITING FOR GAME TO START
    if (pongModel.isReady) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
          Icon(
              Icons.sports_tennis,
              size: 60,
              color: Colors.blueAccent,
            ),
          SizedBox(height: 10),
          Text(
              "Waiting for players...",
              style: TextStyle(fontSize: 16),
            ),
          ],
        ),
      );
    }

    // NOT READY => Show "Ready" + "Create" buttons, plus waiting room list
    return Column(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
          child: Column(
            children: [
              const Text(
                "Welcome to Pong!",
                style: TextStyle(
                  fontSize: 22,  // Slightly smaller title
                  fontWeight: FontWeight.w600,  // Semi-bold instead of bold
                  color: Colors.blueAccent,
                ),
              ),
              const SizedBox(height: 4),  // Reduced spacing
              Text(
                "To place a bet send a tip to the pongbot on Bison Relay",
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,  // Smaller text
                  color: Colors.grey[600],
                  height: 1.3,  // Tighter line height
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 8),

        // The waiting room list should scroll if it has many items,
        // so we give it leftover space with `Expanded`.
        Expanded(
          child: WaitingRoomList(
            pongModel.waitingRooms,
            (roomId) => pongModel.joinWaitingRoom(roomId),
          ),
        ),
      ],
    );
  }
}
