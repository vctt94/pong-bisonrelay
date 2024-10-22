// ignore_for_file: constant_identifier_names

import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:convert/convert.dart';
import 'package:flutter/cupertino.dart';
import 'package:golib_plugin/mock.dart';
import 'package:golib_plugin/util.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:blake_hash/blake_hash.dart';

part 'definitions.g.dart';

@JsonSerializable()
class InitClient {
  @JsonKey(name: 'server_addr')
  final String serverAddr;
  @JsonKey(name: 'log_file')
  final String logFile;
  @JsonKey(name: "msgs_root")
  final String msgsRoot;
  @JsonKey(name: 'debug_level')
  final String debugLevel;
  @JsonKey(name: 'wants_log_ntfns')
  final bool wantsLogNtfns;
 
  // rpc fields
  @JsonKey(name: 'json_rpc_listen')
  final List<String> jsonrpclisten;
  @JsonKey(name: 'rpc_cert_path')
  final String rpccertpath;
  @JsonKey(name: 'rpc_key_path')
  final String rpckeypath;
  @JsonKey(name: 'rpc_issue_client_cert')
  final bool rpcissueclientcert;
    @JsonKey(name: 'rpc_client_ca_path')
  final String rpcclientcapath;
  @JsonKey(name: 'rpc_user')
  final String rpcUser;
  @JsonKey(name: 'rpc_pass')
  final String rpcPass;
  @JsonKey(name: 'rpc_auth_mode')
  final String rpcAuthMode;
  

  InitClient(
    this.serverAddr,
    this.logFile,
    this.msgsRoot,
    this.debugLevel,
    this.wantsLogNtfns,
    this.jsonrpclisten,
    this.rpccertpath,
    this.rpckeypath,
    this.rpcissueclientcert,
    this.rpcclientcapath,
    this.rpcUser,
    this.rpcPass,
    this.rpcAuthMode
  );

  Map<String, dynamic> toJson() => _$InitClientToJson(this);
}

@JsonSerializable()
class IDInit {
  final String nick;
  final String name;
  IDInit(this.nick, this.name);
  factory IDInit.fromJson(Map<String, dynamic> json) => _$IDInitFromJson(json);

  Map<String, dynamic> toJson() => _$IDInitToJson(this);
}

@JsonSerializable()
class LocalInfo {
  final String id;
  final String nick;
  LocalInfo(this.id, this.nick);
  factory LocalInfo.fromJson(Map<String, dynamic> json) =>
      _$LocalInfoFromJson(json);
}

@JsonSerializable()
class ServerCert {
  @JsonKey(name: "inner_fingerprint")
  final String innerFingerprint;
  @JsonKey(name: "outer_fingerprint")
  final String outerFingerprint;
  const ServerCert(this.innerFingerprint, this.outerFingerprint);

  factory ServerCert.fromJson(Map<String, dynamic> json) =>
      _$ServerCertFromJson(json);
}

const connStateOffline = 0;
const connStateCheckingWallet = 1;
const connStateOnline = 2;


@JsonSerializable()
class ServerInfo {
  final String innerFingerprint;
  final String outerFingerprint;
  final String serverAddr;
  const ServerInfo(
      {required this.innerFingerprint,
      required this.outerFingerprint,
      required this.serverAddr});
  const ServerInfo.empty()
      : this(innerFingerprint: "", outerFingerprint: "", serverAddr: "");

  factory ServerInfo.fromJson(Map<String, dynamic> json) =>
      _$ServerInfoFromJson(json);
}

@JsonSerializable()
class RemoteUser {
  final String uid;
  final String nick;

  const RemoteUser(this.uid, this.nick);

  factory RemoteUser.fromJson(Map<String, dynamic> json) =>
      _$RemoteUserFromJson(json);
}


@JsonSerializable()
class PublicIdentity {
  final String name;
  final String nick;
  final String identity;

  PublicIdentity(this.name, this.nick, this.identity);
  factory PublicIdentity.fromJson(Map<String, dynamic> json) =>
      _$PublicIdentityFromJson(json);
}


