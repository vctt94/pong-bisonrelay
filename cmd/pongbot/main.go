package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

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
	certDir = flag.String("serverdir", "/home/pongbot/.pongserver", "Path to server dir")

	flagURL = flag.String("url", "wss://127.0.0.1:7676/ws", "URL of the websocket endpoint")

	flagServerCertPath = flag.String("servercert", "/home/pongbot/brclient/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "/home/pongbot/brclient/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "/home/pongbot/brclient/rpc-client.key", "Path to rpc-client.key file")
	debug              = flag.Bool("debug", false, "Enable debug mode")
)

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelDebug)

	g, gctx := errgroup.WithContext(ctx)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen on port 50051: %v", err)
	}

	// Paths for the generated certificate and key in .pongserver directory
	certPath := filepath.Join(*certDir, "server.cert")
	keyPath := filepath.Join(*certDir, "server.key")

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
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)

	srv := server.NewServer(&zkShortID, server.ServerConfig{Debug: *debug})
	go func() error {
		if err := srv.GameManager.Run(ctx); err != nil {
			return fmt.Errorf("failed to manage games: %v", err)
		}
		return nil
	}()

	// s := grpc.NewServer(grpc.Creds(creds))
	s := grpc.NewServer()
	pong.RegisterPongGameServer(s, srv)

	fmt.Printf("server listening at %v\n", lis.Addr())
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
