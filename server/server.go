package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

const (
	name    = "pong"
	version = "v0.0.0"
)

var (
	fps           = flag.Uint("fps", canvas.DEFAULT_FPS, "")
	flagDCRAmount = flag.Float64("dcramount", 0.00000001, "Amount of DCR to tip the winner")
	flagIsF2p     = flag.Bool("isf2p", true, "allow f2p games")
)

type ServerConfig struct {
	Debug                 slog.Level
	DebugGameManagerLevel slog.Level
	PaymentClient         types.PaymentsServiceClient
	ChatClient            types.ChatServiceClient
}

type Server struct {
	pong.UnimplementedPongGameServer
	sync.Mutex

	debug              slog.Level
	log                slog.Logger
	waitingRoomCreated chan struct{}

	paymentClient   types.PaymentsServiceClient
	chatClient      types.ChatServiceClient
	users           map[zkidentity.ShortID]*Player
	gameManager     *gameManager
	unprocessedTips map[zkidentity.ShortID][]*types.ReceivedTip
}

func NewServer(id *zkidentity.ShortID, cfg ServerConfig) *Server {
	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("[Server]")
	log.SetLevel(cfg.Debug)

	logGM := bknd.Logger("[GM]")
	logGM.SetLevel(cfg.DebugGameManagerLevel)

	return &Server{
		log:   log,
		debug: cfg.Debug,
		gameManager: &gameManager{
			ID:           id,
			games:        make(map[string]*gameInstance),
			waitingRooms: []*WaitingRoom{},
			playerSessions: &PlayerSessions{
				sessions: make(map[zkidentity.ShortID]*Player),
			},
			debug: cfg.DebugGameManagerLevel,
			log:   logGM,
		},
		paymentClient:      cfg.PaymentClient,
		chatClient:         cfg.ChatClient,
		waitingRoomCreated: make(chan struct{}, 1),
		users:              make(map[zkidentity.ShortID]*Player),
		unprocessedTips:    make(map[zkidentity.ShortID][]*types.ReceivedTip),
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

	if !*flagIsF2p {
		minAmt := *flagDCRAmount
		if player.BetAmt < minAmt {
			player.notifier.Send(&pong.NtfnStreamResponse{
				Message: fmt.Sprintf("player needs to place bet higher or equal to: %.8f", minAmt),
			})
			return fmt.Errorf("player needs to place bet higher or equal to: %.8f DCR", minAmt)
		}
	}

	player.stream = stream
	player.ready = true

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
	// playerwr := s.gameManager.waitingRoom.GetPlayer(clientID)
	// if playerwr != nil {
	// 	// XXX return tip
	// 	s.gameManager.waitingRoom.RemovePlayer(clientID)
	// }

	game := s.gameManager.getPlayerGame(clientID)
	// if player not in active game and have unprocessed tips, send them back.
	if game == nil {
		if len(s.unprocessedTips[clientID]) > 0 {
			// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			// defer cancel()
			// s.sendUnprocessedTipToUser(ctx, clientID)
		}
	}
	if game != nil {
		remainingPlayer := game.players[0]
		// Notify the remaining player about the disconnection
		if remainingPlayer.notifier != nil {
			remainingPlayer.notifier.Send(&pong.NtfnStreamResponse{
				Message: "Opponent disconnected. Game over.",
				Started: false,
			})
			s.log.Debugf("Player %s disconnected and cleaned up", clientID)
		}
		s.gameManager.cleanupGameInstance(game)
	}
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx := stream.Context()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	s.log.Debugf("StartNtfnStream called by client %s", clientID)

	player := s.gameManager.playerSessions.GetOrCreateSession(clientID)
	player.notifier = stream

	s.users[clientID] = player

	s.Lock()
	if tips, exists := s.unprocessedTips[clientID]; exists {
		totalDcrAmount := 0.0
		for _, tip := range tips {
			totalDcrAmount += float64(tip.AmountMatoms) / 1e11 // Convert matoms to DCR
		}
		player.BetAmt = totalDcrAmount
		s.log.Debugf("Pending payments applied to client %s, total amount: %.8f", clientID, totalDcrAmount)
	}
	s.Unlock()

	player.notifier.Send(&pong.NtfnStreamResponse{Message: "Notifier stream Initialized", BetAmt: player.BetAmt})

	<-ctx.Done()
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) SendInput(ctx context.Context, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.PlayerId)

	player := s.gameManager.playerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player: %s not found", clientID)
	}
	if player.playerNumber != 1 && player.playerNumber != 2 {
		return nil, fmt.Errorf("player number incorrect, it must be 1 or 2; it is: %d", player.playerNumber)
	}

	game := s.gameManager.getPlayerGame(clientID)
	if game == nil {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	req.PlayerNumber = player.playerNumber
	inputBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	game.Lock()
	defer game.Unlock()

	if !game.running {
		return nil, fmt.Errorf("game has ended for client ID %s", clientID)
	}

	// Send inputBytes to game.inputch
	game.inputch <- inputBytes

	return &pong.GameUpdate{}, nil
}

