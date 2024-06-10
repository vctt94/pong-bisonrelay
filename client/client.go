package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
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
	serverAddr = flag.String("server_addr", "localhost:50051", "The server address in the format of host:port")
	rpcurl     = flag.String("rpcurl", "127.0.0.1:7676", "URL of the RPC endpoint without protocol")
	brdatadir  = flag.String("brdatadir", "", "Directory containing the certificates and keys")
)

var (
	flagServerCertPath string
	flagClientCertPath string
	flagClientKeyPath  string
)

func init() {
	flag.Parse()

	if *brdatadir != "" {
		flagServerCertPath = filepath.Join(*brdatadir, "rpc.cert")
		flagClientCertPath = filepath.Join(*brdatadir, "rpc-client.cert")
		flagClientKeyPath = filepath.Join(*brdatadir, "rpc-client.key")
	} else {
		flagServerCertPath = expandPath("~/.brclient/rpc.cert")
		flagClientCertPath = expandPath("~/.brclient/rpc-client.cert")
		flagClientKeyPath = expandPath("~/.brclient/rpc-client.key")
	}
}

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
		return fmt.Errorf("error creating game started stream: %v", err)
	}

	go func() {
		for {
			started, err := gameStartedStream.Recv()
			pc.playerNumber = started.PlayerNumber
			if err == io.EOF {
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
		return fmt.Errorf("error sending input: %v", err)
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
		return fmt.Errorf("error signaling readiness: %v", err)
	}

	err = pc.initializeStream(ctx)
	if err != nil {
		return fmt.Errorf("error initializing stream: %v", err)
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
		return fmt.Errorf("error creating updates stream: %v", err)
	}

	// Set the stream before starting the goroutine
	pc.stream = stream

	// Use a separate goroutine to handle the stream
	go func() {
		for {
			update, err := stream.Recv()
			if err == io.EOF {
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
	g, gctx := errgroup.WithContext(ctx)

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelInfo)

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL("wss://"+*rpcurl+"/ws"),
		jsonrpc.WithServerTLSCertPath(flagServerCertPath),
		jsonrpc.WithClientTLSCert(flagClientCertPath, flagClientKeyPath),
		jsonrpc.WithClientLog(log),
	)
	if err != nil {
		return err
	}

	chatClient := types.NewChatServiceClient(c)
	var clientID string
	g.Go(func() error { return c.Run(gctx) })

	resp := &types.PublicIdentity{}
	err = chatClient.UserPublicIdentity(ctx, &types.PublicIdentityReq{}, resp)
	if err != nil {
		return fmt.Errorf("failed to get public identity: %w", err)
	}

	clientID = hex.EncodeToString(resp.Identity[:])
	if clientID == "" {
		return fmt.Errorf("client ID is empty after fetching")
	}
	cfg := &PongClientCfg{
		ServerAddr: *serverAddr,
	}
	pongConn, err := grpc.Dial(cfg.ServerAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer pongConn.Close()
	updatesCh := make(chan tea.Msg)

	pc := &pongClient{
		ID:         clientID,
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

	return nil
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
