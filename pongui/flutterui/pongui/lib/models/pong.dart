import 'dart:convert';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/config.dart';
import 'package:pongui/grpc/generated/pong.pbgrpc.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:pongui/models/notifications.dart';

class PongModel extends ChangeNotifier {
  late GrpcPongClient grpcClient;
  late PongGame pongGame;
  final NotificationModel notificationModel;

  String clientId = '';
  String nick = '';
  bool isReady = false;
  bool gameStarted = false;
  bool isConnected = false;
  double betAmt = 0;
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom currentWR = const LocalWaitingRoom("", "", 0.0);
  Map<String, dynamic> gameState = {};

  PongModel(Config cfg, this.notificationModel) {
    _initPongClient(cfg);
  }

  Future<void> _initPongClient(Config cfg) async {
    try {
      if (clientId.isNotEmpty) {
        return;
      }
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

      isConnected = true;
      startListeningToNtfn(grpcClient, clientId);
      notifyListeners();
    } catch (exception) {
      print("Exception: $exception");
      isConnected = false;
      notifyListeners();
    }
  }

  void startListeningToNtfn(GrpcPongClient grpcClient, String clientId) {
    grpcClient.startNtfnStreamRequest(clientId).listen((ntfn) {
      developer.log("Notification Stream Response: $ntfn");

      isConnected = true;
      notifyListeners();

      switch (ntfn.notificationType) {
        case NotificationType.BET_AMOUNT_UPDATE:
        if (ntfn.playerId == clientId) {
          betAmt = ntfn.betAmt;
          notificationModel.showNotification(
            "Bet Amount Updated: ${ntfn.betAmt}",
          );
        }
        case NotificationType.ON_WR_CREATED:
          waitingRooms.add(LocalWaitingRoom(
            ntfn.wr.id,
            ntfn.wr.hostId,
            ntfn.wr.betAmt,
          ));
          notificationModel.showNotification(
            "Waiting room created by ${ntfn.wr.hostId}",
          );
          notifyListeners();
          break;
        case NotificationType.GAME_START:
          gameStarted = true;
          notificationModel.showNotification(
            "Game started with ID: ${ntfn.gameId}",
          );
          notifyListeners();
          break;

        case NotificationType.PLAYER_JOINED_WR:
          if (ntfn.playerId == clientId) {
            currentWR = LocalWaitingRoom(ntfn.wr.id, ntfn.wr.hostId, ntfn.wr.betAmt);
          }
          notificationModel.showNotification("A new player joined the waiting room");
          break;

        case NotificationType.GAME_END:
          notificationModel.showNotification(ntfn.message);
          resetGameState();
          break;

        default:
          developer.log("Unknown notification type: ${ntfn.notificationType}");
      }
    }, onError: (error) {
      errorMessage = "Error in notification stream: ${error.message}";
      developer.log("Error: $error");
      print("Error: $error");
      isConnected = false;
      notifyListeners();
    });
  }

  void resetGameState() {
    isReady = false;
    waitingRooms.removeWhere((room) => room.id == currentWR.id);
    currentWR = const LocalWaitingRoom("", "", 0.0);
    gameStarted = false;
    notifyListeners();
  }

  Future<void> joinWaitingRoom(String id) async {
    try {
      await Golib.JoinWaitingRoom(id);
      errorMessage = '';
      notifyListeners();
    } catch (e) {
      errorMessage = "Error joining waiting room: $e";
      print("Error: $e");
      notifyListeners();
    }
  }

  void toggleReady() {
    if (currentWR.id.isEmpty) {
      var error = "Need to get into a waiting room to get ready.";
      errorMessage = error;
      notifyListeners();
      throw Exception(error);
    }
    grpcClient.startGameStreamRequest(clientId).listen((gameUpdate) {
      var data = utf8.decode(gameUpdate.data);
      var parsedData = json.decode(data) as Map<String, dynamic>;
      gameStarted = true;
      gameState = parsedData;
      errorMessage = '';
      notifyListeners();
    }, onError: (error) {
      developer.log("Error in game stream: $error");
      errorMessage = "Error in game stream: ${error.message}";
      print("Error: $error");
      isConnected = false;
      notifyListeners();
    });
    isReady = !isReady;
    notifyListeners();
  }
}