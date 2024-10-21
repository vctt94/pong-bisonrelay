package golib

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/companyzero/bisonrelay/client"
	"github.com/companyzero/bisonrelay/client/clientdb"
	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/client/rpcserver"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/lockfile"
	"github.com/companyzero/bisonrelay/rates"
	"github.com/companyzero/bisonrelay/rpc"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"golang.org/x/sync/errgroup"
)

const (
	appName = "bruig"
)

type clientCtx struct {
	c         *jsonrpc.WSClient
	lnpc      *client.DcrlnPaymentClient
	rpcserver *rpcserver.Server
	ctx       context.Context
	cancel    func()
	runMtx    sync.Mutex
	runErr    error

	log     slog.Logger
	logBknd *logBackend

	// skipWalletCheckChan is called if we should skip the next wallet
	// check.
	skipWalletCheckChan chan struct{}

	initIDChan   chan iDInit
	certConfChan chan bool

	// confirmPayReqRecvChan is written to by the user to confirm or deny
	// paying to open a chan.
	confirmPayReqRecvChan chan bool

	httpClient *http.Client
	rates      *rates.Rates

	// downloadConfChans tracks confirmation channels about downloads that
	// are about to be initiated.
	downloadConfMtx   sync.Mutex
	downloadConfChans map[zkidentity.ShortID]chan bool

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

func handleInitClient(handle uint32, args initClient) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmtx.Lock()
	defer cmtx.Unlock()
	if cs == nil {
		cs = make(map[uint32]*clientCtx)
	}
	if cs[handle] != nil {
		return errors.New("client already initialized")
	}

	flagURL := "wss://127.0.0.1:7878/ws" // WebSocket URL
	flagServerCertPath := "/home/vctt/.bruig/rpc.cert"
	flagClientCertPath := "/home/vctt/.bruig/rpc-client.cert"
	flagClientKeyPath := "/home/vctt/.bruig/rpc-client.key"

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelDebug)
	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(flagURL),
		jsonrpc.WithServerTLSCertPath(flagServerCertPath),
		jsonrpc.WithClientTLSCert(flagClientCertPath, flagClientKeyPath),
		jsonrpc.WithClientLog(log),
		jsonrpc.WithClientBasicAuth(args.RPCUser, args.RPCPass),
	)
	if err != nil {
		return err
	}
	g, gctx := errgroup.WithContext(ctx)
	// Run JSON-RPC client
	g.Go(func() error { return c.Run(gctx) })

	// Retrieve public identity via JSON-RPC
	chat := types.NewChatServiceClient(c)
	// version := types.NewVersionServiceClient(c)
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}
	// Initialize logging.
	logBknd, err := newLogBackend(args.LogFile, args.DebugLevel)
	if err != nil {
		return err
	}
	logBknd.notify = args.WantsLogNtfns

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	// ctx := context.Background()
	fmt.Printf("args: %+v\n\n", args)
	fmt.Printf("aqui**************\n clientId: %+v\n\n", clientID)
	// Initialize DB.
	db, err := clientdb.New(clientdb.Config{
		Root:          args.DBRoot,
		MsgsRoot:      args.MsgsRoot,
		DownloadsRoot: args.DownloadsDir,
		EmbedsRoot:    args.EmbedsDir,
		Logger:        logBknd.logger("FDDB"),
		ChunkSize:     rpc.MaxChunkSize,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize DB: %v", err)
	}
	// Prune embedded file cache.
	if err = db.PruneEmbeds(0); err != nil {
		return fmt.Errorf("unable to prune cache: %v", err)
	}

	// c, err = client.New(cfg)
	// if err != nil {
	// 	return err
	// }
	var cctx *clientCtx

	cctx = &clientCtx{
		c:      c,
		cancel: cancel,
	}

	go func() {
		// err := c.Run(ctx)
		// if errors.Is(err, context.Canceled) {
		// 	err = nil
		// }
		cctx.runMtx.Lock()
		cctx.runErr = err
		cctx.runMtx.Unlock()
		cmtx.Lock()
		delete(cs, handle)
		cmtx.Unlock()
		notify(NTClientStopped, nil, err)
	}()

	go func() {
		select {
		case <-ctx.Done():
			// case <-c.AddressBookLoaded():
			// 	notify(NTAddressBookLoaded, nil, nil)
		}
	}()

	return nil
}

func handleClientCmd(cc *clientCtx, cmd *cmd) (interface{}, error) {
	// c := cc.c
	// var lnc lnrpc.LightningClient
	// var lnWallet walletrpc.WalletKitClient
	// if cc.lnpc != nil {
	// 	lnc = cc.lnpc.LNRPC()
	// 	lnWallet = cc.lnpc.LNWallet()
	// }

	switch cmd.Type {

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
