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
	"runtime"
	"strings"

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
)

var (
	flagURL = flag.String("url", "wss://127.0.0.1:7676/ws", "URL of the websocket endpoint")

	flagServerCertPath = flag.String("servercert", "/home/pongbot/.brclient/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "/home/pongbot/.brclient/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "/home/pongbot/.brclient/rpc-client.key", "Path to rpc-client.key file")
)

type PongPlugin struct {
	id      string
	name    string
	version string
	config  map[string]interface{}
	logger  slog.Logger

	UpdatesCh  map[string]chan *pong.GameUpdateBytes
	PongClient map[string]pong.PongGame_InitClient
	Stream     map[string]grpctypes.PluginService_InitServer
}

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
	rpcplugin PongPlugin
	gameSrv   *server.GameServer
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
	// Set the stream before starting the goroutine
	s.rpcplugin.Stream[clientID] = stream
	// Listen for context cancellation to handle disconnection
	for range ctx.Done() {
		// s.handleDisconnect(clientID)
		fmt.Printf("client ctx disconnected")
		return ctx.Err()
	}

	return nil
}

func (s *pluginServer) CallAction(req *grpctypes.PluginCallActionStreamRequest, stream grpctypes.PluginService_CallActionServer) error {
	switch req.Action {
	case "ready":
		ctx := stream.Context()

		r := &pong.SignalReadyRequest{
			ClientId: req.User,
		}
		// Signal readiness after stream is initialized
		err := s.gameSrv.SignalReady(r, stream)
		if err != nil {
			return fmt.Errorf("error signaling readiness: %w", err)
		}

		log.Println("Stream initialized successfully")
		for range ctx.Done() {
			// XXX Handle disconnections
			// s.handleDisconnect(clientID)
			fmt.Printf("client ctx disconnected")
			return ctx.Err()
		}

	case "input":
		ctx := stream.Context()

		r := &pong.PlayerInput{
			PlayerId: req.User,
			Input:    string(req.Data),
		}
		_, err := s.gameSrv.SendInput(ctx, r)
		if err != nil {
			return fmt.Errorf("error sending input: %w", err)
		}

	default:
		return fmt.Errorf("unsupported action: %v", req.Action)
	}

	return nil
}

func (s *pluginServer) GetVersion(ctx context.Context, req *grpctypes.PluginVersionRequest) (*grpctypes.PluginVersionResponse, error) {
	// Implement your GetVersion logic here
	return &grpctypes.PluginVersionResponse{
		AppName:    s.rpcplugin.name,
		AppVersion: s.rpcplugin.version,
		GoRuntime:  runtime.Version(),
	}, nil
}

// NewPongPlugin initializes a new PongPlugin
func NewPongPlugin() PongPlugin {
	return PongPlugin{
		name:    "pong",
		version: "0.0.0",

		UpdatesCh:  make(map[string]chan *pong.GameUpdateBytes),
		PongClient: make(map[string]pong.PongGame_InitClient),
		Stream:     make(map[string]grpctypes.PluginService_InitServer),
	}
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
		return err
	}

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(*flagURL),
		jsonrpc.WithServerTLSCertPath(*flagServerCertPath),
		jsonrpc.WithClientTLSCert(*flagClientCertPath, *flagClientKeyPath),
		jsonrpc.WithClientLog(log),
	)
	if err != nil {
		return err
	}
	g.Go(func() error { return c.Run(gctx) })

	chat := types.NewChatServiceClient(c)
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return err
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	var zkShortID zkidentity.ShortID
	copy(zkShortID[:], clientID)
	plugin := NewPongPlugin()
	gameSrv := server.NewServer(&zkShortID, true)

	go func() error {
		if err := gameSrv.ManageGames(ctx); err != nil {
			return err
		}

		return nil
	}()

	srv := &pluginServer{
		rpcplugin: plugin,
		gameSrv:   gameSrv,
	}
	s := grpc.NewServer()
	grpctypes.RegisterPluginServiceServer(s, srv)

	fmt.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return err
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
