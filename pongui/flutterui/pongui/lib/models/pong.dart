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
import 'package:path/path.dart' as path;

// Define a clear enum for game states
enum GameState {
  idle, // Initial state, not in a game or waiting room
  inWaitingRoom, // In a waiting room, not ready
  waitingRoomReady, // In a waiting room and marked as ready
  gameInitialized, // Game has started but not ready to play
  readyToPlay, // Signaled ready to play but waiting for opponent
  countdown, // Countdown in progress
  playing // Active gameplay
}

class PongModel extends ChangeNotifier {
  late GrpcPongClient grpcClient;
  late PongGame pongGame;
  final NotificationModel notificationModel;

  String clientId = '';
  String nick = '';
  int betAmt = 0;
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom? currentWR;
  GameUpdate? gameState;

  // Connection state
  bool isConnected = false;

  // Game related state
  GameState _currentGameState = GameState.idle;
  String currentGameId = '';
  String countdownMessage = '';

  // Getters for the game state
  GameState get currentGameState => _currentGameState;
  bool get isInGame =>
      _currentGameState != GameState.idle &&
      _currentGameState != GameState.inWaitingRoom &&
      _currentGameState != GameState.waitingRoomReady;
  bool get isReady => _currentGameState == GameState.waitingRoomReady;
  bool get isGameStarted =>
      _currentGameState != GameState.idle &&
      _currentGameState != GameState.inWaitingRoom &&
      _currentGameState != GameState.waitingRoomReady;
  bool get isReadyToPlay =>
      _currentGameState == GameState.readyToPlay ||
      _currentGameState == GameState.countdown ||
      _currentGameState == GameState.playing;
  bool get countdownStarted => _currentGameState == GameState.countdown;

  PongModel(Config cfg, this.notificationModel) {
    _initPongClient(cfg);
  }

