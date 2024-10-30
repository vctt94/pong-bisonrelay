package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
)

type PongClientCfg struct {
	ServerAddr    string      // Address of the Pong server
	Log           slog.Logger // Application's logger
	ChatClient    types.ChatServiceClient
	PaymentClient types.PaymentsServiceClient
}
type PongClient struct {
	ctx          context.Context
	ID           string
	playerNumber int32
	cfg          *PongClientCfg
	conn         *grpc.ClientConn
	// game client
	gc pong.PongGameClient
	// br clientrpc
	chat    types.ChatServiceClient
	payment types.PaymentsServiceClient

	stream    pong.PongGame_StartGameStreamClient
	notifier  pong.PongGame_StartNtfnStreamClient
	UpdatesCh chan tea.Msg
}

func (pc *PongClient) StartNotifier() error {
	ctx := context.Background()

	// Creates game start stream so we can notify when the game starts
	gameStartedStream, err := pc.gc.StartNtfnStream(ctx, &pong.StartNtfnStreamRequest{
		ClientId: pc.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating notifier stream: %w", err)
	}
	pc.notifier = gameStartedStream

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ntfn, err := pc.notifier.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					log.Printf("Error receiving notification: %v", err)
					return
				}
				fmt.Printf("ntfn: %+v\n", ntfn)
			}
		}
	}()

	return nil
}

func (pc *PongClient) SignalReady() error {
	ctx := context.Background()

	// Signal readiness after stream is initialized
	stream, err := pc.gc.StartGameStream(ctx, &pong.StartGameStreamRequest{
		ClientId: pc.ID,
	})
	if err != nil {
		return fmt.Errorf("error signaling readiness: %w", err)
	}

	// Set the stream before starting the goroutine
	pc.stream = stream

	// Use a separate goroutine to handle the stream
	go func() {
		for {
			update, err := pc.stream.Recv()
			if err != nil {
				log.Printf("stream receive error: %v", err)
				if errors.Is(err, io.EOF) {
					break
				}

				return
			}
			// fmt.Printf("update :%+v\n", update)
			pc.UpdatesCh <- update
		}
	}()

	return nil
}

func (pc *PongClient) SendInput(input string) error {
	ctx := context.Background()

	_, err := pc.gc.SendInput(ctx, &pong.PlayerInput{
		Input:        input,
		PlayerId:     pc.ID,
		PlayerNumber: pc.playerNumber,
	})
	if err != nil {
		return fmt.Errorf("error sending input: %w", err)
	}
	return nil
}

func (pc *PongClient) GetWRPlayers() ([]*pong.Player, error) {
	ctx := context.Background()

	wr, err := pc.gc.GetWaitingRoom(ctx, &pong.WaitingRoomRequest{})
	if err != nil {
		return nil, fmt.Errorf("error sending input: %w", err)
	}
	return wr.Players, nil
}

func NewPongClient(clientID string, cfg *PongClientCfg) (*PongClient, error) {
	// Establish a gRPC connection to the server using the address in cfg
	pongConn, err := grpc.Dial(cfg.ServerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Initialize the pongClient instance
	pc := &PongClient{
		ID:        clientID,
		cfg:       cfg,
		conn:      pongConn,
		gc:        pong.NewPongGameClient(pongConn),
		chat:      cfg.ChatClient,
		payment:   cfg.PaymentClient,
		UpdatesCh: make(chan tea.Msg),
	}

	return pc, nil
}
