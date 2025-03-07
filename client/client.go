package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/decred/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type UpdatedMsg struct{}

type PongClientCfg struct {
	ServerAddr    string      // Address of the Pong server
	GRPCCertPath  string      // Cert to the grpc server
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

	IsReady bool

	BetAmt       float64
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
	ErrorsCh  chan error
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
					pc.ErrorsCh <- fmt.Errorf("notification stream closed")
					return
				}
				if err != nil {
					pc.ErrorsCh <- fmt.Errorf("notifier stream error: %v", err) // Send error
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
					if ntfn.PlayerId == pc.ID {
						pc.BetAmt = ntfn.BetAmt
						pc.ntfns.notifyBetAmtChanged(ntfn.PlayerId, ntfn.BetAmt, time.Now())
					}
				case pong.NotificationType_ON_PLAYER_READY:
					if ntfn.PlayerId == pc.ID {
						pc.IsReady = true
						pc.UpdatesCh <- true
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
				pc.ErrorsCh <- fmt.Errorf("game stream error: %v", err) // Send error
				if errors.Is(err, io.EOF) {
					break
				}
				return
			}
			// Forward updates to UpdatesCh
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

func (pc *PongClient) CreateWaitingRoom(clientId string, betAmt float64) (*pong.WaitingRoom, error) {
	ctx := context.Background()
	res, err := pc.gc.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomRequest{
		HostId: clientId,
		BetAmt: betAmt,
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

	// Load the credentials from the certificate file
	creds, err := credentials.NewClientTLSFromFile(cfg.GRPCCertPath, "")
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}

	// Dial the gRPC server with TLS credentials
	pongConn, err := grpc.Dial(cfg.ServerAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
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
		ErrorsCh:  make(chan error),
		log:       cfg.Log,

		ntfns: ntfns,
	}

	return pc, nil
}
