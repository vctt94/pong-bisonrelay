package golib

import (
	"encoding/json"

	"github.com/companyzero/bisonrelay/client"
	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/client/resources/simplestore"
	"github.com/companyzero/bisonrelay/rpc"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
)

type initClient struct {
	ServerAddr     string `json:"server_addr"`
	DBRoot         string `json:"dbroot"`
	DownloadsDir   string `json:"downloads_dir"`
	LogFile        string `json:"log_file"`
	DebugLevel     string `json:"debug_level"`
	WantsLogNtfns  bool   `json:"wants_log_ntfns"`
	LogPings       bool   `json:"log_pings"`
	PingIntervalMs int64  `json:"ping_interval_ms"`

	// New fields for RPC configuration
	RPCWebsocketURL   string `json:"rpc_websocket_url"`
	RPCCertPath       string `json:"rpc_cert_path"`
	RPCCLientCertPath string `json:"rpc_client_cert_path"`
	RPCCLientKeyPath  string `json:"rpc_client_key_path"`
	RPCUser           string `json:"rpc_user"`
	RPCPass           string `json:"rpc_pass"`
}

type iDInit struct {
	Nick string `json:"nick"`
	Name string `json:"name"`
}

type localInfo struct {
	ID   clientintf.UserID `json:"id"`
	Nick string            `json:"nick"`
}

type serverCert struct {
	InnerFingerprint string `json:"inner_fingerprint"`
	OuterFingerprint string `json:"outer_fingerprint"`
}

const (
	ConnStateOffline        = 0
	ConnStateCheckingWallet = 1
	ConnStateOnline         = 2
)

type serverSessState struct {
	State          int     `json:"state"`
	CheckWalletErr *string `json:"check_wallet_err"`
}

type pm struct {
	UID       clientintf.UserID `json:"sid"` // sid == source id
	Msg       string            `json:"msg"`
	Mine      bool              `json:"mine"`
	TimeStamp int64             `json:"timestamp"`
	Nick      string            `json:"nick"`
}

type remoteUser struct {
	UID  string `json:"uid"`
	Nick string `json:"nick"`
}

func remoteUserFromPII(pii *zkidentity.PublicIdentity) remoteUser {
	return remoteUser{
		UID:  pii.Identity.String(),
		Nick: pii.Nick,
	}
}

func remoteUserFromRU(ru *client.RemoteUser) remoteUser {
	if ru == nil {
		return remoteUser{}
	}
	return remoteUser{
		UID:  ru.ID().String(),
		Nick: ru.Nick(),
	}
}

type payTipArgs struct {
	UID    clientintf.UserID `json:"uid"`
	Amount float64           `json:"amount"`
}

type account struct {
	Name               string         `json:"name"`
	UnconfirmedBalance dcrutil.Amount `json:"unconfirmed_balance"`
	ConfirmedBalance   dcrutil.Amount `json:"confirmed_balance"`
	InternalKeyCount   uint32         `json:"internal_key_count"`
	ExternalKeyCount   uint32         `json:"external_key_count"`
}

type sendOnChain struct {
	Addr        string         `json:"addr"`
	Amount      dcrutil.Amount `json:"amount"`
	FromAccount string         `json:"from_account"`
}

type writeInvite struct {
	FundAmount  dcrutil.Amount      `json:"fund_amount"`
	FundAccount string              `json:"fund_account"`
	GCID        *zkidentity.ShortID `json:"gc_id"`
	Prepaid     bool                `json:"prepaid"`
}

type generatedKXInvite struct {
	Blob  []byte                    `json:"blob"`
	Funds *rpc.InviteFunds          `json:"funds"`
	Key   *clientintf.PaidInviteKey `json:"key"`
}

type redeemedInviteFunds struct {
	Txid  rpc.TxHash     `json:"txid"`
	Total dcrutil.Amount `json:"total"`
}

type invitation struct {
	Blob   []byte                      `json:"blob"`
	Invite rpc.OOBPublicIdentityInvite `json:"invite"`
}

type fetchResourceArgs struct {
	UID        clientintf.UserID         `json:"uid"`
	Path       []string                  `json:"path"`
	Metadata   map[string]string         `json:"metadata,omitempty"`
	SessionID  clientintf.PagesSessionID `json:"session_id"`
	ParentPage clientintf.PagesSessionID `json:"parent_page"`
	Data       json.RawMessage           `json:"data"`
}

type simpleStoreOrder struct {
	Order simplestore.Order `json:"order"`
	Msg   string            `json:"msg"`
}

type handshakeStage struct {
	UID   clientintf.UserID `json:"uid"`
	Stage string            `json:"stage"`
}

type loadUserHistory struct {
	UID     clientintf.UserID `json:"uid"`
	IsGC    bool              `json:"is_gc"`
	Page    int               `json:"page"`
	PageNum int               `json:"page_num"`
}

type transReset struct {
	Mediator zkidentity.ShortID `json:"mediator"`
	Target   zkidentity.ShortID `json:"target"`
}

type listTransactions struct {
	StartHeight int32 `json:"start_height"`
	EndHeight   int32 `json:"end_height"`
}

type transaction struct {
	TxHash      string `json:"tx_hash"`
	Amount      int64  `json:"amount"`
	BlockHeight int32  `json:"block_height"`
}

type postAndCommentID struct {
	PostID    clientintf.PostID `json:"post_id"`
	CommentID clientintf.ID     `json:"comment_id"`
}

type runState struct {
	DcrlndRunning bool `json:"dcrlnd_running"`
	ClientRunning bool `json:"client_running"`
}

type zipLogsArgs struct {
	IncludeGolib bool   `json:"include_golib"`
	IncludeLn    bool   `json:"include_ln"`
	OnlyLastFile bool   `json:"only_last_file"`
	DestPath     string `json:"dest_path"`
}

type uiNotificationsConfig struct {
	PMs        bool `json:"pms"`
	GCMs       bool `json:"gcms"`
	GCMentions bool `json:"gcmentions"`
}
