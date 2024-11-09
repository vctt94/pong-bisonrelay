package golib

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/lockfile"
	"github.com/companyzero/bisonrelay/rates"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/client"
	"golang.org/x/sync/errgroup"
)

const (
	appName = "pongui"
)

type clientCtx struct {
	ID      *localInfo
	c       *client.PongClient
	ctx     context.Context
	chat    types.ChatServiceClient
	payment types.PaymentsServiceClient
	cancel  func()
	runMtx  sync.Mutex
	runErr  error

	log          slog.Logger
	logBknd      *logBackend
	certConfChan chan bool

	httpClient *http.Client
	rates      *rates.Rates

	// expirationDays are the expirtation days provided by the server when
	// connected
	expirationDays uint64

	serverState atomic.Value
}

var (
	cmtx sync.Mutex
	cs   map[uint32]*clientCtx
	lfs  map[string]*lockfile.LockFile = map[string]*lockfile.LockFile{}

	// The following are debug vars.
	sigUrgCount       atomic.Uint64
	isServerConnected atomic.Bool
)

func handleHello(name string) (string, error) {
	if name == "*bug" {
		return "", fmt.Errorf("name '%s' is an error", name)
	}
	return "hello " + name, nil
}

func isClientRunning(handle uint32) bool {
	cmtx.Lock()
	var res bool
	if cs != nil {
		res = cs[handle] != nil
	}
	cmtx.Unlock()
	return res
}

func handleInitClient(handle uint32, args initClient) (*localInfo, error) {
	cmtx.Lock()
	defer cmtx.Unlock()
	if cs == nil {
		cs = make(map[uint32]*clientCtx)
	}
	if cs[handle] != nil {
		return cs[handle].ID, nil
	}

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelDebug)
	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(args.RPCWebsocketURL),
		jsonrpc.WithServerTLSCertPath(args.RPCCertPath),
		jsonrpc.WithClientTLSCert(args.RPCCLientCertPath, args.RPCCLientKeyPath),
		jsonrpc.WithClientLog(log),
		jsonrpc.WithClientBasicAuth(args.RPCUser, args.RPCPass),
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)
	// Run JSON-RPC client
	g.Go(func() error { return c.Run(gctx) })

	// Retrieve public identity via JSON-RPC
	chat := types.NewChatServiceClient(c)
	payment := types.NewPaymentsServiceClient(c)
	// Initialize clientID
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(gctx, &types.PublicIdentityReq{}, &publicIdentity)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get user public identity: %v", err)
	}
	var id zkidentity.ShortID
	id.FromBytes(publicIdentity.Identity[:])
	localInfo := &localInfo{
		ID:   id,
		Nick: publicIdentity.Nick,
	}
	// Initialize logging.
	logBknd, err := newLogBackend(args.LogFile, args.DebugLevel)
	if err != nil {
		cancel()
		return nil, err
	}
	logBknd.notify = args.WantsLogNtfns
	pc, err := client.NewPongClient(localInfo.ID.String(), &client.PongClientCfg{
		ServerAddr: args.ServerAddr,
		ChatClient: chat,
		Log:        logBknd.logger("client"),
	})
	if err != nil {
		cancel()
		return nil, err
	}
	cctx := &clientCtx{
		ID:      localInfo,
		ctx:     gctx,
		c:       pc,
		cancel:  cancel,
		log:     log,
		logBknd: logBknd,
	}
	cs[handle] = cctx
	go func() {
		// Run JSON-RPC client (it will block until the client is done)
		g.Go(func() error { return receiveLoop(gctx, chat, log) })
		g.Go(func() error { return receiveTipLoop(gctx, payment, log, cctx) })
		// Handle client closure and errors
		if err := g.Wait(); err != nil {
			fmt.Printf("err: %+v\n\n", err)
			cctx.runMtx.Lock()
			cctx.runErr = err
			cctx.runMtx.Unlock()

			// Clean up the client if it stops running
			cmtx.Lock()
			delete(cs, handle)
			cmtx.Unlock()

			// Notify the system that the client stopped
			notify(NTClientStopped, nil, err)
		}
	}()

	return localInfo, nil
}

func handleClientCmd(cc *clientCtx, cmd *cmd) (interface{}, error) {
	chat := cc.chat

	switch cmd.Type {
	case CTGetUserNick:
		resp := &types.UserNickResponse{}
		hexUid := string(cmd.Payload)
		err := chat.UserNick(cc.ctx, &types.UserNickRequest{
			HexUid: strings.Trim(hexUid, `"`),
		}, resp)
		if err != nil {
			return nil, err
		}
		return resp.Nick, nil
	case CTGetWRPlayers:
		wrp, err := cc.c.GetWRPlayers()
		if err != nil {
			return nil, err
		}
		res := make([]*player, len(wrp))
		for i, p := range wrp {
			res[i], err = playerFromServer(p)
			if err != nil {
				return nil, err
			}
		}
		return res, nil
	case CTGetWaitingRooms:
		rooms, err := cc.c.GetWaitingRooms()
		if err != nil {
			return nil, err
		}
		res := make([]*waitingRoom, len(rooms))
		for i, r := range rooms {
			players := make([]*player, len(r.Players))
			for i, p := range r.Players {
				var id zkidentity.ShortID
				err := id.FromString(p.Uid)
				if err != nil {
					return nil, err
				}

				players[i], err = playerFromServer(p)
				if err != nil {
					return nil, err
				}
			}
			res[i] = &waitingRoom{
				ID:     r.Id,
				HostID: r.HostId,
				BetAmt: r.BetAmt,
			}
		}
		return res, nil

	case CTStopClient:
		cc.cancel()
		return nil, nil
	}
	return nil, nil

}

func handleCreateLockFile(rootDir string) error {
	filePath := filepath.Join(rootDir, clientintf.LockFileName)

	cmtx.Lock()
	defer cmtx.Unlock()

	lf := lfs[filePath]
	if lf != nil {
		// Already running on this DB from this process.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	lf, err := lockfile.Create(ctx, filePath)
	cancel()
	if err != nil {
		return fmt.Errorf("unable to create lockfile %q: %v", filePath, err)
	}
	lfs[filePath] = lf
	return nil
}

func handleCloseLockFile(rootDir string) error {
	filePath := filepath.Join(rootDir, clientintf.LockFileName)

	cmtx.Lock()
	lf := lfs[filePath]
	delete(lfs, filePath)
	cmtx.Unlock()

	if lf == nil {
		return fmt.Errorf("nil lockfile")
	}
	return lf.Close()
}
