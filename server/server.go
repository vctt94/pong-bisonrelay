package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientplugin/grpctypes"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/ndabAP/ping-pong/engine"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

var (
	serverLogger  = log.New(os.Stdout, "[SERVER] ", 0)
	fps           = flag.Uint("fps", canvas.DEFAULT_FPS, "")
	flagDCRAmount = flag.Float64("dcramount", 0, "Amount of DCR to tip the winner")
)

type GameServer struct {
	pong.UnimplementedPongGameServer
	ID             *zkidentity.ShortID
	mu             sync.Mutex
	clientReady    chan string
	games          map[string]*gameInstance
	waitingRoom    *WaitingRoom
	playerSessions *PlayerSessions
	paymentService types.PaymentsServiceClient
	dcrAmount      float64
	debug          bool
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
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s *GameServer) SendInput(ctx context.Context, in *pong.PlayerInput) (*pong.GameUpdate, error) {
	clientID := in.PlayerId
	gameInstance, player, exists := s.findGameInstanceAndPlayerByClientID(clientID)
	if !exists {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	in.PlayerNumber = player.PlayerNumber
	inputBytes, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}
	gameInstance.inputch <- inputBytes

	return &pong.GameUpdate{}, nil
}

func (s *GameServer) cleanupGameInstance(instance *gameInstance) {
	if !instance.cleanedUp {
		instance.cleanedUp = true
		instance.cancel()
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

func (s *GameServer) handleDisconnect(clientID string) {
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
			remainingPlayer.notifier.Send(&grpctypes.PluginStartStreamResponse{
				Message: "Opponent disconnected. Game over.",
				Started: false,
			})
		}
	}

	// Remove player session
	s.playerSessions.RemovePlayer(clientID)
	serverLogger.Printf("Player %s disconnected and cleaned up", clientID)
}

func (s *GameServer) SignalReady(req *pong.SignalReadyRequest, stream grpctypes.PluginService_CallActionServer) error {
	clientID := req.ClientId

	serverLogger.Printf("SignalReady called by client ID: %s", clientID)

	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}
	if player.notifier == nil {
		return fmt.Errorf("player notifier nil %s", clientID)
	}

	if err := player.notifier.Send(&grpctypes.PluginStartStreamResponse{Message: "Client signaling ready"}); err != nil {
		serverLogger.Printf("Failed to send game start notification to player %s: %v", player.ID, err)
	}
	player.stream = stream
	s.playerSessions.AddOrUpdatePlayer(player)
	serverLogger.Printf("Player %s stream initialized in StreamUpdates", clientID)

	s.waitingRoom.AddPlayer(player)
	s.clientReady <- clientID
	serverLogger.Printf("Player %s added to waiting room. Current ready players: %v", player.ID, s.waitingRoom.queue)

	return nil
}

func (s *GameServer) findGameInstanceAndPlayerByClientID(clientID string) (*gameInstance, *Player, bool) {
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

func NewServer(id *zkidentity.ShortID, debug bool) *GameServer {
	return &GameServer{
		ID:             id,
		clientReady:    make(chan string, 10),
		games:          make(map[string]*gameInstance),
		waitingRoom:    NewWaitingRoom(),
		playerSessions: NewPlayerSessions(),
		debug:          debug,
	}
}

func (s *GameServer) ManageGames(ctx context.Context) error {
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
			return nil
		}
	}
}

func (s *GameServer) Init(req *pong.GameStartedStreamRequest, stream grpctypes.PluginService_InitServer) error {
	ctx := stream.Context()
	clientID := req.ClientId

	serverLogger.Printf("Init called by client")

	player, exists := s.playerSessions.GetPlayer(clientID)
	if !exists {
		player = NewPlayer(clientID)
		player.notifier = stream
		s.playerSessions.AddOrUpdatePlayer(player)
		serverLogger.Printf("Player %s registered in Init", clientID)
	} else {
		player.notifier = stream
	}

	player.notifier.Send(&grpctypes.PluginStartStreamResponse{Message: "Notifier stream Initialized"})

	// Listen for context cancellation to handle disconnection
	for range ctx.Done() {
		s.handleDisconnect(clientID)
		return ctx.Err()
	}

	return nil
}

func (s *GameServer) startGame(ctx context.Context, players []*Player) error {
	gameID := generateGameID()
	serverLogger.Printf("Starting new game with ID: %s", gameID)

	newGameInstance := s.startNewGame(ctx, players)
	s.mu.Lock()
	s.games[gameID] = newGameInstance
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *Player) {
			defer wg.Done()
			serverLogger.Printf("Notifying player %s that game %s started", player.ID, gameID)
			if player.notifier == nil {
				return
			}
			if err := player.notifier.Send(&grpctypes.PluginStartStreamResponse{Message: "Game has started with ID: " + gameID, Started: true, PlayerNumber: player.PlayerNumber}); err != nil {
				serverLogger.Printf("Failed to send game start notification to player %s: %v", player.ID, err)
				return
			}
			for {
				select {
				case <-ctx.Done():
					s.handleDisconnect(player.ID)
					return
				case frame, ok := <-newGameInstance.framesch:
					if !ok {
						return
					}
					if err := player.stream.Send(&grpctypes.PluginCallActionStreamResponse{Response: frame}); err != nil {
						fmt.Printf("err: %+v\n\n", err)
						s.handleDisconnect(player.ID)
						return
					}
				}
			}
		}(player)
	}

	wg.Wait()
	return nil
}

func (s *GameServer) startNewGame(ctx context.Context, players []*Player) *gameInstance {
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(3, 3),
	)

	s.mu.Lock()
	players[0].PlayerNumber = 1
	players[1].PlayerNumber = 2
	s.mu.Unlock()

	canvasEngine := canvas.New(game)
	canvasEngine.SetDebug(s.debug).SetFPS(*fps)

	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 10)
	roundResult := make(chan int32)
	instanceCtx, cancel := context.WithCancel(ctx)
	instance := &gameInstance{
		engine:      canvasEngine,
		framesch:    framesch,
		inputch:     inputch,
		roundResult: roundResult,
		running:     true,
		ctx:         instanceCtx,
		cancel:      cancel,
		players:     players,
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
		canvasEngine.NewRound(instance.ctx, instance.framesch, instance.inputch, instance.roundResult)
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

func (s *GameServer) handleRoundResult(playerNumber int32, instance *gameInstance) {
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

	if s.dcrAmount == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &types.TipUserRequest{
		User:        winner.ID,
		DcrAmount:   s.dcrAmount,
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
