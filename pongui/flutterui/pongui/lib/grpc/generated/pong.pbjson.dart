//
//  Generated code. Do not modify.
//  source: pong.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use notificationTypeDescriptor instead')
const NotificationType$json = {
  '1': 'NotificationType',
  '2': [
    {'1': 'UNKNOWN', '2': 0},
    {'1': 'MESSAGE', '2': 1},
    {'1': 'GAME_START', '2': 2},
    {'1': 'GAME_END', '2': 3},
    {'1': 'OPPONENT_DISCONNECTED', '2': 4},
    {'1': 'BET_AMOUNT_UPDATE', '2': 5},
    {'1': 'PLAYER_JOINED_WR', '2': 6},
    {'1': 'ON_WR_CREATED', '2': 7},
  ],
};

/// Descriptor for `NotificationType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List notificationTypeDescriptor = $convert.base64Decode(
    'ChBOb3RpZmljYXRpb25UeXBlEgsKB1VOS05PV04QABILCgdNRVNTQUdFEAESDgoKR0FNRV9TVE'
    'FSVBACEgwKCEdBTUVfRU5EEAMSGQoVT1BQT05FTlRfRElTQ09OTkVDVEVEEAQSFQoRQkVUX0FN'
    'T1VOVF9VUERBVEUQBRIUChBQTEFZRVJfSk9JTkVEX1dSEAYSEQoNT05fV1JfQ1JFQVRFRBAH');

@$core.Deprecated('Use startNtfnStreamRequestDescriptor instead')
const StartNtfnStreamRequest$json = {
  '1': 'StartNtfnStreamRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `StartNtfnStreamRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startNtfnStreamRequestDescriptor = $convert.base64Decode(
    'ChZTdGFydE50Zm5TdHJlYW1SZXF1ZXN0EhsKCWNsaWVudF9pZBgBIAEoCVIIY2xpZW50SWQ=');

