import 'dart:async';
import 'dart:developer' as developer;
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/models/newconfig.dart';
import 'package:provider/provider.dart';
import 'package:window_manager/window_manager.dart';

import 'package:pongui/config.dart';
import 'package:pongui/models/pong.dart';
import 'package:pongui/screens/home.dart';
import 'package:pongui/screens/newconfig.dart';

Future<void> runNewConfigApp(List<String> args) async {
  final newConfig = NewConfigModel(args);

  runApp(
    MaterialApp(
      title: 'New RPC Configuration',
      home: NewConfigScreen(
        newConfig: newConfig,
        onConfigSaved: () async {
          // Load the updated configuration
          Config cfg = await configFromArgs(args);

          // Navigate back to the main app
          runMainApp(cfg);
        },
      ),
    ),
  );
}

void main(List<String> args) async {
  try {
    WidgetsFlutterBinding.ensureInitialized();
    if (Platform.isLinux || Platform.isWindows || Platform.isMacOS) {
      await windowManager.ensureInitialized();
    }

    developer.log("Platform: ${Golib.majorPlatform}/${Golib.minorPlatform}");
    Golib.platformVersion
        .then((value) => developer.log("Platform Version: $value"));
    Config cfg = await configFromArgs(args);
    runMainApp(cfg);
  } catch (exception) {
    print(exception);
    developer.log("Error: $exception");
    if (exception == usageException) {
      exit(0);
    } else if (exception == newConfigNeededException) {
      runNewConfigApp(args);
      return;
    }
  }
}

Future<void> runMainApp(Config cfg) async {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (context) => PongModel(cfg)),
      ],
      child: MyApp(cfg),
    ),
  );
}

class MyApp extends StatelessWidget {
  final Config cfg;
  const MyApp(this.cfg, {super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Pong Game App',
      theme: ThemeData.dark().copyWith(
        scaffoldBackgroundColor: const Color.fromARGB(255, 25, 23, 44),
        primaryColor: Colors.blueAccent,
      ),
      routes: {
        '/': (context) => const HomeScreen(),
        '/settings': (context) => NewConfigScreen(
              newConfig: NewConfigModel.fromConfig(cfg),
              onConfigSaved: () => runMainApp(cfg),
            ),
      },
      initialRoute: '/',
    );
  }
}
