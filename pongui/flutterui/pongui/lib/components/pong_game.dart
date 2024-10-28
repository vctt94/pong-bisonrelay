import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:pongui/grpc/grpc_client.dart';

class PongGame {
  final GrpcPongClient grpcClient; // gRPC client instance
  final String clientId;
  Timer? _inputTimer; // Timer to repeatedly send input

  PongGame(this.clientId, this.grpcClient);

  Widget buildWidget(Map<String, dynamic> gameState, FocusNode focusNode) {
    return GestureDetector(
      onPanUpdate: handlePaddleMovement,
      onTap: () => focusNode.requestFocus(),
      child: Focus(
        child: KeyboardListener(
          focusNode: focusNode..requestFocus(),
          onKeyEvent: (KeyEvent event) {
            if (event is KeyDownEvent) {
              String keyLabel = event.logicalKey.keyLabel;
              handleInput(clientId, keyLabel, true); // Start handling input
            } else if (event is KeyUpEvent) {
              _inputTimer?.cancel(); // Stop handling input when key is released
            }
          },
          child: LayoutBuilder(
            builder: (context, constraints) {
              return CustomPaint(
                size: Size(constraints.maxWidth, constraints.maxHeight),
                painter: PongPainter(gameState),
              );
            },
          ),
        ),
      ),
    );
  }

  void handlePaddleMovement(DragUpdateDetails details) {
    double deltaY = details.delta.dy;
    String data = deltaY < 0 ? 'ArrowUp' : 'ArrowDown';
    grpcClient.sendInput(clientId, data);
  }

  Future<void> handleInput(String clientId, String data, bool isKeyDown) async {
    if (isKeyDown) {
      // Send the first input immediately
      await _sendKeyInput(data);

      // Start the timer to continuously send input
      _inputTimer?.cancel(); // Cancel any existing timer
      _inputTimer = Timer.periodic(Duration(milliseconds: 100), (timer) {
        _sendKeyInput(data); // Send input at intervals
      });
    } else {
      // Stop the timer when the key is released
      _inputTimer?.cancel();
    }
  }

  Future<void> _sendKeyInput(String data) async {
    try {
      String action;

      // Translate raw key label to action
      if (data == 'W' || data == 'Arrow Up') {
        action = 'ArrowUp';
      } else if (data == 'S' || data == 'Arrow Down') {
        action = 'ArrowDown';
      } else {
        // Ignore unhandled keys
        return;
      }

      // Send the action via gRPC
      await grpcClient.sendInput(clientId, action);
    } catch (e) {
      print(e);
      // Handle error if needed
    }
  }

  @override
  String get name => 'Pong';
}

class PongPainter extends CustomPainter {
  final Map<String, dynamic> gameState;

  PongPainter(this.gameState);

  @override
  void paint(Canvas canvas, Size size) {
    // Extract game dimensions
    double gameWidth = (gameState['gameWidth'] as num?)?.toDouble() ?? 80.0;
    double gameHeight = (gameState['gameHeight'] as num?)?.toDouble() ?? 40.0;

    // Calculate scaling factors
    double scaleX = size.width / gameWidth;
    double scaleY = size.height / gameHeight;

    // Extract and scale paddle 1 properties
    double paddle1X = 0.0; // Paddle 1 is on the left edge
    double paddle1Y = (gameState['p1Y'] as num?)?.toDouble() ?? 0.0;
    double paddle1Width = (gameState['p1Width'] as num?)?.toDouble() ?? 1.0;
    double paddle1Height = (gameState['p1Height'] as num?)?.toDouble() ?? 5.0;

    // Scale paddle 1 properties
    paddle1X *= scaleX;
    paddle1Y *= scaleY;
    paddle1Width *= scaleX;
    paddle1Height *= scaleY;

    // Extract and scale paddle 2 properties
    double paddle2X = (gameState['p2X'] as num?)?.toDouble() ?? gameWidth - 1.0;
    double paddle2Y = (gameState['p2Y'] as num?)?.toDouble() ?? 0.0;
    double paddle2Width = (gameState['p2Width'] as num?)?.toDouble() ?? 1.0;
    double paddle2Height = (gameState['p2Height'] as num?)?.toDouble() ?? 5.0;

    // Scale paddle 2 properties
    paddle2X *= scaleX;
    paddle2Y *= scaleY;
    paddle2Width *= scaleX;
    paddle2Height *= scaleY;

    // Extract and scale ball properties
    double ballX = (gameState['ballX'] as num?)?.toDouble() ?? gameWidth / 2;
    double ballY = (gameState['ballY'] as num?)?.toDouble() ?? gameHeight / 2;
    double ballWidth = (gameState['ballWidth'] as num?)?.toDouble() ?? 1.0;
    double ballHeight = (gameState['ballHeight'] as num?)?.toDouble() ?? 1.0;

    // Scale ball properties
    ballX *= scaleX;
    ballY *= scaleY;
    ballWidth *= scaleX;
    ballHeight *= scaleY;

    // Paint object for drawing
    var paint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill;

    // Draw background
    canvas.drawRect(
      Rect.fromLTWH(0.0, 0.0, size.width, size.height),
      Paint()..color = Colors.black,
    );

    // Draw Paddle 1
    canvas.drawRect(
      Rect.fromLTWH(paddle1X, paddle1Y, paddle1Width, paddle1Height),
      paint,
    );

    // Draw Paddle 2
    canvas.drawRect(
      Rect.fromLTWH(paddle2X, paddle2Y, paddle2Width, paddle2Height),
      paint,
    );

    // Draw the ball
    canvas.drawRect(
      Rect.fromLTWH(ballX, ballY, ballWidth, ballHeight),
      paint,
    );
  }

  @override
  bool shouldRepaint(PongPainter oldDelegate) {
    // Repaint whenever the game state changes
    return oldDelegate.gameState != gameState;
  }
}
