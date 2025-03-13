import 'package:flutter/material.dart';
import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/config.dart';
import 'package:pongui/models/newconfig.dart';
import 'dart:io';
import 'package:path/path.dart' as path;

class NewConfigScreen extends StatefulWidget {
  final NewConfigModel newConfig;
  final Future<void> Function() onConfigSaved;

  const NewConfigScreen({
    Key? key,
    required this.newConfig,
    required this.onConfigSaved,
  }) : super(key: key);

  @override
  _NewConfigScreenState createState() => _NewConfigScreenState();
}

class _NewConfigScreenState extends State<NewConfigScreen> {
  final _formKey = GlobalKey<FormState>();

  // Controllers for all fields
  late TextEditingController _serverAddrController;
  late TextEditingController _grpcCertPathController;
  late TextEditingController _rpcCertPathController;
  late TextEditingController _rpcClientCertPathController;
  late TextEditingController _rpcClientKeyPathController;
  late TextEditingController _rpcWebsocketURLController;
  late TextEditingController _debugLevelController;
  late TextEditingController _userController;
  late TextEditingController _passController;

  // Checkbox state for wantsLogNtfns
  late bool _wantsLogNtfns;

  // Placeholder certificate content
  static const String placeholderCertContent = '''-----BEGIN CERTIFICATE-----
MIIBkzCCATmgAwIBAgIRAOCyLu1U/ZKyD33nXFPgJOQwCgYIKoZIzj0EAwIwFjEU
MBIGA1UEChMLUG9uZyBTZXJ2ZXIwHhcNMjUwMTMxMTg0NzQwWhcNMjYwMTMxMTg0
NzQwWjAWMRQwEgYDVQQKEwtQb25nIFNlcnZlcjBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABLpaje+KDrdAe77RwOaxYAkxRmlDg1cbLspf1riFhskIUyfILM1r8zPd
Ql10MGxeKipbE3LCPOn5BV0KVGxfb2mjaDBmMA4GA1UdDwEB/wQEAwICpDATBgNV
HSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQLw3WW
CxxXpNfuuDgGuZ3c8IX0rDAPBgNVHREECDAGhwRog7QdMAoGCCqGSM49BAMCA0gA
MEUCIEWR7Iw7ua6WAQuIf8Yf0lNzP6s2NczAR0W4uP8zuVK0AiEA6ruxkMcv4CHw
Aq6RDElOTqAlDbNAuV8b/joQjIDLwqA=
-----END CERTIFICATE-----''';

  @override
  void initState() {
    super.initState();

    // Initialize controllers with existing values
    _serverAddrController =
        TextEditingController(text: widget.newConfig.serverAddr);
    _grpcCertPathController =
        TextEditingController(text: widget.newConfig.grpcCertPath);
    _rpcCertPathController =
        TextEditingController(text: widget.newConfig.rpcCertPath);
    _rpcClientCertPathController =
        TextEditingController(text: widget.newConfig.rpcClientCertPath);
    _rpcClientKeyPathController =
        TextEditingController(text: widget.newConfig.rpcClientKeyPath);
    _rpcWebsocketURLController =
        TextEditingController(text: widget.newConfig.rpcWebsocketURL);
    _debugLevelController =
        TextEditingController(text: widget.newConfig.debugLevel);
    _userController = TextEditingController(text: widget.newConfig.rpcUser);
    _passController = TextEditingController(text: widget.newConfig.rpcPass);

    _wantsLogNtfns = widget.newConfig.wantsLogNtfns;
  }

  // Helper function to ensure gRPC certificate exists
  Future<String> _ensureGrpcCertExists() async {
    // Get the target directory and certificate path
    final appDataDir = await defaultAppDataDir();
    final certPath = path.join(appDataDir, 'server.cert');

    // Ensure the directory exists
    final directory = Directory(path.dirname(certPath));
    if (!directory.existsSync()) {
      await directory.create(recursive: true);
    }

    // Create the certificate file if it doesn't exist
    final certFile = File(certPath);
    if (!certFile.existsSync()) {
      await certFile.writeAsString(placeholderCertContent);
    }

    // Return the path to the certificate file
    return certPath;
  }

  Future<void> _saveConfig() async {
    if (_formKey.currentState!.validate()) {
      try {
        // Ensure the gRPC certificate exists and get its path
        final certPath = await _ensureGrpcCertExists();

        // If the user hasn't specified a custom path, use the default
        if (_grpcCertPathController.text.isEmpty) {
          _grpcCertPathController.text = certPath;
        }

        // Update the config model
        widget.newConfig.serverAddr = _serverAddrController.text;
        widget.newConfig.grpcCertPath = _grpcCertPathController.text;
        widget.newConfig.rpcCertPath = _rpcCertPathController.text;
        widget.newConfig.rpcClientCertPath = _rpcClientCertPathController.text;
        widget.newConfig.rpcClientKeyPath = _rpcClientKeyPathController.text;
        widget.newConfig.rpcWebsocketURL = _rpcWebsocketURLController.text;
        widget.newConfig.debugLevel = _debugLevelController.text;
        widget.newConfig.rpcUser = _userController.text;
        widget.newConfig.rpcPass = _passController.text;
        widget.newConfig.wantsLogNtfns = _wantsLogNtfns;

        final configFilePath = await widget.newConfig.getConfigFilePath();
        await widget.newConfig.saveConfig(configFilePath);

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Config saved!')),
        );

        // Notify the caller that the config has been saved, so they can reload.
        await widget.onConfigSaved();
      } catch (e) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error saving config: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return SharedLayout(
      title: "Settings Screen",
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              children: [
                _buildTextField(
                  controller: _serverAddrController,
                  label: 'Server Address',
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter the server address';
                    }
                    return null;
                  },
                ),
                _buildTextField(
                  controller: _grpcCertPathController,
                  label: 'gRPC server Cert Path',
                  validator: (value) => null,
                ),
                _buildTextField(
                  controller: _rpcCertPathController,
                  label: 'RPC Cert Path',
                  validator: (value) => null,
                ),
                _buildTextField(
                  controller: _rpcClientCertPathController,
                  label: 'RPC Client Cert Path',
                  validator: (value) => null,
                ),
                _buildTextField(
                  controller: _rpcClientKeyPathController,
                  label: 'RPC Client Key Path',
                  validator: (value) => null,
                ),
                _buildTextField(
                  controller: _rpcWebsocketURLController,
                  label: 'RPC WebSocket URL',
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter the WebSocket URL';
                    }
                    return null;
                  },
                ),
                _buildTextField(
                  controller: _debugLevelController,
                  label: 'Debug Level',
                  validator: (value) => null,
                ),
                _buildTextField(
                  controller: _userController,
                  label: 'RPC User',
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter the RPC User';
                    }
                    return null;
                  },
                ),
                _buildTextField(
                  controller: _passController,
                  label: 'RPC Password',
                  obscureText: true,
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter the RPC Password';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 10),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Log Notifications'),
                    Switch(
                      value: _wantsLogNtfns,
                      onChanged: (value) {
                        setState(() {
                          _wantsLogNtfns = value;
                        });
                      },
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                ElevatedButton(
                  onPressed: _saveConfig,
                  child: const Text('Save Config'),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    bool obscureText = false,
    String? Function(String?)? validator,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: TextFormField(
        controller: controller,
        decoration: InputDecoration(labelText: label),
        obscureText: obscureText,
        validator: validator,
      ),
    );
  }
}
