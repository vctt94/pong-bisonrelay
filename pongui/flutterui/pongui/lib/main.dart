import 'dart:async';
import 'dart:io';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:path_provider/path_provider.dart';
import 'package:pongui/config.dart';
import 'package:pongui/screens/newconfig.dart';

void main(List<String> args) async {
  try {
    // Ensure the platform bindings are initialized.
    WidgetsFlutterBinding.ensureInitialized();

    // Load the configuration from the args.
    mainConfigFilename = await configFileName(args);
    Config cfg = await configFromArgs(args);

    // Example of using config values in initClient
    // InitClient initArgs = InitClient(
    //   "127.0.0.1:7878", 
    //   cfg.rpcCertPath,
    //   cfg.rpcKeyPath,
    //   cfg.debugLevel,
    //   cfg.wantsLogNtfns,
    //   ["127.0.0.1:7878"],
    //   "/home/vctt/.bruig/rpc-client.cert",
    //   cfg.rpcKeyPath,
    //   cfg.rpcissueclientcert,
    //   cfg.rpcClientCApath,
    //   cfg.rpcUser,
    //   cfg.rpcPass,
    //   cfg.rpcAuthMode,
    // );
          InitClient initArgs = InitClient(
             "127.0.0.1:7878",
   "",
   "",
   "debug",
   true,
  ["127.0.0.1:7878"],
   "/home/vctt/.bruig/rpc-client.cert",
   "/home/vctt/.bruig/rpc-client.key",
   true,
   "/home/vctt/.bruig/rpc-ca.cert",
   cfg.rpcUser,
   cfg.rpcPass,
   "basic");

developer.log("InitClient args: $initArgs");
    await Golib.initClient(initArgs);
    
    // You can now pass cfg to your runApp method to initialize your app.
    runMainApp(cfg);

  } catch (exception) {
    developer.log("Error: $exception");
    if (exception == usageException) {
      exit(0);
    }
    if (exception == newConfigNeededException) {
      runNewConfigApp(args);
      print("exception!!");
      print(newConfigNeededException);
      return;
    }
    // runFatalErrorApp(exception);
  }
}

Future<void> runMainApp(Config cfg) async {
  runApp(MyApp(cfg));
}

class MyApp extends StatelessWidget {
  final Config config;
  
  MyApp(this.config);

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'My Flutter App',
      home: Scaffold(
        appBar: AppBar(
          title: Text('Configured App'),
        ),
        body: Center(
          child: Text('Server Address: ${config.serverAddr}'),
        ),
      ),
    );
  }
}