@JsonSerializable()
class Account {
  final String name;
  @JsonKey(name: "unconfirmed_balance")
  final int unconfirmedBalance;
  @JsonKey(name: "confirmed_balance")
  final int confirmedBalance;
  @JsonKey(name: "internal_key_count")
  final int internalKeyCount;
  @JsonKey(name: "external_key_count")
  final int externalKeyCount;

  Account(this.name, this.unconfirmedBalance, this.confirmedBalance,
      this.internalKeyCount, this.externalKeyCount);

  factory Account.fromJson(Map<String, dynamic> json) =>
      _$AccountFromJson(json);
}

@JsonSerializable()
class LogEntry {
  final String from;
  final String message;
  final bool internal;
  final int timestamp;
  LogEntry(this.from, this.message, this.internal, this.timestamp);

  factory LogEntry.fromJson(Map<String, dynamic> json) =>
      _$LogEntryFromJson(json);
}

@JsonSerializable()
class SendOnChain {
  final String addr;
  final int amount;
  @JsonKey(name: "from_account")
  final String fromAccount;

  SendOnChain(this.addr, this.amount, this.fromAccount);
  Map<String, dynamic> toJson() => _$SendOnChainToJson(this);
}

@JsonSerializable()
class LoadUserHistory {
  final String uid;
  @JsonKey(name: "is_gc")
  final bool isGC;
  final int page;
  @JsonKey(name: "page_num")
  final int pageNum;

  LoadUserHistory(this.uid, this.isGC, this.page, this.pageNum);
  Map<String, dynamic> toJson() => _$LoadUserHistoryToJson(this);
}

@JsonSerializable()
class WriteInvite {
  @JsonKey(name: "fund_amount")
  final int fundAmount;
  @JsonKey(name: "fund_account")
  final String fundAccount;
  @JsonKey(name: "gc_id")
  final String? gcid;
  final bool prepaid;

  WriteInvite(this.fundAmount, this.fundAccount, this.gcid, this.prepaid);
  Map<String, dynamic> toJson() => _$WriteInviteToJson(this);
}

@JsonSerializable()
class RedeemedInviteFunds {
  final String txid;
  final int total;

  RedeemedInviteFunds(this.txid, this.total);
  factory RedeemedInviteFunds.fromJson(Map<String, dynamic> json) =>
      _$RedeemedInviteFundsFromJson(json);
}

enum OnboardStage {
  @JsonValue("fetching_invite")
  stageFetchingInvite,
  @JsonValue("invite_unpaid")
  stageInviteUnpaid,
  @JsonValue("invite_no_funds")
  stageInviteNoFunds,
  @JsonValue("invite_fetch_timeout")
  stageInviteFetchTimeout,
  @JsonValue("redeeming_funds")
  stageRedeemingFunds,
  @JsonValue("waiting_out_mined")
  stageWaitingOutMined,
  @JsonValue("waiting_funds_confirm")
  stageWaitingFundsConfirm,
  @JsonValue("opening_outbound")
  stageOpeningOutbound,
  @JsonValue("waiting_out_confirm")
  stageWaitingOutConfirm,
  @JsonValue("opening_inbound")
  stageOpeningInbound,
  @JsonValue("initial_kx")
  stageInitialKX,
  @JsonValue("done")
  stageOnboardDone,
}

@JsonSerializable()
class RunState {
  @JsonKey(name: "dcrlnd_running")
  final bool dcrlndRunning;
  @JsonKey(name: "client_running")
  final bool clientRunning;

  RunState({required this.dcrlndRunning, required this.clientRunning});
  factory RunState.fromJson(Map<String, dynamic> json) =>
      _$RunStateFromJson(json);
}

@JsonSerializable()
class ZipLogsArgs {
  @JsonKey(name: "include_golib")
  final bool includeGolib;
  @JsonKey(name: "include_ln")
  final bool includeLn;
  @JsonKey(name: "only_last_file")
  final bool onlyLastFile;
  @JsonKey(name: "dest_path")
  final String destPath;

