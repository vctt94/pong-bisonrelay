package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

const (
	name    = "pong"
	version = "v0.0.0"
)

type ServerConfig struct {
	ServerDir string

	MinBetAmt             float64
	IsF2P                 bool
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
	isF2P              bool
	minBetAmt          float64
	waitingRoomCreated chan struct{}

	paymentClient types.PaymentsServiceClient
	chatClient    types.ChatServiceClient
	users         map[*zkidentity.ShortID]*ponggame.Player
	gameManager   *ponggame.GameManager

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
		log.Errorf("Failed to open database: %v\n", err)
		os.Exit(1)
	}

	s := &Server{
		log:                log,
		debug:              cfg.Debug,
		db:                 db,
		paymentClient:      cfg.PaymentClient,
		chatClient:         cfg.ChatClient,
		isF2P:              cfg.IsF2P,
		minBetAmt:          cfg.MinBetAmt,
		waitingRoomCreated: make(chan struct{}, 1),
		users:              make(map[*zkidentity.ShortID]*ponggame.Player),
		gameManager: &ponggame.GameManager{
			ID:             id,
			Games:          make(map[string]*ponggame.GameInstance),
			WaitingRooms:   []*ponggame.WaitingRoom{},
			PlayerSessions: &ponggame.PlayerSessions{Sessions: make(map[zkidentity.ShortID]*ponggame.Player)},
			Debug:          cfg.DebugGameManagerLevel,
			Log:            logGM,
		},
	}
	s.gameManager.OnWaitingRoomRemoved = s.handleWaitingRoomRemoved

	if cfg.HTTPPort != "" {
		// Set up HTTP server for db calls
		mux := http.NewServeMux()
		mux.HandleFunc("/fetchTipsByClientID", s.handleFetchTipsByClientIDHandler)
		mux.HandleFunc("/fetchAllUnprocessedTips", s.handleFetchAllUnprocessedTipsHandler)
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

	s.log.Debugf("Client %s called StartGameStream", req.ClientId)

	gameStreamReq := &ponggame.StartGameStreamRequest{
		ClientID: clientID,
		Stream:   stream,
		MinBet:   s.minBetAmt,
		IsF2P:    s.isF2P,
		Log:      s.log,
	}

	_, err := s.gameManager.StartGameStream(gameStreamReq)
	if err != nil {
		return err
	}

	// Wait for context to end and handle disconnection
	<-ctx.Done()
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) handleDisconnect(clientID zkidentity.ShortID) {
	// Remove player from sessions
	playerSession := s.gameManager.PlayerSessions.GetPlayer(clientID)
	if playerSession != nil {
		s.gameManager.PlayerSessions.RemovePlayer(clientID)
	}

	// Handle waiting room disconnection
	s.gameManager.HandleWaitingRoomDisconnection(clientID, s.log)

	// Handle game disconnection
	s.gameManager.HandleGameDisconnection(clientID, s.log)

	// Handle unprocessed tips
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.handleReturnUnprocessedTips(ctx, clientID, s.paymentClient, s.log); err != nil {
		s.log.Errorf("Error returning unprocessed tips for client %s: %v", clientID.String(), err)
	}
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	var clientID zkidentity.ShortID
	clientID.FromString(req.ClientId)
	s.log.Debugf("StartNtfnStream called by client %s", clientID)

	// Add to active streams
	s.activeStreams.Store(clientID, cancel)
	defer s.activeStreams.Delete(clientID)

	// Create player session
	player := s.gameManager.PlayerSessions.CreateSession(clientID)
	s.Lock()
	player.NotifierStream = stream
	s.users[&clientID] = player
	s.Unlock()

	// Fetch unprocessed tips
	totalDcrAmount, _, err := s.handleFetchTotalUnprocessedTips(ctx, clientID)
	if err != nil {
		s.log.Errorf("Failed to fetch unprocessed tips for client %s: %v", clientID, err)
		return err
	}

	// Update player's bet amount and notify
	player.BetAmt = totalDcrAmount
	s.log.Debugf("Pending payments applied to client %s, total amount: %.8f", clientID, totalDcrAmount)

	s.users[&clientID].NotifierStream.Send(&pong.NtfnStreamResponse{
		NotificationType: pong.NotificationType_BET_AMOUNT_UPDATE,
		Message:          "Notifier stream Initialized",
		BetAmt:           player.BetAmt,
		PlayerId:         player.ID.String(),
	})
	// Wait for disconnection
	<-ctx.Done()
	s.handleDisconnect(clientID)
	return ctx.Err()
}

func (s *Server) SendInput(ctx context.Context, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	var clientID zkidentity.ShortID
	clientID.FromString(req.PlayerId)

	s.log.Debugf("Received input from player %s", clientID)

	// Delegate to GameManager
	return s.gameManager.HandlePlayerInput(clientID, req)
}

func (s *Server) ManageWaitingRoom(ctx context.Context, wr *ponggame.WaitingRoom) error {
	defer s.gameManager.RemoveWaitingRoom(wr.ID)

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Exited ManageWaitingRoom: %s (context cancelled)", wr.ID)
			return nil

		case <-time.After(time.Second):
			players, ready := wr.ReadyPlayers()
			if ready {
				s.log.Infof("Game starting with players: %v and %v", players[0].ID, players[1].ID)

				go s.handleGameLifecycle(ctx, players, wr.BetAmount) // Start game lifecycle in a goroutine
				return nil
			}
		}
	}
}

