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
	debug        = flag.Bool("debug", false, "")
	fps          = flag.Uint("fps", canvas.DEFAULT_FPS, "")
)

type server struct {
	pong.UnimplementedPongGameServer
	engine   *canvas.CanvasEngine
	framesch chan []byte
	inputch  chan []byte
}

func (s *server) SendInput(ctx context.Context, in *pong.PlayerInput) (*pong.GameUpdate, error) {
	// Convert and forward input to the game engine
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

			// Wrap the bytes in a GameUpdateBytes message
			gameUpdateBytes := &pong.GameUpdateBytes{
				Data: frame,
			}

			// Send the update to the client
			if err := stream.Send(gameUpdateBytes); err != nil {
				return err
			}
		}
	}
}

func (s *server) prepareGameUpdate(frame []byte) (*pong.GameUpdate, error) {
	// Assuming frame is already a serialized pong.GameUpdate for simplicity
	var update pong.GameUpdate
	if err := json.Unmarshal(frame, &update); err != nil {
		return nil, fmt.Errorf("error unmarshalling game update: %v", err)
	}
	return &update, nil
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
	canvasEngine.SetDebug(*debug).SetFPS(*fps)

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
