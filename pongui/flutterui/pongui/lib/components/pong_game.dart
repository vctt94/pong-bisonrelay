import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';

class PongGame {
  final GrpcPongClient grpcClient; // gRPC client instance
  final String clientId;

  PongGame(this.clientId, this.grpcClient);

  Widget buildWidget(GameUpdate gameState, FocusNode focusNode) {
    return GestureDetector(
      onPanUpdate: handlePaddleMovement,
      onTap: () => focusNode.requestFocus(),
      child: Focus(
        child: KeyboardListener(
          focusNode: focusNode..requestFocus(),
          onKeyEvent: (KeyEvent event) {
            if (event is KeyDownEvent || event is KeyRepeatEvent) {
              String keyLabel = event.logicalKey.keyLabel;
              handleInput(clientId, keyLabel);
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

  Future<void> handleInput(String clientId, String data) async {
    await _sendKeyInput(data);
  }

  Future<void> _sendKeyInput(String data) async {
    try {
      String action;

      if (data == 'W' || data == 'Arrow Up') {
        action = 'ArrowUp';
      } else if (data == 'S' || data == 'Arrow Down') {
        action = 'ArrowDown';
      } else {
        return;
      }
      await grpcClient.sendInput(clientId, action);
    } catch (e) {
      print(e);
    }
  }

  @override
  String get name => 'Pong';
}

class PongPainter extends CustomPainter {
  final GameUpdate gameState;

  PongPainter(this.gameState);

  @override
  void paint(Canvas canvas, Size size) {
    // Extract game dimensions
    double gameWidth = gameState.gameWidth;
    double gameHeight = gameState.gameHeight;

    // Calculate scaling factors
    double scaleX = size.width / gameWidth;
    double scaleY = size.height / gameHeight;

    // Paint object for drawing
    var paint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill;

    // Draw background
    canvas.drawRect(
      Rect.fromLTWH(0.0, 0.0, size.width, size.height),
      Paint()..color = Colors.black,
    );

    // Extract and scale paddle 1 properties
    double paddle1X = 0.0; // Paddle 1 is on the left edge
    double paddle1Y = gameState.p1Y;
    double paddle1Width = gameState.p1Width;
    double paddle1Height = gameState.p1Height;

    // Scale paddle 1 properties
    paddle1X *= scaleX;
    paddle1Y *= scaleY;
    paddle1Width *= scaleX;
    paddle1Height *= scaleY;

    // Extract and scale paddle 2 properties
    double paddle2X = gameState.p2X;
    double paddle2Y = gameState.p2Y;
    double paddle2Width = gameState.p2Width;
    double paddle2Height = gameState.p2Height;

    // Scale paddle 2 properties
    paddle2X *= scaleX;
    paddle2Y *= scaleY;
    paddle2Width *= scaleX;
    paddle2Height *= scaleY;

    // Extract and scale ball properties
    double ballX = gameState.ballX;
    double ballY = gameState.ballY;
    double ballWidth = gameState.ballWidth;
    double ballHeight = gameState.ballHeight;

    // Scale ball properties
    ballX *= scaleX;
    ballY *= scaleY;
    ballWidth *= scaleX;
    ballHeight *= scaleY;

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

    // Draw scores
    int p1Score = gameState.p1Score;
    int p2Score = gameState.p2Score;

    // Create text painters for scores
    final p1ScoreTextPainter = TextPainter(
      text: TextSpan(
        text: '$p1Score',
        style: TextStyle(
            color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
      ),
      textDirection: TextDirection.ltr,
    );

    final p2ScoreTextPainter = TextPainter(
      text: TextSpan(
        text: '$p2Score',
        style: TextStyle(
            color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
      ),
      textDirection: TextDirection.ltr,
    );

    // Layout the text
    p1ScoreTextPainter.layout();
    p2ScoreTextPainter.layout();

    // Position and draw the scores at the top of the screen
    p1ScoreTextPainter.paint(
        canvas, Offset(size.width * 0.25 - p1ScoreTextPainter.width / 2, 20));
    p2ScoreTextPainter.paint(
        canvas, Offset(size.width * 0.75 - p2ScoreTextPainter.width / 2, 20));
  }

  @override
  bool shouldRepaint(PongPainter oldDelegate) {
    // Repaint whenever the game state changes
    return true;
  }
}
