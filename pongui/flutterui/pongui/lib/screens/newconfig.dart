import 'package:flutter/material.dart';
import 'package:pongui/models/newconfig.dart';

class NewConfigScreen extends StatefulWidget {
  final NewConfigModel newConfig;

  const NewConfigScreen(this.newConfig, {Key? key}) : super(key: key);

  @override
  _NewConfigScreenState createState() => _NewConfigScreenState();
}

class _NewConfigScreenState extends State<NewConfigScreen> {
  final _formKey = GlobalKey<FormState>();

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
          child: Column(
            children: [
              TextFormField(
                decoration: const InputDecoration(labelText: 'RPC User'),
                onChanged: (value) {
                  widget.newConfig.rpcUser = value;
                },
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter RPC User';
                  }
                  return null;
                },
              ),
              TextFormField(
                decoration: const InputDecoration(labelText: 'RPC Password'),
                obscureText: true,
                onChanged: (value) {
                  widget.newConfig.rpcPass = value;
                },
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Please enter RPC Password';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 20),
              ElevatedButton(
                onPressed: () async {
                  if (_formKey.currentState!.validate()) {
                    final configFilePath =
                        await widget.newConfig.getConfigFilePath();
                    await widget.newConfig.saveConfig(configFilePath);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Config saved!')),
                    );
                  }
                },
                child: const Text('Save Config'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

Future<void> runNewConfigApp(List<String> args) async {
  final newConfig = NewConfigModel(args);

  runApp(
    MaterialApp(
      title: 'New RPC Configuration',
      home: NewConfigScreen(newConfig),
    ),
  );
}
