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
	"strings"
	"sync"

	"pingpongexample/pongrpc/grpc/pong"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ID = zkidentity.ShortID

type appMode int

const (
	chatMode appMode = iota
	gameMode
)

var (
	serverAddr         = flag.String("server_addr", "localhost:50051", "The server address in the format of host:port")
	flagURL            = flag.String("url", "wss://127.0.0.1:7676/ws", "URL of the websocket endpoint")
	flagServerCertPath = flag.String("servercert", "~/.brclient/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "~/.brclient/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "~/.brclient/rpc-client.key", "Path to rpc-client.key file")
)

type PongClientCfg struct {
	ServerAddr string      // Address of the Pong server
	Log        slog.Logger // Application's logger
}

type pongClient struct {
	ID string

	playerNumber int32
	cfg          *PongClientCfg
	conn         *grpc.ClientConn
	pongClient   pong.PongGameClient
}

func (pc *pongClient) SendInput(input string) error {
	// Example client ID; replace "yourClientID" with the actual client ID
	ctx := attachClientIDToContext(context.Background(), pc.ID)

	_, err := pc.pongClient.SendInput(ctx, &pong.PlayerInput{
		Input:    input,
		PlayerId: pc.ID,
	})
	if err != nil {
		pc.cfg.Log.Errorf("Error sending input: %v", err)
		return fmt.Errorf("error sending input: %v", err)
	}
	return nil
}

func (pc *pongClient) StreamUpdates() error {
	// Use the same client ID as in SendInput
	ctx := attachClientIDToContext(context.Background(), pc.ID)

	stream, err := pc.pongClient.StreamUpdates(ctx, &pong.GameStreamRequest{})
	if err != nil {
		pc.cfg.Log.Errorf("Error creating updates stream: %v", err)
		return fmt.Errorf("error creating updates stream: %v", err)
	}
	for {
		update, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			pc.cfg.Log.Errorf("Error receiving update: %v", err)
			return fmt.Errorf("error receiving update: %v", err)
		}
		// Process the game update here
		fmt.Printf("Game Update: %+v\n", update)
	}
	return nil
}

// GameUpdateMsg is used to send game state updates through the Bubble Tea program
type GameUpdateMsg *pong.GameUpdate

type model struct {
	mode           appMode
	gameStateMutex sync.Mutex
	gameState      *pong.GameUpdate
	err            error
	ctx            context.Context
	cancel         context.CancelFunc
	pc             *pongClient
	chatClient     *types.ChatServiceClient // Assuming this is the correct type for your chat client
	versionClient  *types.VersionServiceClient
}

