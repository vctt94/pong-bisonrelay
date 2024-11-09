package server

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
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
	ServerDir string

	Debug                 slog.Level
	DebugGameManagerLevel slog.Level
	PaymentClient         types.PaymentsServiceClient
	ChatClient            types.ChatServiceClient
	HTTPPort              string
}

type Server struct {
	pong.UnimplementedPongGameServer
	sync.RWMutex

	debug              slog.Level
	log                slog.Logger
	waitingRoomCreated chan struct{}

	paymentClient types.PaymentsServiceClient
	chatClient    types.ChatServiceClient
	users         map[zkidentity.ShortID]*Player
	gameManager   *gameManager

	httpServer    *http.Server
	activeStreams sync.Map
	db            serverdb.ServerDB
}

func NewServer(id *zkidentity.ShortID, cfg ServerConfig) *Server {
	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("[Server]")
	log.SetLevel(cfg.Debug)

	logGM := bknd.Logger("[GM]")
	logGM.SetLevel(cfg.DebugGameManagerLevel)

	dbPath := filepath.Join(cfg.ServerDir, "server.db")
	db, err := serverdb.NewBoltDB(dbPath)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}

	s := &Server{
		log:                log,
		debug:              cfg.Debug,
		db:                 db,
		paymentClient:      cfg.PaymentClient,
		chatClient:         cfg.ChatClient,
		waitingRoomCreated: make(chan struct{}, 1),
		users:              make(map[zkidentity.ShortID]*Player),
		gameManager: &gameManager{
			ID:             id,
			games:          make(map[string]*gameInstance),
			waitingRooms:   []*WaitingRoom{},
			playerSessions: &PlayerSessions{sessions: make(map[zkidentity.ShortID]*Player)},
			debug:          cfg.DebugGameManagerLevel,
			log:            logGM,
		},
	}

	if cfg.HTTPPort != "" {
		// Set up HTTP server for db calls
		mux := http.NewServeMux()
		mux.HandleFunc("/fetchTipsByClientID", s.FetchTipsByClientIDHandler)
		mux.HandleFunc("/fetchAllUnprocessedTips", s.FetchAllUnprocessedTipsHandler)
		s.httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
			Handler: mux,
		}

		go func() {
			s.log.Infof("Starting HTTP server on port %s", cfg.HTTPPort)
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.log.Errorf("HTTP server error: %v", err)
			}
		}()
	}

	return s
}

