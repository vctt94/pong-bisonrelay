package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

const (
	name    = "pong"
	version = "v0.0.0"
)

var (
	serverLogger  = log.New(os.Stdout, "[SERVER] ", 0)
	fps           = flag.Uint("fps", canvas.DEFAULT_FPS, "")
	flagDCRAmount = flag.Float64("dcramount", 0, "Amount of DCR to tip the winner")
)

type ServerConfig struct {
	Debug         bool
	PaymentClient types.PaymentsServiceClient
	ChatClient    types.ChatServiceClient
}

type Server struct {
	pong.UnimplementedPongGameServer
	clientReady chan zkidentity.ShortID

	paymentClient types.PaymentsServiceClient
	chatClient    types.ChatServiceClient
	users         map[zkidentity.ShortID]*Player
	gameManager   *gameManager
	debug         bool
}

func NewServer(id *zkidentity.ShortID, cfg ServerConfig) *Server {
	manager := &gameManager{
		ID:    id,
		games: make(map[string]*gameInstance),
		waitingRoom: &WaitingRoom{
			queue: make([]*Player, 0),
		},
		playerSessions: &PlayerSessions{
			sessions: make(map[zkidentity.ShortID]*Player),
		},
		debug: cfg.Debug,
	}

	return &Server{
		debug:         cfg.Debug,
		gameManager:   manager,
		paymentClient: cfg.PaymentClient,
		chatClient:    cfg.ChatClient,
		clientReady:   make(chan zkidentity.ShortID, 10),
		users:         make(map[zkidentity.ShortID]*Player),
	}
}

func (s *Server) StartGameStream(req *pong.StartGameStreamRequest, stream pong.PongGame_StartGameStreamServer) error {
	ctx := stream.Context()
	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)

	player := s.gameManager.playerSessions.GetPlayer(clientID)
	if player == nil {
		return fmt.Errorf("player not found for client ID %s", clientID)
	}
	if player.notifier == nil {
		return fmt.Errorf("player notifier nil %s", clientID)
	}

	player.stream = stream

	s.gameManager.waitingRoom.AddPlayer(player)
	s.clientReady <- clientID
	serverLogger.Printf("Player %s added to waiting room. Current ready players: %v", player.ID, s.gameManager.waitingRoom.getWaitingRoom())

	for range ctx.Done() {
		s.handleDisconnect(clientID)
		fmt.Printf("client ctx disconnected")
		return ctx.Err()
	}
	return nil
}

func (s *Server) handleDisconnect(clientID zkidentity.ShortID) {
	playerSession := s.gameManager.playerSessions.GetPlayer(clientID)
	if playerSession != nil {
		s.gameManager.playerSessions.RemovePlayer(clientID)
	}

	game := s.gameManager.getPlayerGame(clientID)
	if game != nil {
		remainingPlayer := game.players[0]
		// Notify the remaining player about the disconnection
		if remainingPlayer.notifier != nil {
			remainingPlayer.notifier.Send(&pong.NtfnStreamResponse{
				Message: "Opponent disconnected. Game over.",
				Started: false,
			})
			serverLogger.Printf("Player %s disconnected and cleaned up", clientID)
		}
		s.gameManager.cleanupGameInstance(game)
	}
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx := stream.Context()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	serverLogger.Printf("StartNtfnStream called by client %s", clientID)

	player := s.gameManager.playerSessions.GetOrCreateSession(clientID)
	player.notifier = stream

	s.users[clientID] = player

	player.notifier.Send(&pong.NtfnStreamResponse{Message: "Notifier stream Initialized"})

	<-ctx.Done() // The context was canceled (client disconnected)
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) SendInput(ctx context.Context, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.PlayerId)
	game := s.gameManager.getPlayerGame(clientID)
	if game == nil {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	player := s.gameManager.playerSessions.GetPlayer(clientID)
	if player.playerNumber != 1 && player.playerNumber != 2 {
		return nil, fmt.Errorf("player number incorrect, it must be 1 or 2; it is: %d", player.playerNumber)
	}

	req.PlayerNumber = player.playerNumber
	inputBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}
	game.inputch <- inputBytes

	return &pong.GameUpdate{}, nil
}

func (s *Server) Run(ctx context.Context) error {
	for {
		select {
		case clientID := <-s.clientReady:

			serverLogger.Printf("Received client ready signal for client ID: %s", clientID)
			if players, ready := s.gameManager.waitingRoom.ReadyPlayers(); ready {
				serverLogger.Printf("Starting game with players: %v and %v", players[0].ID, players[1].ID)
				go func(players []*Player) {
					game := s.gameManager.startGame(ctx, players)
					go game.Run()

					var wg sync.WaitGroup
					for _, player := range players {
						wg.Add(1)
						go func(player *Player) {
							defer wg.Done()
							serverLogger.Printf("Notifying player %s that game %s started", player.ID, game.id)
							if player.notifier == nil {
								return
							}
							if err := player.notifier.Send(&pong.NtfnStreamResponse{Message: "Game has started with ID: " + game.id, Started: true}); err != nil {
								serverLogger.Printf("Failed to send game start notification to player %s: %v", player.ID, err)
								return
							}
							for {
								select {
								case <-ctx.Done():
									s.handleDisconnect(player.ID)
									return
								case frame, ok := <-game.framesch:
									if !ok {
										return
									}
									if err := player.stream.Send(&pong.GameUpdateBytes{Data: frame}); err != nil {
										s.handleDisconnect(player.ID)
										return
									}
								}
							}
						}(player)
					}
				}(players)
			} else {
				serverLogger.Printf("Not enough players ready. Current ready players: %v", s.gameManager.waitingRoom.length())
			}
		case <-ctx.Done():
			return nil
		}
	}
}
