package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/companyzero/bisonrelay/clientplugin/grpctypes"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/jrick/logrotate/rotator"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	certDir = flag.String("serverdir", "/home/pongbot/.pongserver", "Path to server dir")

	flagURL = flag.String("url", "wss://127.0.0.1:7676/ws", "URL of the websocket endpoint")

	flagServerCertPath = flag.String("servercert", "/home/pongbot/.brclient/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "/home/pongbot/.brclient/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "/home/pongbot/.brclient/rpc-client.key", "Path to rpc-client.key file")
)

func (s *pluginServer) Render(ctx context.Context, data *grpctypes.RenderRequest) (*grpctypes.RenderResponse, error) {
	gameState := pong.GameUpdate{}
	if err := json.Unmarshal(data.Data, &gameState); err != nil {
		return nil, fmt.Errorf("failed to decode game state: %v", err)
	}

	var gameView strings.Builder
	for y := 0; y < int(gameState.GameHeight); y++ {
		for x := 0; x < int(gameState.GameWidth); x++ {
			ballX := int(gameState.BallX)
			ballY := int(gameState.BallY)
			switch {
			case x == ballX && y == ballY:
				gameView.WriteString("O")
			case x == 0 && y >= int(gameState.P1Y) && y < int(gameState.P1Y)+int(gameState.P1Height):
				gameView.WriteString("|")
			case x == int(gameState.GameWidth)-1 && y >= int(gameState.P2Y) && y < int(gameState.P2Y)+int(gameState.P2Height):
				gameView.WriteString("|")
			default:
				gameView.WriteString(" ")
			}
		}
		gameView.WriteString("\n")
	}
	gameView.WriteString(fmt.Sprintf("Score: %d - %d\n", gameState.P1Score, gameState.P2Score))
	gameView.WriteString("Controls: W/S and Arrow Keys - Q to quit")

	return &grpctypes.RenderResponse{
		Data: gameView.String(),
	}, nil
}

type pluginServer struct {
	grpctypes.UnimplementedPluginServiceServer
	gameSrv *server.GameServer
}

func (s *pluginServer) Init(req *grpctypes.PluginStartStreamRequest, stream grpctypes.PluginService_InitServer) error {
	ctx := stream.Context()
	clientID := req.ClientId

	fmt.Printf("Init called by client: %+v\n", clientID)

	in := &pong.GameStartedStreamRequest{
		ClientId: clientID,
	}
	err := s.gameSrv.Init(in, stream)
	if err != nil {
		return err
	}
	// Listen for context cancellation to handle disconnection
	for range ctx.Done() {
		// s.handleDisconnect(clientID)
		fmt.Printf("client ctx disconnected")
		return ctx.Err()
	}

	return nil
}

func (s *pluginServer) CallAction(req *grpctypes.PluginCallActionStreamRequest, stream grpctypes.PluginService_CallActionServer) error {
	if req.Action != "ready" {
		return fmt.Errorf("unsupported action: %v", req.Action)
	}

	ctx := stream.Context()

	r := &pong.SignalReadyRequest{
		ClientId: req.User,
	}

	// Signal readiness and start the game update stream
	err := s.gameSrv.SignalReady(r, stream)
	if err != nil {
		return fmt.Errorf("error signaling readiness: %w", err)
	}

	log.Println("Stream initialized successfully")

	// Keep the stream open until the context is done
	<-ctx.Done()
	fmt.Printf("Client %s disconnected\n", req.User)
	return ctx.Err()
}

func (s *pluginServer) SendInput(ctx context.Context, req *grpctypes.PluginInputRequest) (*grpctypes.PluginInputResponse, error) {
	// Process the input
	r := &pong.PlayerInput{
		PlayerId: req.User,
		Input:    string(req.Data),
	}

	// Send the input to the game server
	_, err := s.gameSrv.SendInput(ctx, r)
	if err != nil {
		return &grpctypes.PluginInputResponse{
			Success: false,
			Message: fmt.Sprintf("error sending input: %v", err),
		}, nil
	}

	return &grpctypes.PluginInputResponse{
		Success: true,
		Message: "Input processed successfully",
	}, nil
}

func (s *pluginServer) GetVersion(ctx context.Context, req *grpctypes.PluginVersionRequest) (*grpctypes.PluginVersionResponse, error) {
	// Implement your GetVersion logic here
	return s.gameSrv.GetVersion(ctx, req), nil
}

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
	var zkShortID zkidentity.ShortID              // Assuming this is initialized correctly in your full code
	gameSrv := server.NewServer(&zkShortID, true) // Initialize the GameServer here

	// Check if the TLS certificate and key exist; if not, generate them
	if _, err := os.Stat(certPath); os.IsNotExist(err) || func() bool {
		_, err := os.Stat(keyPath)
		return os.IsNotExist(err)
	}() {
		if err := gameSrv.GenerateNewTLSCertPair("Pong Server", time.Now().Add(365*24*time.Hour), []string{"localhost"}, certPath, keyPath); err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %v", err)
		}
		fmt.Println("Generated new self-signed certificate and key")
	}

	// Load the server certificate and key
	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %v", err)
	}

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(*flagURL),
		jsonrpc.WithServerTLSCertPath(*flagServerCertPath),
		jsonrpc.WithClientTLSCert(*flagClientCertPath, *flagClientKeyPath),
		jsonrpc.WithClientLog(log),
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
	copy(zkShortID[:], clientID)

	go func() error {
		if err := gameSrv.ManageGames(ctx); err != nil {
			return fmt.Errorf("failed to manage games: %v", err)
		}
		return nil
	}()

	s := grpc.NewServer(grpc.Creds(creds))
	grpctypes.RegisterPluginServiceServer(s, &pluginServer{
		gameSrv: gameSrv,
	})

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

type logWriter struct {
	r *rotator.Rotator
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return l.r.Write(p)
}
