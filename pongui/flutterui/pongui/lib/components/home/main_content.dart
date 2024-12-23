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
      return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.sports_tennis,
            size: 60,
            color: Colors.blueAccent,
          ),
          const SizedBox(height: 10),
          const Text(
            "Waiting for players...",
            style: TextStyle(fontSize: 16),
          ),
        ],
      );
    }

    // NOT READY => Show "Ready" + "Create" buttons, plus waiting room list
    return Column(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            FilledButton(
              onPressed: pongModel.toggleReady,
              child: Text(
                pongModel.isReady ? "Cancel Ready" : "Ready",
              ),
            ),
            const SizedBox(width: 16),
            FilledButton(
              onPressed: pongModel.createWaitingRoom,
              child: const Text("Create Waiting Room"),
            ),
          ],
        ),
        const SizedBox(height: 16),

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
