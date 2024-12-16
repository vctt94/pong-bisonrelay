import 'package:flutter/material.dart';
import 'package:pongui/config.dart';
import 'package:pongui/main.dart';
import 'package:pongui/models/newconfig.dart';

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

  @override
  void dispose() {
    // Dispose all controllers
    _serverAddrController.dispose();
    _grpcCertPathController.dispose();
    _rpcCertPathController.dispose();
    _rpcClientCertPathController.dispose();
    _rpcClientKeyPathController.dispose();
    _rpcWebsocketURLController.dispose();
    _debugLevelController.dispose();
    _userController.dispose();
    _passController.dispose();
    super.dispose();
  }

  Future<void> _saveConfig() async {
    if (_formKey.currentState!.validate()) {
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
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('New RPC Configuration'),
      ),
      body: Padding(
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
                  label: 'grpc server Cert Path',
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