func initialModel(pc *pongClient, chatClient *types.ChatServiceClient, versionClient *types.VersionServiceClient) model {
	ctx, cancel := context.WithCancel(context.Background())
	return model{
		mode:          chatMode,
		ctx:           ctx,
		cancel:        cancel,
		pc:            pc,
		chatClient:    chatClient,
		versionClient: versionClient,
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {

		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if the game is in game mode
		if m.mode == gameMode {
			// Directly handle game inputs here
			switch msg.String() {
			case "w", "s", "up", "down":
				// Call handleGameInput for handling game-related inputs
				return m, m.handleGameInput(msg)
			}
		}

		switch msg.Type {
		case tea.KeyEsc:
			// Toggle between modes
			if m.mode == chatMode {
				m.mode = gameMode
			} else if m.mode == gameMode {
				m.mode = chatMode
			}
		case tea.KeySpace:
			// Handle space separately if needed, e.g., to make the client ready
			if m.mode == gameMode {
				return m, m.makeClientReady()
			}
		case tea.KeyRunes:
			// Handle character input for quitting the game
			if msg.String() == "q" {
				return m, tea.Quit
			}
		}
		return m, nil
	case types.VersionResponse:
		fmt.Printf("AppName: %s\nversion: %s\nGoRuntime: %s\n", msg.AppName, msg.AppVersion, msg.GoRuntime)
		// Handle other message types as before...
	case GameUpdateMsg:
		m.gameStateMutex.Lock()
		m.gameState = msg
		m.gameStateMutex.Unlock()
		return m, nil
	}

	// Keep the rest of your switch cases as they are...

	return m, nil
}

func (m *model) makeClientReady() tea.Cmd {
	// Example: Signal to the server that this client is ready. Adjust according to your server's API.
	log.Println("Client signaling readiness")
	// Replace the following with the actual call to your server.
	go func() {
		log.Printf("pongClient signaling readiness: %v", m.pc.pongClient)
		_, err := m.pc.pongClient.SignalReady(context.Background(), &pong.SignalReadyRequest{ClientId: m.pc.ID})
		if err != nil {
			log.Printf("Error signaling readiness: %v", err)
		}
	}()
	return nil
}

func (m *model) handleGameInput(msg tea.KeyMsg) tea.Cmd {
	// Example: Handling up and down inputs
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

func (m model) View() string {
	var b strings.Builder
	if m.mode == chatMode {
		fmt.Fprintln(&b, "Chat mode: Press space to join the game. 'q' to quit.")
		// Include logic to display chat messages
	} else if m.mode == gameMode {
		fmt.Fprintln(&b, "Game mode: 'q' to return to chat.")
		if m.gameState == nil {
			return "Waiting for game to start..."
		}

		var gameView strings.Builder
		for y := 0; y < int(m.gameState.GameHeight); y++ {
			for x := 0; x < int(m.gameState.GameWidth); x++ {
				// Round ball's position before comparing
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
		gameView.WriteString("Controls: W/S and O/L - Q to quit")

		return gameView.String()
	}

	return b.String()
}

func (pc *pongClient) receiveUpdates(m *model, p *tea.Program) error {
	ctx := attachClientIDToContext(context.Background(), pc.ID)
	stream, err := pc.pongClient.StreamUpdates(ctx, &pong.GameStreamRequest{PlayerId: pc.ID})
	if err != nil {
		log.Fatalf("could not subscribe to game updates: %v", err)
		return err
	}

	for {
		updateBytes, err := stream.Recv()
		if err == io.EOF || m.ctx.Err() != nil {
			break // Stream closed by server or context canceled
		}
		if err != nil {
			log.Fatalf("error receiving game update bytes: %v", err)
			return err
		}

		var update pong.GameUpdate
		if err := json.Unmarshal(updateBytes.Data, &update); err != nil {
			log.Fatalf("error unmarshalling game update: %v", err)
			return err
		}

		p.Send(GameUpdateMsg(&update)) // Send update to Bubble Tea program
	}
	return nil
}

// attachClientIDToContext creates a new context with the client-id metadata.
func attachClientIDToContext(ctx context.Context, clientID string) context.Context {
	md := metadata.New(map[string]string{
		"client-id": clientID, // The key must match what the server expects
	})
	return metadata.NewOutgoingContext(ctx, md)
}

func realMain() error {

	flag.Parse()
	*flagServerCertPath = expandPath(*flagServerCertPath)
	*flagClientCertPath = expandPath(*flagClientCertPath)
	*flagClientKeyPath = expandPath(*flagClientKeyPath)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, gctx := errgroup.WithContext(ctx)

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelInfo)

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(*flagURL),
		jsonrpc.WithServerTLSCertPath(*flagServerCertPath),
		jsonrpc.WithClientTLSCert(*flagClientCertPath, *flagClientKeyPath),
		jsonrpc.WithClientLog(log),
	)
	if err != nil {
		return err
	}

	versionClient := types.NewVersionServiceClient(c)
	var clientID string
	g.Go(func() error { return c.Run(gctx) })

	resp := &types.PublicIdentity{}
	err = versionClient.Public(ctx, &types.PublicIdentityReq{}, resp)
	if err != nil {
		return fmt.Errorf("failed to get public identity: %w", err)
	}

	clientID = hex.EncodeToString(resp.Identity[:])

	// Check if clientID is still empty here, which it shouldn't be now
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

	pc := &pongClient{
		ID:         clientID,
		cfg:        cfg,
		conn:       pongConn,
		pongClient: pong.NewPongGameClient(pongConn),
	}

	m := initialModel(pc, nil, nil)
	p := tea.NewProgram(m)

	go pc.receiveUpdates(&m, p)

	// Start the Bubble Tea program
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
