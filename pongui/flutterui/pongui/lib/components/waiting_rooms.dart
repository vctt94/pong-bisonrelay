import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';

class WaitingRoomList extends StatelessWidget {
  final List<LocalWaitingRoom> waitingRooms;
  final Function(String roomId) onJoinRoom;

  const WaitingRoomList(this.waitingRooms, this.onJoinRoom, {Key? key})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    if (waitingRooms.isEmpty) {
      return const Center(
        child: Text(
          'No active waiting rooms',
          style: TextStyle(fontSize: 18, color: Colors.white70),
        ),
      );
    }

    return ListView.builder(
      itemCount: waitingRooms.length,
      itemBuilder: (context, index) {
        final wr = waitingRooms[index];
        return Card(
          color: Colors.deepPurpleAccent,
          margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
          child: ListTile(
            leading: const Icon(Icons.person, color: Colors.white),
            title: Text(
              wr.host,
              style: const TextStyle(color: Colors.white),
            ),
            subtitle: Text(
              'Bet: ${wr.betAmt} DCR',
              style: const TextStyle(color: Colors.white70),
            ),
            trailing: ElevatedButton(
              onPressed: () => onJoinRoom(wr.id),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blueAccent,
              ),
              child: const Text("Join"),
            ),
          ),
        );
      },
    );
  }
}
