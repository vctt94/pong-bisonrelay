package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/slog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

func realMain() error {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)
	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelInfo)

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(*flagURL),
		jsonrpc.WithServerTLSCertPath(*flagServerCertPath),
		jsonrpc.WithClientTLSCert(*flagClientCertPath, *flagClientKeyPath),
		jsonrpc.WithClientLog(log),
	)
	if err != nil {
		return err
	}

	chatClient := types.NewChatServiceClient(c)
	paymentService := types.NewPaymentsServiceClient(c)

	var clientID string
	g.Go(func() error { return c.Run(gctx) })
	g.Go(func() error { return receivePaymentLoop(gctx, paymentService, log) })

	resp := &types.PublicIdentity{}
	err = chatClient.UserPublicIdentity(ctx, &types.PublicIdentityReq{}, resp)
	if err != nil {
		return fmt.Errorf("failed to get public identity: %w", err)
	}

	clientID = hex.EncodeToString(resp.Identity[:])

	if clientID == "" {
		return fmt.Errorf("client ID is empty after fetching")
	}
	srv := newServer(clientID)

	go srv.manageGames(ctx)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Errorf("failed to listen: %v", err)
		return err
	}
	grpcServer := grpc.NewServer()
	pong.RegisterPongGameServer(grpcServer, srv)
	fmt.Println("server listening at", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorf("failed to serve: %v", err)
		return err
	}
	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
