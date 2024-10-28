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
    {'1': 'started', '3': 1, '4': 1, '5': 8, '10': 'started'},
    {'1': 'player_number', '3': 2, '4': 1, '5': 5, '10': 'playerNumber'},
    {'1': 'message', '3': 3, '4': 1, '5': 9, '10': 'message'},
    {'1': 'client_id', '3': 4, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `NtfnStreamResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ntfnStreamResponseDescriptor = $convert.base64Decode(
    'ChJOdGZuU3RyZWFtUmVzcG9uc2USGAoHc3RhcnRlZBgBIAEoCFIHc3RhcnRlZBIjCg1wbGF5ZX'
    'JfbnVtYmVyGAIgASgFUgxwbGF5ZXJOdW1iZXISGAoHbWVzc2FnZRgDIAEoCVIHbWVzc2FnZRIb'
    'CgljbGllbnRfaWQYBCABKAlSCGNsaWVudElk');

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
    {'1': 'playerId', '3': 1, '4': 1, '5': 9, '10': 'playerId'},
    {'1': 'input', '3': 2, '4': 1, '5': 9, '10': 'input'},
    {'1': 'playerNumber', '3': 3, '4': 1, '5': 5, '10': 'playerNumber'},
  ],
};

/// Descriptor for `PlayerInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List playerInputDescriptor = $convert.base64Decode(
    'CgtQbGF5ZXJJbnB1dBIaCghwbGF5ZXJJZBgBIAEoCVIIcGxheWVySWQSFAoFaW5wdXQYAiABKA'
    'lSBWlucHV0EiIKDHBsYXllck51bWJlchgDIAEoBVIMcGxheWVyTnVtYmVy');

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

