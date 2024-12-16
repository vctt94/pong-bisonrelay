import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:pongui/components/app_bar.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/components/waiting_rooms.dart';
import 'package:pongui/grpc/generated/pong.pbgrpc.dart'; // If needed

class HomeScreen extends StatelessWidget {
  final bool isReady;
  final bool isLoading;
  final bool gameStarted;
  final String errorMessage;
  final PongGame pongGame;
  final Map<String, dynamic> gameState;
  final LocalWaitingRoom currentWR;
  final double betAmount;
  final List<LocalWaitingRoom> waitingRooms;
  final TextEditingController roomIdController;
  final String serverAddr;
  final String clientId;

  final VoidCallback _createWaitingRoom;
  final VoidCallback _toggleReady;
  final VoidCallback _retryGameStream;
  final Function(String) _handleJoinRoom;

  const HomeScreen({
    Key? key,
    required this.isReady,
    required this.isLoading,
    required this.gameStarted,
    required this.errorMessage,
    required this.pongGame,
    required this.gameState,
    required this.currentWR,
    required this.betAmount,
    required this.waitingRooms,
    required this.roomIdController,
    required this.serverAddr,
    required this.clientId,
    required VoidCallback createWaitingRoom,
    required VoidCallback toggleReady,
    required VoidCallback retryGameStream,
    required Function(String) handleJoinRoom,
  })  : _createWaitingRoom = createWaitingRoom,
        _toggleReady = toggleReady,
        _retryGameStream = retryGameStream,
        _handleJoinRoom = handleJoinRoom,
        super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: PongGameAppBar(
        isReady: isReady,
        currentWR: currentWR,
        createWaitingRoom: _createWaitingRoom,
        toggleReady: _toggleReady,
        betAmount: betAmount,
      ),
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: <Widget>[
            DrawerHeader(
              decoration: BoxDecoration(color: Colors.blueAccent),
              child: Text(
                'Game Menu',
                style: TextStyle(color: Colors.white, fontSize: 24),
              ),
            ),
            ListTile(
              leading: Icon(Icons.home),
              title: Text('Home'),
              onTap: () {
                Navigator.pop(context);
                Navigator.pushNamed(context, '/');
              },
            ),
            ListTile(
              leading: Icon(Icons.leaderboard),
              title: Text('Leaderboard'),
              onTap: () {
                Navigator.pop(context);
                Navigator.pushNamed(context, '/leaderboard');
              },
            ),
            ListTile(
              leading: Icon(Icons.settings),
              title: Text('Settings'),
              onTap: () {
                Navigator.pop(context);
                Navigator.pushNamed(context, '/settings');
              },
            ),
          ],
        ),
      ),
      body: isLoading
          ? Center(child: CircularProgressIndicator())
          : Stack(
              children: [
                Center(
                  child: errorMessage.isNotEmpty
                      ? AlertDialog(
                          title: Text('Connection Error'),
                          content: Text(errorMessage),
                          actions: [
                            TextButton(
                              onPressed: _retryGameStream,
                              child: Text('Retry'),
                            ),
                          ],
                        )
                      : isReady
                          ? gameStarted
                              ? pongGame.buildWidget(
                                  gameState,
                                  FocusNode(),
                                )
                              : Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: [
                                    Icon(
                                      Icons.sports_tennis,
                                      size: 100,
                                      color: Colors.blueAccent,
                                    ),
                                    SizedBox(height: 20),
                                    Text(
                                      'Waiting for all players to get ready...',
                                      style: TextStyle(fontSize: 18),
                                    ),
                                    if (currentWR.id.isNotEmpty)
                                      Text(
                                        'Joined Room: ${currentWR.id}',
                                        style: TextStyle(
                                          fontSize: 16,
                                          color: Colors.white70,
                                        ),
                                      ),
                                  ],
                                )
                          : Padding(
                              padding: const EdgeInsets.all(16.0),
                              child: SingleChildScrollView(
                                child: Column(
                                  children: [
                                    Padding(
                                      padding: const EdgeInsets.symmetric(
                                          vertical: 10.0),
                                      child: Row(
                                        children: [
                                          Expanded(
                                            child: TextField(
                                              controller: roomIdController,
                                              decoration: InputDecoration(
                                                labelText: 'Enter Room ID',
                                                border: OutlineInputBorder(),
                                              ),
                                            ),
                                          ),
                                          SizedBox(width: 10),
                                          ElevatedButton(
                                            onPressed: () {
                                              _handleJoinRoom(
                                                  roomIdController.text);
                                            },
                                            child: Text('Join Room'),
                                            style: ElevatedButton.styleFrom(
                                              backgroundColor:
                                                  Colors.blueAccent,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    ConstrainedBox(
                                      constraints: BoxConstraints(
                                          maxHeight: MediaQuery.of(context)
                                                  .size
                                                  .height -
                                              200),
                                      child: WaitingRoomList(
                                          waitingRooms, _handleJoinRoom),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                ),
                Positioned(
                  bottom: 0,
                  left: 0,
                  right: 0,
                  child: Container(
                    padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                    color: Colors.black54,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Connected to Server: $serverAddr',
                          style: TextStyle(fontSize: 16, color: Colors.white70),
                        ),
                        SizedBox(height: 5),
                        Text(
                          'Client ID: $clientId',
                          style: TextStyle(fontSize: 16, color: Colors.white70),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
    );
  }
}
