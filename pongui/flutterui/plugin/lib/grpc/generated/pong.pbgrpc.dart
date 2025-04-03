//
//  Generated code. Do not modify.
//  source: pong.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'pong.pb.dart' as $0;

export 'pong.pb.dart';

@$pb.GrpcServiceName('pong.PongGame')
class PongGameClient extends $grpc.Client {
  static final _$sendInput = $grpc.ClientMethod<$0.PlayerInput, $0.GameUpdate>(
      '/pong.PongGame/SendInput',
      ($0.PlayerInput value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.GameUpdate.fromBuffer(value));
  static final _$startGameStream = $grpc.ClientMethod<$0.StartGameStreamRequest, $0.GameUpdateBytes>(
      '/pong.PongGame/StartGameStream',
      ($0.StartGameStreamRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.GameUpdateBytes.fromBuffer(value));
  static final _$startNtfnStream = $grpc.ClientMethod<$0.StartNtfnStreamRequest, $0.NtfnStreamResponse>(
      '/pong.PongGame/StartNtfnStream',
      ($0.StartNtfnStreamRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.NtfnStreamResponse.fromBuffer(value));
  static final _$unreadyGameStream = $grpc.ClientMethod<$0.UnreadyGameStreamRequest, $0.UnreadyGameStreamResponse>(
      '/pong.PongGame/UnreadyGameStream',
      ($0.UnreadyGameStreamRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.UnreadyGameStreamResponse.fromBuffer(value));
  static final _$signalReadyToPlay = $grpc.ClientMethod<$0.SignalReadyToPlayRequest, $0.SignalReadyToPlayResponse>(
      '/pong.PongGame/SignalReadyToPlay',
      ($0.SignalReadyToPlayRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.SignalReadyToPlayResponse.fromBuffer(value));
  static final _$getWaitingRoom = $grpc.ClientMethod<$0.WaitingRoomRequest, $0.WaitingRoomResponse>(
      '/pong.PongGame/GetWaitingRoom',
      ($0.WaitingRoomRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.WaitingRoomResponse.fromBuffer(value));
  static final _$getWaitingRooms = $grpc.ClientMethod<$0.WaitingRoomsRequest, $0.WaitingRoomsResponse>(
      '/pong.PongGame/GetWaitingRooms',
      ($0.WaitingRoomsRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.WaitingRoomsResponse.fromBuffer(value));
  static final _$createWaitingRoom = $grpc.ClientMethod<$0.CreateWaitingRoomRequest, $0.CreateWaitingRoomResponse>(
      '/pong.PongGame/CreateWaitingRoom',
      ($0.CreateWaitingRoomRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.CreateWaitingRoomResponse.fromBuffer(value));
  static final _$joinWaitingRoom = $grpc.ClientMethod<$0.JoinWaitingRoomRequest, $0.JoinWaitingRoomResponse>(
      '/pong.PongGame/JoinWaitingRoom',
      ($0.JoinWaitingRoomRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.JoinWaitingRoomResponse.fromBuffer(value));
  static final _$leaveWaitingRoom = $grpc.ClientMethod<$0.LeaveWaitingRoomRequest, $0.LeaveWaitingRoomResponse>(
      '/pong.PongGame/LeaveWaitingRoom',
      ($0.LeaveWaitingRoomRequest value) => value.writeToBuffer(),
      ($core.List<$core.int> value) => $0.LeaveWaitingRoomResponse.fromBuffer(value));

  PongGameClient($grpc.ClientChannel channel,
      {$grpc.CallOptions? options,
      $core.Iterable<$grpc.ClientInterceptor>? interceptors})
      : super(channel, options: options,
        interceptors: interceptors);

  $grpc.ResponseFuture<$0.GameUpdate> sendInput($0.PlayerInput request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$sendInput, request, options: options);
  }

  $grpc.ResponseStream<$0.GameUpdateBytes> startGameStream($0.StartGameStreamRequest request, {$grpc.CallOptions? options}) {
    return $createStreamingCall(_$startGameStream, $async.Stream.fromIterable([request]), options: options);
  }

  $grpc.ResponseStream<$0.NtfnStreamResponse> startNtfnStream($0.StartNtfnStreamRequest request, {$grpc.CallOptions? options}) {
    return $createStreamingCall(_$startNtfnStream, $async.Stream.fromIterable([request]), options: options);
  }

  $grpc.ResponseFuture<$0.UnreadyGameStreamResponse> unreadyGameStream($0.UnreadyGameStreamRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$unreadyGameStream, request, options: options);
  }

  $grpc.ResponseFuture<$0.SignalReadyToPlayResponse> signalReadyToPlay($0.SignalReadyToPlayRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$signalReadyToPlay, request, options: options);
  }

  $grpc.ResponseFuture<$0.WaitingRoomResponse> getWaitingRoom($0.WaitingRoomRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWaitingRoom, request, options: options);
  }

  $grpc.ResponseFuture<$0.WaitingRoomsResponse> getWaitingRooms($0.WaitingRoomsRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$getWaitingRooms, request, options: options);
  }

  $grpc.ResponseFuture<$0.CreateWaitingRoomResponse> createWaitingRoom($0.CreateWaitingRoomRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$createWaitingRoom, request, options: options);
  }

  $grpc.ResponseFuture<$0.JoinWaitingRoomResponse> joinWaitingRoom($0.JoinWaitingRoomRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$joinWaitingRoom, request, options: options);
  }

  $grpc.ResponseFuture<$0.LeaveWaitingRoomResponse> leaveWaitingRoom($0.LeaveWaitingRoomRequest request, {$grpc.CallOptions? options}) {
    return $createUnaryCall(_$leaveWaitingRoom, request, options: options);
  }
}

@$pb.GrpcServiceName('pong.PongGame')
abstract class PongGameServiceBase extends $grpc.Service {
  $core.String get $name => 'pong.PongGame';

  PongGameServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.PlayerInput, $0.GameUpdate>(
        'SendInput',
        sendInput_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PlayerInput.fromBuffer(value),
        ($0.GameUpdate value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.StartGameStreamRequest, $0.GameUpdateBytes>(
        'StartGameStream',
        startGameStream_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.StartGameStreamRequest.fromBuffer(value),
        ($0.GameUpdateBytes value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.StartNtfnStreamRequest, $0.NtfnStreamResponse>(
        'StartNtfnStream',
        startNtfnStream_Pre,
        false,
        true,
        ($core.List<$core.int> value) => $0.StartNtfnStreamRequest.fromBuffer(value),
        ($0.NtfnStreamResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UnreadyGameStreamRequest, $0.UnreadyGameStreamResponse>(
        'UnreadyGameStream',
        unreadyGameStream_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.UnreadyGameStreamRequest.fromBuffer(value),
        ($0.UnreadyGameStreamResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SignalReadyToPlayRequest, $0.SignalReadyToPlayResponse>(
        'SignalReadyToPlay',
        signalReadyToPlay_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.SignalReadyToPlayRequest.fromBuffer(value),
        ($0.SignalReadyToPlayResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.WaitingRoomRequest, $0.WaitingRoomResponse>(
        'GetWaitingRoom',
        getWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.WaitingRoomRequest.fromBuffer(value),
        ($0.WaitingRoomResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.WaitingRoomsRequest, $0.WaitingRoomsResponse>(
        'GetWaitingRooms',
        getWaitingRooms_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.WaitingRoomsRequest.fromBuffer(value),
        ($0.WaitingRoomsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateWaitingRoomRequest, $0.CreateWaitingRoomResponse>(
        'CreateWaitingRoom',
        createWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.CreateWaitingRoomRequest.fromBuffer(value),
        ($0.CreateWaitingRoomResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.JoinWaitingRoomRequest, $0.JoinWaitingRoomResponse>(
        'JoinWaitingRoom',
        joinWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.JoinWaitingRoomRequest.fromBuffer(value),
        ($0.JoinWaitingRoomResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LeaveWaitingRoomRequest, $0.LeaveWaitingRoomResponse>(
        'LeaveWaitingRoom',
        leaveWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.LeaveWaitingRoomRequest.fromBuffer(value),
        ($0.LeaveWaitingRoomResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.GameUpdate> sendInput_Pre($grpc.ServiceCall call, $async.Future<$0.PlayerInput> request) async {
    return sendInput(call, await request);
  }

  $async.Stream<$0.GameUpdateBytes> startGameStream_Pre($grpc.ServiceCall call, $async.Future<$0.StartGameStreamRequest> request) async* {
    yield* startGameStream(call, await request);
  }

  $async.Stream<$0.NtfnStreamResponse> startNtfnStream_Pre($grpc.ServiceCall call, $async.Future<$0.StartNtfnStreamRequest> request) async* {
    yield* startNtfnStream(call, await request);
  }

  $async.Future<$0.UnreadyGameStreamResponse> unreadyGameStream_Pre($grpc.ServiceCall call, $async.Future<$0.UnreadyGameStreamRequest> request) async {
    return unreadyGameStream(call, await request);
  }

  $async.Future<$0.SignalReadyToPlayResponse> signalReadyToPlay_Pre($grpc.ServiceCall call, $async.Future<$0.SignalReadyToPlayRequest> request) async {
    return signalReadyToPlay(call, await request);
  }

  $async.Future<$0.WaitingRoomResponse> getWaitingRoom_Pre($grpc.ServiceCall call, $async.Future<$0.WaitingRoomRequest> request) async {
    return getWaitingRoom(call, await request);
  }

  $async.Future<$0.WaitingRoomsResponse> getWaitingRooms_Pre($grpc.ServiceCall call, $async.Future<$0.WaitingRoomsRequest> request) async {
    return getWaitingRooms(call, await request);
  }

  $async.Future<$0.CreateWaitingRoomResponse> createWaitingRoom_Pre($grpc.ServiceCall call, $async.Future<$0.CreateWaitingRoomRequest> request) async {
    return createWaitingRoom(call, await request);
  }

  $async.Future<$0.JoinWaitingRoomResponse> joinWaitingRoom_Pre($grpc.ServiceCall call, $async.Future<$0.JoinWaitingRoomRequest> request) async {
    return joinWaitingRoom(call, await request);
  }

  $async.Future<$0.LeaveWaitingRoomResponse> leaveWaitingRoom_Pre($grpc.ServiceCall call, $async.Future<$0.LeaveWaitingRoomRequest> request) async {
    return leaveWaitingRoom(call, await request);
  }

  $async.Future<$0.GameUpdate> sendInput($grpc.ServiceCall call, $0.PlayerInput request);
  $async.Stream<$0.GameUpdateBytes> startGameStream($grpc.ServiceCall call, $0.StartGameStreamRequest request);
  $async.Stream<$0.NtfnStreamResponse> startNtfnStream($grpc.ServiceCall call, $0.StartNtfnStreamRequest request);
  $async.Future<$0.UnreadyGameStreamResponse> unreadyGameStream($grpc.ServiceCall call, $0.UnreadyGameStreamRequest request);
  $async.Future<$0.SignalReadyToPlayResponse> signalReadyToPlay($grpc.ServiceCall call, $0.SignalReadyToPlayRequest request);
  $async.Future<$0.WaitingRoomResponse> getWaitingRoom($grpc.ServiceCall call, $0.WaitingRoomRequest request);
  $async.Future<$0.WaitingRoomsResponse> getWaitingRooms($grpc.ServiceCall call, $0.WaitingRoomsRequest request);
  $async.Future<$0.CreateWaitingRoomResponse> createWaitingRoom($grpc.ServiceCall call, $0.CreateWaitingRoomRequest request);
  $async.Future<$0.JoinWaitingRoomResponse> joinWaitingRoom($grpc.ServiceCall call, $0.JoinWaitingRoomRequest request);
  $async.Future<$0.LeaveWaitingRoomResponse> leaveWaitingRoom($grpc.ServiceCall call, $0.LeaveWaitingRoomRequest request);
}