func (s *Server) ackUnprocessedTipFromPlayer(ctx context.Context, clientID zkidentity.ShortID) {
	s.Lock()
	defer s.Unlock()
	if tips, exists := s.unprocessedTips[clientID]; exists {
		for _, tip := range tips {
			ackRes := &types.AckResponse{}
			err := s.paymentClient.AckTipReceived(ctx, &types.AckRequest{SequenceId: tip.SequenceId}, ackRes)
			if err != nil {
				s.log.Debugf("Failed to acknowledge tip for player %s: %v", clientID, err)
			} else {
				s.log.Debugf("Acknowledged tip with SequenceId %d for player %s", tip.SequenceId, clientID)
			}
		}
		// Remove acknowledged tips for this player
		delete(s.unprocessedTips, clientID)
	}
}

func (s *Server) ManageWaitingRoom(ctx context.Context, wr *WaitingRoom) error {
	var err error
	var game *gameInstance
	for {
		if players, ready := wr.ReadyPlayers(); ready {
			// remove wr after players are ready and removed from wr.
			s.gameManager.RemoveWaitingRoom(wr.ID)

			s.log.Debugf("Starting game with players: %v and %v", players[0].ID, players[1].ID)
			go func(players []*Player) {
				// Start the game with the ready players
				game, err = s.gameManager.startGame(ctx, players)
				if err != nil {
					return
				}
				go game.Run()

				var wg sync.WaitGroup
				for _, player := range players {
					wg.Add(1)
					go func(player *Player) {
						defer wg.Done()
						s.log.Debugf("Notifying player %s that game %s started", player.ID, game.id)
						if player.notifier == nil {
							return
						}

						// Send game start notification
						if err = player.notifier.Send(&pong.NtfnStreamResponse{
							Message: "Game has started with ID: " + game.id,
							Started: true,
							GameId:  game.id,
						}); err != nil {
							s.log.Debugf("Failed to send game start notification to player %s: %v", player.ID, err)
							return
						}

						// Game loop to handle frames and end game logic
						for {
							select {
							case <-ctx.Done():
								s.handleDisconnect(player.ID)
								return
							case frame, ok := <-game.framesch:
								if !ok {
									return
								}
								// Send game frame update
								if err := player.stream.Send(&pong.GameUpdateBytes{Data: frame}); err != nil {
									s.handleDisconnect(player.ID)
									return
								}
							}
						}
					}(player)
				}

				// Wait for all player routines to finish
				wg.Wait()

				// pay winner
				winner := game.winner
				if winner != nil {
					paymentReq := &types.TipUserRequest{
						User:        winner.String(),
						DcrAmount:   game.betAmt,
						MaxAttempts: 3,
					}
					resp := &types.TipUserResponse{}
					if err = s.paymentClient.TipUser(ctx, paymentReq, resp); err != nil {
						s.log.Errorf("Failed to send bet to winner %s: %v", winner.String(), err)
					} else {
						s.log.Debugf("Try sending total bet amount %.8f to winner %s", game.betAmt, winner.String())
						// XXX Store on db and ack only after successfully send it back
						for _, player := range players {
							s.ackUnprocessedTipFromPlayer(ctx, player.ID)
						}
					}
				}
			}(players)
			return err
		} else {
			s.log.Debugf("Not enough players ready. Current ready players: %v", wr.length())
		}
		time.Sleep(time.Second)
	}
}