func (s *Server) StartGameStream(req *pong.StartGameStreamRequest, stream pong.PongGame_StartGameStreamServer) error {
	ctx := stream.Context()
	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)

	s.log.Debugf("client %s called StartGameStream", req.ClientId)
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
				NotificationType: pong.NotificationType_MESSAGE,
				Message:          fmt.Sprintf("player needs to place bet higher or equal to: %.8f", minAmt),
			})
			return fmt.Errorf("player needs to place bet higher or equal to: %.8f DCR", minAmt)
		}
	}

	player.stream = stream
	player.ready = true

	<-ctx.Done()
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) handleDisconnect(clientID zkidentity.ShortID) {
	playerSession := s.gameManager.playerSessions.GetPlayer(clientID)
	if playerSession != nil {
		s.gameManager.playerSessions.RemovePlayer(clientID)
	}

	waitingRoom := s.gameManager.GetWaitingRoomFromPlayer(clientID)
	if waitingRoom != nil {
		// Notify remaining players in the waiting room about disconnection if needed
		remainingPlayers := getRemainingPlayersInWaitingRoom(waitingRoom, clientID)
		for _, player := range remainingPlayers {
			if player.notifier != nil {
				player.notifier.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
					Message:          "Opponent left the waiting room.",
					Started:          false,
				})
			}
		}

		// Return the tip if necessary
		s.log.Debugf("Player %s disconnected; removing waiting room %s", clientID, waitingRoom.ID)
		waitingRoom.cancel()
		s.gameManager.RemoveWaitingRoom(waitingRoom.ID)
	}

	game := s.gameManager.getPlayerGame(clientID)
	// if player not in active game and have unprocessed tips, send them back.
	if game == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		tips, err := s.db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnprocessed)
		if err != nil {
			s.log.Errorf("failed to fetch unprocessed user %s tips: %v", clientID.String(), err)
		}
		if len(tips) > 0 {
			totalDcrAmount := 0.0
			for _, w := range tips {
				totalDcrAmount += float64(w.Tip.AmountMatoms) / 1e11 // Convert matoms to DCR
			}
			paymentReq := &types.TipUserRequest{
				User:        clientID.String(),
				DcrAmount:   totalDcrAmount,
				MaxAttempts: 3,
			}
			resp := &types.TipUserResponse{}
			if err = s.paymentClient.TipUser(ctx, paymentReq, resp); err != nil {
				s.log.Errorf("failed to send bet to winner %s: %v", clientID.String(), err)
			} else {
				s.log.Infof("unprocessed tip returned to user %s: %.8f", clientID.String(), totalDcrAmount)
				for _, w := range tips {
					tipID := make([]byte, 8)
					binary.BigEndian.PutUint64(tipID, w.Tip.SequenceId)
					if err := s.db.UpdateTipStatus(ctx, clientID.Bytes(), tipID, serverdb.StatusSending); err != nil {
						s.log.Errorf("failed to update tip status to sending from player %s: %w", clientID.String(), err)
					}
				}
			}
		}
	}
	if game != nil {
		remainingPlayer := getRemainingPlayerInGame(game, clientID)
		if remainingPlayer != nil && remainingPlayer.notifier != nil {
			remainingPlayer.notifier.Send(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
				Message:          "Opponent disconnected. Game over.",
				Started:          false,
			})
		}
		s.log.Debugf("Player %s disconnected and cleaned up", clientID)
		for gameID, g := range s.gameManager.games {
			if g == game {
				delete(s.gameManager.games, gameID)
				s.log.Debugf("Game %s cleaned up", gameID)
				break
			}
		}
	}
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	s.log.Debugf("StartNtfnStream called by client %s", clientID)

	s.activeStreams.Store(clientID, cancel)

	// Cleanup the stream from activeStreams on exit
	defer s.activeStreams.Delete(clientID)

	player := s.gameManager.playerSessions.GetOrCreateSession(clientID)
	player.notifier = stream
	s.Lock()
	s.users[clientID] = player
	s.Unlock()

	// Fetch unprocessed tips from the database
	tipsResult, err := s.db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnprocessed)
	if err != nil {
		s.log.Errorf("Failed to fetch unprocessed tips for client %s: %v", clientID, err)
		return err
	}

	totalDcrAmount := 0.0
	for _, w := range tipsResult {
		totalDcrAmount += float64(w.Tip.AmountMatoms) / 1e11 // Convert matoms to DCR
	}
	player.BetAmt = totalDcrAmount
	s.log.Debugf("Pending payments applied to client %s, total amount: %.8f", clientID, totalDcrAmount)

	s.RLock()
	s.users[clientID].notifier.Send(&pong.NtfnStreamResponse{
		NotificationType: pong.NotificationType_BET_AMOUNT_UPDATE,
		Message:          "Notifier stream Initialized",
		BetAmt:           player.BetAmt,
		PlayerId:         player.ID.String(),
	})
	s.RUnlock()

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

func (s *Server) ManageWaitingRoom(ctx context.Context, wr *WaitingRoom) error {
	defer s.gameManager.RemoveWaitingRoom(wr.ID) // Ensure waiting room removal

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Exited ManageWaitingRoom: %s (context cancelled)", wr.ID)
			return nil

		case <-time.After(time.Second):
			if players, ready := wr.ReadyPlayers(); ready {
				s.log.Infof("Game starting with players: %v and %v", players[0].ID, players[1].ID)

				go s.handleGameLifecycle(ctx, players, wr.BetAmount) // Start game lifecycle in a goroutine
				return nil
			}
		}
	}
}

