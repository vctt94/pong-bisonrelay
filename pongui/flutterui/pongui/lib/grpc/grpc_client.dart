import 'dart:io';
import 'package:golib_plugin/grpc/generated/pong.pbgrpc.dart';
import 'package:grpc/grpc.dart';

class GrpcPongClient {
  late ClientChannel _channel;
  late PongGameClient _client;

  GrpcPongClient(String serverAddress, int port, {String? tlsCertPath}) {
    // Set up credentials based on whether TLS is being used
    final credentials = (tlsCertPath != null && tlsCertPath.isNotEmpty)
        ? _createSecureCredentials(tlsCertPath)
        : const ChannelCredentials.insecure();

    // Initialize the gRPC channel and client stub
    _channel = ClientChannel(
      serverAddress,
      port: port,
      options: ChannelOptions(
        credentials: credentials,
      ),
    );
    _client = PongGameClient(_channel);
  }

  // Helper method to create secure credentials
  ChannelCredentials _createSecureCredentials(String certPath) {
    try {
      final cert = File(certPath).readAsBytesSync();
      return ChannelCredentials.secure(
        certificates: cert,
        authority: null, // Add authority if required
      );
    } catch (e) {
      throw Exception('Failed to read TLS certificate: $e');
    }
  }

  // Call Init on the PluginService and listen to the stream
  Stream<NtfnStreamResponse> startNtfnStreamRequest(String clientId) async* {
    final request = StartNtfnStreamRequest()..clientId = clientId;

    try {
      final responseStream = _client.startNtfnStream(request);
      await for (var response in responseStream) {
        yield response; // Yield each response back to the caller
      }
    } catch (e) {
      print('Error during Init: $e');
      rethrow;
    }
  }

  // Call Action on the PluginService
  Stream<GameUpdateBytes> startGameStreamRequest(String clientId) async* {
    final request = StartGameStreamRequest()..clientId = clientId;

    try {
      final responseStream = _client.startGameStream(request);
      await for (var response in responseStream) {
        yield response;
      }
    } catch (e) {
      print('Error during CallAction: $e');
      rethrow;
    }
  }

  // In GrpcPongClient
  Future<void> sendInput(String clientId, String inputData) async {
    // Implement the gRPC call to send input without expecting a stream response
    final request = PlayerInput()
      ..input = inputData
      ..playerId = clientId;

    await _client.sendInput(request);
  }

  // GetVersion method (unary call)
  // Future<PluginVersionResponse> getVersion() async {
  //   final request = PluginVersionRequest();

  //   try {
  //     final response = await _stub.getVersion(request);
  //     print(response);
  //     return response; // Return the response
  //   } catch (e) {
  //     print('Error during GetVersion: $e');
  //     rethrow;
  //   }
  // }

  // Optionally, clean up the gRPC connection
  Future<void> shutdown() async {
    await _channel.shutdown();
  }
}