@$core.Deprecated('Use ntfnStreamResponseDescriptor instead')
const NtfnStreamResponse$json = {
  '1': 'NtfnStreamResponse',
  '2': [
    {'1': 'notification_type', '3': 1, '4': 1, '5': 14, '6': '.pong.NotificationType', '10': 'notificationType'},
    {'1': 'started', '3': 2, '4': 1, '5': 8, '10': 'started'},
    {'1': 'game_id', '3': 3, '4': 1, '5': 9, '10': 'gameId'},
    {'1': 'message', '3': 4, '4': 1, '5': 9, '10': 'message'},
    {'1': 'betAmt', '3': 5, '4': 1, '5': 1, '10': 'betAmt'},
    {'1': 'player_number', '3': 6, '4': 1, '5': 5, '10': 'playerNumber'},
    {'1': 'player_id', '3': 7, '4': 1, '5': 9, '10': 'playerId'},
    {'1': 'room_id', '3': 8, '4': 1, '5': 9, '10': 'roomId'},
    {'1': 'wr', '3': 9, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `NtfnStreamResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ntfnStreamResponseDescriptor = $convert.base64Decode(
    'ChJOdGZuU3RyZWFtUmVzcG9uc2USQwoRbm90aWZpY2F0aW9uX3R5cGUYASABKA4yFi5wb25nLk'
    '5vdGlmaWNhdGlvblR5cGVSEG5vdGlmaWNhdGlvblR5cGUSGAoHc3RhcnRlZBgCIAEoCFIHc3Rh'
    'cnRlZBIXCgdnYW1lX2lkGAMgASgJUgZnYW1lSWQSGAoHbWVzc2FnZRgEIAEoCVIHbWVzc2FnZR'
    'IWCgZiZXRBbXQYBSABKAFSBmJldEFtdBIjCg1wbGF5ZXJfbnVtYmVyGAYgASgFUgxwbGF5ZXJO'
    'dW1iZXISGwoJcGxheWVyX2lkGAcgASgJUghwbGF5ZXJJZBIXCgdyb29tX2lkGAggASgJUgZyb2'
    '9tSWQSIQoCd3IYCSABKAsyES5wb25nLldhaXRpbmdSb29tUgJ3cg==');

@$core.Deprecated('Use waitingRoomsRequestDescriptor instead')
const WaitingRoomsRequest$json = {
  '1': 'WaitingRoomsRequest',
  '2': [
    {'1': 'room_id', '3': 1, '4': 1, '5': 9, '10': 'roomId'},
  ],
};

/// Descriptor for `WaitingRoomsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomsRequestDescriptor = $convert.base64Decode(
    'ChNXYWl0aW5nUm9vbXNSZXF1ZXN0EhcKB3Jvb21faWQYASABKAlSBnJvb21JZA==');

@$core.Deprecated('Use waitingRoomsResponseDescriptor instead')
const WaitingRoomsResponse$json = {
  '1': 'WaitingRoomsResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 3, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `WaitingRoomsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomsResponseDescriptor = $convert.base64Decode(
    'ChRXYWl0aW5nUm9vbXNSZXNwb25zZRIhCgJ3chgBIAMoCzIRLnBvbmcuV2FpdGluZ1Jvb21SAn'
    'dy');

@$core.Deprecated('Use joinWaitingRoomRequestDescriptor instead')
const JoinWaitingRoomRequest$json = {
  '1': 'JoinWaitingRoomRequest',
  '2': [
    {'1': 'room_id', '3': 1, '4': 1, '5': 9, '10': 'roomId'},
    {'1': 'client_id', '3': 2, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `JoinWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChZKb2luV2FpdGluZ1Jvb21SZXF1ZXN0EhcKB3Jvb21faWQYASABKAlSBnJvb21JZBIbCgljbG'
    'llbnRfaWQYAiABKAlSCGNsaWVudElk');

@$core.Deprecated('Use joinWaitingRoomResponseDescriptor instead')
const JoinWaitingRoomResponse$json = {
  '1': 'JoinWaitingRoomResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `JoinWaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitingRoomResponseDescriptor = $convert.base64Decode(
    'ChdKb2luV2FpdGluZ1Jvb21SZXNwb25zZRIhCgJ3chgBIAEoCzIRLnBvbmcuV2FpdGluZ1Jvb2'
    '1SAndy');

@$core.Deprecated('Use createWaitingRoomRequestDescriptor instead')
const CreateWaitingRoomRequest$json = {
  '1': 'CreateWaitingRoomRequest',
  '2': [
    {'1': 'host_id', '3': 1, '4': 1, '5': 9, '10': 'hostId'},
    {'1': 'betAmt', '3': 2, '4': 1, '5': 1, '10': 'betAmt'},
  ],
};

/// Descriptor for `CreateWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChhDcmVhdGVXYWl0aW5nUm9vbVJlcXVlc3QSFwoHaG9zdF9pZBgBIAEoCVIGaG9zdElkEhYKBm'
    'JldEFtdBgCIAEoAVIGYmV0QW10');

@$core.Deprecated('Use createWaitingRoomResponseDescriptor instead')
const CreateWaitingRoomResponse$json = {
  '1': 'CreateWaitingRoomResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `CreateWaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWaitingRoomResponseDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVXYWl0aW5nUm9vbVJlc3BvbnNlEiEKAndyGAEgASgLMhEucG9uZy5XYWl0aW5nUm'
    '9vbVICd3I=');

@$core.Deprecated('Use waitingRoomDescriptor instead')
const WaitingRoom$json = {
  '1': 'WaitingRoom',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'host_id', '3': 2, '4': 1, '5': 9, '10': 'hostId'},
    {'1': 'players', '3': 3, '4': 3, '5': 11, '6': '.pong.Player', '10': 'players'},
    {'1': 'bet_amt', '3': 4, '4': 1, '5': 1, '10': 'betAmt'},
  ],
};

/// Descriptor for `WaitingRoom`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomDescriptor = $convert.base64Decode(
    'CgtXYWl0aW5nUm9vbRIOCgJpZBgBIAEoCVICaWQSFwoHaG9zdF9pZBgCIAEoCVIGaG9zdElkEi'
    'YKB3BsYXllcnMYAyADKAsyDC5wb25nLlBsYXllclIHcGxheWVycxIXCgdiZXRfYW10GAQgASgB'
    'UgZiZXRBbXQ=');

@$core.Deprecated('Use waitingRoomRequestDescriptor instead')
const WaitingRoomRequest$json = {
  '1': 'WaitingRoomRequest',
};

/// Descriptor for `WaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomRequestDescriptor = $convert.base64Decode(
    'ChJXYWl0aW5nUm9vbVJlcXVlc3Q=');

@$core.Deprecated('Use waitingRoomResponseDescriptor instead')
const WaitingRoomResponse$json = {
  '1': 'WaitingRoomResponse',
  '2': [
    {'1': 'players', '3': 1, '4': 3, '5': 11, '6': '.pong.Player', '10': 'players'},
  ],
};

/// Descriptor for `WaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomResponseDescriptor = $convert.base64Decode(
    'ChNXYWl0aW5nUm9vbVJlc3BvbnNlEiYKB3BsYXllcnMYASADKAsyDC5wb25nLlBsYXllclIHcG'
    'xheWVycw==');

