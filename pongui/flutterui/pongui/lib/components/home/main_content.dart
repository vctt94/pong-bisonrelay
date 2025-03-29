import 'package:flutter/material.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';
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
      if (pongModel.gameState == null) {
        return const Center(
            child: CircularProgressIndicator(
          valueColor: AlwaysStoppedAnimation<Color>(Colors.blueAccent),
        ));
      }

      return Center(
        child: pongModel.pongGame.buildWidget(
          pongModel.gameState!,
          FocusNode(),
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
            SizedBox(height: 16),
            Text(
              "Waiting for players...",
              style: TextStyle(
                fontSize: 18,
                color: Colors.white,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      );
    }

    return Column(
      children: [
        // Welcome section
        const SizedBox(height: 40),
        const Text(
          "Welcome to Pong!",
          style: TextStyle(
            fontSize: 32,
            fontWeight: FontWeight.bold,
            color: Colors.blueAccent,
          ),
        ),
        const SizedBox(height: 16),
        const Padding(
          padding: EdgeInsets.symmetric(horizontal: 24.0),
          child: Text(
            "To place a bet send a tip to pongbot on Bisonn Relay.",
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 16,
              color: Colors.white,
              height: 1.4,
            ),
          ),
        ),

        // Waiting rooms or empty state
        Expanded(
          child: pongModel.waitingRooms.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const SizedBox(height: 40),
                      Container(
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: const Color(0xFF1B1E2C).withOpacity(0.6),
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          Icons.sports_esports,
                          size: 64,
                          color: Colors.grey.shade400,
                        ),
                      ),
                      const SizedBox(height: 24),
                      const Text(
                        'No active waiting rooms',
                        style: TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.w500,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 8),
                      const Text(
                        'Create a room to start playing!',
                        style: TextStyle(
                          fontSize: 16,
                          color: Colors.white70,
                        ),
                      ),
                    ],
                  ),
                )
              : WaitingRoomList(
                  pongModel.waitingRooms,
                  currentRoomId: pongModel.currentWR?.id,
                  onJoinRoom: (roomId) => pongModel.joinWaitingRoom(roomId),
                ),
        ),
      ],
    );
  }
}
