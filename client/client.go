package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/decred/slog"

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

	// Notifications tracks handlers for client events. If nil, the client
	// will initialize a new notification manager. Specifying a
	// notification manager in the config is useful to ensure no
	// notifications are lost due to race conditions in client
	// initialization.
	Notifications *NotificationManager
}
type PongClient struct {
	sync.RWMutex
	ID string

	playerNumber int32
	cfg          *PongClientCfg
	conn         *grpc.ClientConn
	// game client
	gc pong.PongGameClient
	// br clientrpc
	chat    types.ChatServiceClient
	payment types.PaymentsServiceClient

	ntfns *NotificationManager

	log       slog.Logger
	stream    pong.PongGame_StartGameStreamClient
	notifier  pong.PongGame_StartNtfnStreamClient
	UpdatesCh chan tea.Msg
	GameCh    chan *pong.GameUpdateBytes
}

func (pc *PongClient) StartNotifier(ctx context.Context) error {
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
				pc.log.Infof("ntfn stream closed")
				return
			default:
				ntfn, err := pc.notifier.Recv()
				if errors.Is(err, io.EOF) {
					pc.log.Infof("ntfn stream closed")
					return
				}
				if err != nil {
					pc.log.Errorf("err: %v", err)
					return
				}

				// Handle notifications based on NotificationType
				switch ntfn.NotificationType {
				case pong.NotificationType_ON_WR_CREATED:
					pc.ntfns.notifyOnWRCreated(ntfn.Wr, time.Now())
				case pong.NotificationType_MESSAGE:
				case pong.NotificationType_PLAYER_JOINED_WR:
					pc.ntfns.notifyPlayerJoinedWR(ntfn.Wr, time.Now())
				case pong.NotificationType_GAME_START:
					if ntfn.Started {
						pc.ntfns.notifyGameStarted(ntfn.GameId, time.Now())
					}
				case pong.NotificationType_GAME_END:
					pc.ntfns.notifyGameEnded(ntfn.GameId, ntfn.Message, time.Now())
					pc.log.Infof("%s", ntfn.Message)
				case pong.NotificationType_OPPONENT_DISCONNECTED:
				case pong.NotificationType_BET_AMOUNT_UPDATE:
					if ntfn.BetAmt > 0 {
						pc.ntfns.notifyBetAmtChanged(ntfn.PlayerId, ntfn.BetAmt, time.Now())
					}
				default:
					pc.log.Warnf("Unknown notification type: %d", ntfn.NotificationType)
				}
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
				pc.log.Errorf("stream receive error: %v", err)
				if errors.Is(err, io.EOF) {
					break
				}

				return
			}
			// game stream update
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
		return nil, fmt.Errorf("error getting wr: %w", err)
	}
	go func() { pc.UpdatesCh <- res.Wr }()

	return res.Wr, nil
}

func (pc *PongClient) GetWRPlayers() ([]*pong.Player, error) {
	ctx := context.Background()

	wr, err := pc.gc.GetWaitingRoom(ctx, &pong.WaitingRoomRequest{})
	if err != nil {
		return nil, fmt.Errorf("error getting wr players: %w", err)
	}
	return wr.Players, nil
}

func (pc *PongClient) CreatewaitingRoom(ctx context.Context) (*pong.WaitingRoom, error) {
	res, err := pc.gc.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomRequest{
		HostId: pc.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating wr: %w", err)
	}
	return res.Wr, nil
}

func (pc *PongClient) JoinWaitingRoom(roomID string) (*pong.JoinWaitingRoomResponse, error) {
	ctx := context.Background()
	res, err := pc.gc.JoinWaitingRoom(ctx, &pong.JoinWaitingRoomRequest{
		ClientId: pc.ID,
		RoomId:   roomID,
	})
	if err != nil {
		return nil, fmt.Errorf("error joining wr: %w", err)
	}
	return res, nil
}

func NewPongClient(clientID string, cfg *PongClientCfg) (*PongClient, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("client must have logger")
	}
	// Establish a gRPC connection to the server using the address in cfg
	pongConn, err := grpc.Dial(cfg.ServerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	ntfns := cfg.Notifications
	if ntfns == nil {
		ntfns = NewNotificationManager()
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
		log:       cfg.Log,

		ntfns: ntfns,
	}

	return pc, nil
}