@$core.Deprecated('Use playerDescriptor instead')
const Player$json = {
  '1': 'Player',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '10': 'uid'},
    {'1': 'nick', '3': 2, '4': 1, '5': 9, '10': 'nick'},
    {'1': 'bet_amt', '3': 3, '4': 1, '5': 1, '10': 'betAmt'},
    {'1': 'number', '3': 4, '4': 1, '5': 5, '10': 'number'},
    {'1': 'score', '3': 5, '4': 1, '5': 5, '10': 'score'},
    {'1': 'ready', '3': 6, '4': 1, '5': 8, '10': 'ready'},
  ],
};

/// Descriptor for `Player`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List playerDescriptor = $convert.base64Decode(
    'CgZQbGF5ZXISEAoDdWlkGAEgASgJUgN1aWQSEgoEbmljaxgCIAEoCVIEbmljaxIXCgdiZXRfYW'
    '10GAMgASgBUgZiZXRBbXQSFgoGbnVtYmVyGAQgASgFUgZudW1iZXISFAoFc2NvcmUYBSABKAVS'
    'BXNjb3JlEhQKBXJlYWR5GAYgASgIUgVyZWFkeQ==');

@$core.Deprecated('Use startGameStreamRequestDescriptor instead')
const StartGameStreamRequest$json = {
  '1': 'StartGameStreamRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `StartGameStreamRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startGameStreamRequestDescriptor = $convert.base64Decode(
    'ChZTdGFydEdhbWVTdHJlYW1SZXF1ZXN0EhsKCWNsaWVudF9pZBgBIAEoCVIIY2xpZW50SWQ=');

