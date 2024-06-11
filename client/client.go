package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"

	"github.com/decred/slog"
	"google.golang.org/grpc"
)

type ID = zkidentity.ShortID

type appMode int

const (
	gameIdle appMode = iota
	gameMode
)

var (
	// serverAddr = flag.String("server_addr", "104.131.180.29:50051", "The server address in the format of host:port")

	serverAddr = flag.String("server_addr", "localhost:50051", "The server address in the format of host:port")
	brdatadir  = flag.String("brdatadir", "", "Directory containing the certificates and keys")
)

type PongClientCfg struct {
	ServerAddr string      // Address of the Pong server
	Log        slog.Logger // Application's logger
}

type pongClient struct {
	ID           string
	playerNumber int32
	cfg          *PongClientCfg
	conn         *grpc.ClientConn
	pongClient   pong.PongGameClient
	stream       pong.PongGame_StreamUpdatesClient
	updatesCh    chan tea.Msg
}

type GameStartedMsg struct {
	Started      bool
	PlayerNumber int32
}

func (pc *pongClient) StartNotifier() error {
	ctx := attachClientIDToContext(context.Background(), pc.ID)

	// Creates game start stream so we can notify when the game starts
	gameStartedStream, err := pc.pongClient.StartNotifier(ctx, &pong.GameStartedStreamRequest{})
	if err != nil {
		return fmt.Errorf("error creating game started stream: %w", err)
	}

	go func() {
		for {
			started, err := gameStartedStream.Recv()
			pc.playerNumber = started.PlayerNumber
			pc.ID = started.ClientId
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				log.Printf("Error receiving game started notification: %v", err)
				return
			}
			pc.updatesCh <- GameStartedMsg{Started: started.Started}
		}
	}()

	return nil
}

func (pc *pongClient) SendInput(input string) error {
	ctx := attachClientIDToContext(context.Background(), pc.ID)

	_, err := pc.pongClient.SendInput(ctx, &pong.PlayerInput{
		Input:        input,
		PlayerId:     pc.ID,
		PlayerNumber: pc.playerNumber,
	})
	if err != nil {
		return fmt.Errorf("error sending input: %w", err)
	}
	return nil
}

type GameUpdateMsg *pong.GameUpdateBytes

type model struct {
	mode           appMode
	gameStateMutex sync.Mutex
	gameState      *pong.GameUpdate
	err            error
	ctx            context.Context
	cancel         context.CancelFunc
	pc             *pongClient
	chatClient     *types.ChatServiceClient
	versionClient  *types.VersionServiceClient
}

func initialModel(pc *pongClient, chatClient *types.ChatServiceClient, versionClient *types.VersionServiceClient) *model {
	ctx, cancel := context.WithCancel(context.Background())
	return &model{
		mode:          gameIdle,
		ctx:           ctx,
		cancel:        cancel,
		pc:            pc,
		chatClient:    chatClient,
		versionClient: versionClient,
	}
}

func (m *model) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		for msg := range m.pc.updatesCh {
			return msg
		}
		return nil
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.listenForUpdates(), func() tea.Msg {
		for msg := range m.pc.updatesCh {
			return msg
		}
		return nil
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeySpace:
			m.mode = gameMode
			return m, m.makeClientReady()
		case tea.KeyEsc:
			return m, tea.Quit
		}
		if m.mode == gameMode {
			switch msg.String() {
			case "w", "s", "up", "down":
				return m, m.handleGameInput(msg)
			}
		}
		return m, nil
	case GameUpdateMsg:
		m.gameStateMutex.Lock()
		var gameUpdate pong.GameUpdate
		if err := json.Unmarshal(msg.Data, &gameUpdate); err != nil {
			m.err = err
			m.gameStateMutex.Unlock()
			return m, nil
		}
		m.gameState = &gameUpdate
		m.gameStateMutex.Unlock()
		return m, m.listenForUpdates()
	case GameStartedMsg:
		m.gameState = &pong.GameUpdate{}
		return m, m.listenForUpdates()
	case types.VersionResponse:
		fmt.Printf("AppName: %s\nversion: %s\nGoRuntime: %s\n", msg.AppName, msg.AppVersion, msg.GoRuntime)
	}
	return m, nil
}

func (m *model) makeClientReady() tea.Cmd {
	log.Println("Client signaling readiness")
	go func() {
		err := m.pc.SignalReady()
		if err != nil {
			log.Printf("Error signaling readiness: %v", err)
		}
	}()
	return nil
}

