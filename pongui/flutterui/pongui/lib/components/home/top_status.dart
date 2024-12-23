import 'package:flutter/material.dart';
import 'package:pongui/models/pong.dart';

class TopStatusCard extends StatelessWidget {
  final PongModel pongModel;

  const TopStatusCard({
    Key? key,
    required this.pongModel,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Compact Status Section
        Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              Card(
                elevation: 2,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(12.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            "Bet: ${pongModel.betAmt}",
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                          Text(
                            pongModel.isReady
                                ? (pongModel.gameStarted ? "In Game" : "Ready")
                                : "Not Ready",
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ],
                      ),
                      if (pongModel.currentWR != null) ...[
                        const SizedBox(height: 12),
                        Divider(color: Colors.grey.shade400),
                        const SizedBox(height: 12),
                        Text(
                          "Current Waiting Room:",
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        const SizedBox(height: 8),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              "Room ID: ${pongModel.currentWR?.id}",
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                            Text(
                              "Bet: ${pongModel.currentWR?.betAmt}",
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          "Players: ${pongModel.currentWR?.players?.length ?? 0} / 2",
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),

        // Error Message
        if (pongModel.errorMessage.isNotEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Card(
              color: Colors.red.shade100,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: Row(
                  children: [
                    const Icon(Icons.error, color: Colors.red),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        pongModel.errorMessage,
                        style: Theme.of(context)
                            .textTheme
                            .bodyMedium
                            ?.copyWith(color: Colors.red),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
      ],
    );
  }
}
