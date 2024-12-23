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
      title: "Pong Game - Home",
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Compact Status Section
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Card(
              elevation: 2,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              child: Padding(
                padding: const EdgeInsets.all(12.0),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      "Bet: ${pongModel.betAmt}",
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                    Text(
                      pongModel.isReady
                          ? (pongModel.gameStarted ? "In Game" : "Ready")
                          : "Not Ready",
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ],
                ),
              ),
            ),
          ),
          // Error Message
          if (pongModel.errorMessage.isNotEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Card(
                color: Colors.red.shade100,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Row(
                    children: [
                      const Icon(Icons.error, color: Colors.red),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          pongModel.errorMessage,
                          style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: Colors.red),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          const SizedBox(height: 16),
          // Main Content
          Expanded(
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  if (pongModel.gameStarted)
                    Expanded(
                      child: Center(
                        child: pongModel.pongGame.buildWidget(
                          pongModel.gameState,
                          FocusNode(),
                        ),
                      ),
                    )
                  else if (pongModel.isReady) ...[
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
                    if (pongModel.currentWR != null)
                      Text(
                        "Room: ${pongModel.currentWR?.id}",
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                  ] else ...[
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
                    const Text(
                      "Waiting Rooms",
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                    ),
                    Expanded(
                      child: pongModel.waitingRooms.isEmpty
                          ? const Center(
                              child: Text(
                                "No waiting rooms available.",
                                style: TextStyle(fontSize: 16, color: Colors.grey),
                              ),
                            )
                          : ListView.builder(
                              itemCount: pongModel.waitingRooms.length,
                              itemBuilder: (context, index) {
                                final room = pongModel.waitingRooms[index];
                                return Card(
                                  margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 16),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: ListTile(
                                    title: Text("Room: ${room.id}"),
                                    subtitle: Text("Bet: ${room.betAmt}"),
                                    trailing: IconButton(
                                      icon: const Icon(Icons.add),
                                      onPressed: () => pongModel.joinWaitingRoom(room.id),
                                    ),
                                  ),
                                );
                              },
                            ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
