import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:pongui/models/pong.dart';

class SharedLayout extends StatelessWidget {
  final String title;
  final Widget child;

  const SharedLayout({
    super.key,
    required this.title,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    final pongModel = Provider.of<PongModel>(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        leading: Navigator.of(context).canPop()
            ? IconButton(
                icon: const Icon(Icons.arrow_back),
                onPressed: () {
                  Navigator.of(context).pop();
                },
              )
            : null,
      ),
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            const DrawerHeader(
              decoration: BoxDecoration(color: Colors.blueAccent),
              child: Text('Menu',
                  style: TextStyle(color: Colors.white, fontSize: 24)),
            ),
            ListTile(
              leading: const Icon(Icons.home),
              title: const Text('Home'),
              onTap: () {
                Navigator.of(context).pushReplacementNamed('/');
              },
            ),
            ListTile(
              leading: const Icon(Icons.settings),
              title: const Text('Settings'),
              onTap: () {
                Navigator.of(context).pushNamed('/settings');
              },
            ),
          ],
        ),
      ),
      body: Column(
        children: [
          Expanded(child: child),
          // Footer Section
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.grey.shade900,
              borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(
                      pongModel.isConnected ? Icons.cloud_done : Icons.cloud_off,
                      color: pongModel.isConnected ? Colors.green : Colors.red,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      pongModel.isConnected ? "Connected" : "Disconnected",
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: pongModel.isConnected ? Colors.green : Colors.red,
                          ),
                    ),
                  ],
                ),
                Text(
                  "Client ID: ${pongModel.clientId}",
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Colors.white70),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
