package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
)

type UpdatedMsg struct{}

type PongClientCfg struct {
	ServerAddr    string      // Address of the Pong server
	Log           slog.Logger // Application's logger
	ChatClient    types.ChatServiceClient
	PaymentClient types.PaymentsServiceClient
}
type PongClient struct {
	sync.RWMutex
	ctx       context.Context
	ID        string
	BetAmount float64

	CurrentWR *pong.WaitingRoom

	WaitingRooms []*pong.WaitingRoom

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
	GameCh    chan *pong.GameUpdateBytes
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
				if ntfn.BetAmt > 0 {
					log.Printf("Current Bet Amount: %.8f DCR\n", ntfn.BetAmt)
					pc.BetAmount = ntfn.BetAmt
					pc.UpdatesCh <- UpdatedMsg{}
				}
				if ntfn.Started {
					pc.UpdatesCh <- UpdatedMsg{}
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
			go func() { pc.UpdatesCh <- update }()
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

func (pc *PongClient) GetWaitingRooms() ([]*pong.WaitingRoom, error) {
	ctx := context.Background()

	res, err := pc.gc.GetWaitingRooms(ctx, &pong.WaitingRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("error sending input: %w", err)
	}
	pc.Lock()
	pc.WaitingRooms = res.Wr
	pc.Unlock()
	go func() { pc.UpdatesCh <- UpdatedMsg{} }()

	return res.Wr, nil
}

func (pc *PongClient) GetWRPlayers() ([]*pong.Player, error) {
	ctx := context.Background()

	wr, err := pc.gc.GetWaitingRoom(ctx, &pong.WaitingRoomRequest{})
	if err != nil {
		return nil, fmt.Errorf("error sending input: %w", err)
	}
	return wr.Players, nil
}

func (pc *PongClient) CreatewaitingRoom(ctx context.Context) (*pong.WaitingRoom, error) {
	res, err := pc.gc.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomResquest{
		HostId: pc.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("error sending input: %w", err)
	}
	pc.CurrentWR = res.Wr
	go func() { pc.UpdatesCh <- UpdatedMsg{} }()
	return res.Wr, nil
}

func (pc *PongClient) JoinWaitingRoom(ctx context.Context, roomID string) (*pong.JoinWaitingRoomResponse, error) {
	res, err := pc.gc.JoinWaitingRoom(ctx, &pong.JoinWaitingRoomRequest{
		ClientId: pc.ID,
		RoomId:   roomID,
	})
	if err != nil {
		return nil, fmt.Errorf("error sending input: %w", err)
	}
	go func() { pc.UpdatesCh <- UpdatedMsg{} }()
	return res, nil
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
