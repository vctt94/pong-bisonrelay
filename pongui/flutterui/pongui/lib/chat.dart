import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

class ChatClientApp extends StatefulWidget {
  @override
  _ChatClientAppState createState() => _ChatClientAppState();
}

class _ChatClientAppState extends State<ChatClientApp> {
  final String apiUrl = "http://127.0.0.1:8080"; // Your Go backend URL
  TextEditingController userController = TextEditingController();
  TextEditingController messageController = TextEditingController();
  List<String> messages = [];

  Future<void> sendMessage(String user, String message) async {
    var url = Uri.parse('$apiUrl/send');
    var response = await http.post(
      url,
      headers: {"Content-Type": "application/json"},
      body: jsonEncode({"user": user, "message": message}),
    );
    if (response.statusCode == 200) {
      setState(() {
        messages.add("-> $user: $message");
      });
    }
  }

  Future<void> receiveMessages() async {
    var url = Uri.parse('$apiUrl/receive');
    var response = await http.get(url);
    if (response.statusCode == 200) {
      List<dynamic> receivedMessages = jsonDecode(response.body);
      setState(() {
        messages.addAll(receivedMessages.map((msg) => "<- ${msg['user']}: ${msg['message']}").toList());
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Chat Client')),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              itemCount: messages.length,
              itemBuilder: (context, index) {
                return ListTile(title: Text(messages[index]));
              },
            ),
          ),
          _messageInputWidget(),
        ],
      ),
    );
  }

  Widget _messageInputWidget() {
    return Padding(
      padding: const EdgeInsets.all(8.0),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: userController,
              decoration: InputDecoration(labelText: 'User'),
            ),
          ),
          Expanded(
            child: TextField(
              controller: messageController,
              decoration: InputDecoration(labelText: 'Message'),
            ),
          ),
          IconButton(
            icon: Icon(Icons.send),
            onPressed: () {
              sendMessage(userController.text, messageController.text);
              userController.clear();
              messageController.clear();
            },
          ),
        ],
      ),
    );
  }
}

void main() {
  runApp(MaterialApp(home: ChatClientApp()));
}
