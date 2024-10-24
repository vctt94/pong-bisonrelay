package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

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
	Debug bool
	// Add other configuration fields as needed
}

type Server struct {
	pong.UnimplementedPongGameServer
	GameManager *GameServer
}

func NewServer(id *zkidentity.ShortID, cfg ServerConfig) *Server {
	game := &GameServer{
		ID:             id,
		clientReady:    make(chan string, 10),
		games:          make(map[string]*gameInstance),
		waitingRoom:    NewWaitingRoom(),
		playerSessions: NewPlayerSessions(),
		debug:          cfg.Debug,
	}
	return &Server{
		GameManager: game,
	}
}

func (s *Server) StartGameStream(req *pong.StartGameStreamRequest, stream pong.PongGame_StartGameStreamServer) error {
	ctx := stream.Context()
	id := req.ClientId
	err := s.GameManager.startGameStream(&pong.StartGameStreamRequest{
		ClientId: id,
	}, stream)
	if err != nil {
		return err
	}
	for range ctx.Done() {
		// s.handleDisconnect(clientID)
		fmt.Printf("client ctx disconnected")
		return ctx.Err()
	}
	return nil
}

func (s *Server) StartNtfnStream(req *pong.StartNtfnStreamRequest, stream pong.PongGame_StartNtfnStreamServer) error {
	err := s.GameManager.startNtfnStream(req, stream)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) SendInput(ctx context.Context, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	update, err := s.GameManager.sendInput(ctx, req)
	if err != nil {
		return nil, err
	}

	return update, nil
}
