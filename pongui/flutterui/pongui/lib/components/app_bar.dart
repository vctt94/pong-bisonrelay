import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';

class PongGameAppBar extends StatelessWidget implements PreferredSizeWidget {
  final bool isReady;
  final LocalWaitingRoom currentWR;
  final VoidCallback createWaitingRoom;
  final VoidCallback toggleReady;
  final double betAmount;

  const PongGameAppBar({
    Key? key,
    required this.isReady,
    required this.currentWR,
    required this.createWaitingRoom,
    required this.toggleReady,
    required this.betAmount,
  }) : super(key: key);

  @override
  Size get preferredSize => const Size.fromHeight(80.0);

  @override
  Widget build(BuildContext context) {
    return AppBar(
      toolbarHeight: preferredSize.height,
      title: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Pong Game'),
          Text(
            'Status: ${isReady ? "Ready" : "Not Ready"}\n'
            '${currentWR.id.isNotEmpty ? "Joined Room: ${currentWR.id}" : "No Room Joined"}',
            style: const TextStyle(fontSize: 14, color: Colors.white70),
          ),
        ],
      ),
      actions: [
        if (betAmount > 0)
          Padding(
            padding: const EdgeInsets.only(right: 10.0),
            child: ElevatedButton(
              onPressed: createWaitingRoom,
              child: const Text('Create Waiting Room'),
            ),
          ),
        if (currentWR.id.isNotEmpty && !isReady)
          Padding(
            padding: const EdgeInsets.only(right: 10.0),
            child: ElevatedButton(
              onPressed: toggleReady,
              child: const Text('Ready'),
            ),
          ),
      ],
    );
  }
}
