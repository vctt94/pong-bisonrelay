import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:ini/ini.dart' as ini;
import 'package:path/path.dart' as path;
import 'package:pongui/config.dart';

class NewConfigModel extends ChangeNotifier {
  String rpcUser = 'defaultuser';
  String rpcPass = 'defaultpass';
  String serverAddr = '104.131.180.29:50051';
  String grpcCertPath = '';
  String rpcCertPath = '';
  String rpcClientCertPath = '';
  String rpcClientKeyPath = '';
  String rpcWebsocketURL = 'wss://127.0.0.1:7676/ws';
  String debugLevel = 'info';
  bool wantsLogNtfns = false;

  final List<String> appArgs;

  NewConfigModel(this.appArgs) {
    // Assign default file paths based on app data directory
    _initializeDefaults();
  }

  factory NewConfigModel.fromConfig(Config config) {
    return NewConfigModel([])
      ..rpcUser = config.rpcUser
      ..rpcPass = config.rpcPass
      ..serverAddr = config.serverAddr
      ..grpcCertPath = config.grpcCertPath
      ..rpcCertPath = config.rpcCertPath
      ..rpcClientCertPath = config.rpcClientCertPath
      ..rpcClientKeyPath = config.rpcClientKeyPath
      ..rpcWebsocketURL = config.rpcWebsocketURL
      ..debugLevel = config.debugLevel
      ..wantsLogNtfns = config.wantsLogNtfns;
  }

  Future<void> _initializeDefaults() async {
    var brDataDir = await defaultAppDataBRUIGDir();
    var appDataDir = await defaultAppDataDir();

    // Set default paths for certificates and keys
    rpcCertPath =
        rpcCertPath.isEmpty ? path.join(brDataDir, 'rpc.cert') : rpcCertPath;
    rpcClientCertPath = rpcClientCertPath.isEmpty
        ? path.join(brDataDir, 'rpc-client.cert')
        : rpcClientCertPath;
    rpcClientKeyPath = rpcClientKeyPath.isEmpty
        ? path.join(brDataDir, 'rpc-client.key')
        : rpcClientKeyPath;
    grpcCertPath = grpcCertPath.isEmpty
        ? path.join(appDataDir, 'server.cert')
        : grpcCertPath;

    // Notify listeners of any changes
    notifyListeners();
  }

  // Get the application data directory path
  Future<String> appDataDir() async =>
      path.dirname(await configFileName(appArgs));

// Save current configuration values to a file
  Future<void> saveConfig(String configFilePath) async {
    ini.Config config;

    // Check if the config file exists; otherwise, create a new one
    final configFile = File(configFilePath);
    if (configFile.existsSync()) {
      config = ini.Config.fromStrings(await configFile.readAsLines());
    } else {
      // Create a new config instance
      config = ini.Config();
      // Ensure required sections exist in the new config
    }
    print('Existing sections: ${config.sections().join(', ')}');
    // Ensure the 'clientrpc' section exists

    if (!config.hasSection('clientrpc')) {
      config.addSection('clientrpc');
    }
    if (!config.hasSection('log')) {
      config.addSection('log');
    }

    // Set the configuration values
    config.set('default', 'server', serverAddr);
    config.set('default', 'grpccertpath', grpcCertPath);

    config.set('clientrpc', 'rpcuser', rpcUser);
    config.set('clientrpc', 'rpcpass', rpcPass);
    config.set('clientrpc', 'rpcwebsocketurl', rpcWebsocketURL);
    config.set('clientrpc', 'rpccertpath', rpcCertPath);
    config.set('clientrpc', 'rpcclientcertpath', rpcClientCertPath);
    config.set('clientrpc', 'rpcclientkeypath', rpcClientKeyPath);
    config.set('clientrpc', 'wantsLogNtfns', wantsLogNtfns ? '1' : '0');

    config.set('log', 'debuglevel', debugLevel);

    // Ensure the config directory exists
    final configFileDir = path.dirname(configFilePath);
    final directory = Directory(configFileDir);
    if (!directory.existsSync()) {
      await directory.create(recursive: true);
    }

    // Write the updated config to the file
    await configFile.writeAsString(config.toString());
  }

  // Get the path to the config file
  Future<String> getConfigFilePath() async {
    var dataDir = await appDataDir();
    return path.join(dataDir, 'pongui.conf');
  }
}
