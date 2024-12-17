import 'dart:convert';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/config.dart';
import 'package:pongui/grpc/generated/pong.pbgrpc.dart';
import 'package:pongui/grpc/grpc_client.dart';

class PongModel extends ChangeNotifier {
  late GrpcPongClient grpcClient;
  late PongGame pongGame;

  String clientId = '';
  String nick = '';
  bool isReady = false;
  bool gameStarted = false;
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom currentWR = const LocalWaitingRoom("", "", 0.0);
  Map<String, dynamic> gameState = {};

  PongModel(Config cfg) {
    _initPongClient(cfg);
  }

  Future<void> _initPongClient(Config cfg) async {
    try {
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

      clientId = localInfo.id;
      nick = localInfo.nick;
      var rooms = await Golib.getWaitingRooms();
      waitingRooms = rooms;
      List<String> parts = cfg.serverAddr.split(":");
      String ipAddress = parts[0];
      int port = int.parse(parts[1]);
      grpcClient =
          GrpcPongClient(ipAddress, port, tlsCertPath: cfg.grpcCertPath);
      print("Connecting to gRPC server: $ipAddress:$port");
      pongGame = PongGame(clientId, grpcClient);

      startListeningToNtfn(grpcClient, clientId);
      notifyListeners();
    } catch (exception) {
      print("Exception: $exception");
    }
  }

  void startListeningToNtfn(GrpcPongClient grpcClient, String clientId) {
    grpcClient.startNtfnStreamRequest(clientId).listen((ntfn) {
      print(ntfn);

      developer.log("Notification Stream Response: $ntfn");

      switch (ntfn.notificationType) {
        case NotificationType.ON_WR_CREATED:
          waitingRooms.add(LocalWaitingRoom(
            ntfn.wr.id,
            ntfn.wr.hostId,
            ntfn.wr.betAmt,
          ));
          notifyListeners();
          break;
        case NotificationType.GAME_START:
          // Handle game start notification
          gameStarted = true;
          gameState = {
            'gameId': ntfn.gameId,
            'message': ntfn.message,
            'started': ntfn.started,
          };
          errorMessage = ''; // Clear any previous errors
          print("Game Started: ${ntfn.message}");
          notifyListeners(); // Notify UI to update
          break;
        default:
          developer.log("Unknown notification type: ${ntfn.notificationType}");
      }
    }, onError: (error) {
      errorMessage = "Error: ${error.message}";
      print(errorMessage);
      notifyListeners();
      print("ERROR: $error");
      developer.log("Error in notification stream: $error");
    });
  }

  void startGameStream() {
    grpcClient.startGameStreamRequest(clientId).listen((response) {
      print(response.data);
      var data = utf8.decode(response.data);
      gameState = json.decode(data);
      gameStarted = true;
      errorMessage = '';
      notifyListeners();
    }, onError: (error) {
      errorMessage = "Error: ${error.message}";
      gameStarted = false;
      notifyListeners();
    });
  }

  Future<void> joinWaitingRoom(String id) async {
    try {
      currentWR = await Golib.JoinWaitingRoom(id);
      print("Successfully joined: $currentWR");
      errorMessage = ''; // Clear any previous error messages
      notifyListeners();
    } catch (e) {
      errorMessage = "Error joining waiting room: $e";
      print("Error: $e");
      notifyListeners();
    }
  }

  void toggleReady() {
    if (currentWR.id == "") {
      return;
    }
    grpcClient.startGameStreamRequest(clientId).listen((gameUpdate) {
      var data = utf8.decode(gameUpdate.data);
      var parsedData = json.decode(data) as Map<String, dynamic>;
      gameStarted = true;
      gameState = parsedData;
      errorMessage = '';
    }, onError: (error) {
      developer.log("Error in notification stream: $error");
    });
    isReady = !isReady;
    notifyListeners();
  }
}
