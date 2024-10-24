import 'dart:async';
import 'dart:io';
import 'dart:developer' as developer;
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/config.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:pongui/screens/newconfig.dart';

void main(List<String> args) async {
  WidgetsFlutterBinding.ensureInitialized();

  // Pass args and initialize config
  runApp(MyApp(args));
}

class MyApp extends StatefulWidget {
  final List<String> args;

  MyApp(this.args);

  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  Config? config;
  String serverAddr = '';
  bool isLoading = true;
  String errorMessage = '';
  bool isReady = false;
  bool gameStarted = false;
  GrpcPongClient? grpcClient;
  String clientId = '';
  Map<String, dynamic> gameState = {};
  late PongGame pongGame;

  @override
  void initState() {
    super.initState();
    _initializeApp(widget.args);
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

  Future<void> _initializeApp(List<String> args) async {
    try {
      // Load the configuration from the args.
      final filename = await configFileName(args);
      final cfg = await configFromArgs(args);

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
      clientId = await Golib.initClient(initArgs);
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
      developer.log("Error: $exception");
      if (exception == usageException) {
        exit(0);
      } else if (exception == newConfigNeededException) {
        runNewConfigApp(widget.args);
      } else {
        setState(() {
          errorMessage = 'Error: $exception';
          isLoading = false;
        });
      }
    }
  }

  void _startGameStream() {
    if (grpcClient != null && clientId.isNotEmpty) {
      setState(() {
        isReady = true;
      });
      grpcClient!.startGameStreamRequest(clientId).listen((response) {
        var data = utf8.decode(response.data);
        var parsedData = json.decode(data) as Map<String, dynamic>; // Decode the JSON
        setState(() {
          gameState = parsedData;
        });
        developer.log("Game Stream Started: $response");
      }, onError: (error) {
        developer.log("Error in game stream: $error");
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Pong Game App',
      home: Scaffold(
        appBar: AppBar(
          title: Text(
            'Pong Game',
            style: TextStyle(color: const Color.fromARGB(255, 202, 202, 202)), // Set text color to white
          ),
          backgroundColor: const Color.fromARGB(255, 25, 23, 44),
        ),
        body: isLoading
            ? Center(child: CircularProgressIndicator())
            : errorMessage.isNotEmpty
                ? Center(child: Text(errorMessage))
                : Stack(
                    children: [
                      // Display server address and client ID at the top left
                      Positioned(
                        top: 10,
                        left: 10,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Connected to Server: $serverAddr',
                              style: TextStyle(
                                fontSize: 16,
                                color: Colors.black54,
                              ),
                            ),
                            SizedBox(height: 5),
                            Text(
                              'Client ID: $clientId',
                              style: TextStyle(
                                fontSize: 16,
                                color: Colors.black54,
                              ),
                            ),
                          ],
                        ),
                      ),

                      // Display Pong Game if ready
                      Center(
                        child: isReady
                            ? gameStarted ?  pongGame.buildWidget(
                                gameState,
                                FocusNode()
                              ) : Column(
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
                                    style: TextStyle(
                                      fontSize: 18,
                                      color: Colors.blueAccent,
                                    ),
                                  ),
                                ],
                              )
                            : ElevatedButton(
                                onPressed: _startGameStream,
                                style: ElevatedButton.styleFrom(
                                  padding: EdgeInsets.symmetric(
                                      horizontal: 40, vertical: 20),
                                  backgroundColor: Colors.blueAccent,
                                ),
                                child: Text(
                                  'Start Game',
                                  style: TextStyle(
                                    fontSize: 18,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                      ),
                    ],
                  ),
      ),
    );
  }
}