  Future<void> _initPongClient(Config cfg) async {
    try {
      if (clientId.isNotEmpty) {
        return;
      }
      
      // Create log file path in app data directory
      final appDataDir = await defaultAppDataDir();
      final logFilePath = path.join(appDataDir, "logs", "pongui.log");
      
      InitClient initArgs = InitClient(
        cfg.serverAddr,
        cfg.grpcCertPath,
        appDataDir,
        logFilePath,
        "",
        cfg.debugLevel,
        cfg.wantsLogNtfns,
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
            betAmt = ntfn.betAmt.toInt();
            notifyListeners();
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
          if (_currentGameState == GameState.idle ||
              _currentGameState == GameState.inWaitingRoom ||
              _currentGameState == GameState.waitingRoomReady) {
            _currentGameState = GameState.gameInitialized;
          }
          // can set current wr as null after game starting
          currentWR = null;
          notificationModel.showNotification(
            "Game started with ID: ${ntfn.gameId}",
          );
          notifyListeners();
          break;

        case NotificationType.GAME_READY_TO_PLAY:
          // Store the game ID when we receive the ready to play notification
          currentGameId = ntfn.gameId;
          if (_currentGameState == GameState.idle ||
              _currentGameState == GameState.inWaitingRoom ||
              _currentGameState == GameState.waitingRoomReady) {
            _currentGameState = GameState.gameInitialized;
          }
          notificationModel.showNotification(
              "Game is ready! Signal when you're ready to play.");
          notifyListeners();
          break;

        case NotificationType.COUNTDOWN_UPDATE:
          countdownMessage = ntfn.message;
          _currentGameState = GameState.countdown;

          // When countdown reaches 0, transition to playing state
          if (ntfn.message.contains("0")) {
            _currentGameState = GameState.playing;
          }

          notificationModel.showNotification(ntfn.message);
          notifyListeners();
          break;

        case NotificationType.PLAYER_JOINED_WR:
          if (ntfn.playerId == clientId) {
            currentWR = LocalWaitingRoom.fromProto(ntfn.wr);
            _currentGameState = GameState.inWaitingRoom;
          }
          notificationModel
              .showNotification("A new player joined the waiting room");
          notifyListeners();
          break;

        case NotificationType.GAME_END:
          notificationModel.showNotification(ntfn.message);
          resetGameState();
          break;

        case NotificationType.ON_WR_REMOVED:
          // Handle the waiting room removal
          waitingRooms.removeWhere((room) => room.id == ntfn.roomId);

          // If we were in this waiting room, reset our state
          if (currentWR != null && currentWR!.id == ntfn.roomId) {
            currentWR = null;
            _currentGameState = GameState.idle;
          }

          notificationModel.showNotification(
            "Waiting room removed: ${ntfn.roomId}",
          );
          notifyListeners();
          break;

        case NotificationType.OPPONENT_DISCONNECTED:
          if (_currentGameState == GameState.playing) {
            _currentGameState = GameState.idle;
          }
          currentWR = LocalWaitingRoom.fromProto(ntfn.wr);
          notificationModel.showNotification(ntfn.message);
          notifyListeners();
          break;

        case NotificationType.ON_PLAYER_READY:
          // Check if this is a ready to play notification for game
          if (ntfn.gameId.isNotEmpty) {
            String playerName =
                ntfn.playerId == clientId ? "You are" : "Opponent is";
            notificationModel.showNotification("$playerName ready to play!");

            // If this is our own ready signal, update our state
            if (ntfn.playerId == clientId) {
              _currentGameState = GameState.readyToPlay;
            }
          }
          // Otherwise handle waiting room ready state
          else if (currentWR != null) {
            // Find the player in the current waiting room and update their ready status
            for (var i = 0; i < currentWR!.players.length; i++) {
              if (currentWR!.players[i].uid == ntfn.playerId) {
                currentWR!.players[i].ready = ntfn.ready;

                // If this is our own ready signal, update our state
                if (ntfn.playerId == clientId) {
                  _currentGameState = ntfn.ready
                      ? GameState.waitingRoomReady
                      : GameState.inWaitingRoom;
                }
                break;
              }
            }

            // Show notification
            String playerName = ntfn.playerId;
            String readyStatus = ntfn.ready ? "ready" : "not ready";
            notificationModel.showNotification(
              "Player $playerName is now $readyStatus",
            );
          }
          notifyListeners();
          break;

        default:
          developer.log("Unknown notification type: ${ntfn.notificationType}");
      }
    }, onError: (error) {
      errorMessage = "Error in notification stream: ${error.message}";
      developer.log("Error: $error");
      // XXX this is not correct, need to check if error is eof
      isConnected = false;
      notifyListeners();
    });
  }

  void resetGameState() {
    _currentGameState = GameState.idle;
    currentWR = null;
    betAmt = 0;
    currentGameId = '';
    countdownMessage = '';
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
      _currentGameState = GameState.inWaitingRoom;
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
      _currentGameState = GameState.inWaitingRoom;
      errorMessage = '';
      notifyListeners();
    } catch (e) {
      errorMessage = "Error joining waiting room: $e";
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

    if (_currentGameState != GameState.waitingRoomReady) {
      // Player is getting ready
      grpcClient.startGameStreamRequest(clientId).listen((gameUpdateBytes) {
        final update = GameUpdate.fromBuffer(gameUpdateBytes.data);
        gameState = update;
        errorMessage = '';
        notifyListeners();
      }, onError: (error) {
        developer.log("Error in game stream: $error");
        errorMessage = "Error in game stream: ${error.message}";
        notifyListeners();
      });

      _currentGameState = GameState.waitingRoomReady;
    } else {
      // Player is unreadying
      try {
        grpcClient.unreadyGameStream(clientId);
        _currentGameState = GameState.inWaitingRoom;
      } catch (error) {
        developer.log("Error in unready game stream: $error");
        errorMessage = "Error in unready game stream: $error";
        notifyListeners();
        return;
      }
    }

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
      _currentGameState = GameState.idle;
      errorMessage = '';
      notifyListeners();

      notificationModel.showNotification("Left waiting room successfully");
    } catch (e) {
      errorMessage = "Error leaving waiting room: $e";
      developer.log("Error leaving waiting room: $e");
      notifyListeners();
    }
  }

  // Signal that the player is ready to play
  Future<void> signalReadyToPlay() async {
    try {
      if (currentGameId.isEmpty) {
        errorMessage = "No active game found";
        notifyListeners();
        return;
      }

      final response =
          await grpcClient.signalReadyToPlay(clientId, currentGameId);

      if (response.success) {
        _currentGameState = GameState.readyToPlay;
        notificationModel.showNotification("You are ready to play!");
      } else {
        errorMessage = response.message;
      }

      notifyListeners();
    } catch (e) {
      errorMessage = "Error signaling ready to play: $e";
      notifyListeners();
    }
  }
}
