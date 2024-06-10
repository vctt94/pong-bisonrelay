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
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	flagURL            = flag.String("url", "wss://127.0.0.1:7777/ws", "URL of the websocket endpoint")
	flagServerCertPath = flag.String("servercert", "/home/pokerbot/brclient/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "/home/pokerbot/brclient/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "/home/pokerbot/brclient/rpc-client.key", "Path to rpc-client.key file")
)

var (
	serverLogger = log.New(os.Stdout, "[SERVER] ", 0)
	debug        = flag.Bool("debug", false, "")
	fps          = flag.Uint("fps", canvas.DEFAULT_FPS, "")
)

type server struct {
	pong.UnimplementedPongGameServer
	ID             string
	mu             sync.Mutex
	clientReady    chan string
	games          map[string]*gameInstance
	waitingRoom    *WaitingRoom
	playerSessions *PlayerSessions
	paymentService types.PaymentsServiceClient
}

type GameStartNotification struct {
	GameID  string
	Players []*Player
}

type gameInstance struct {
	engine      *canvas.CanvasEngine
	framesch    chan []byte
	inputch     chan []byte
	roundResult chan int32
	players     []*Player
	cleanedUp   bool
	running     bool
}

func (s *server) SendInput(ctx context.Context, in *pong.PlayerInput) (*pong.GameUpdate, error) {
	clientID, err := getClientIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	gameInstance, player, exists := s.findGameInstanceAndPlayerByClientID(clientID)
	if !exists {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	in.PlayerNumber = player.PlayerNumber
	inputBytes, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %v", err)
	}
	gameInstance.inputch <- inputBytes

	return &pong.GameUpdate{}, nil
}

func (s *server) StreamUpdates(req *pong.GameStreamRequest, stream pong.PongGame_StreamUpdatesServer) error {
	ctx := stream.Context()
	clientID, err := getClientIDFromContext(ctx)
	if err != nil {
		return err
	}

	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}
	player.stream = stream
	s.playerSessions.AddOrUpdatePlayer(player)
	serverLogger.Printf("Player %s stream initialized in StreamUpdates", clientID)

	s.waitingRoom.AddPlayer(player)
	s.clientReady <- clientID

	gameInstance, _, exists := s.findGameInstanceAndPlayerByClientID(clientID)
	if !exists {
		return fmt.Errorf("no game instance found for client ID %s", clientID)
	}

	for {
		select {
		case <-ctx.Done():
			s.handleDisconnect(clientID)
			return ctx.Err()
		case frame, ok := <-gameInstance.framesch:
			if !ok {
				return nil
			}
			if err := stream.Send(&pong.GameUpdateBytes{Data: frame}); err != nil {
				s.handleDisconnect(clientID)
				return err
			}
		}
	}
}

func (s *server) cleanupGameInstance(instance *gameInstance) {
	if !instance.cleanedUp {
		instance.cleanedUp = true
		close(instance.framesch)
		close(instance.inputch)
		close(instance.roundResult)
	}

	for gameID, game := range s.games {
		if game == instance {
			delete(s.games, gameID)
			serverLogger.Printf("[SERVER] Game %s cleaned up", gameID)
			break
		}
	}
}

func (s *server) handleDisconnect(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var instanceToCleanup *gameInstance
	var remainingPlayer *Player

	for _, game := range s.games {
		for i, player := range game.players {
			if player.ID == clientID {
				// Remove the player from the game
				game.players = append(game.players[:i], game.players[i+1:]...)
				if len(game.players) == 0 {
					instanceToCleanup = game
				} else {
					game.running = false
					remainingPlayer = game.players[0]
				}
				break
			}
		}
	}

	if instanceToCleanup != nil {
		s.cleanupGameInstance(instanceToCleanup)
	} else if remainingPlayer != nil {
		// Notify the remaining player about the disconnection
		if remainingPlayer.stream != nil {
			remainingPlayer.startNotifier.Send(&pong.GameStartedStreamResponse{
				Message: "Opponent disconnected. Game over.",
				Started: false,
			})
		}
	}

	// Remove player session
	s.playerSessions.RemovePlayer(clientID)
	serverLogger.Printf("Player %s disconnected and cleaned up", clientID)
}

func (s *server) SignalReady(ctx context.Context, req *pong.SignalReadyRequest) (*pong.SignalReadyResponse, error) {
	clientID, err := getClientIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	serverLogger.Printf("SignalReady called by client ID: %s", clientID)

	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		return &pong.SignalReadyResponse{}, fmt.Errorf("player not found for client ID %s", player.ID)
	}

	s.waitingRoom.AddPlayer(player)
	s.clientReady <- clientID

	serverLogger.Printf("Player %s added to waiting room. Current ready players: %v", clientID, s.waitingRoom.queue)

	return &pong.SignalReadyResponse{}, nil
}

