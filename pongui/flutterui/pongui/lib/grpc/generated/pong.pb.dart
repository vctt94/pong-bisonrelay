//
//  Generated code. Do not modify.
//  source: pong.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

class WaitingRoomRequest extends $pb.GeneratedMessage {
  factory WaitingRoomRequest() => create();
  WaitingRoomRequest._() : super();
  factory WaitingRoomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoomRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoomRequest clone() => WaitingRoomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoomRequest copyWith(void Function(WaitingRoomRequest) updates) => super.copyWith((message) => updates(message as WaitingRoomRequest)) as WaitingRoomRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomRequest create() => WaitingRoomRequest._();
  WaitingRoomRequest createEmptyInstance() => create();
  static $pb.PbList<WaitingRoomRequest> createRepeated() => $pb.PbList<WaitingRoomRequest>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoomRequest>(create);
  static WaitingRoomRequest? _defaultInstance;
}

class Player extends $pb.GeneratedMessage {
  factory Player({
    $core.String? playerId,
    $core.String? nick,
    $core.double? betAmount,
  }) {
    final $result = create();
    if (playerId != null) {
      $result.playerId = playerId;
    }
    if (nick != null) {
      $result.nick = nick;
    }
    if (betAmount != null) {
      $result.betAmount = betAmount;
    }
    return $result;
  }
  Player._() : super();
  factory Player.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Player.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Player', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'playerId')
    ..aOS(2, _omitFieldNames ? '' : 'nick')
    ..a<$core.double>(3, _omitFieldNames ? '' : 'betAmount', $pb.PbFieldType.OD, protoName: 'betAmount')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Player clone() => Player()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Player copyWith(void Function(Player) updates) => super.copyWith((message) => updates(message as Player)) as Player;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Player create() => Player._();
  Player createEmptyInstance() => create();
  static $pb.PbList<Player> createRepeated() => $pb.PbList<Player>();
  @$core.pragma('dart2js:noInline')
  static Player getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Player>(create);
  static Player? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get playerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set playerId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPlayerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlayerId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get nick => $_getSZ(1);
  @$pb.TagNumber(2)
  set nick($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNick() => $_has(1);
  @$pb.TagNumber(2)
  void clearNick() => clearField(2);

  @$pb.TagNumber(3)
  $core.double get betAmount => $_getN(2);
  @$pb.TagNumber(3)
  set betAmount($core.double v) { $_setDouble(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBetAmount() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAmount() => clearField(3);
}

class WaitingRoomResponse extends $pb.GeneratedMessage {
  factory WaitingRoomResponse({
    $core.Iterable<Player>? players,
  }) {
    final $result = create();
    if (players != null) {
      $result.players.addAll(players);
    }
    return $result;
  }
  WaitingRoomResponse._() : super();
  factory WaitingRoomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoomResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..pc<Player>(1, _omitFieldNames ? '' : 'players', $pb.PbFieldType.PM, subBuilder: Player.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoomResponse clone() => WaitingRoomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoomResponse copyWith(void Function(WaitingRoomResponse) updates) => super.copyWith((message) => updates(message as WaitingRoomResponse)) as WaitingRoomResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomResponse create() => WaitingRoomResponse._();
  WaitingRoomResponse createEmptyInstance() => create();
  static $pb.PbList<WaitingRoomResponse> createRepeated() => $pb.PbList<WaitingRoomResponse>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoomResponse>(create);
  static WaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Player> get players => $_getList(0);
}

class StartNtfnStreamRequest extends $pb.GeneratedMessage {
  factory StartNtfnStreamRequest({
    $core.String? clientId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  StartNtfnStreamRequest._() : super();
  factory StartNtfnStreamRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StartNtfnStreamRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'StartNtfnStreamRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StartNtfnStreamRequest clone() => StartNtfnStreamRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StartNtfnStreamRequest copyWith(void Function(StartNtfnStreamRequest) updates) => super.copyWith((message) => updates(message as StartNtfnStreamRequest)) as StartNtfnStreamRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartNtfnStreamRequest create() => StartNtfnStreamRequest._();
  StartNtfnStreamRequest createEmptyInstance() => create();
  static $pb.PbList<StartNtfnStreamRequest> createRepeated() => $pb.PbList<StartNtfnStreamRequest>();
  @$core.pragma('dart2js:noInline')
  static StartNtfnStreamRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartNtfnStreamRequest>(create);
  static StartNtfnStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);
}

class NtfnStreamResponse extends $pb.GeneratedMessage {
  factory NtfnStreamResponse({
    $core.bool? started,
    $core.int? playerNumber,
    $core.String? message,
    $core.String? clientId,
  }) {
    final $result = create();
    if (started != null) {
      $result.started = started;
    }
    if (playerNumber != null) {
      $result.playerNumber = playerNumber;
    }
    if (message != null) {
      $result.message = message;
    }
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  NtfnStreamResponse._() : super();
  factory NtfnStreamResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory NtfnStreamResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'NtfnStreamResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'started')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'playerNumber', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'message')
    ..aOS(4, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  NtfnStreamResponse clone() => NtfnStreamResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  NtfnStreamResponse copyWith(void Function(NtfnStreamResponse) updates) => super.copyWith((message) => updates(message as NtfnStreamResponse)) as NtfnStreamResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NtfnStreamResponse create() => NtfnStreamResponse._();
  NtfnStreamResponse createEmptyInstance() => create();
  static $pb.PbList<NtfnStreamResponse> createRepeated() => $pb.PbList<NtfnStreamResponse>();
  @$core.pragma('dart2js:noInline')
  static NtfnStreamResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NtfnStreamResponse>(create);
  static NtfnStreamResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get started => $_getBF(0);
  @$pb.TagNumber(1)
  set started($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasStarted() => $_has(0);
  @$pb.TagNumber(1)
  void clearStarted() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get playerNumber => $_getIZ(1);
  @$pb.TagNumber(2)
  set playerNumber($core.int v) { $_setSignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasPlayerNumber() => $_has(1);
  @$pb.TagNumber(2)
  void clearPlayerNumber() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get message => $_getSZ(2);
  @$pb.TagNumber(3)
  set message($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMessage() => $_has(2);
  @$pb.TagNumber(3)
  void clearMessage() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get clientId => $_getSZ(3);
  @$pb.TagNumber(4)
  set clientId($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasClientId() => $_has(3);
  @$pb.TagNumber(4)
  void clearClientId() => clearField(4);
}

/// SignalReadyRequest contains information about the client signaling readiness
class StartGameStreamRequest extends $pb.GeneratedMessage {
  factory StartGameStreamRequest({
    $core.String? clientId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  StartGameStreamRequest._() : super();
  factory StartGameStreamRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StartGameStreamRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'StartGameStreamRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StartGameStreamRequest clone() => StartGameStreamRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StartGameStreamRequest copyWith(void Function(StartGameStreamRequest) updates) => super.copyWith((message) => updates(message as StartGameStreamRequest)) as StartGameStreamRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartGameStreamRequest create() => StartGameStreamRequest._();
  StartGameStreamRequest createEmptyInstance() => create();
  static $pb.PbList<StartGameStreamRequest> createRepeated() => $pb.PbList<StartGameStreamRequest>();
  @$core.pragma('dart2js:noInline')
  static StartGameStreamRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartGameStreamRequest>(create);
  static StartGameStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);
}

class GameUpdateBytes extends $pb.GeneratedMessage {
  factory GameUpdateBytes({
    $core.List<$core.int>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  GameUpdateBytes._() : super();
  factory GameUpdateBytes.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GameUpdateBytes.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GameUpdateBytes', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GameUpdateBytes clone() => GameUpdateBytes()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GameUpdateBytes copyWith(void Function(GameUpdateBytes) updates) => super.copyWith((message) => updates(message as GameUpdateBytes)) as GameUpdateBytes;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GameUpdateBytes create() => GameUpdateBytes._();
  GameUpdateBytes createEmptyInstance() => create();
  static $pb.PbList<GameUpdateBytes> createRepeated() => $pb.PbList<GameUpdateBytes>();
  @$core.pragma('dart2js:noInline')
  static GameUpdateBytes getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GameUpdateBytes>(create);
  static GameUpdateBytes? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
}

class PlayerInput extends $pb.GeneratedMessage {
  factory PlayerInput({
    $core.String? playerId,
    $core.String? input,
    $core.int? playerNumber,
  }) {
    final $result = create();
    if (playerId != null) {
      $result.playerId = playerId;
    }
    if (input != null) {
      $result.input = input;
    }
    if (playerNumber != null) {
      $result.playerNumber = playerNumber;
    }
    return $result;
  }
  PlayerInput._() : super();
  factory PlayerInput.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PlayerInput.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PlayerInput', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'playerId', protoName: 'playerId')
    ..aOS(2, _omitFieldNames ? '' : 'input')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'playerNumber', $pb.PbFieldType.O3, protoName: 'playerNumber')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PlayerInput clone() => PlayerInput()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PlayerInput copyWith(void Function(PlayerInput) updates) => super.copyWith((message) => updates(message as PlayerInput)) as PlayerInput;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PlayerInput create() => PlayerInput._();
  PlayerInput createEmptyInstance() => create();
  static $pb.PbList<PlayerInput> createRepeated() => $pb.PbList<PlayerInput>();
  @$core.pragma('dart2js:noInline')
  static PlayerInput getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PlayerInput>(create);
  static PlayerInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get playerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set playerId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPlayerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlayerId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get input => $_getSZ(1);
  @$pb.TagNumber(2)
  set input($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasInput() => $_has(1);
  @$pb.TagNumber(2)
  void clearInput() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get playerNumber => $_getIZ(2);
  @$pb.TagNumber(3)
  set playerNumber($core.int v) { $_setSignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPlayerNumber() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlayerNumber() => clearField(3);
}

class GameUpdate extends $pb.GeneratedMessage {
  factory GameUpdate({
    $core.double? ballX,
    $core.double? ballY,
    $core.double? p1X,
    $core.double? p1Y,
    $core.double? p2X,
    $core.double? p2Y,
    $core.double? p1YVelocity,
    $core.double? p2YVelocity,
    $core.double? ballXVelocity,
    $core.double? ballYVelocity,
    $core.double? fps,
    $core.double? tps,
    $core.double? gameWidth,
    $core.double? gameHeight,
    $core.double? p1Width,
    $core.double? p1Height,
    $core.double? p2Width,
    $core.double? p2Height,
    $core.double? ballWidth,
    $core.double? ballHeight,
    $core.int? p1Score,
    $core.int? p2Score,
    $core.String? error,
    $core.bool? debug,
  }) {
    final $result = create();
    if (ballX != null) {
      $result.ballX = ballX;
    }
    if (ballY != null) {
      $result.ballY = ballY;
    }
    if (p1X != null) {
      $result.p1X = p1X;
    }
    if (p1Y != null) {
      $result.p1Y = p1Y;
    }
    if (p2X != null) {
      $result.p2X = p2X;
    }
    if (p2Y != null) {
      $result.p2Y = p2Y;
    }
    if (p1YVelocity != null) {
      $result.p1YVelocity = p1YVelocity;
    }
    if (p2YVelocity != null) {
      $result.p2YVelocity = p2YVelocity;
    }
    if (ballXVelocity != null) {
      $result.ballXVelocity = ballXVelocity;
    }
    if (ballYVelocity != null) {
      $result.ballYVelocity = ballYVelocity;
    }
    if (fps != null) {
      $result.fps = fps;
    }
    if (tps != null) {
      $result.tps = tps;
    }
    if (gameWidth != null) {
      $result.gameWidth = gameWidth;
    }
    if (gameHeight != null) {
      $result.gameHeight = gameHeight;
    }
    if (p1Width != null) {
      $result.p1Width = p1Width;
    }
    if (p1Height != null) {
      $result.p1Height = p1Height;
    }
    if (p2Width != null) {
      $result.p2Width = p2Width;
    }
    if (p2Height != null) {
      $result.p2Height = p2Height;
    }
    if (ballWidth != null) {
      $result.ballWidth = ballWidth;
    }
    if (ballHeight != null) {
      $result.ballHeight = ballHeight;
    }
    if (p1Score != null) {
      $result.p1Score = p1Score;
    }
    if (p2Score != null) {
      $result.p2Score = p2Score;
    }
    if (error != null) {
      $result.error = error;
    }
    if (debug != null) {
      $result.debug = debug;
    }
    return $result;
  }
  GameUpdate._() : super();
  factory GameUpdate.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GameUpdate.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GameUpdate', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..a<$core.double>(1, _omitFieldNames ? '' : 'ballX', $pb.PbFieldType.OD, protoName: 'ballX')
    ..a<$core.double>(2, _omitFieldNames ? '' : 'ballY', $pb.PbFieldType.OD, protoName: 'ballY')
    ..a<$core.double>(3, _omitFieldNames ? '' : 'p1X', $pb.PbFieldType.OD, protoName: 'p1X')
    ..a<$core.double>(4, _omitFieldNames ? '' : 'p1Y', $pb.PbFieldType.OD, protoName: 'p1Y')
    ..a<$core.double>(5, _omitFieldNames ? '' : 'p2X', $pb.PbFieldType.OD, protoName: 'p2X')
    ..a<$core.double>(6, _omitFieldNames ? '' : 'p2Y', $pb.PbFieldType.OD, protoName: 'p2Y')
    ..a<$core.double>(7, _omitFieldNames ? '' : 'p1YVelocity', $pb.PbFieldType.OD, protoName: 'p1YVelocity')
    ..a<$core.double>(8, _omitFieldNames ? '' : 'p2YVelocity', $pb.PbFieldType.OD, protoName: 'p2YVelocity')
    ..a<$core.double>(9, _omitFieldNames ? '' : 'ballXVelocity', $pb.PbFieldType.OD, protoName: 'ballXVelocity')
    ..a<$core.double>(10, _omitFieldNames ? '' : 'ballYVelocity', $pb.PbFieldType.OD, protoName: 'ballYVelocity')
    ..a<$core.double>(11, _omitFieldNames ? '' : 'fps', $pb.PbFieldType.OD)
    ..a<$core.double>(12, _omitFieldNames ? '' : 'tps', $pb.PbFieldType.OD)
    ..a<$core.double>(13, _omitFieldNames ? '' : 'gameWidth', $pb.PbFieldType.OD, protoName: 'gameWidth')
    ..a<$core.double>(14, _omitFieldNames ? '' : 'gameHeight', $pb.PbFieldType.OD, protoName: 'gameHeight')
    ..a<$core.double>(15, _omitFieldNames ? '' : 'p1Width', $pb.PbFieldType.OD, protoName: 'p1Width')
    ..a<$core.double>(16, _omitFieldNames ? '' : 'p1Height', $pb.PbFieldType.OD, protoName: 'p1Height')
    ..a<$core.double>(17, _omitFieldNames ? '' : 'p2Width', $pb.PbFieldType.OD, protoName: 'p2Width')
    ..a<$core.double>(18, _omitFieldNames ? '' : 'p2Height', $pb.PbFieldType.OD, protoName: 'p2Height')
    ..a<$core.double>(19, _omitFieldNames ? '' : 'ballWidth', $pb.PbFieldType.OD, protoName: 'ballWidth')
    ..a<$core.double>(20, _omitFieldNames ? '' : 'ballHeight', $pb.PbFieldType.OD, protoName: 'ballHeight')
    ..a<$core.int>(21, _omitFieldNames ? '' : 'p1Score', $pb.PbFieldType.O3, protoName: 'p1Score')
    ..a<$core.int>(22, _omitFieldNames ? '' : 'p2Score', $pb.PbFieldType.O3, protoName: 'p2Score')
    ..aOS(23, _omitFieldNames ? '' : 'error')
    ..aOB(24, _omitFieldNames ? '' : 'debug')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GameUpdate clone() => GameUpdate()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GameUpdate copyWith(void Function(GameUpdate) updates) => super.copyWith((message) => updates(message as GameUpdate)) as GameUpdate;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GameUpdate create() => GameUpdate._();
  GameUpdate createEmptyInstance() => create();
  static $pb.PbList<GameUpdate> createRepeated() => $pb.PbList<GameUpdate>();
  @$core.pragma('dart2js:noInline')
  static GameUpdate getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GameUpdate>(create);
  static GameUpdate? _defaultInstance;

  @$pb.TagNumber(1)
  $core.double get ballX => $_getN(0);
  @$pb.TagNumber(1)
  set ballX($core.double v) { $_setDouble(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasBallX() => $_has(0);
  @$pb.TagNumber(1)
  void clearBallX() => clearField(1);

  @$pb.TagNumber(2)
  $core.double get ballY => $_getN(1);
  @$pb.TagNumber(2)
  set ballY($core.double v) { $_setDouble(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBallY() => $_has(1);
  @$pb.TagNumber(2)
  void clearBallY() => clearField(2);

  @$pb.TagNumber(3)
  $core.double get p1X => $_getN(2);
  @$pb.TagNumber(3)
  set p1X($core.double v) { $_setDouble(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasP1X() => $_has(2);
  @$pb.TagNumber(3)
  void clearP1X() => clearField(3);

  @$pb.TagNumber(4)
  $core.double get p1Y => $_getN(3);
  @$pb.TagNumber(4)
  set p1Y($core.double v) { $_setDouble(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasP1Y() => $_has(3);
  @$pb.TagNumber(4)
  void clearP1Y() => clearField(4);

  @$pb.TagNumber(5)
  $core.double get p2X => $_getN(4);
  @$pb.TagNumber(5)
  set p2X($core.double v) { $_setDouble(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasP2X() => $_has(4);
  @$pb.TagNumber(5)
  void clearP2X() => clearField(5);

  @$pb.TagNumber(6)
  $core.double get p2Y => $_getN(5);
  @$pb.TagNumber(6)
  set p2Y($core.double v) { $_setDouble(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasP2Y() => $_has(5);
  @$pb.TagNumber(6)
  void clearP2Y() => clearField(6);

  @$pb.TagNumber(7)
  $core.double get p1YVelocity => $_getN(6);
  @$pb.TagNumber(7)
  set p1YVelocity($core.double v) { $_setDouble(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasP1YVelocity() => $_has(6);
  @$pb.TagNumber(7)
  void clearP1YVelocity() => clearField(7);

  @$pb.TagNumber(8)
  $core.double get p2YVelocity => $_getN(7);
  @$pb.TagNumber(8)
  set p2YVelocity($core.double v) { $_setDouble(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasP2YVelocity() => $_has(7);
  @$pb.TagNumber(8)
  void clearP2YVelocity() => clearField(8);

  @$pb.TagNumber(9)
  $core.double get ballXVelocity => $_getN(8);
  @$pb.TagNumber(9)
  set ballXVelocity($core.double v) { $_setDouble(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasBallXVelocity() => $_has(8);
  @$pb.TagNumber(9)
  void clearBallXVelocity() => clearField(9);

  @$pb.TagNumber(10)
  $core.double get ballYVelocity => $_getN(9);
  @$pb.TagNumber(10)
  set ballYVelocity($core.double v) { $_setDouble(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasBallYVelocity() => $_has(9);
  @$pb.TagNumber(10)
  void clearBallYVelocity() => clearField(10);

  @$pb.TagNumber(11)
  $core.double get fps => $_getN(10);
  @$pb.TagNumber(11)
  set fps($core.double v) { $_setDouble(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasFps() => $_has(10);
  @$pb.TagNumber(11)
  void clearFps() => clearField(11);

  @$pb.TagNumber(12)
  $core.double get tps => $_getN(11);
  @$pb.TagNumber(12)
  set tps($core.double v) { $_setDouble(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasTps() => $_has(11);
  @$pb.TagNumber(12)
  void clearTps() => clearField(12);

  @$pb.TagNumber(13)
  $core.double get gameWidth => $_getN(12);
  @$pb.TagNumber(13)
  set gameWidth($core.double v) { $_setDouble(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasGameWidth() => $_has(12);
  @$pb.TagNumber(13)
  void clearGameWidth() => clearField(13);

  @$pb.TagNumber(14)
  $core.double get gameHeight => $_getN(13);
  @$pb.TagNumber(14)
  set gameHeight($core.double v) { $_setDouble(13, v); }
  @$pb.TagNumber(14)
  $core.bool hasGameHeight() => $_has(13);
  @$pb.TagNumber(14)
  void clearGameHeight() => clearField(14);

  @$pb.TagNumber(15)
  $core.double get p1Width => $_getN(14);
  @$pb.TagNumber(15)
  set p1Width($core.double v) { $_setDouble(14, v); }
  @$pb.TagNumber(15)
  $core.bool hasP1Width() => $_has(14);
  @$pb.TagNumber(15)
  void clearP1Width() => clearField(15);

  @$pb.TagNumber(16)
  $core.double get p1Height => $_getN(15);
  @$pb.TagNumber(16)
  set p1Height($core.double v) { $_setDouble(15, v); }
  @$pb.TagNumber(16)
  $core.bool hasP1Height() => $_has(15);
  @$pb.TagNumber(16)
  void clearP1Height() => clearField(16);

  @$pb.TagNumber(17)
  $core.double get p2Width => $_getN(16);
  @$pb.TagNumber(17)
  set p2Width($core.double v) { $_setDouble(16, v); }
  @$pb.TagNumber(17)
  $core.bool hasP2Width() => $_has(16);
  @$pb.TagNumber(17)
  void clearP2Width() => clearField(17);

  @$pb.TagNumber(18)
  $core.double get p2Height => $_getN(17);
  @$pb.TagNumber(18)
  set p2Height($core.double v) { $_setDouble(17, v); }
  @$pb.TagNumber(18)
  $core.bool hasP2Height() => $_has(17);
  @$pb.TagNumber(18)
  void clearP2Height() => clearField(18);

  @$pb.TagNumber(19)
  $core.double get ballWidth => $_getN(18);
  @$pb.TagNumber(19)
  set ballWidth($core.double v) { $_setDouble(18, v); }
  @$pb.TagNumber(19)
  $core.bool hasBallWidth() => $_has(18);
  @$pb.TagNumber(19)
  void clearBallWidth() => clearField(19);

  @$pb.TagNumber(20)
  $core.double get ballHeight => $_getN(19);
  @$pb.TagNumber(20)
  set ballHeight($core.double v) { $_setDouble(19, v); }
  @$pb.TagNumber(20)
  $core.bool hasBallHeight() => $_has(19);
  @$pb.TagNumber(20)
  void clearBallHeight() => clearField(20);

  @$pb.TagNumber(21)
  $core.int get p1Score => $_getIZ(20);
  @$pb.TagNumber(21)
  set p1Score($core.int v) { $_setSignedInt32(20, v); }
  @$pb.TagNumber(21)
  $core.bool hasP1Score() => $_has(20);
  @$pb.TagNumber(21)
  void clearP1Score() => clearField(21);

  @$pb.TagNumber(22)
  $core.int get p2Score => $_getIZ(21);
  @$pb.TagNumber(22)
  set p2Score($core.int v) { $_setSignedInt32(21, v); }
  @$pb.TagNumber(22)
  $core.bool hasP2Score() => $_has(21);
  @$pb.TagNumber(22)
  void clearP2Score() => clearField(22);

  /// Optional: if you want to send error messages or debug information
  @$pb.TagNumber(23)
  $core.String get error => $_getSZ(22);
  @$pb.TagNumber(23)
  set error($core.String v) { $_setString(22, v); }
  @$pb.TagNumber(23)
  $core.bool hasError() => $_has(22);
  @$pb.TagNumber(23)
  void clearError() => clearField(23);

  @$pb.TagNumber(24)
  $core.bool get debug => $_getBF(23);
  @$pb.TagNumber(24)
  set debug($core.bool v) { $_setBool(23, v); }
  @$pb.TagNumber(24)
  $core.bool hasDebug() => $_has(23);
  @$pb.TagNumber(24)
  void clearDebug() => clearField(24);
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
