import 'package:flutter/material.dart';
import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/models/pong.dart';
import 'package:provider/provider.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final pongModel = Provider.of<PongModel>(context);

    return SharedLayout(
      title: "Home Screen",
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (pongModel.errorMessage.isNotEmpty)
            Text(
              pongModel.errorMessage,
              style: const TextStyle(color: Colors.red),
            ),
          // Show the game widget when the game has started
          if (pongModel.gameStarted)
            Expanded(
              child: Center(
                child: pongModel.pongGame.buildWidget(
                  pongModel.gameState,
                  FocusNode(),
                ),
              ),
            )
          // Show waiting message and Pong icon if player is ready but game hasn't started
          else if (pongModel.isReady) ...[
            Icon(
              Icons.sports_tennis,
              size: 100,
              color: Colors.blueAccent,
            ),
            const SizedBox(height: 20),
            const Text(
              "Waiting for all players to get ready...",
              style: TextStyle(fontSize: 18),
            ),
            if (pongModel.currentWR.id.isNotEmpty)
              Text(
                "Joined Room: ${pongModel.currentWR.id}",
                style: const TextStyle(
                  fontSize: 16,
                  color: Colors.white70,
                ),
              ),
          ]
          // Show the default screen when player is not ready
          else ...[
            ElevatedButton(
              onPressed: pongModel.toggleReady,
              child: Text(pongModel.isReady ? "Cancel Ready" : "Ready"),
            ),
            ElevatedButton(
              onPressed: pongModel.startGameStream,
              child: const Text("Start Game"),
            ),
            Expanded(
              child: ListView.builder(
                itemCount: pongModel.waitingRooms.length,
                itemBuilder: (context, index) {
                  final room = pongModel.waitingRooms[index];
                  return ListTile(
                    title: Text("Room ID: ${room.id}"),
                    subtitle: Text("Bet: ${room.betAmount}"),
                    onTap: () => pongModel.joinWaitingRoom(room.id),
                  );
                },
              ),
            ),
          ],
        ],
      ),
    );
  }
}