func (s *Server) handleGameLifecycle(ctx context.Context, players []*Player, betAmt float64) {
	game, err := s.gameManager.startGame(ctx, players)
	if err != nil {
		s.log.Errorf("Failed to start game: %v", err)
		return
	}
	defer func() {
		// reset player status
		for _, g := range s.gameManager.games {
			if g == game {
				for _, player := range game.players {
					player.score = 0
					player.playerNumber = 0
				}
			}
		}
		// remove game from gameManager after it ended
		for gameID, g := range s.gameManager.games {
			if g == game {
				delete(s.gameManager.games, gameID)
				s.log.Infof("Game %s cleaned up", gameID)
				break
			}
		}
	}()

	game.Run()

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *Player) {
			defer wg.Done()
			if player.notifier != nil {
				err := player.notifier.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_GAME_START,
					Message:          "Game started with ID: " + game.id,
					Started:          true,
					GameId:           game.id,
				})
				if err != nil {
					s.log.Warnf("Failed to notify player %s: %v", player.ID, err)
				}
			}
			s.sendGameUpdates(ctx, player, game)
		}(player)
	}

	wg.Wait() // Wait for both players' streams to finish

	s.handleGameEnd(ctx, game, players, betAmt)
}

func (s *Server) sendGameUpdates(ctx context.Context, player *Player, game *gameInstance) {
	for {
		select {
		case <-ctx.Done():
			s.handleDisconnect(player.ID)
			return
		case frame, ok := <-game.framesch:
			if !ok {
				return // Game has ended, exit
			}
			err := player.stream.Send(&pong.GameUpdateBytes{Data: frame})
			if err != nil {
				s.handleDisconnect(player.ID)
				return
			}
		}
	}
}

func (s *Server) handleGameEnd(ctx context.Context, game *gameInstance, players []*Player, betAmt float64) {
	winner := game.winner
	var winnerID string
	if winner != nil {
		winnerID = winner.String()
		s.log.Infof("Game ended. Winner: %s", winnerID)
	} else {
		s.log.Infof("Game ended in a draw.")
	}

	totalBet := betAmt * 2
	// Notify players of game outcome
	for _, player := range players {
		message := "Game ended in a draw."
		if winner != nil && player.ID == *winner {
			message = fmt.Sprintf("Congratulations, you won and received: %.8f", totalBet)
		} else if winner != nil {
			message = fmt.Sprintf("Sorry, you lost and lose: %.8f", betAmt)
		}
		player.notifier.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.id,
		})
	}

	// Transfer bet amount to winner
	if winner != nil {
		resp := &types.TipUserResponse{}
		err := s.paymentClient.TipUser(ctx, &types.TipUserRequest{
			User:        winner.String(),
			DcrAmount:   totalBet,
			MaxAttempts: 3,
		}, resp)
		if err != nil {
			s.log.Errorf("Failed to transfer bet amount to winner %s: %v", winner.String(), err)
			return
		}

		s.log.Infof("transfering %.8f to winner %s", totalBet, winner.String())
		for _, player := range players {
			unprocessedTips, err := s.db.FetchReceivedTipsByUID(ctx, player.ID, serverdb.StatusUnprocessed)
			if err != nil {
				s.log.Errorf("Failed to fetch unprocessed tips for player %s: %v", player.ID, err)
			}
			for _, w := range unprocessedTips {
				tipID := make([]byte, 8)
				binary.BigEndian.PutUint64(tipID, w.Tip.SequenceId)
				err := s.db.UpdateTipStatus(ctx, player.ID.Bytes(), tipID, serverdb.StatusSending)
				if err != nil {
					s.log.Errorf("Failed to update tip status for player %s: %v", player.ID, err)
				}
			}
		}
	}
}

