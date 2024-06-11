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
	"github.com/jrick/logrotate/rotator"
	"github.com/vctt94/pong-bisonrelay/bot"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server"
	"google.golang.org/grpc"
)

const ()

func realMain() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Setup logging
	logDir := filepath.Join(cfg.DataDir, "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, "bot.log")
	logFd, err := rotator.New(logPath, 32*1024, true, 0)
	if err != nil {
		return err
	}
	defer logFd.Close()

	logBknd := slog.NewBackend(&logWriter{logFd}, slog.WithFlags(slog.LUTC))
	botLog := logBknd.Logger("BOT")
	pmLog := logBknd.Logger("PM")

	pmLog.SetLevel(slog.LevelDebug)
	botLog.SetLevel(slog.LevelDebug)

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("BRLY")
	log.SetLevel(slog.LevelDebug)

	pmChan := make(chan types.ReceivedPM)

	botCfg := bot.Config{
		DataDir: cfg.DataDir,
		Log:     botLog,

		URL:            cfg.URL,
		ServerCertPath: cfg.ServerCertPath,
		ClientCertPath: cfg.ClientCertPath,
		ClientKeyPath:  cfg.ClientKeyPath,

		PMChan: pmChan,
		PMLog:  pmLog,
	}

	bot, err := bot.New(botCfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	bot.GetPublicIdentity(ctx, req, &publicIdentity)

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)

	srv := server.NewServer(&zkShortID)

	grpcServer := grpc.NewServer()
	pong.RegisterPongGameServer(grpcServer, srv)
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Errorf("failed to listen: %v", err)
		return err
	}
	fmt.Println("server listening at", lis.Addr())

	// Run the gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Errorf("failed to serve: %v", err)
		}
	}()

	bot.RegisterGameServer(srv)

	// Launch handler
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pm := <-pmChan:
				nick := escapeNick(pm.Nick)
				if pm.Msg == nil {
					pmLog.Tracef("empty message from %v", nick)
					continue
				}
			}
		}
	}()

	return bot.Run()
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type logWriter struct {
	r *rotator.Rotator
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return l.r.Write(p)
}
