import 'dart:convert';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/config.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';
import 'package:golib_plugin/grpc/generated/pong.pbgrpc.dart';
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
  int betAmt = 0;
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom? currentWR;
  GameUpdate? gameState;

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
      // XXX this is not correct, need to check if error is eof
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
            print("Bet amount updated: ${ntfn.betAmt}");
            betAmt = ntfn.betAmt.toInt();
          }
          break;

        case NotificationType.ON_WR_CREATED:
          waitingRooms.add(LocalWaitingRoom.fromProto(ntfn.wr));
          notificationModel.showNotification(
            "Waiting room created by ${ntfn.wr.hostId}",
          );
          notifyListeners();
          break;

        case NotificationType.GAME_START:
          gameStarted = true;
          // can set current wr as null after game starting
          currentWR = null;
          notificationModel.showNotification(
            "Game started with ID: ${ntfn.gameId}",
          );
          notifyListeners();
          break;

        case NotificationType.PLAYER_JOINED_WR:
          if (ntfn.playerId == clientId) {
            currentWR = LocalWaitingRoom.fromProto(ntfn.wr);
          }
          notificationModel
              .showNotification("A new player joined the waiting room");
          break;

        case NotificationType.GAME_END:
          notificationModel.showNotification(ntfn.message);
          resetGameState();
          break;

        case NotificationType.ON_WR_REMOVED:
          // Handle the waiting room removal
          waitingRooms.removeWhere((room) => room.id == ntfn.roomId);
          currentWR = null;
          notificationModel.showNotification(
            "Waiting room removed: ${ntfn.roomId}",
          );
          notifyListeners();
          break;

        case NotificationType.OPPONENT_DISCONNECTED:
          gameStarted = false;
          currentWR = LocalWaitingRoom.fromProto(ntfn.wr);
          notificationModel.showNotification(ntfn.message);
          notifyListeners();
          break;

        case NotificationType.ON_PLAYER_READY:
          // Update the waiting room with the player's ready status
          if (currentWR != null) {
            // Find the player in the current waiting room and update their ready status
            for (var i = 0; i < currentWR!.players.length; i++) {
              if (currentWR!.players[i].uid == ntfn.playerId) {
                currentWR!.players[i].ready = ntfn.ready;
                break;
              }
            }

            // Show notification
            String playerName = ntfn.playerId;
            String readyStatus = ntfn.ready ? "ready" : "not ready";
            notificationModel.showNotification(
              "Player $playerName is now $readyStatus",
            );
            notifyListeners();
          }
          break;

        default:
          developer.log("Unknown notification type: ${ntfn.notificationType}");
      }
    }, onError: (error) {
      errorMessage = "Error in notification stream: ${error.message}";
      developer.log("Error: $error");
      print("Error: $error");
      // XXX this is not correct, need to check if error is eof
      isConnected = false;
      notifyListeners();
    });
  }

  void resetGameState() {
    isReady = false;
    currentWR = null;
    gameStarted = false;
    betAmt = 0;
    notifyListeners();
  }

  void clearErrorMessage() {
    errorMessage = '';
    notifyListeners();
  }

  Future<void> createWaitingRoom() async {
    try {
      if (betAmt <= 0) {
        errorMessage = "bet amount needs to be higher than 0";
        notifyListeners();
        return;
      }

      CreateWaitingRoomArgs createRoomArgs =
          CreateWaitingRoomArgs(clientId, betAmt);

      developer.log("CreateWaitingRoom args: $createRoomArgs");
      var roomInfo = await Golib.CreateWaitingRoom(createRoomArgs);

      // Update the model state
      currentWR = roomInfo;
      errorMessage = '';
      notifyListeners();

      notificationModel.showNotification(
        "Waiting room created with Bet Amount: ${roomInfo.betAmt}",
      );
    } catch (e) {
      errorMessage = "Error creating waiting room: $e";
      developer.log("Error creating waiting room: $e");
      notifyListeners();
    }
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
    if (currentWR == null) {
      var error = "Need to get into a waiting room to get ready.";
      errorMessage = error;
      notifyListeners();
      throw Exception(error);
    }

    if (!isReady) {
      // Player is getting ready
      grpcClient.startGameStreamRequest(clientId).listen((gameUpdateBytes) {
        final update = GameUpdate.fromBuffer(gameUpdateBytes.data);
        gameStarted = true;
        gameState = update;
        errorMessage = '';
        notifyListeners();
      }, onError: (error) {
        developer.log("Error in game stream: $error");
        errorMessage = "Error in game stream: ${error.message}";
        print("Error: $error");
        notifyListeners();
      });
    } else {
      // Player is unreadying
      try {
        grpcClient.unreadyGameStream(clientId);
      } catch (error) {
        developer.log("Error in unready game stream: $error");
        errorMessage = "Error in unready game stream: $error";
        print("Error: $error");
        notifyListeners();
        return;
      }
    }

    isReady = !isReady;
    notifyListeners();
  }

  Future<void> leaveWaitingRoom() async {
    if (currentWR == null) {
      return;
    }

    try {
      await Golib.LeaveWaitingRoom(currentWR!.id);

      // Reset waiting room state
      currentWR = null;
      isReady = false;
      errorMessage = '';
      notifyListeners();

      notificationModel.showNotification("Left waiting room successfully");
    } catch (e) {
      errorMessage = "Error leaving waiting room: $e";
      developer.log("Error leaving waiting room: $e");
      notifyListeners();
    }
  }
}
