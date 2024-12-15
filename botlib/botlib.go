package botlib

import (
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/decred/slog"
)

func NewJSONRPCClient(cfg *BotConfig, log slog.Logger) (*jsonrpc.WSClient, error) {
	return jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(cfg.RPCURL),
		jsonrpc.WithServerTLSCertPath(cfg.ServerCertPath),
		jsonrpc.WithClientTLSCert(cfg.ClientCertPath, cfg.ClientKeyPath),
		jsonrpc.WithClientLog(log),
		jsonrpc.WithClientBasicAuth(cfg.RPCUser, cfg.RPCPass),
	)
}
