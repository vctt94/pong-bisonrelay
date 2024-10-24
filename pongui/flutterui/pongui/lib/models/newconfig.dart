import 'dart:io';
import 'package:flutter/cupertino.dart';
import 'package:ini/ini.dart' as ini;
import 'package:path/path.dart' as path;
import 'package:pongui/config.dart';

class NewConfigModel extends ChangeNotifier {
  String rpcUser = '';
  String rpcPass = '';
  final List<String> appArgs;
  NewConfigModel(this.appArgs);

  // Get the application data directory path
  Future<String> appDataDir() async =>
      path.dirname(await configFileName(appArgs));

  // Generate a new config file with the provided rpcUser and rpcPass
  Future<Config> generateConfig() async {
    var dataDir = await appDataDir();

    // XXX Needs fixing
    // Create a new config object with relevant fields
    var cfg = Config.filled(
      serverAddr: 'localhost:50051', // Example, adjust if needed
      // rpcCertPath: path.join(dataDir, 'cert.pem'),  // Example
      // rpcClientKeyPath: path.join(dataDir, 'key.pem'),    // Example
      rpcUser: rpcUser,
      rpcPass: rpcPass,
      debugLevel: "info",
      // rpcAuthMode: "basic",
    );

    // Save the new config to file
    await cfg.saveNewConfig(await getConfigFilePath());
    return cfg;
  }

  // Save config to a file
  Future<void> saveConfig(String configFilePath) async {
    ini.Config config;

    // Check if the config file exists and load it, otherwise create a new one
    if (File(configFilePath).existsSync()) {
      config = ini.Config.fromStrings(File(configFilePath).readAsLinesSync());
    } else {
      config = ini.Config();
    }

    // Ensure the 'clientrpc' section exists
    if (!config.hasSection('clientrpc')) {
      config.addSection('clientrpc');
    }

    // Set rpcuser and rpcpass in the 'clientrpc' section
    config.set('clientrpc', 'rpcuser', rpcUser);
    config.set('clientrpc', 'rpcpass', rpcPass);

    // Ensure the directory for the config file exists
    final configFileDir = path.dirname(configFilePath);
    final directory = Directory(configFileDir);

    if (!directory.existsSync()) {
      await directory.create(recursive: true); // Create the directory if it doesn't exist
    }

    // Write the updated config to the file
    await File(configFilePath).writeAsString(config.toString());
  }

  // Get the path to the config file
  Future<String> getConfigFilePath() async {
    var dataDir = await appDataDir();
    return path.join(dataDir, 'pongui.conf');
  }
}
