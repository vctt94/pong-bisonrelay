package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/botlib"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func realMain() error {
	cfg, err := botlib.LoadBotConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	debugLevel := server.GetDebugLevel(cfg.Debug)
	debugGameManager := server.GetDebugLevel(cfg.Debug)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("[Bot]")
	log.SetLevel(debugLevel)

	botlib.SetupSignalHandler(cancel, log)

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	g, gctx := errgroup.WithContext(ctx)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %v", err)
	}

	c, err := botlib.NewJSONRPCClient(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create JSON-RPC client: %w", err)
	}
	g.Go(func() error { return c.Run(gctx) })

	chat := types.NewChatServiceClient(c)
	payment := types.NewPaymentsServiceClient(c)

	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	if err := chat.UserPublicIdentity(ctx, req, &publicIdentity); err != nil {
		return fmt.Errorf("failed to get public identity: %w", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)

	srv := server.NewServer(&zkShortID, server.ServerConfig{
		DebugGameManagerLevel: debugGameManager,
		Debug:                 debugLevel,
		PaymentClient:         payment,
		ChatClient:            chat,
		ServerDir:             cfg.DataDir,
		HTTPPort:              "8888",
	})

	g.Go(func() error { return srv.Run(gctx) })
	g.Go(func() error { return srv.SendTipProgressLoop(gctx) })
	g.Go(func() error { return srv.ReceiveTipLoop(gctx) })

	certPath := filepath.Join(cfg.DataDir, "server.cert")
	keyPath := filepath.Join(cfg.DataDir, "server.key")
	if err := botlib.EnsureTLSCert(certPath, keyPath, cfg.GRPCHost); err != nil {
		return fmt.Errorf("failed to ensure TLS cert: %w", err)
	}

	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	pong.RegisterPongGameServer(grpcServer, srv)

	log.Infof("server listening at %v", lis.Addr())
	go func() {
		<-ctx.Done()
		log.Info("shutting down gRPC server gracefully")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %v", err)
	}

	return g.Wait()
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