func (s *server) findGameInstanceAndPlayerByClientID(clientID string) (*gameInstance, *Player, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, game := range s.games {
		for _, player := range game.players {
			if player.ID == clientID {
				return game, player, true
			}
		}
	}
	return nil, nil, false
}

func newServer(id string) *server {
	return &server{
		ID:             id,
		clientReady:    make(chan string, 10),
		games:          make(map[string]*gameInstance),
		waitingRoom:    NewWaitingRoom(),
		playerSessions: NewPlayerSessions(),
	}
}

func (s *server) manageGames(ctx context.Context) {
	for {
		select {
		case clientID := <-s.clientReady:
			serverLogger.Printf("Received client ready signal for client ID: %s", clientID)
			if players, ready := s.waitingRoom.ReadyPlayers(); ready {
				serverLogger.Printf("Starting game with players: %v and %v", players[0].ID, players[1].ID)
				s.startGame(ctx, players)
			} else {
				serverLogger.Printf("Not enough players ready. Current ready players: %v", s.waitingRoom.queue)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) StartNotifier(req *pong.GameStartedStreamRequest, stream pong.PongGame_StartNotifierServer) error {
	ctx := stream.Context()
	clientID, err := getClientIDFromContext(ctx)
	if err != nil {
		return err
	}

	serverLogger.Printf("StartNotifier called by client ID: %s", clientID)

	var player *Player
	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		player = NewPlayer(clientID, nil)
		s.playerSessions.AddOrUpdatePlayer(player)
		serverLogger.Printf("Player %s registered in StartNotifier", clientID)
	}
	player.startNotifier = stream

	for {
		select {
		case <-ctx.Done():
			s.handleDisconnect(clientID)
			return ctx.Err()
		}
	}
}

func (s *server) startGame(ctx context.Context, players []*Player) {
	gameID := generateGameID()
	serverLogger.Printf("Starting new game with ID: %s", gameID)

	newGameInstance := s.startNewGame(ctx)
	players[0].PlayerNumber = 1
	players[1].PlayerNumber = 2
	newGameInstance.players = players

	s.mu.Lock()
	s.games[gameID] = newGameInstance
	s.mu.Unlock()

	for _, player := range players {
		serverLogger.Printf("Notifying player %s that game %s started", player.ID, gameID)
		if player.startNotifier == nil {
			serverLogger.Panic("startNotifier nil")
		}
		if player.startNotifier != nil {
			if err := player.startNotifier.Send(&pong.GameStartedStreamResponse{Message: "Game has started with ID: " + gameID, Started: true, PlayerNumber: player.PlayerNumber}); err != nil {
				serverLogger.Printf("Failed to send game start notification to player %s: %v", player.ID, err)
			}
		}
	}
}

func (s *server) startNewGame(ctx context.Context) *gameInstance {
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(3, 3),
	)

	canvasEngine := canvas.New(game)
	canvasEngine.SetDebug(*debug).SetFPS(*fps)

	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 10)
	roundResult := make(chan int32)
	instance := &gameInstance{
		engine:      canvasEngine,
		framesch:    framesch,
		inputch:     inputch,
		roundResult: roundResult,
		running:     true,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				serverLogger.Printf("Recovered from panic in NewRound: %v", r)
			}
		}()
		if !instance.running {
			return
		}
		canvasEngine.NewRound(ctx, instance.framesch, instance.inputch, instance.roundResult)
	}()

	go func() {
		for winnerID := range roundResult {
			if !instance.running {
				return
			}
			s.handleRoundResult(winnerID, instance)
		}
	}()

	return instance
}

func (s *server) handleRoundResult(playerNumber int32, instance *gameInstance) {
	var winner *Player
	for _, player := range instance.players {
		if player.PlayerNumber == playerNumber {
			winner = player
			break
		}
	}

	if winner == nil {
		serverLogger.Printf("Winner not found in game instance")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &types.TipUserRequest{
		User:        winner.ID,
		DcrAmount:   0.00000001,
		MaxAttempts: 3,
	}
	var res types.TipUserResponse
	err := s.paymentService.TipUser(ctx, req, &res)
	if err != nil {
		serverLogger.Printf("Failed to send payment to winner %s: %v", winner.ID, err)
		return
	}

	serverLogger.Printf("Successfully sent payment to winner %s", winner.ID)
}

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
	srv.paymentService = paymentService

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
