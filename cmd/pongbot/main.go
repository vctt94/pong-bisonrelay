package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func realMain() error {
	cfg, err := loadConfig()
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

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Infof("Received shutdown signal")
		cancel()
	}()

	// Create DataDir directory if not exists
	if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %v", err)
		}
	}

	g, gctx := errgroup.WithContext(ctx)

	// Start gRPC server
	// XXX Port can also come from config
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen on port 50051: %v", err)
	}

	// Initialize JSON-RPC client
	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(cfg.URL),
		jsonrpc.WithServerTLSCertPath(cfg.ServerCertPath),
		jsonrpc.WithClientTLSCert(cfg.ClientCertPath, cfg.ClientKeyPath),
		jsonrpc.WithClientLog(log),
		jsonrpc.WithClientBasicAuth(cfg.RPCUser, cfg.RPCPass),
	)
	if err != nil {
		return fmt.Errorf("failed to create WS client: %w", err)
	}
	g.Go(func() error { return c.Run(gctx) })

	// Chat and Payment clients
	chat := types.NewChatServiceClient(c)
	payment := types.NewPaymentsServiceClient(c)

	// Retrieve public identity
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	if err := chat.UserPublicIdentity(ctx, req, &publicIdentity); err != nil {
		return fmt.Errorf("failed to get public identity: %w", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)

	// Create Pong server
	srv := server.NewServer(&zkShortID, server.ServerConfig{
		DebugGameManagerLevel: debugGameManager,
		Debug:                 debugLevel,
		PaymentClient:         payment,
		ChatClient:            chat,
		ServerDir:             cfg.DataDir,
		HTTPPort:              "8888",
	})

	// Run server
	g.Go(func() error { return srv.Run(gctx) })
	g.Go(func() error { return srv.SendTipProgressLoop(gctx) })
	g.Go(func() error { return srv.ReceiveTipLoop(gctx) })

	// load tls cert
	certPath := filepath.Join(cfg.DataDir, "server.cert")
	keyPath := filepath.Join(cfg.DataDir, "server.key")
	// Check if the TLS certificate and key exist; if not, generate them
	if _, err := os.Stat(certPath); os.IsNotExist(err) || func() bool {
		_, err := os.Stat(keyPath)
		return os.IsNotExist(err)
	}() {
		if err := srv.GenerateNewTLSCertPair("Pong Server", time.Now().Add(365*24*time.Hour), []string{"localhost"}, certPath, keyPath); err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %v", err)
		}
		fmt.Println("Generated new self-signed certificate and key")
	}

	// Load the server certificate and key
	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %v", err)
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