func (s *Server) sendGameUpdates(ctx context.Context, player *ponggame.Player, game *ponggame.GameInstance) {
	for {
		select {
		case <-ctx.Done():
			s.handleDisconnect(*player.ID)
			return
		case frame, ok := <-game.Framesch:
			if !ok {
				return // Game has ended, exit
			}
			err := player.GameStream.Send(&pong.GameUpdateBytes{Data: frame})
			if err != nil {
				s.handleDisconnect(*player.ID)
				return
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

			s.gameManager.Lock()
			for _, wr := range s.gameManager.WaitingRooms {
				if wr.Ctx.Err() == nil { // Only manage rooms with active contexts
					s.log.Debugf("Managing waiting room: %s", wr.ID)
					go s.ManageWaitingRoom(wr.Ctx, wr)
				}
			}
			s.gameManager.Unlock()

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *Server) GetWaitingRooms(ctx context.Context, req *pong.WaitingRoomsRequest) (*pong.WaitingRoomsResponse, error) {
	s.Lock()
	defer s.Unlock()

	pongWaitingRooms := make([]*pong.WaitingRoom, len(s.gameManager.WaitingRooms))
	for i, room := range s.gameManager.WaitingRooms {
		wr, err := room.Marshal()
		if err != nil {
			return nil, err
		}
		pongWaitingRooms[i] = wr
	}

	return &pong.WaitingRoomsResponse{
		Wr: pongWaitingRooms,
	}, nil
}

func (s *Server) JoinWaitingRoom(ctx context.Context, req *pong.JoinWaitingRoomRequest) (*pong.JoinWaitingRoomResponse, error) {
	var uid zkidentity.ShortID
	err := uid.FromString(req.ClientId)
	if err != nil {
		return nil, err
	}
	player := s.gameManager.PlayerSessions.GetPlayer(uid)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", req.ClientId)
	}

	wr := s.gameManager.GetWaitingRoom(req.RoomId)
	if wr == nil {
		return nil, fmt.Errorf("waiting room not found: %s", req.RoomId)
	}

	// Check if BetAmt matches
	if player.BetAmt != wr.BetAmount {
		return nil, fmt.Errorf("bet amount mismatch. Room Bet: %.8f, Player Bet: %.8f", wr.BetAmount, player.BetAmt)
	}

	wr.AddPlayer(player)

	pwr, err := wr.Marshal()
	if err != nil {
		return nil, err
	}
	for _, p := range wr.Players {
		p.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_PLAYER_JOINED_WR,
			Message:          fmt.Sprintf("New player joined Waiting Room: %s", player.Nick),
			PlayerId:         p.ID.String(),
			RoomId:           wr.ID,
			Wr:               pwr,
		})
	}

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

	hostPlayer := s.gameManager.PlayerSessions.Sessions[hostID]
	if hostPlayer == nil {
		return nil, fmt.Errorf("player not found: %s", req.HostId)
	}
	if hostPlayer.BetAmt != req.BetAmt {
		return nil, fmt.Errorf("server and request mismatch. request amt: %.8f, server amt: %.8f", req.BetAmt, hostPlayer.BetAmt)
	}
	if !s.isF2P && req.BetAmt == 0 {
		return nil, fmt.Errorf("bet needs to be higher than 0")
	}
	if !s.isF2P && req.BetAmt < s.minBetAmt {
		return nil, fmt.Errorf("bet needs to be higher than %.8f", s.minBetAmt)
	}

	s.log.Debugf("creating waiting room. Host ID: %s", hostID)

	// Check if a Waiting Room with the same Host ID already exists
	s.gameManager.Lock()
	defer s.gameManager.Unlock()
	for _, room := range s.gameManager.WaitingRooms {
		if *room.HostID == hostID {
			return nil, fmt.Errorf("A Waiting Room already exists for Host ID: %s", hostID)
		}
	}

	wr, err := ponggame.NewWaitingRoom(hostPlayer, req.BetAmt)
	if err != nil {
		return nil, fmt.Errorf("failed to create waiting room: %v", err)
	}
	s.gameManager.WaitingRooms = append(s.gameManager.WaitingRooms, wr)
	s.log.Debugf("waiting room created. Total rooms: %d", len(s.gameManager.WaitingRooms))

	// Signal that a new Waiting Room has been created
	select {
	case s.waitingRoomCreated <- struct{}{}:
	default:
		// Non-blocking send to avoid deadlock in case of rapid room creations
	}

	pongWR, err := wr.Marshal()
	if err != nil {
		return nil, err
	}

	for _, user := range s.users {
		user.NotifierStream.Send(&pong.NtfnStreamResponse{
			Wr:               pongWR,
			NotificationType: pong.NotificationType_ON_WR_CREATED,
		})
	}

	return &pong.CreateWaitingRoomResponse{
		Wr: pongWR,
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
	for _, wr := range s.gameManager.WaitingRooms {
		wr.Cancel() // Cancel each waiting room context
	}
	s.users = nil
	s.gameManager.WaitingRooms = nil // Clear all waiting rooms
	s.gameManager.Unlock()

	s.log.Info("Server shut down completed.")
	return nil
}
