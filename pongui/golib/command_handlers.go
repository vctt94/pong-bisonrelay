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

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/lockfile"
	"github.com/companyzero/bisonrelay/rates"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"golang.org/x/sync/errgroup"
)

const (
	appName = "bruig"
)

type clientCtx struct {
	c      *jsonrpc.WSClient
	cancel func()
	runMtx sync.Mutex
	runErr error

	log          slog.Logger
	logBknd      *logBackend
	initIDChan   chan iDInit
	certConfChan chan bool

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

	cmtx.Lock()
	defer cmtx.Unlock()
	if cs == nil {
		cs = make(map[uint32]*clientCtx)
	}
	if cs[handle] != nil {
		return errors.New("client already initialized")
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
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	fmt.Printf("args: %+v\n\n", args)
	g, gctx := errgroup.WithContext(ctx)
	// Run JSON-RPC client
	g.Go(func() error { return c.Run(gctx) })

	// Retrieve public identity via JSON-RPC
	chat := types.NewChatServiceClient(c)
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(gctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}
	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	// ctx := context.Background()
	fmt.Printf("aqui**************\n clientId: %+v\n\n", clientID)
	// Initialize logging.
	logBknd, err := newLogBackend(args.LogFile, args.DebugLevel)
	if err != nil {
		return err
	}
	logBknd.notify = args.WantsLogNtfns

	var cctx *clientCtx

	cctx = &clientCtx{
		c:      c,
		cancel: cancel,
	}

	go func() {
		// Run JSON-RPC client (it will block until the client is done)
		g.Go(func() error { return receiveLoop(gctx, chat, log) })

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
