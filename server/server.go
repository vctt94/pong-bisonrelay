package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	canvas "pingpongexample/pong"
	"pingpongexample/pongrpc/grpc/types/pong"
	"sync"

	"github.com/google/uuid"
	"github.com/ndabAP/ping-pong/engine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	serverLogger = log.New(os.Stdout, "[SERVER] ", 0)
	debug        = flag.Bool("debug", false, "")
	fps          = flag.Uint("fps", canvas.DEFAULT_FPS, "")
)

type server struct {
	pong.UnimplementedPongGameServer
	mu             sync.Mutex
	clientReady    chan string              // Channel to signal a client is ready
	games          map[string]*gameInstance // Map to hold game instances, indexed by a game ID
	waitingClients []string
}

type gameInstance struct {
	engine   *canvas.CanvasEngine
	framesch chan []byte
	inputch  chan []byte
	players  []string // Track players (client IDs) in this game
}

func (s *server) SendInput(ctx context.Context, in *pong.PlayerInput) (*pong.GameUpdate, error) {
	// Example: Determine client ID and game instance (implementation depends on your client ID strategy)
	clientID, err := getClientIDFromContext(ctx) // Implement this based on your authentication/identification scheme
	if err != nil {
		return nil, err
	}
	gameInstance, exists := s.findGameInstanceByClientID(clientID)
	if !exists {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	// Forward input to the correct game instance
	input := fmt.Sprintf("%s", in.Input)
	gameInstance.inputch <- []byte(input)

	return &pong.GameUpdate{}, nil
}

func (s *server) StreamUpdates(req *pong.GameStreamRequest, stream pong.PongGame_StreamUpdatesServer) error {
	clientID, err := getClientIDFromStream(stream)
	if err != nil {
		return err
	}

	// Initially, check if a game instance exists.
	gameInstance, exists := s.findGameInstanceByClientID(clientID)
	if !exists {
		// Wait for the game instance to become available. This might involve a loop with a condition variable or a channel that notifies when the game is ready.
		// For illustration purposes only. Implement proper synchronization based on your application's architecture.
		for {
			gameInstance, exists = s.findGameInstanceByClientID(clientID)
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

func (s *server) findGameInstanceByClientID(clientID string) (*gameInstance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, game := range s.games {
		for _, playerID := range game.players {
			if playerID == clientID {
				return game, true
			}
		}
	}
	return nil, false
}

func newServer() *server {
	return &server{
		clientReady:    make(chan string, 10), // Buffer based on expected simultaneous ready signals
		games:          make(map[string]*gameInstance),
		waitingClients: make([]string, 0), // Initialize the waitingClients slice
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
	// Clients call this method to signal readiness
	fmt.Printf("Client %s is ready\n", clientID)

	s.clientReady <- clientID
}

func allGamesFull(map[string]*gameInstance) bool {
	return false
}

func (s *server) manageGames(ctx context.Context) {
	for {
		select {
		case clientID := <-s.clientReady:
			s.waitingClients = append(s.waitingClients, clientID)
			s.checkAndStartGame(ctx)
		case <-ctx.Done():
			return // Exit if the context is canceled
		}
	}
}

func (s *server) checkAndStartGame(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we have enough clients ready for a game
	if len(s.waitingClients) >= 1 {
		// Extract the first two clients
		player1 := s.waitingClients[0]
		s.waitingClients = s.waitingClients[1:] // Remove them from the waiting list

		// Start a new game with these clients
		newGame := s.startNewGame(ctx)
		newGame.players = []string{player1}

		// Notify these clients that a game has started
		// Implementation depends on your game logic
	}

	// Unlock if you used locking
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
		players:  make([]string, 0, 2),
	}
	s.games[gameID] = instance

	return instance
}

func main() {
	flag.Parse()

	srv := newServer()

	// Initialize and start managing games in the background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go srv.manageGames(ctx)

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pong.RegisterPongGameServer(grpcServer, srv)
	log.Println("server listening at", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
