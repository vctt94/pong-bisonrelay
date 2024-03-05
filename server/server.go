package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	canvas "pingpongexample/pong"
	"pingpongexample/pongrpc/grpc/pong"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/slog"
	"github.com/google/uuid"
	"github.com/ndabAP/ping-pong/engine"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	clientReady    chan string              // Channel to signal a client is ready
	games          map[string]*gameInstance // Map to hold game instances, indexed by a game ID
	waitingClients []*Player
	paymentService types.PaymentsServiceClient
}

type Player struct {
	ID           string
	PlayerNumber int32 // 1 for player 1, 2 for player 2
	stream       pong.PongGame_StreamUpdatesServer
}

type gameInstance struct {
	engine   *canvas.CanvasEngine
	framesch chan []byte
	inputch  chan []byte
	players  []*Player
}

func (s *server) SendInput(ctx context.Context, in *pong.PlayerInput) (*pong.GameUpdate, error) {
	// Example: Determine client ID and game instance (implementation depends on your client ID strategy)
	clientID, err := getClientIDFromContext(ctx) // Implement this based on your authentication/identification scheme
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
	// Forward input to the correct game instance

	gameInstance.inputch <- inputBytes

	return &pong.GameUpdate{}, nil
}

func (s *server) StreamUpdates(req *pong.GameStreamRequest, stream pong.PongGame_StreamUpdatesServer) error {
	clientID, err := getClientIDFromStream(stream)
	if err != nil {
		return err
	}

	// Initially, check if a game instance exists.
	gameInstance, _, exists := s.findGameInstanceAndPlayerByClientID(clientID)
	if !exists {
		for {
			gameInstance, _, exists = s.findGameInstanceAndPlayerByClientID(clientID)
			if exists {
				break // Exit the loop when the game instance becomes available.
			}
		}
	}

	// Proceed with streaming updates to the client now that a game instance is guaranteed to exist.
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case frame := <-gameInstance.framesch:
			gameUpdateBytes := &pong.GameUpdateBytes{Data: frame}
			if err := stream.Send(gameUpdateBytes); err != nil {
				return err
			}
		}
	}
}

func (s *server) SignalReady(ctx context.Context, req *pong.SignalReadyRequest) (*pong.SignalReadyResponse, error) {
	clientID := req.ClientId
	serverLogger.Printf("SignalReady called by client ID: %s", clientID)

	// Signal that the client is ready. You might want to add the client to a list of ready clients or directly initiate game logic here.
	s.signalClientReady(clientID)

	// Return an empty response or include any relevant data.
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

func newServer(id string, paymentService types.PaymentsServiceClient) *server {
	return &server{
		ID:             id,
		clientReady:    make(chan string, 10), // Buffer based on expected simultaneous ready signals
		games:          make(map[string]*gameInstance),
		waitingClients: make([]*Player, 0), // Initialize the waitingClients slice
		paymentService: paymentService,
	}
}

func getClientIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	fmt.Printf("md: %+v\n\n", md)
	if !ok {
		return "", fmt.Errorf("no metadata found in context")
	}

	clientIDs, ok := md["client-id"] // Assuming the client ID is passed under the key "client-id"
	if !ok || len(clientIDs) == 0 {
		return "", fmt.Errorf("client-id not found in metadata")
	}

	return clientIDs[0], nil // Return the first client ID from the metadata
}

// getClientIDFromStream extracts the client ID from the stream's context.
func getClientIDFromStream(stream grpc.ServerStream) (string, error) {
	return getClientIDFromContext(stream.Context())
}

func (s *server) signalClientReady(clientID string) {
	fmt.Printf("Client %s is ready\n", clientID)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the client is already waiting
	for _, waitingClient := range s.waitingClients {
		if waitingClient.ID == clientID {
			fmt.Printf("Client %s is already waiting\n", clientID)
			return // Client is already in the list, so just return
		}
	}

	s.clientReady <- clientID
}
func (s *server) manageGames(ctx context.Context) {
	for {
		select {
		case clientID := <-s.clientReady:
			s.mu.Lock()

			// Always append the new client as a waiting client
			playerNumber := int32(len(s.waitingClients)) + 1 // This will be 1 or 2, depending on the position in the waitingClients slice

			s.waitingClients = append(s.waitingClients, &Player{ID: clientID, PlayerNumber: playerNumber})
			s.mu.Unlock()

			// If we have 2 or more clients waiting, start a game
			if len(s.waitingClients) >= 2 {
				s.checkAndStartGame(ctx)
			}
		case <-ctx.Done():
			return // Exit if the context is canceled
		}
	}
}

func (s *server) checkAndStartGame(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we have enough clients ready for a game
	if len(s.waitingClients) >= 2 {
		// Extract the first two clients

		players := []*Player{s.waitingClients[0], s.waitingClients[1]}
		// Remove them from the waiting list
		s.waitingClients = s.waitingClients[2:]
		// Notify both clients

		// Start a new game with these clients
		newGame := s.startNewGame(ctx)
		newGame.players = players

		// Notify these clients that a game has started
		// Implementation depends on your game logic
	}
}

func generateGameID() string {
	return uuid.New().String()
}

func (s *server) startNewGame(ctx context.Context) *gameInstance {
	// Initialize game engine
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(3, 3),
	)

	canvasEngine := canvas.New(game)
	canvasEngine.SetDebug(*debug).SetFPS(*fps)

	// Set up channels for the new game instance
	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 10)

	// Start the game engine for this instance
	go canvasEngine.NewRound(ctx, framesch, inputch)

	// Create a new game instance and add it to the map
	gameID := generateGameID() // Implement this function to generate unique IDs
	instance := &gameInstance{
		engine:   canvasEngine,
		framesch: framesch,
		inputch:  inputch,
	}
	s.games[gameID] = instance

	return instance
}

func receivePaymentLoop(ctx context.Context, payment types.PaymentsServiceClient, log slog.Logger) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest
	for {
		// Keep requesting a new stream if the connection breaks. Also
		// request any messages received since the last one we acked.
		streamReq := types.TipProgressRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := payment.TipProgress(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			log.Warn("Error while obtaining PM stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}

		for {
			var tip types.TipProgressEvent
			err := stream.Recv(&tip)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				log.Warnf("Error while receiving stream: %v", err)
				break
			}

			ruid := hex.EncodeToString(tip.Uid)
			fmt.Printf("received from id: %+v\n\n", ruid)
			// Ack to client that message is processed.
			ackReq.SequenceId = tip.SequenceId
			err = payment.AckTipProgress(ctx, &ackReq, &ackRes)
			if err != nil {
				log.Warnf("Error while ack'ing received pm: %v", err)
				break
			}
		}

		time.Sleep(time.Second)
	}
}

func realMain() error {
	flag.Parse()

	// Initialize and start managing games in the background
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

	// Check if clientID is still empty here, which it shouldn't be now
	if clientID == "" {
		return fmt.Errorf("client ID is empty after fetching")
	}
	fmt.Printf("client:%+v\n\n", clientID)
	srv := newServer(clientID, paymentService)

	go srv.manageGames(ctx)

	// Set up gRPC server
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
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