func (s *Server) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case <-s.waitingRoomCreated:
			// Handle the new waiting room(s) when triggered
			s.log.Debugf("New waiting room created")

			s.gameManager.RLock()
			for _, wr := range s.gameManager.waitingRooms {
				s.log.Debugf("Found waiting room with ID: %s, HostID: %s, BetAmount: %.8f",
					wr.ID, wr.hostID, wr.BetAmount)

				// Start managing the waiting room
				go s.ManageWaitingRoom(ctx, wr)
			}
			s.gameManager.RUnlock()

		default:
			// Optional: Add sleep to avoid tight looping
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (s *Server) GetWaitingRoom(ctx context.Context, req *pong.WaitingRoomRequest) (*pong.WaitingRoomResponse, error) {
	// wrp := s.gameManager.waitingRoom.GetPlayers()

	// var players []*pong.Player
	// for _, p := range wrp {
	// 	players = append(players, &pong.Player{
	// 		Uid:       p.ID.String(),
	// 		Nick:      p.Nick,
	// 		BetAmount: p.BetAmt,
	// 	})
	// }
	// return &pong.WaitingRoomResponse{
	// 	Players: players,
	// }, nil
	return nil, nil
}

func (s *Server) GetWaitingRooms(ctx context.Context, req *pong.WaitingRoomsRequest) (*pong.WaitingRoomsResponse, error) {
	s.Lock()
	defer s.Unlock()

	wrp := s.gameManager.waitingRooms

	// Convert []*WaitingRoom to []*pong.WaitingRoom
	pongWaitingRooms := make([]*pong.WaitingRoom, len(wrp))
	for i, room := range wrp {
		pongPlayers := make([]*pong.Player, len(room.players))
		for j, player := range room.players {
			pongPlayers[j] = &pong.Player{
				Uid:       player.ID.String(),
				Nick:      player.Nick,
				BetAmount: player.BetAmt,
			}
		}
		pongWaitingRooms[i] = &pong.WaitingRoom{
			Id:      room.ID,
			HostId:  room.hostID.String(),
			Players: pongPlayers,
			BetAmt:  room.BetAmount,
		}
	}

	return &pong.WaitingRoomsResponse{
		Wr: pongWaitingRooms,
	}, nil
}
func (s *Server) JoinWaitingRoom(ctx context.Context, req *pong.JoinWaitingRoomRequest) (*pong.JoinWaitingRoomResponse, error) {
	var uid zkidentity.ShortID
	s.log.Debugf("client: %s entering room: %s", req.ClientId, req.RoomId)

	err := uid.FromString(req.ClientId)
	if err != nil {
		return nil, err
	}
	player := s.gameManager.playerSessions.GetPlayer(uid)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}
	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	wr.AddPlayer(player)

	return &pong.JoinWaitingRoomResponse{}, nil
}

func (s *Server) CreateWaitingRoom(ctx context.Context, req *pong.CreateWaitingRoomResquest) (*pong.CreateWaitingRoomResponse, error) {
	var hostID zkidentity.ShortID
	err := hostID.FromString(req.HostId)
	if err != nil {
		return nil, err
	}
	player := s.gameManager.playerSessions.GetPlayer(hostID)

	s.log.Debugf("creating waiting room. Host id: %s", hostID)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.HostId)
	}
	if !(*flagIsF2p) {
		if req.BetAmt <= 0 {
			return nil, fmt.Errorf("bet needs to be higher than 0: %.8f", req.BetAmt)
		}
	}
	id, err := generateRandomID()
	if err != nil {
		return nil, fmt.Errorf("error generating id: %w", err)
	}
	wr := &WaitingRoom{
		ID:        id,
		hostID:    hostID,
		BetAmount: player.BetAmt,
		players:   []*Player{player},
	}
	s.Lock()
	nwr := append(s.gameManager.waitingRooms, wr)
	s.gameManager.waitingRooms = nwr
	s.Unlock()
	s.log.Debugf("waiting room created. waiting room: %d", len(s.gameManager.waitingRooms))
	// Signal that a new waiting room has been created
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
		// Non-blocking send to avoid deadlock in case of rapid room creations
	}

	pp := &pong.Player{
		Uid:       player.ID.String(),
		BetAmount: player.BetAmt,
		Nick:      player.Nick,
	}
	pongWR := &pong.WaitingRoom{
		Id:      wr.ID,
		HostId:  wr.hostID.String(),
		Players: []*pong.Player{pp},
		BetAmt:  wr.BetAmount,
	}
	return &pong.CreateWaitingRoomResponse{
		Wr: pongWR,
	}, nil
}
