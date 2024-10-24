import 'dart:async';
import 'dart:io';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/config.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:pongui/screens/newconfig.dart';

void main(List<String> args) async {
  WidgetsFlutterBinding.ensureInitialized();

  // Pass args and initialize config
  runApp(MyApp(args));
}

class MyApp extends StatefulWidget {
  final List<String> args;

  MyApp(this.args);

  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  Config? config;
  String serverAddr = '';
  bool isLoading = true;
  String errorMessage = '';

  @override
  void initState() {
    super.initState();
    _initializeApp(widget.args);
  }

  void _startListeningToStreams(GrpcPongClient grpcClient, String clientId) {
    // Start notification stream
    grpcClient.startNtfnStreamRequest(clientId).listen((response) {
      developer.log("Notification Stream Response: ${response}");
      // Handle the response (e.g., update UI or handle game state)
    }, onError: (error) {
      developer.log("Error in notification stream: $error");
    });
  }

  Future<void> _initializeApp(List<String> args) async {
    try {
      // Load the configuration from the args.
      final filename = await configFileName(args);
      final cfg = await configFromArgs(args);

      InitClient initArgs = InitClient(
        cfg.serverAddr,
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
      String clientId = await Golib.initClient(initArgs);
      List<String> parts = cfg.serverAddr.split(":");
      String ipAddress = parts[0]; // "127.0.0.1"
      int port = int.parse(parts[1]); // 50051 as an integer
      final grpcClient = GrpcPongClient(
        ipAddress,
        port, // Assuming you have the port in the config
        // tlsCertPath: cfg.rpcCertPath, // Pass the certificate path if using TLS
      );

      setState(() {
        config = cfg;
        serverAddr = cfg.serverAddr;
        isLoading = false;
      });
      _startListeningToStreams(grpcClient, clientId);
    } catch (exception) {
      developer.log("Error: $exception");
      if (exception == usageException) {
        exit(0);
      } else if (exception == newConfigNeededException) {
        runNewConfigApp(widget.args);
      } else {
        setState(() {
          errorMessage = 'Error: $exception';
          isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Configured App',
      home: Scaffold(
        appBar: AppBar(
          title: Text('Configured App'),
        ),
        body: isLoading
            ? Center(child: CircularProgressIndicator())
            : errorMessage.isNotEmpty
                ? Center(child: Text(errorMessage))
                : Center(
                    child: Text('Server Address: $serverAddr'),
                  ),
      ),
    );
  }
}
