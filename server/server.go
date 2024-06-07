package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"

	canvas "github.com/vctt94/pong-bisonrelay/pong"

	"github.com/ndabAP/ping-pong/engine"
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
}

type GameStartNotification struct {
	GameID  string
	Players []*Player
}

type gameInstance struct {
	engine   *canvas.CanvasEngine
	framesch chan []byte
	inputch  chan []byte
	players  []*Player
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

	s.mu.Lock()
	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		player = NewPlayer(clientID, stream)
		s.playerSessions.AddOrUpdatePlayer(player)
		serverLogger.Printf("Player %s registered and stream initialized in StreamUpdates", clientID)
	} else {
		player.stream = stream
		s.playerSessions.AddOrUpdatePlayer(player)
		serverLogger.Printf("Player %s stream initialized in StreamUpdates", clientID)
	}
	s.mu.Unlock()

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

func (s *server) handleDisconnect(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for gameID, game := range s.games {
		for i, player := range game.players {
			if player.ID == clientID {
				game.players = append(game.players[:i], game.players[i+1:]...)
				if len(game.players) == 0 {
					close(game.framesch)
					close(game.inputch)
					delete(s.games, gameID)
					log.Printf("[SERVER] Game %s cleaned up after player %s disconnected", gameID, clientID)
				}
				return
			}
		}
	}
}

func (s *server) SignalReady(ctx context.Context, req *pong.SignalReadyRequest) (*pong.SignalReadyResponse, error) {
	clientID := req.ClientId
	serverLogger.Printf("SignalReady called by client ID: %s", clientID)

	s.mu.Lock()
	player, exists := s.playerSessions.GetPlayer(clientID)
	s.mu.Unlock()

	if !exists {
		// serverLogger.Panic("player not found for client ID")
		player = NewPlayer(clientID, nil)
		s.playerSessions.AddOrUpdatePlayer(player)
		serverLogger.Printf("Player %s registered in SignalReady", clientID)
	}

	s.waitingRoom.AddPlayer(player)
	s.clientReady <- clientID

	// Ensure both players' streams are initialized before starting the game
	if players, ready := s.waitingRoom.ReadyPlayers(); ready {
		for _, p := range players {
			if p.stream == nil {
				serverLogger.Printf("Player %s stream not initialized, waiting...", p.ID)
				return &pong.SignalReadyResponse{}, nil
			}
		}

		serverLogger.Printf("Starting game with players: %v and %v", players[0].ID, players[1].ID)
		s.startGameWithPlayers(ctx, players)
	} else {
		serverLogger.Printf("Waiting for more players. Current ready players: %v", s.waitingRoom.queue)
	}

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
			if player, exists := s.waitingRoom.GetPlayer(clientID); exists && player.stream != nil {
				serverLogger.Printf("Player %s is ready", clientID)
				if players, ready := s.waitingRoom.ReadyPlayers(); ready {
					serverLogger.Printf("Ready to start a game with players: %v and %v", players[0].ID, players[1].ID)
					s.startGameWithPlayers(ctx, players)
				} else {
					serverLogger.Printf("Waiting for more players. Current ready players: %v", s.waitingRoom.queue)
				}
			} else {
				fmt.Printf("Waiting for player %s to initialize stream\n", clientID)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *server) NotifyGameStarted(req *pong.GameStartedStreamRequest, stream pong.PongGame_NotifyGameStartedServer) error {
	clientID := req.ClientId
	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}

	// Store the stream in the player's startNotifier
	s.mu.Lock()
	player.startNotifier = stream
	s.mu.Unlock()

	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *server) startGameWithPlayers(ctx context.Context, players []*Player) {
	gameID := generateGameID()
	serverLogger.Printf("Starting new game with ID: %s", gameID)

	newGameInstance := s.startNewGame(ctx)
	newGameInstance.players = players

	s.mu.Lock()
	s.games[gameID] = newGameInstance
	s.mu.Unlock()

	for _, player := range players {
		serverLogger.Printf("Notifying player %s that game %s started", player.ID, gameID)
		if player.startNotifier != nil {
			if err := player.startNotifier.Send(&pong.GameStartedStreamResponse{Message: "Game has started with ID: " + gameID}); err != nil {
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
	instance := &gameInstance{
		engine:   canvasEngine,
		framesch: framesch,
		inputch:  inputch,
	}

	go func() {
		canvasEngine.NewRound(ctx, instance.framesch, instance.inputch)
	}()

	return instance
}
