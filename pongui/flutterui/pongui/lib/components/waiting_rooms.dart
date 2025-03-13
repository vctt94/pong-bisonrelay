import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';

class WaitingRoomList extends StatelessWidget {
  final List<LocalWaitingRoom> waitingRooms;
  final Function(String roomId) onJoinRoom;
  final String? currentRoomId;

  const WaitingRoomList(this.waitingRooms, this.onJoinRoom,
      {this.currentRoomId, Key? key})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    if (waitingRooms.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: const [
            Icon(
              Icons.sports_esports,
              size: 64,
              color: Colors.white54,
            ),
            SizedBox(height: 16),
            Text(
              'No active waiting rooms',
              style: TextStyle(fontSize: 18, color: Colors.white70),
            ),
            SizedBox(height: 8),
            Text(
              'Create a room to start playing!',
              style: TextStyle(fontSize: 14, color: Colors.white54),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      itemCount: waitingRooms.length,
      padding: const EdgeInsets.all(12),
      itemBuilder: (context, index) {
        final wr = waitingRooms[index];
        final bool isCurrentRoom = currentRoomId == wr.id;

        return Card(
          elevation: 4,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: isCurrentRoom
                ? const BorderSide(color: Colors.greenAccent, width: 2)
                : BorderSide.none,
          ),
          margin: const EdgeInsets.symmetric(vertical: 8),
          child: Container(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
              gradient: LinearGradient(
                colors: isCurrentRoom
                    ? [Colors.deepPurple.shade800, Colors.deepPurple.shade600]
                    : [
                        Colors.deepPurpleAccent.shade700,
                        Colors.deepPurpleAccent
                      ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
            ),
            child: ListTile(
              contentPadding:
                  const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              leading: CircleAvatar(
                backgroundColor: Colors.white24,
                child: Icon(
                  Icons.person,
                  color: Colors.white,
                ),
              ),
              title: Text(
                wr.host,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
              subtitle: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const SizedBox(height: 6),
                  Row(
                    children: [
                      Icon(Icons.attach_money, size: 16, color: Colors.amber),
                      const SizedBox(width: 4),
                      Text(
                        'Bet: ${wr.betAmt / 1e11} DCR',
                        style: const TextStyle(color: Colors.white70),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      Icon(Icons.tag, size: 16, color: Colors.white54),
                      const SizedBox(width: 4),
                      Text(
                        'Room ID: ${wr.id}',
                        style: const TextStyle(
                            color: Colors.white60, fontSize: 12),
                      ),
                    ],
                  ),
                ],
              ),
              trailing: currentRoomId != wr.id
                  ? currentRoomId == null
                      ? ElevatedButton(
                          onPressed: () => onJoinRoom(wr.id),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blueAccent,
                            foregroundColor: Colors.white,
                            elevation: 2,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                            padding: const EdgeInsets.symmetric(
                                horizontal: 16, vertical: 10),
                          ),
                          child: const Text("Join",
                              style: TextStyle(fontWeight: FontWeight.bold)),
                        )
                      : null
                  : Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 6),
                      decoration: BoxDecoration(
                        color: Colors.green.withOpacity(0.3),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.greenAccent, width: 1),
                      ),
                      child: const Text(
                        "Joined",
                        style: TextStyle(
                          color: Colors.greenAccent,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
            ),
          ),
        );
      },
    );
  }
}
