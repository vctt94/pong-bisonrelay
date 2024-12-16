import 'dart:async';
import 'dart:io';
import 'dart:developer' as developer;
import 'dart:convert';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/components/waiting_rooms.dart';
import 'package:pongui/config.dart';
import 'package:pongui/grpc/generated/pong.pbgrpc.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:pongui/models/newconfig.dart';
import 'package:pongui/screens/home.dart';
import 'package:pongui/screens/newconfig.dart';
import 'package:window_manager/window_manager.dart';

final Random random = Random();

void main(List<String> args) async {
  try {
    WidgetsFlutterBinding.ensureInitialized();
    if (Platform.isLinux || Platform.isWindows || Platform.isMacOS) {
      await windowManager.ensureInitialized();
    }

    developer.log("Platform: ${Golib.majorPlatform}/${Golib.minorPlatform}");
    Golib.platformVersion
        .then((value) => developer.log("Platform Version: $value"));
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
  late Config config;
  List<LocalWaitingRoom> waitingRooms = [];
  String serverAddr = '';
  bool isLoading = true;
  String errorMessage = '';
  bool isReady = false;
  bool gameStarted = false;
  GrpcPongClient? grpcClient;
  String nick = '';
  String clientId = '';
  double betAmount = 0.0;
  Map<String, dynamic> gameState = {};
  LocalWaitingRoom currentWR = const LocalWaitingRoom("", "", 0.0);
  late PongGame pongGame;

  @override
  void initState() {
    super.initState();
    initClient();
    windowManager.addListener(this);
  }

  void _startListeningToNtfn(GrpcPongClient grpcClient, String clientId) {
    grpcClient.startNtfnStreamRequest(clientId).listen((ntfn) {
      developer.log("Notification Stream Response: $ntfn");

      switch (ntfn.notificationType) {
        case NotificationType.ON_WR_CREATED:
          setState(() {
            waitingRooms.add(LocalWaitingRoom(
              ntfn.wr.id,
              ntfn.wr.hostId,
              ntfn.wr.betAmt,
            ));
          });
          break;
        default:
          developer.log("Unknown notification type: ${ntfn.notificationType}");
      }
    }, onError: (error) {
      developer.log("Error in notification stream: $error");
    });
  }

  Future<void> _onConfigSaved() async {
    runMainApp(config);
  }

  void _toggleReady() {
    if (currentWR.id == "") {
      return;
    }
    grpcClient?.startGameStreamRequest(clientId).listen((gameUpdate) {
      var data = utf8.decode(gameUpdate.data);
      var parsedData = json.decode(data) as Map<String, dynamic>;
      setState(() {
        gameState = parsedData;
        gameStarted = true;
        errorMessage = '';
      });
    }, onError: (error) {
      developer.log("Error in notification stream: $error");
    });
    setState(() {
      isReady = !isReady;
    });
  }

  Future<void> initClient() async {
    try {
      var cfg = widget.cfg;
      InitClient initArgs = InitClient(
        cfg.serverAddr,
        cfg.grpcCertPath,
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
      var rooms = await Golib.getWaitingRooms();
      setState(() {
        waitingRooms = rooms;
      });
      List<String> parts = cfg.serverAddr.split(":");
      String ipAddress = parts[0];
      int port = int.parse(parts[1]);
      grpcClient = GrpcPongClient(ipAddress, port);
      pongGame = PongGame(clientId, grpcClient!);

      setState(() {
        config = cfg;
        serverAddr = cfg.serverAddr;
        isLoading = false;
      });
      _startListeningToNtfn(grpcClient!, clientId);
    } catch (exception) {
      print("Exception: $exception");
    }
  }

  void _startGameStream() {
    if (grpcClient != null && clientId.isNotEmpty) {
      setState(() {
        isReady = true;
      });
      grpcClient!.startGameStreamRequest(clientId).listen((response) {
        var data = utf8.decode(response.data);
        var parsedData = json.decode(data) as Map<String, dynamic>;
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
    });
    _startGameStream();
  }

  void _handleJoinRoom(String id) async {
    var wr = await Golib.JoinWaitingRoom(id);
    setState(() {
      currentWR = wr;
    });
  }

  Future<void> _createWaitingRoom() async {
    try {
      if (grpcClient != null && clientId.isNotEmpty) {
        isReady = true;
      }

      developer.log("Waiting room created for Client ID: $clientId");
    } catch (error) {
      developer.log("Error creating waiting room: $error");
      setState(() {
        errorMessage = "Failed to create waiting room. Please try again.";
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final TextEditingController roomIdController = TextEditingController();

    return MaterialApp(
      title: 'Pong Game App',
      theme: ThemeData.dark().copyWith(
        scaffoldBackgroundColor: const Color.fromARGB(255, 25, 23, 44),
        primaryColor: Colors.blueAccent,
      ),
      // Define routes for navigation
      routes: {
        '/': (context) => HomeScreen(
              isReady: isReady,
              isLoading: isLoading,
              gameStarted: gameStarted,
              errorMessage: errorMessage,
              pongGame: pongGame,
              gameState: gameState,
              currentWR: currentWR,
              betAmount: betAmount,
              waitingRooms: waitingRooms,
              roomIdController: roomIdController,
              serverAddr: serverAddr,
              clientId: clientId,
              createWaitingRoom: _createWaitingRoom,
              toggleReady: _toggleReady,
              retryGameStream: _retryGameStream,
              handleJoinRoom: _handleJoinRoom,
            ),
        // '/leaderboard': (context) => LeaderboardScreen(),
        '/settings': (context) => NewConfigScreen(
              newConfig: NewConfigModel.fromConfig(config),
              onConfigSaved: _onConfigSaved,
            ),
      },
      // Initial route is '/'
      initialRoute: '/',
    );
  }
}
