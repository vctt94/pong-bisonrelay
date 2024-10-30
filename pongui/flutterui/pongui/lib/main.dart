import 'dart:async';
import 'dart:io';
import 'dart:developer' as developer;
import 'dart:convert';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/config.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:pongui/screens/newconfig.dart';
import 'package:window_manager/window_manager.dart';

final Random random = Random();

void main(List<String> args) async {
  try {
    WidgetsFlutterBinding.ensureInitialized();
    if (Platform.isLinux || Platform.isWindows || Platform.isMacOS) {
      windowManager.ensureInitialized();
    }

    developer.log("Platform: ${Golib.majorPlatform}/${Golib.minorPlatform}");
    Golib.platformVersion
        .then((value) => developer.log("Platform Version: $value"));
    // Pass args and initialize config
    Config cfg = await configFromArgs(args);
    runMainApp(cfg);
  } catch (exception) {
    print(exception);
    developer.log("Error: $exception");
    if (exception == usageException) {
      exit(0);
    } else if (exception == newConfigNeededException) {
      runNewConfigApp(args);
      return;
    } else {
      // runFatalErrorApp(exception);
    }
  }
}

void runMainApp(Config cfg) {
  runApp(MyApp(cfg));
}

class MyApp extends StatefulWidget {
  final Config cfg;

  const MyApp(this.cfg);

  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> with WindowListener {
  Config? config;
  List<Player> waitingPlayers = [];
  String serverAddr = '';
  bool isLoading = true;
  String errorMessage = '';
  bool isReady = false;
  bool gameStarted = false;
  GrpcPongClient? grpcClient;
  String nick = '';
  String clientId = '';
  Map<String, dynamic> gameState = {};
  late PongGame pongGame;

  @override
  void initState() {
    super.initState();
    initClient();
    windowManager.setPreventClose(true);
    windowManager.addListener(this);
  }

  void _startListeningToStreams(GrpcPongClient grpcClient, String clientId) {
    // Start notification stream
    grpcClient.startNtfnStreamRequest(clientId).listen((response) {
      if (response.started) {
        setState(() {
          gameStarted = true;
        });
      }
      developer.log("Notification Stream Response: ${response}");
      // Handle the response (e.g., update UI or handle game state)
    }, onError: (error) {
      developer.log("Error in notification stream: $error");
    });
  }

  Future<void> initClient() async {
    try {
      // Load the configuration from the args.
      var cfg = widget.cfg;

      InitClient initArgs = InitClient(
        cfg.serverAddr,
        "",
        "",
        "debug",
        true,
        cfg.rpcWebsocketURL,
        cfg.rpcCertPath,
        cfg.rpcClientCertPath,
        cfg.rpcClientKeyPath,
        cfg.rpcUser,
        cfg.rpcPass,
      );

      developer.log("InitClient args: $initArgs");
      var localInfo = await Golib.initClient(initArgs);

      setState(() {
        clientId = localInfo.id;
        nick = localInfo.nick;
      });
      var players = await Golib.getWRPlayers();
      setState(() {
        waitingPlayers = players;
      });
      print(waitingPlayers.length);
      List<String> parts = cfg.serverAddr.split(":");
      String ipAddress = parts[0]; // "127.0.0.1"
      int port = int.parse(parts[1]); // 50051 as an integer
      grpcClient = GrpcPongClient(
        ipAddress,
        port, // Assuming you have the port in the config
      );
      pongGame = PongGame(clientId, grpcClient!);

      setState(() {
        config = cfg;
        serverAddr = cfg.serverAddr;
        isLoading = false;
      });
      _startListeningToStreams(grpcClient!, clientId);
    } catch (exception) {
      print("**********************exception***********");
      print(exception);
    }
  }

  void _startGameStream() {
    if (grpcClient != null && clientId.isNotEmpty) {
      setState(() {
        isReady = true;
      });
      grpcClient!.startGameStreamRequest(clientId).listen((response) {
        var data = utf8.decode(response.data);
        var parsedData =
            json.decode(data) as Map<String, dynamic>; // Decode the JSON
        setState(() {
          gameState = parsedData;
          gameStarted = true;
          errorMessage = '';
        });
        developer.log("Game Stream Started: $response");
      }, onError: (error) {
        setState(() {
          errorMessage = "Error: ${error.message}";
          isReady = false;
          gameStarted = false;
        });
        developer.log("Error in game stream: $error");
      });
    }
  }

  void _retryGameStream() {
    setState(() {
      errorMessage = '';
      // Retry logic here
    });
    _startGameStream();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pong Game App',
      theme: ThemeData.dark().copyWith(
        scaffoldBackgroundColor: const Color.fromARGB(255, 25, 23, 44),
        primaryColor: Colors.blueAccent,
      ),
      home: Scaffold(
        appBar: AppBar(
          title: Text('Pong Game'),
          actions: [
            IconButton(
              icon: Icon(Icons.settings),
              onPressed: () {
                // Open settings screen
              },
            ),
          ],
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
                onTap: () => Navigator.pop(context),
              ),
              ListTile(
                leading: Icon(Icons.leaderboard),
                title: Text('Leaderboard'),
                onTap: () => Navigator.pop(context),
              ),
              ListTile(
                leading: Icon(Icons.settings),
                title: Text('Settings'),
                onTap: () => Navigator.pop(context),
              ),
            ],
          ),
        ),
        body: isLoading
            ? Center(child: CircularProgressIndicator())
            : Stack(
                children: [
                  Positioned(
                    top: 10,
                    left: 10,
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
                                        'Waiting for another player...',
                                        style: TextStyle(fontSize: 18),
                                      ),
                                    ],
                                  )
                            : Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  // Display waiting room players if there are any
                                  if (waitingPlayers.isNotEmpty)
                                    SizedBox(
                                      height:
                                          200, // Set a fixed height for the waiting list area
                                      child: ListView.builder(
                                        shrinkWrap: true,
                                        itemCount: waitingPlayers.length,
                                        itemBuilder: (context, index) {
                                          final player = waitingPlayers[index];
                                          print(player);
                                          return ListTile(
                                            title: Text(player.uid),
                                            subtitle: Text(
                                                'Bet: ${player.betAmount} DCR'),
                                          );
                                        },
                                      ),
                                    ),
                                  if (waitingPlayers.isNotEmpty)
                                    SizedBox(height: 20),
                                  // Start Game Button
                                  ElevatedButton.icon(
                                    onPressed: _startGameStream,
                                    icon: Icon(Icons.play_arrow),
                                    label: Text('Start Game'),
                                    style: ElevatedButton.styleFrom(
                                      padding: EdgeInsets.symmetric(
                                          horizontal: 40, vertical: 20),
                                    ),
                                  ),
                                ],
                              ),
                  ),
                ],
              ),
      ),
    );
  }
}