  ZipLogsArgs(
      this.includeGolib, this.includeLn, this.onlyLastFile, this.destPath);
  Map<String, dynamic> toJson() => _$ZipLogsArgsToJson(this);
}

const UINtfnPM = "pm";
const UINtfnGCM = "gcm";
const UINtfnGCMMention = "gcmmention";
const UINtfnMultiple = "multiple";

@JsonSerializable()
class UINotification {
  final String type;
  final String text;
  final int count;
  final String from;

  UINotification(this.type, this.text, this.count, this.from);
  factory UINotification.fromJson(Map<String, dynamic> json) =>
      _$UINotificationFromJson(json);
}

@JsonSerializable()
class UINotificationsConfig {
  final bool pms;
  final bool gcms;
  @JsonKey(name: "gcmentions")
  final bool gcMentions;

  UINotificationsConfig(this.pms, this.gcms, this.gcMentions);
  factory UINotificationsConfig.disabled() =>
      UINotificationsConfig(false, false, false);
  factory UINotificationsConfig.fromJson(Map<String, dynamic> json) =>
      _$UINotificationsConfigFromJson(json);
  Map<String, dynamic> toJson() => _$UINotificationsConfigToJson(this);
}

mixin NtfStreams {
  StreamController<RemoteUser> ntfAcceptedInvites =
      StreamController<RemoteUser>();
  Stream<RemoteUser> acceptedInvites() => ntfAcceptedInvites.stream;


  StreamController<String> ntfLogLines = StreamController<String>();
  Stream<String> logLines() => ntfLogLines.stream;

  StreamController<int> ntfRescanProgress = StreamController<int>();
  Stream<int> rescanWalletProgress() => ntfRescanProgress.stream;

  StreamController<UINotification> ntfUINotifications =
      StreamController<UINotification>();
  Stream<UINotification> uiNotifications() => ntfUINotifications.stream;

  handleNotifications(int cmd, bool isError, String jsonPayload) {
    dynamic payload;
    if (jsonPayload != "") {
      payload = jsonDecode(jsonPayload);
    }

    switch (cmd) {


      // case NTPM:
      //   isError
      //       ? ntfChatEvents.addError(payload)
      //       : ntfChatEvents.add(PM.fromJson(payload));
      //   break;


      default:
        debugPrint("Received unknown notification ${cmd.toRadixString(16)}");
    }
  }
}

abstract class PluginPlatform {
  Future<String?> get platformVersion => throw "unimplemented";
  String get majorPlatform => "unknown-major-plat";
  String get minorPlatform => "unknown-minor-plat";
  Future<void> setTag(String tag) async => throw "unimplemented";
  Future<void> hello() async => throw "unimplemented";
  Future<String> getURL(String url) async => throw "unimplemented";
  Future<String> nextTime() async => throw "unimplemented";
  Future<void> writeStr(String s) async => throw "unimplemented";
  Stream<String> readStream() async* {
    throw "unimplemented";
  }

  // These are only implemented in android.
  Future<void> startForegroundSvc() => throw "unimplemented";
  Future<void> stopForegroundSvc() => throw "unimplemented";
  Future<void> setNtfnsEnabled(bool enabled) => throw "unimplemented";

  Future<dynamic> asyncCall(int cmd, dynamic payload) async =>
      throw "unimplemented";

  Future<String> asyncHello(String name) async {
    var r = await asyncCall(CTHello, name);
    return r as String;
  }

  Future<void> initClient(InitClient args) async {
    await asyncCall(CTInitClient, args);
  }
    Future<void> createLockFile(String rootDir) async =>
      await asyncCall(CTCreateLockFile, rootDir);
  Future<void> closeLockFile(String rootDir) async =>
      await asyncCall(CTCloseLockFile, rootDir);

}

const int CTUnknown = 0x00;
const int CTHello = 0x01;
const int CTInitClient = 0x02;
const int CTInvite = 0x03;
const int CTCreateLockFile = 0x04;
const int CTCloseLockFile = 0x05;

const int notificationsStartID = 0x1000;
// const int NTPM = 0x1004;