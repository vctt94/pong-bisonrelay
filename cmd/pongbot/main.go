package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	serverDir = flag.String("serverdir", "/home/pongbot/.pongserver", "Path to server dir")

	flagURL = flag.String("url", "wss://127.0.0.1:7676/ws", "URL of the websocket endpoint")

	flagServerCertPath  = flag.String("servercert", "/home/pongbot/.brclient/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath  = flag.String("clientcert", "/home/pongbot/.brclient/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath   = flag.String("clientkey", "/home/pongbot/.brclient/rpc-client.key", "Path to rpc-client.key file")
	debugStr            = flag.String("debug", "debug", "Enable debug mode")
	debugGameManagerStr = flag.String("debuggamemanager", "debug", "Enable debug mode for game manager")
)

func realMain() error {
	debugLevel := server.GetDebugLevel(*debugStr)
	debugGameManager := server.GetDebugLevel(*debugGameManagerStr)

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

	g, gctx := errgroup.WithContext(ctx)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen on port 50051: %v", err)
	}

	// Paths for the generated certificate and key in .pongserver directory
	certPath := filepath.Join(*serverDir, "server.cert")
	keyPath := filepath.Join(*serverDir, "server.key")

	// Initialize the Pong plugin and GameServer

	// Check if the TLS certificate and key exist; if not, generate them
	if _, err := os.Stat(certPath); os.IsNotExist(err) || func() bool {
		_, err := os.Stat(keyPath)
		return os.IsNotExist(err)
	}() {
		// if err := gameSrv.GenerateNewTLSCertPair("Pong Server", time.Now().Add(365*24*time.Hour), []string{"localhost"}, certPath, keyPath); err != nil {
		// 	return fmt.Errorf("failed to generate self-signed certificate: %v", err)
		// }
		// fmt.Println("Generated new self-signed certificate and key")
	}

	// Load the server certificate and key
	// creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to load TLS credentials: %v", err)
	// }

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(*flagURL),
		jsonrpc.WithServerTLSCertPath(*flagServerCertPath),
		jsonrpc.WithClientTLSCert(*flagClientCertPath, *flagClientKeyPath),
		jsonrpc.WithClientLog(log),
		jsonrpc.WithClientBasicAuth("rpcuser", "rpcpass"),
	)
	if err != nil {
		return fmt.Errorf("failed to create WS client: %v", err)
	}
	g.Go(func() error { return c.Run(gctx) })

	chat := types.NewChatServiceClient(c)
	payment := types.NewPaymentsServiceClient(c)
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)

	srv := server.NewServer(&zkShortID, server.ServerConfig{
		DebugGameManagerLevel: debugGameManager,
		Debug:                 debugLevel,
		PaymentClient:         payment,
		ChatClient:            chat,
		ServerDir:             *serverDir,
		HTTPPort:              "8888",
	})

	// Run server
	g.Go(func() error { return srv.Run(gctx) })
	g.Go(func() error { return srv.SendTipProgressLoop(gctx) })
	g.Go(func() error { return srv.ReceiveTipLoop(gctx) })

	s := grpc.NewServer()
	pong.RegisterPongGameServer(s, srv)

	log.Infof("server listening at %v", lis.Addr())
	go func() {
		<-ctx.Done()
		log.Info("shutting down gRPC server gracefully")
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
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