func (m *model) handleGameInput(msg tea.KeyMsg) tea.Cmd {
	return func() tea.Msg {
		var input string
		switch msg.String() {
		case "w", "up":
			input = "ArrowUp"
		case "s", "down":
			input = "ArrowDown"
		}
		if input != "" {
			err := m.pc.SendInput(input)
			if err != nil {
				log.Printf("Error sending game input: %v", err)
			}
		}
		return nil
	}
}

func (m *model) View() string {
	var b strings.Builder
	if m.mode == gameIdle {
		fmt.Fprintln(&b, "Idle mode: Press space to get ready for the game")
		fmt.Fprintln(&b, "Idle mode: Press esc to quit.")
	} else if m.mode == gameMode {
		fmt.Fprintln(&b, "Game mode: 'q' to return to chat.")
		if m.gameState == nil {
			return "Waiting for game to start..."
		}

		var gameView strings.Builder
		for y := 0; y < int(m.gameState.GameHeight); y++ {
			for x := 0; x < int(m.gameState.GameWidth); x++ {
				ballX := int(m.gameState.BallX)
				ballY := int(m.gameState.BallY)
				switch {
				case x == ballX && y == ballY:
					gameView.WriteString("O")
				case x == 0 && y >= int(m.gameState.P1Y) && y < int(m.gameState.P1Y)+int(m.gameState.P1Height):
					gameView.WriteString("|")
				case x == int(m.gameState.GameWidth)-1 && y >= int(m.gameState.P2Y) && y < int(m.gameState.P2Y)+int(m.gameState.P2Height):
					gameView.WriteString("|")
				default:
					gameView.WriteString(" ")
				}
			}
			gameView.WriteString("\n")
		}
		gameView.WriteString(fmt.Sprintf("Score: %d - %d\n", m.gameState.P1Score, m.gameState.P2Score))
		gameView.WriteString("Controls: W/S and Arrow Keys - Q to quit")

		return gameView.String()
	}

	return b.String()
}

func (pc *pongClient) SignalReady() error {
	ctx := attachClientIDToContext(context.Background(), pc.ID)

	// Signal readiness after stream is initialized
	_, err := pc.pongClient.SignalReady(ctx, &pong.SignalReadyRequest{})
	if err != nil {
		return fmt.Errorf("error signaling readiness: %w", err)
	}

	err = pc.initializeStream(ctx)
	if err != nil {
		return fmt.Errorf("error initializing stream: %w", err)
	}

	log.Println("Stream initialized successfully")
	return nil
}

func (pc *pongClient) initializeStream(ctx context.Context) error {
	if pc.pongClient == nil {
		return fmt.Errorf("pongClient is nil")
	}

	// Initialize the stream
	stream, err := pc.pongClient.StreamUpdates(ctx, &pong.GameStreamRequest{})
	if err != nil {
		return fmt.Errorf("error creating updates stream: %w", err)
	}

	// Set the stream before starting the goroutine
	pc.stream = stream

	// Use a separate goroutine to handle the stream
	go func() {
		for {
			update, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return
			}
			pc.updatesCh <- GameUpdateMsg(update)
		}
	}()

	return nil
}

func realMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// g, _ := errgroup.WithContext(ctx)

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelInfo)

	cfg := &PongClientCfg{
		ServerAddr: *serverAddr,
	}
	pongConn, err := grpc.Dial(cfg.ServerAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer pongConn.Close()

	// Create a channel to listen for signals (e.g., SIGINT, SIGTERM).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	updatesCh := make(chan tea.Msg)

	pc := &pongClient{
		// ID:         clientID,
		cfg:        cfg,
		conn:       pongConn,
		pongClient: pong.NewPongGameClient(pongConn),
		updatesCh:  updatesCh,
	}

	// Perform authentication during initialization
	err = pc.StartNotifier()
	if err != nil {
		return fmt.Errorf("failed to StartNotifier: %w", err)
	}

	m := initialModel(pc, nil, nil)
	defer m.cancel()

	p := tea.NewProgram(m)

	if err := p.Start(); err != nil {
		return err
	}

	// Wait for a termination signal or context cancellation.
	select {
	case <-sigCh:
		log.Info("termination signal received")
		cancel() // Clean up resources.
	case <-ctx.Done():
		log.Info("context cancelled")
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