@$core.Deprecated('Use gameUpdateBytesDescriptor instead')
const GameUpdateBytes$json = {
  '1': 'GameUpdateBytes',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `GameUpdateBytes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gameUpdateBytesDescriptor = $convert.base64Decode(
    'Cg9HYW1lVXBkYXRlQnl0ZXMSEgoEZGF0YRgBIAEoDFIEZGF0YQ==');

@$core.Deprecated('Use playerInputDescriptor instead')
const PlayerInput$json = {
  '1': 'PlayerInput',
  '2': [
    {'1': 'player_id', '3': 1, '4': 1, '5': 9, '10': 'playerId'},
    {'1': 'input', '3': 2, '4': 1, '5': 9, '10': 'input'},
    {'1': 'player_number', '3': 3, '4': 1, '5': 5, '10': 'playerNumber'},
  ],
};

/// Descriptor for `PlayerInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List playerInputDescriptor = $convert.base64Decode(
    'CgtQbGF5ZXJJbnB1dBIbCglwbGF5ZXJfaWQYASABKAlSCHBsYXllcklkEhQKBWlucHV0GAIgAS'
    'gJUgVpbnB1dBIjCg1wbGF5ZXJfbnVtYmVyGAMgASgFUgxwbGF5ZXJOdW1iZXI=');

@$core.Deprecated('Use gameUpdateDescriptor instead')
const GameUpdate$json = {
  '1': 'GameUpdate',
  '2': [
    {'1': 'gameWidth', '3': 13, '4': 1, '5': 1, '10': 'gameWidth'},
    {'1': 'gameHeight', '3': 14, '4': 1, '5': 1, '10': 'gameHeight'},
    {'1': 'p1Width', '3': 15, '4': 1, '5': 1, '10': 'p1Width'},
    {'1': 'p1Height', '3': 16, '4': 1, '5': 1, '10': 'p1Height'},
    {'1': 'p2Width', '3': 17, '4': 1, '5': 1, '10': 'p2Width'},
    {'1': 'p2Height', '3': 18, '4': 1, '5': 1, '10': 'p2Height'},
    {'1': 'ballWidth', '3': 19, '4': 1, '5': 1, '10': 'ballWidth'},
    {'1': 'ballHeight', '3': 20, '4': 1, '5': 1, '10': 'ballHeight'},
    {'1': 'p1Score', '3': 21, '4': 1, '5': 5, '10': 'p1Score'},
    {'1': 'p2Score', '3': 22, '4': 1, '5': 5, '10': 'p2Score'},
    {'1': 'ballX', '3': 1, '4': 1, '5': 1, '10': 'ballX'},
    {'1': 'ballY', '3': 2, '4': 1, '5': 1, '10': 'ballY'},
    {'1': 'p1X', '3': 3, '4': 1, '5': 1, '10': 'p1X'},
    {'1': 'p1Y', '3': 4, '4': 1, '5': 1, '10': 'p1Y'},
    {'1': 'p2X', '3': 5, '4': 1, '5': 1, '10': 'p2X'},
    {'1': 'p2Y', '3': 6, '4': 1, '5': 1, '10': 'p2Y'},
    {'1': 'p1YVelocity', '3': 7, '4': 1, '5': 1, '10': 'p1YVelocity'},
    {'1': 'p2YVelocity', '3': 8, '4': 1, '5': 1, '10': 'p2YVelocity'},
    {'1': 'ballXVelocity', '3': 9, '4': 1, '5': 1, '10': 'ballXVelocity'},
    {'1': 'ballYVelocity', '3': 10, '4': 1, '5': 1, '10': 'ballYVelocity'},
    {'1': 'fps', '3': 11, '4': 1, '5': 1, '10': 'fps'},
    {'1': 'tps', '3': 12, '4': 1, '5': 1, '10': 'tps'},
    {'1': 'error', '3': 23, '4': 1, '5': 9, '10': 'error'},
    {'1': 'debug', '3': 24, '4': 1, '5': 8, '10': 'debug'},
  ],
};

/// Descriptor for `GameUpdate`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gameUpdateDescriptor = $convert.base64Decode(
    'CgpHYW1lVXBkYXRlEhwKCWdhbWVXaWR0aBgNIAEoAVIJZ2FtZVdpZHRoEh4KCmdhbWVIZWlnaH'
    'QYDiABKAFSCmdhbWVIZWlnaHQSGAoHcDFXaWR0aBgPIAEoAVIHcDFXaWR0aBIaCghwMUhlaWdo'
    'dBgQIAEoAVIIcDFIZWlnaHQSGAoHcDJXaWR0aBgRIAEoAVIHcDJXaWR0aBIaCghwMkhlaWdodB'
    'gSIAEoAVIIcDJIZWlnaHQSHAoJYmFsbFdpZHRoGBMgASgBUgliYWxsV2lkdGgSHgoKYmFsbEhl'
    'aWdodBgUIAEoAVIKYmFsbEhlaWdodBIYCgdwMVNjb3JlGBUgASgFUgdwMVNjb3JlEhgKB3AyU2'
    'NvcmUYFiABKAVSB3AyU2NvcmUSFAoFYmFsbFgYASABKAFSBWJhbGxYEhQKBWJhbGxZGAIgASgB'
    'UgViYWxsWRIQCgNwMVgYAyABKAFSA3AxWBIQCgNwMVkYBCABKAFSA3AxWRIQCgNwMlgYBSABKA'
    'FSA3AyWBIQCgNwMlkYBiABKAFSA3AyWRIgCgtwMVlWZWxvY2l0eRgHIAEoAVILcDFZVmVsb2Np'
    'dHkSIAoLcDJZVmVsb2NpdHkYCCABKAFSC3AyWVZlbG9jaXR5EiQKDWJhbGxYVmVsb2NpdHkYCS'
    'ABKAFSDWJhbGxYVmVsb2NpdHkSJAoNYmFsbFlWZWxvY2l0eRgKIAEoAVINYmFsbFlWZWxvY2l0'
    'eRIQCgNmcHMYCyABKAFSA2ZwcxIQCgN0cHMYDCABKAFSA3RwcxIUCgVlcnJvchgXIAEoCVIFZX'
    'Jyb3ISFAoFZGVidWcYGCABKAhSBWRlYnVn');