func (s *Server) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the server's Shutdown method
			if err := s.Shutdown(ctx); err != nil {
				s.log.Errorf("Error during server shutdown: %v", err)
			}

			return nil

		case <-s.waitingRoomCreated:
			s.log.Debugf("New waiting room created")

			s.gameManager.RLock()
			for _, wr := range s.gameManager.waitingRooms {
				if wr.ctx.Err() == nil { // Only manage rooms with active contexts
					s.log.Debugf("Managing waiting room: %s", wr.ID)
					go s.ManageWaitingRoom(wr.ctx, wr)
				}
			}
			s.gameManager.RUnlock()

		default:
			time.Sleep(100 * time.Millisecond)
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

	pwr, err := wr.ToPongWaitingRoom()
	if err != nil {
		return nil, err
	}

	host := s.gameManager.playerSessions.GetPlayer(wr.hostID)
	host.notifier.Send(&pong.NtfnStreamResponse{
		NotificationType: pong.NotificationType_PLAYER_JOINED_WR,
		Message:          fmt.Sprintf("new player joined your waiting room: %s", player.Nick),
		PlayerId:         player.ID.String(),
		RoomId:           wr.ID,
		Wr:               pwr,
	})

	return &pong.JoinWaitingRoomResponse{
		Wr: pwr,
	}, nil
}

func (s *Server) CreateWaitingRoom(ctx context.Context, req *pong.CreateWaitingRoomRequest) (*pong.CreateWaitingRoomResponse, error) {
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
	roomCtx, cancel := context.WithCancel(context.Background())
	wr := &WaitingRoom{
		ctx:       roomCtx,
		cancel:    cancel,
		ID:        id,
		hostID:    hostID,
		BetAmount: player.BetAmt,
		players:   []*Player{player},
	}
	s.Lock()
	s.gameManager.waitingRooms = append(s.gameManager.waitingRooms, wr)
	s.Unlock()
	s.log.Debugf("waiting room created. waiting room count: %d", len(s.gameManager.waitingRooms))

	// Signal that a new waiting room has been created
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
		// Non-blocking send to avoid deadlock in case of rapid room creations
	}

	pongWR, err := wr.ToPongWaitingRoom()
	if err != nil {
		return nil, err
	}

	s.RLock()
	for _, user := range s.users {
		user.notifier.Send(&pong.NtfnStreamResponse{
			Wr:               pongWR,
			NotificationType: pong.NotificationType_ON_WR_CREATED,
		})
	}
	s.RUnlock()
	pwr, err := wr.ToPongWaitingRoom()
	if err != nil {
		return nil, err
	}
	return &pong.CreateWaitingRoomResponse{
		Wr: pwr,
	}, nil
}

// Shutdown forcefully shuts down the server, closing HTTP server, database, waiting rooms, and games.
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop HTTP server
	if s.httpServer != nil {
		s.log.Info("Shutting down HTTP server...")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.log.Errorf("Error shutting down HTTP server: %v", err)
		}
	}

	// Close the database
	s.log.Info("Closing database...")
	if err := s.db.Close(); err != nil {
		s.log.Errorf("Error closing database: %v", err)
	}

	s.log.Info("Canceling active gRPC streams...")
	s.activeStreams.Range(func(key, value interface{}) bool {
		if cancel, ok := value.(context.CancelFunc); ok {
			cancel()
		}
		return true
	})

	// Gracefully close all waiting rooms and ongoing games
	s.log.Info("Shutting down waiting rooms and games...")
	s.gameManager.Lock()
	for _, wr := range s.gameManager.waitingRooms {
		wr.cancel() // Cancel each waiting room context
	}
	s.users = nil
	s.gameManager.waitingRooms = nil // Clear all waiting rooms
	s.gameManager.Unlock()

	s.log.Info("Server shut down completed.")
	return nil
}
