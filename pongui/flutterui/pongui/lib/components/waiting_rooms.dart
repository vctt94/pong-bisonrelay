import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';

class WaitingRoomList extends StatelessWidget {
  final List<WaitingRoom> waitingRooms;

  const WaitingRoomList(this.waitingRooms);

  @override
  Widget build(BuildContext context) {
    return waitingRooms.isNotEmpty
        ? Column(
            children: [
              Text(
                'Waiting Rooms',
                style: TextStyle(fontSize: 24, color: Colors.white70),
              ),
              SizedBox(height: 10),
              Expanded(
                child: ListView.builder(
                  shrinkWrap: true,
                  itemCount: waitingRooms.length,
                  itemBuilder: (context, index) {
                    final wr = waitingRooms[index];
                    return Card(
                      color: Colors.deepPurpleAccent,
                      margin: EdgeInsets.symmetric(vertical: 8, horizontal: 16),
                      child: ListTile(
                        leading: Icon(Icons.person, color: Colors.white),
                        title: Text(
                          wr.host,
                          style: TextStyle(color: Colors.white),
                        ),
                        subtitle: Text(
                          'Bet: ${wr.betAmount} DCR',
                          style: TextStyle(color: Colors.white70),
                        ),
                        trailing: ElevatedButton(
                          onPressed: () {
                            // Join waiting room logic
                          },
                          child: Text("Join"),
                          style: ElevatedButton.styleFrom(
                            // primaryColor: Colors.green,
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          )
        : Center(
            child: Text(
              'No active waiting rooms',
              style: TextStyle(fontSize: 18, color: Colors.white70),
            ),
          );
  }
}
