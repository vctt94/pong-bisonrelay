package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/decred/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
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

	// For reconnection handling
	ctx          context.Context
	cancelFunc   context.CancelFunc
	reconnecting bool
	reconnectMu  sync.Mutex
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
				if err != nil {
					if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "transport is closing") ||
						strings.Contains(err.Error(), "connection is being forcefully terminated") {
						// Connection lost
						pc.log.Warnf("Lost connection to server (notification stream): %v", err)

						// Try to reconnect
						reconnectErr := pc.reconnect()
						if reconnectErr != nil {
							pc.ErrorsCh <- fmt.Errorf("failed to reconnect: %v", reconnectErr)
						}
						return // This goroutine ends, but a new one will be started by reconnect()
					}

					pc.ErrorsCh <- fmt.Errorf("notifier stream error: %v", err)
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
				if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "transport is closing") ||
					strings.Contains(err.Error(), "connection is being forcefully terminated") {
					// Connection lost
					pc.log.Warnf("Lost connection to server (game stream): %v", err)

					// Try to reconnect
					reconnectErr := pc.reconnect()
					if reconnectErr != nil {
						pc.ErrorsCh <- fmt.Errorf("failed to reconnect: %v", reconnectErr)
					}
					return // This goroutine ends, but a new one will be started by reconnect()
				}

				pc.ErrorsCh <- fmt.Errorf("game stream error: %v", err)
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

func (pc *PongClient) reconnect() error {
	pc.reconnectMu.Lock()
	if pc.reconnecting {
		pc.reconnectMu.Unlock()
		return nil // Already reconnecting
	}
	pc.reconnecting = true
	pc.reconnectMu.Unlock()

	defer func() {
		pc.reconnectMu.Lock()
		pc.reconnecting = false
		pc.reconnectMu.Unlock()
	}()

	pc.log.Infof("Attempting to reconnect to server...")

	// Load credentials
	creds, err := credentials.NewClientTLSFromFile(pc.cfg.GRPCCertPath, "")
	if err != nil {
		return fmt.Errorf("failed to load credentials for reconnection: %w", err)
	}

	// Close existing connection if it's still around
	if pc.conn != nil {
		pc.conn.Close()
	}

	// Implement exponential backoff for reconnection attempts
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second
	for i := 0; i < 10; i++ { // Try 10 times before giving up
		// Check if context was canceled
		if pc.ctx.Err() != nil {
			return pc.ctx.Err()
		}

		// Attempt to reconnect
		pongConn, err := grpc.Dial(pc.cfg.ServerAddr,
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second, // Send pings every 60 seconds instead of 10
				Timeout:             10 * time.Second, // Wait 20 seconds for ping ack
				PermitWithoutStream: false,            // Allow pings when there are no active streams
			}),
		)

		if err == nil {
			// Successfully reconnected
			pc.conn = pongConn
			pc.gc = pong.NewPongGameClient(pongConn)

			// Re-establish streams
			err = pc.StartNotifier(pc.ctx)
			if err != nil {
				pc.log.Errorf("Failed to restart notifier after reconnection: %v", err)
				// Close connection and try again
				pongConn.Close()
			} else {
				// If we were in a game, we need to re-establish the game stream
				if pc.stream != nil {
					err = pc.SignalReady()
					if err != nil {
						pc.log.Errorf("Failed to restart game stream after reconnection: %v", err)
						// Continue with the reconnected client even if we couldn't restart the game stream
					}
				}

				pc.log.Infof("Successfully reconnected to server")
				// Send notification that we've reconnected
				pc.UpdatesCh <- UpdatedMsg{}
				return nil
			}
		}

		// Sleep with backoff before retrying
		select {
		case <-pc.ctx.Done():
			return pc.ctx.Err()
		case <-time.After(backoff):
			// Increase backoff for next attempt, but cap it
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return fmt.Errorf("failed to reconnect after multiple attempts")
}

func NewPongClient(clientID string, cfg *PongClientCfg) (*PongClient, error) {
	if cfg.Log == nil {
		return nil, fmt.Errorf("client must have logger")
	}

	// Create a cancelable context for the client
	ctx, cancel := context.WithCancel(context.Background())

	// Load the credentials from the certificate file
	creds, err := credentials.NewClientTLSFromFile(cfg.GRPCCertPath, "")
	if err != nil {
		cancel() // Clean up the context
		log.Fatalf("Failed to load credentials: %v", err)
	}

	// Add connection options with healthchecking to detect disconnection faster
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second, // Send pings every 60 seconds instead of 10
			Timeout: 10 * time.Second, // Wait 20 seconds for ping ack
		}),
	}

	// Dial the gRPC server with TLS credentials
	pongConn, err := grpc.Dial(cfg.ServerAddr, dialOpts...)
	if err != nil {
		cancel() // Clean up the context
		log.Fatalf("Failed to connect to server: %v", err)
	}

	ntfns := cfg.Notifications
	if ntfns == nil {
		ntfns = NewNotificationManager()
	}

	// Initialize the pongClient instance
	pc := &PongClient{
		ID:         clientID,
		cfg:        cfg,
		conn:       pongConn,
		gc:         pong.NewPongGameClient(pongConn),
		chat:       cfg.ChatClient,
		payment:    cfg.PaymentClient,
		UpdatesCh:  make(chan tea.Msg),
		ErrorsCh:   make(chan error),
		log:        cfg.Log,
		ntfns:      ntfns,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	return pc, nil
}

func (pc *PongClient) Cleanup() {
	if pc.cancelFunc != nil {
		pc.cancelFunc()
	}
	if pc.conn != nil {
		pc.conn.Close()
	}
}
