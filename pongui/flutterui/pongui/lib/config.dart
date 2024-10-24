import 'dart:io';
import 'package:args/args.dart';
import 'package:ini/ini.dart' as ini;
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as path;

const APPNAME = "pongui";
String mainConfigFilename = "";

class Config {
  late final String serverAddr;
  late final String rpcCertPath;
  late final String rpcClientCertPath;
  late final String rpcClientKeyPath;
  late final String rpcWebsocketURL;
  late final String debugLevel;
  late final String rpcUser;
  late final String rpcPass;
  late final bool wantsLogNtfns;  // Field for log notifications

  Config();
  
  Config.filled({
    this.serverAddr = "",
    this.rpcCertPath = "",
    this.rpcClientCertPath = "",
    this.rpcClientKeyPath = "",
    this.debugLevel = "info",
    this.rpcWebsocketURL = "",
    this.rpcUser = "",
    this.rpcPass = "",
    this.wantsLogNtfns = false,
  });

  // Save a new config from scratch
  Future<void> saveNewConfig(String filepath) async {
    var f = ini.Config.fromString("\n[clientrpc]\n");
    set(String section, String opt, String val) =>
        val != "" ? f.set(section, opt, val) : null;

    set("default", "server", serverAddr);
    set("clientrpc", "rpccertpath", rpcCertPath);
    set("log", "debuglevel", debugLevel);
    set("clientrpc", "rpcwebsocketurl", rpcWebsocketURL);
    set("clientrpc", "rpcclientcertpath", rpcClientCertPath);
    set("clientrpc", "rpcclientkeypath", rpcClientKeyPath);
    set("clientrpc", "rpccertpath", rpcCertPath);
    set("clientrpc", "rpcuser", rpcUser);
    set("clientrpc", "rpcpass", rpcPass);
    set("clientrpc", "wantsLogNtfns", wantsLogNtfns ? "1" : "0");

    // Write the config file
    await File(filepath).parent.create(recursive: true);
    await File(filepath).writeAsString(f.toString());
  }

  // Load existing config
  static Future<Config> loadConfig(String filepath) async {
    var f = ini.Config.fromStrings(File(filepath).readAsLinesSync());

    return Config.filled(
      serverAddr: f.get("default", "server") ?? "localhost:443",
      rpcCertPath: f.get("clientrpc", "rpccertpath") ?? "",
      rpcClientCertPath: f.get("clientrpc", "rpcclientcertpath") ?? "",
      rpcClientKeyPath: f.get("clientrpc", "rpcclientkeypath") ?? "",
      debugLevel: f.get("log", "debuglevel") ?? "info",
      rpcUser: f.get("clientrpc", "rpcuser") ?? "",
      rpcPass: f.get("clientrpc", "rpcpass") ?? "",
      rpcWebsocketURL: f.get("clientrpc", "rpcwebsocketurl") ?? "",
      wantsLogNtfns: f.get("clientrpc", "wantsLogNtfns") == "1",
    );
  }
}

// Function to get the default app data directory based on the platform
Future<String> defaultAppDataDir() async {
  if (Platform.isLinux) {
    final home = Platform.environment["HOME"];
    if (home != null && home != "") {
      return path.join(home, ".$APPNAME");
    }
  } else if (Platform.isWindows && Platform.environment.containsKey("LOCALAPPDATA")) {
    return path.join(Platform.environment["LOCALAPPDATA"]!, APPNAME);
  } else if (Platform.isMacOS) {
    final baseDir = (await getApplicationSupportDirectory()).parent.path;
    return path.join(baseDir, APPNAME);
  }

  final dir = await getApplicationSupportDirectory();
  return dir.path;
}

final usageException = Exception("Usage Displayed");
final newConfigNeededException = Exception("Config needed");


Future<Config> loadConfig(String filepath) async {
  var f = ini.Config.fromStrings(File(filepath).readAsLinesSync());
  var appDataDir = await defaultAppDataDir();
  var iniAppData = f.get("default", "root");
  
  // If the app data directory is defined in the config, use it
  if (iniAppData != null && iniAppData != "") {
    appDataDir = cleanAndExpandPath(iniAppData);
  }

  String getPath(String section, String option, String def) {
    var iniVal = f.get(section, option);
    if (iniVal == null || iniVal == "") {
      return def;
    }
    return cleanAndExpandPath(iniVal);
  }

  bool getBool(String section, String opt) {
    var v = f.get(section, opt);
    return v == "yes" || v == "true" || v == "1";
  }

  // Creating and populating the Config instance with relevant fields
  var c = Config.filled(
    serverAddr: f.get("default", "server") ?? "localhost:50051",
    debugLevel: f.get("log", "debuglevel") ?? "info",
    rpcWebsocketURL: f.get("clientrpc", "rpcwebsocketurl") ?? "",
    rpcCertPath: getPath("clientrpc", "rpccertpath", ""),
    rpcClientCertPath: getPath("clientrpc", "rpcclientcertpath", ""),
    rpcClientKeyPath: getPath("clientrpc", "rpcclientkeypath", ""),
    rpcUser: f.get("clientrpc", "rpcuser") ?? "",
    rpcPass: f.get("clientrpc", "rpcpass") ?? "",
    wantsLogNtfns: getBool("clientrpc", "wantsLogNtfns")
  );

  return c;
}

String homeDir() {
  var env = Platform.environment;
  if (Platform.isWindows) {
    return env['UserProfile'] ?? "";
  } else {
    return env['HOME'] ?? "";
  }
}

String cleanAndExpandPath(String p) {
  if (p == "") {
    return p;
  }

  if (p.startsWith("~")) {
    p = homeDir() + p.substring(1);
  }

  return path.canonicalize(p);
}

Future<ArgParser> appArgParser() async {
  var defaultCfgFile = path.join(await defaultAppDataDir(), "$APPNAME.conf");
  var p = ArgParser();
  p.addFlag("help", abbr: "h", help: "Display usage info", negatable: false);
  p.addOption("configfile",
      abbr: "c", defaultsTo: defaultCfgFile, help: "Path to config file");
  return p;
}

Future<String> configFileName(List<String> args) async {
  var p = await appArgParser();
  var res = p.parse(args);
  return res["configfile"];
}

// Function to load config using arguments
Future<Config> configFromArgs(List<String> args) async {
  var p = await appArgParser();
  var res = p.parse(args);

  if (res["help"]) {
    // ignore: avoid_print
    print(p.usage);
    throw usageException;
  }

  var cfgFilePath = res["configfile"];
  if (!File(cfgFilePath).existsSync()) {
    throw newConfigNeededException;
  }

  return loadConfig(cfgFilePath);
}
