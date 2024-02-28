package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	canvas "pingpongexample/pong"
	"pingpongexample/pongrpc/grpc/types/pong"

	"github.com/ndabAP/ping-pong/engine"
	"google.golang.org/grpc"
)

var (
	serverLogger = log.New(os.Stdout, "[SERVER] ", 0)

	debug = flag.Bool("debug", false, "")
	fps   = flag.Uint("fps", canvas.DEFAULT_FPS, "")
)

type server struct {
	pong.UnimplementedPongGameServer
	engine   *canvas.CanvasEngine
	framesch chan []byte
	inputch  chan []byte
}

func (s *server) SendInput(ctx context.Context, in *pong.PlayerInput) (*pong.GameUpdate, error) {
	// Example: Convert and forward input to the game engine
	input := fmt.Sprintf("%s", in.Input) // Assuming Key is something like "ArrowUp" or "ArrowDown"
	s.inputch <- []byte(input)

	// Return a placeholder response for now
	return &pong.GameUpdate{}, nil
}

func (s *server) StreamUpdates(req *pong.GameStreamRequest, stream pong.PongGame_StreamUpdatesServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case frame := <-s.framesch:
			// Deserialize the game state from JSON
			var engineOutput *canvas.CanvasEngine // Assuming GameState is the structured representation of your game state
			if err := json.Unmarshal(frame, &engineOutput); err != nil {
				serverLogger.Printf("error unmarshalling frame: %v", err)
				continue // or handle the error appropriately
			}

			// Convert the game state to a pong.GameUpdate message
			update, err := convertToGameUpdate(s.engine)
			if err != nil {
				return err
			}

			// Send the update to the client
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

func convertToGameUpdate(engineOutput *canvas.CanvasEngine) (*pong.GameUpdate, error) {
	// Assuming engineOutput.MarshalJSON() returns ([]byte, error)
	// But since we're directly converting, we won't use MarshalJSON here.

	return &pong.GameUpdate{
		GameWidth:     int32(engineOutput.Game.Width),
		GameHeight:    int32(engineOutput.Game.Height),
		P1Width:       int32(engineOutput.Game.P1.Width),
		P1Height:      int32(engineOutput.Game.P1.Height),
		P2Width:       int32(engineOutput.Game.P2.Width),
		P2Height:      int32(engineOutput.Game.P2.Height),
		BallWidth:     int32(engineOutput.Game.Ball.Width),
		BallHeight:    int32(engineOutput.Game.Ball.Height),
		P1Score:       int32(engineOutput.P1Score),
		P2Score:       int32(engineOutput.P2Score),
		BallX:         int32(engineOutput.BallX),
		BallY:         int32(engineOutput.BallY),
		P1X:           int32(engineOutput.P1X),
		P1Y:           int32(engineOutput.P1Y),
		P2X:           int32(engineOutput.P2X),
		P2Y:           int32(engineOutput.P2Y),
		P1YVelocity:   int32(engineOutput.P1YVelocity),
		P2YVelocity:   int32(engineOutput.P2YVelocity),
		BallXVelocity: int32(engineOutput.BallXVelocity),
		BallYVelocity: int32(engineOutput.BallYVelocity),
		Fps:           float32(engineOutput.FPS),
		Tps:           float32(engineOutput.TPS),
	}, nil
}

func main() {
	// Initialize game engine
	game := engine.NewGame(
		80,
		40,
		engine.NewPlayer(10, 15),
		engine.NewPlayer(10, 15),
		engine.NewBall(3, 3),
	)

	canvasEngine := canvas.New(game)
	canvasEngine.SetDebug(false).SetFPS(60)

	// Set up channels
	framesch := make(chan []byte, 100) // Buffer based on expected frame rate and network delay
	inputch := make(chan []byte, 10)   // Buffer based on expected input rate

	// Start the game engine in a new goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go canvasEngine.NewRound(ctx, framesch, inputch)

	// Set up gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	srv := &server{
		engine:   canvasEngine,
		framesch: framesch,
		inputch:  inputch,
	}
	pong.RegisterPongGameServer(grpcServer, srv)
	log.Println("server listening at", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
