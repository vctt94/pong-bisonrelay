package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vctt94/pong-bisonrelay/client"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/bisonbotkit/botclient"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

type ID = zkidentity.ShortID

type appMode int

var isF2p = true

const (
	gameIdle appMode = iota
	gameMode
	listRooms
	createRoom
	joinRoom
)

var (
	// serverAddr = flag.String("server_addr", "104.131.180.29:50051", "The server address in the format of host:port")
	serverAddr         = flag.String("server_addr", "", "The server address in the format of host:port")
	datadir            = flag.String("datadir", "", "Directory to load config file from")
	flagURL            = flag.String("url", "", "URL of the websocket endpoint")
	flagServerCertPath = flag.String("servercert", "", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "", "Path to rpc-client.key file")
	rpcUser            = flag.String("rpcuser", "", "RPC user for basic authentication")
	rpcPass            = flag.String("rpcpass", "", "RPC password for basic authentication")
	grpcServerCert     = flag.String("grpcservercert", "", "Path to grpc server.cert file")
)

type appstate struct {
	sync.Mutex
	mode              appMode
	gameState         *pong.GameUpdate
	ctx               context.Context
	err               error
	cancel            context.CancelFunc
	pc                *client.PongClient
	selectedRoomIndex int
	msgCh             chan tea.Msg
	viewport          viewport.Model
	createdWRChan     chan struct{}
	betAmtChangedChan chan struct{}

	isGameRunning bool
	log           slog.Logger
	players       []*pong.Player

	// player current bet amt
	betAmount float64

	currentWR *pong.WaitingRoom

	waitingRooms []*pong.WaitingRoom

	notification string
}

func (m *appstate) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		// Start a goroutine to listen for updates
		go func() {
			for msg := range m.pc.UpdatesCh {
				m.msgCh <- msg
			}
		}()
		return nil
	}
}

func (m *appstate) listenForErrors() tea.Cmd {
	return func() tea.Msg {
		go func() {
			for err := range m.pc.ErrorsCh {
				m.msgCh <- fmt.Sprintf("Error: %v", err)
			}
		}()
		return nil
	}
}

func (m *appstate) Init() tea.Cmd {
	m.msgCh = make(chan tea.Msg)

	// Initialize the viewport with zero dimensions; we'll set them upon receiving the window size
	m.viewport = viewport.New(0, 0)

	return tea.Batch(
		m.listenForUpdates(),
		m.listenForErrors(),
		tea.EnterAltScreen, // Optional: Use the alternate screen buffer
	)
}

func (m *appstate) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Lock()
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.Unlock()
		return m, nil
	case client.UpdatedMsg:
		// Simply return the model to refresh the view
		return m, m.waitForMsg()
	case tea.KeyMsg:
		switch msg.String() {
		case "l":
			// Switch to list rooms mode
			m.mode = listRooms
			m.listWaitingRooms()
			return m, nil
		case "c":
			// Switch to create room mode if player has a bet
			if m.betAmount > 0 || isF2p {
				err := m.createRoom()
				if err != nil {
					m.notification = fmt.Sprintf("Error creating room: %v", err)
				}
				return m, nil
			} else {
				m.notification = "Bet amount must be > 0 to create a room."
			}
		case "j":
			// Switch to join room mode
			m.mode = joinRoom
			m.selectedRoomIndex = 0
			m.listWaitingRooms()
			return m, nil
		case "w", "s":
			if m.mode == gameMode {
				return m, m.handleGameInput(msg)
			}
		case "up":
			// Check mode to avoid duplicate handling of "up" key
			if m.mode == joinRoom && m.selectedRoomIndex > 0 {
				m.selectedRoomIndex--
			} else if m.mode == gameMode {
				return m, m.handleGameInput(msg)
			}
			return m, nil

		case "down":
			// Check mode to avoid duplicate handling of "down" key
			if m.mode == joinRoom && m.selectedRoomIndex < len(m.waitingRooms)-1 {
				m.selectedRoomIndex++
			} else if m.mode == gameMode {
				return m, m.handleGameInput(msg)
			}
			return m, nil
		case "enter":
			if m.mode == joinRoom && len(m.waitingRooms) > 0 {
				selectedRoom := m.waitingRooms[m.selectedRoomIndex]
				err := m.joinRoom(selectedRoom.Id)
				if err != nil {
					m.notification = fmt.Sprintf("Error joining room: %v", err)
				}
			}
			return m, nil
		case "q":
			// Leave the current waiting room
			if m.currentWR != nil && !m.isGameRunning {
				err := m.leaveRoom()
				if err != nil {
					m.notification = fmt.Sprintf("Error leaving room: %v", err)
				}
				return m, nil
			}
		}

		if m.mode == gameIdle && msg.Type == tea.KeyEsc {
			m.cancel()
			return m, tea.Quit
		}
		if msg.Type == tea.KeyF2 {
			m.mode = gameMode
			return m, nil
		}
		if msg.Type == tea.KeySpace {
			if m.pc.IsReady {
				// If already ready, set to unready
				err := m.makeClientUnready()
				if err != nil {
					m.notification = fmt.Sprintf("Error signaling unreadiness: %v", err)
					return m, nil
				}
			} else {
				// If not ready, set to ready
				m.mode = gameMode
				err := m.makeClientReady()
				if err != nil {
					m.notification = fmt.Sprintf("Error signaling readiness: %v", err)
					return m, nil
				}
			}
			return m, nil
		}
		if msg.Type == tea.KeyEsc {
			m.mode = gameIdle
			return m, nil
		}
	case *pong.GameUpdateBytes:
		var gameUpdate pong.GameUpdate
		if err := json.Unmarshal(msg.Data, &gameUpdate); err != nil {
			m.err = err
			return m, nil
		}
		m.Lock()
		m.gameState = &gameUpdate
		m.Unlock()

		return m, m.waitForMsg()
	case string:
		if strings.HasPrefix(msg, "Error:") {
			m.notification = msg
			return m, nil
		}

	}
	return m, m.waitForMsg()
}

func (m *appstate) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgCh
	}
}

func (m *appstate) listWaitingRooms() error {
	wr, err := m.pc.GetWaitingRooms()
	if err != nil {
		return err
	}
	m.waitingRooms = wr
	return nil
}

func (m *appstate) createRoom() error {
	var err error
	_, err = m.pc.CreateWaitingRoom(m.pc.ID, m.pc.BetAmt)
	if err != nil {
		m.log.Errorf("Error creating room: %v", err)
		return err
	}

	m.mode = gameMode

	return nil
}
func (m *appstate) joinRoom(roomID string) error {

	// Send request to join the specified room
	res, err := m.pc.JoinWaitingRoom(roomID)
	if err != nil {
		return err
	}

	m.currentWR = res.Wr
	m.mode = gameMode

	return nil
}

func (m *appstate) makeClientReady() error {
	return m.pc.SignalReady()
}

func (m *appstate) makeClientUnready() error {
	return m.pc.SignalUnready()
}

func (m *appstate) handleGameInput(msg tea.KeyMsg) tea.Cmd {
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
				m.log.Debugf("Error sending game input: %v", err)
				return err
			}
		}
		return nil
	}
}

func (m *appstate) leaveRoom() error {
	if m.currentWR == nil {
		return fmt.Errorf("not in a waiting room")
	}

	err := m.pc.LeaveWaitingRoom(m.currentWR.Id)
	if err != nil {
		return err
	}

	m.currentWR = nil
	m.mode = gameIdle
	m.notification = "Successfully left the waiting room"
	return nil
}

func (m *appstate) View() string {
	var b strings.Builder

	// Show the header and controls only if the game is not in game mode
	if !m.isGameRunning {
		// Build the header
		b.WriteString("========== Pong Game Client ==========\n\n")

		if m.notification != "" {
			b.WriteString(fmt.Sprintf("🔔 Notification: %s\n\n", m.notification))
		} else {
			b.WriteString("🔔 No new notifications.\n\n")
		}

		b.WriteString(fmt.Sprintf("👤 Player ID: %s\n", m.pc.ID))
		b.WriteString(fmt.Sprintf("💵 Bet Amount: %.8f\n", m.betAmount))
		b.WriteString(fmt.Sprintf("✅ Status Ready: %t\n", m.pc.IsReady))

		// Display the current room or show a placeholder if not in a room
		if m.currentWR != nil {
			b.WriteString(fmt.Sprintf("🏠 Current Room: %s\n\n", m.currentWR.Id))
		} else {
			b.WriteString("🏠 Current Room: None\n\n")
		}

		// Instructions
		b.WriteString("===== Controls =====\n")
		b.WriteString("Use the following keys to navigate:\n")
		b.WriteString("[L] - List rooms\n")
		b.WriteString("[C] - Create room\n")
		b.WriteString("[J] - Join room\n")
		b.WriteString("[Q] - Leave current room\n")
		b.WriteString("[Esc] - Exit\n")
		b.WriteString("====================\n\n")

		if !m.isGameRunning && m.currentWR != nil {
			if m.pc.IsReady {
				b.WriteString("[Space] - Toggle ready status (currently READY)\n")
			} else {
				b.WriteString("[Space] - Toggle ready status (currently NOT READY)\n")
			}
		}
	}

	// Switch based on the current mode
	switch m.mode {
	case gameIdle:
		b.WriteString("\n[Idle Mode]\n")

	case gameMode:
		b.WriteString("\n[Game Mode]\n")
		b.WriteString("Press 'Esc' to return to the main menu.\n")
		b.WriteString("Use W/S or Arrow Keys to move.\n\n")

		if m.gameState != nil {
			var gameView strings.Builder

			// Calculate header and footer sizes
			headerLines := countLines(b.String())
			footerLines := 2 // For the score and any additional messages

			// Calculate available space
			availableHeight := m.viewport.Height - headerLines - footerLines
			availableWidth := m.viewport.Width

			// Minimum game size constraints
			const minGameHeight = 5
			const minGameWidth = 10

			if availableHeight < minGameHeight || availableWidth < minGameWidth {
				b.WriteString("\n[Warning] Terminal window is too small to display the game.\n")
				b.WriteString("Please resize your window or use a larger terminal.\n")
				return b.String()
			}

			// Original game dimensions
			gameHeight := int(m.gameState.GameHeight)
			gameWidth := int(m.gameState.GameWidth)

			// Calculate scaling factors for width and height
			scaleY := float64(availableHeight) / float64(gameHeight)
			scaleX := float64(availableWidth) / float64(gameWidth)

			// Use the smaller scaling factor to ensure the game fits in both dimensions
			scale := math.Min(scaleX, scaleY)
			scale = math.Min(scale, 1.0) // Prevent upscaling

			// Scale the game elements
			scaledGameHeight := int(float64(gameHeight) * scale)
			scaledGameWidth := int(float64(gameWidth) * scale)

			// Ensure scaled dimensions do not exceed available space
			if scaledGameHeight > availableHeight {
				scaledGameHeight = availableHeight
			}
			if scaledGameWidth > availableWidth {
				scaledGameWidth = availableWidth
			}

			// Scale ball position
			ballX := int(math.Round(float64(m.gameState.BallX) * scale))
			ballY := int(math.Round(float64(m.gameState.BallY) * scale))

			// Scale paddle positions and sizes
			p1Y := int(math.Round(float64(m.gameState.P1Y) * scale))
			p1Height := int(math.Round(float64(m.gameState.P1Height) * scale))

			p2Y := int(math.Round(float64(m.gameState.P2Y) * scale))
			p2Height := int(math.Round(float64(m.gameState.P2Height) * scale))

			// Ensure positions are within bounds
			if ballX >= scaledGameWidth {
				ballX = scaledGameWidth - 1
			}
			if ballY >= scaledGameHeight {
				ballY = scaledGameHeight - 1
			}
			if p1Y+p1Height > scaledGameHeight {
				p1Height = scaledGameHeight - p1Y
			}
			if p2Y+p2Height > scaledGameHeight {
				p2Height = scaledGameHeight - p2Y
			}

			// Drawing the game
			for y := 0; y < scaledGameHeight; y++ {
				for x := 0; x < scaledGameWidth; x++ {
					switch {
					case x == ballX && y == ballY:
						gameView.WriteString("O")
					case x == 0 && y >= p1Y && y < p1Y+p1Height:
						gameView.WriteString("|")
					case x == scaledGameWidth-1 && y >= p2Y && y < p2Y+p2Height:
						gameView.WriteString("|")
					default:
						gameView.WriteString(" ")
					}
				}
				gameView.WriteString("\n")
			}

			// Append the score
			gameView.WriteString(fmt.Sprintf("Score: %d - %d\n", m.gameState.P1Score, m.gameState.P2Score))
			b.WriteString(gameView.String())
		} else {
			b.WriteString("Waiting for game to start... Not all players are ready.\nHit [Space] to get ready\n")
		}

	case listRooms:
		b.WriteString("\n[List Rooms Mode]\n")
		if len(m.waitingRooms) > 0 {
			for i, room := range m.waitingRooms {
				b.WriteString(fmt.Sprintf("%d: Room ID %s - Bet Price: %.8f\n", i+1, room.Id, float64(room.BetAmt)/1e11))
			}
		} else {
			b.WriteString("No rooms available.\n")
		}
		b.WriteString("Press 'esc' to go back to the main menu.\n")

	case createRoom:
		b.WriteString("\n[Create Room Mode]\n")
		b.WriteString("Creating a new room...\n")

	case joinRoom:
		b.WriteString("\n[Join Room Mode]\n")
		b.WriteString("Select a room to join. Use [up]/[down] to navigate and [enter] to join.\n")
		b.WriteString("Press [esc] to go back to the main menu.\n")

		if len(m.waitingRooms) > 0 {
			for i, room := range m.waitingRooms {
				indicator := " " // Indicator for selected room
				if i == m.selectedRoomIndex {
					indicator = ">" // Mark the selected room
				}
				b.WriteString(fmt.Sprintf("%s %d: Room ID %s - Bet Price: %.8f\n", indicator, i+1, room.Id, float64(room.BetAmt)/1e11))
			}
		} else {
			b.WriteString("No rooms available.\n")
		}

	default:
		b.WriteString("\nUnknown mode.\n")
	}

	// Set the viewport content to the built string
	m.viewport.SetContent(b.String())

	// Return the viewport's view
	return m.viewport.View()
}

func realMain() error {
	flag.Parse()
	if *datadir == "" {
		*datadir = utils.AppDataDir("pongclient", false)
	}
	cfg, err := config.LoadClientConfig(*datadir, "pongclient.conf")
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		os.Exit(1)
	}

	// Apply overrides from flags
	if *flagURL != "" {
		cfg.RPCURL = *flagURL
	}
	if *flagServerCertPath != "" {
		cfg.ServerCertPath = *flagServerCertPath
	}
	if *flagClientCertPath != "" {
		cfg.ClientCertPath = *flagClientCertPath
	}
	if *flagClientKeyPath != "" {
		cfg.ClientKeyPath = *flagClientKeyPath
	}
	if *rpcUser != "" {
		cfg.RPCUser = *rpcUser
	}
	if *rpcPass != "" {
		cfg.RPCPass = *rpcPass
	}
	if *serverAddr != "" {
		cfg.ServerAddr = *serverAddr
	}
	if *grpcServerCert != "" {
		cfg.GRPCServerCert = *grpcServerCert
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, gctx := errgroup.WithContext(ctx)

	useStdout := false
	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(*datadir, "logs", "pongclient.log"),
		DebugLevel:     cfg.Debug,
		MaxLogFiles:    10,
		MaxBufferLines: 1000,
		UseStdout:      &useStdout,
	})
	log := logBackend.Logger("Bot")
	c, err := botclient.NewClient(cfg, logBackend)
	if err != nil {
		return err
	}
	g.Go(func() error { return c.RPCClient.Run(gctx) })

	var zkShortID zkidentity.ShortID
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = c.Chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	copy(zkShortID[:], clientID)
	as := &appstate{
		ctx:    ctx,
		cancel: cancel,
		log:    log,
		mode:   gameIdle,
	}
	// Setup notification handlers.
	ntfns := client.NewNotificationManager()
	ntfns.RegisterSync(client.OnWRCreatedNtfn(func(wr *pong.WaitingRoom, ts time.Time) {
		as.Lock()
		as.waitingRooms = append(as.waitingRooms, wr)
		for _, p := range as.players {
			if p.Uid == clientID {
				as.currentWR = wr
				as.betAmount = float64(wr.BetAmt) / 1e11
				as.mode = gameMode
			}
		}
		as.Unlock()
		as.notification = fmt.Sprintf("New waiting room created: %s", wr.Id)

		go func() {
			as.msgCh <- client.UpdatedMsg{}
			select {
			case as.createdWRChan <- struct{}{}:
			case <-as.ctx.Done():
			}
		}()
	}))

	ntfns.Register(client.OnBetAmtChangedNtfn(func(playerID string, betAmt int64, ts time.Time) {
		// Update bet amount for the player in the local state (e.g., as.Players).
		if clientID == playerID {
			as.notification = "bet amount updated"
			as.betAmount = float64(betAmt) / 1e11
			as.msgCh <- client.UpdatedMsg{}
		}
		for i, p := range as.players {
			if p.Uid == playerID {
				as.Lock()
				as.players[i].BetAmt = betAmt
				as.Unlock()

				break
			}
		}
		go func() {
			select {
			case as.betAmtChangedChan <- struct{}{}:
			case <-as.ctx.Done():
			}
		}()
	}))

	ntfns.Register(client.OnGameStartedNtfn(func(id string, ts time.Time) {
		as.mode = gameMode
		as.isGameRunning = true
		as.notification = fmt.Sprintf("game started with ID %s", id)
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	ntfns.Register(client.OnPlayerJoinedNtfn(func(wr *pong.WaitingRoom, ts time.Time) {
		as.currentWR = wr
		as.notification = "new player joined your waiting room"
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	ntfns.Register(client.OnGameEndedNtfn(func(gameID, msg string, ts time.Time) {
		as.notification = fmt.Sprintf("game %s ended\n%s", gameID, msg)
		as.betAmount = 0
		as.isGameRunning = false
		as.mode = gameIdle
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	ntfns.Register(client.OnPlayerLeftNtfn(func(wr *pong.WaitingRoom, playerID string, ts time.Time) {
		if playerID == clientID {
			as.currentWR = nil
			as.notification = "You left the waiting room"
		} else {
			as.currentWR = wr
			as.notification = fmt.Sprintf("Player %s left the waiting room", playerID)
		}
		go func() {
			as.msgCh <- client.UpdatedMsg{}
		}()
	}))

	pc, err := client.NewPongClient(clientID, &client.PongClientCfg{
		ServerAddr:    cfg.ServerAddr,
		Notifications: ntfns,
		Log:           log,
		GRPCCertPath:  cfg.GRPCServerCert,
	})
	if err != nil {
		return fmt.Errorf("failed to create pong client: %v", err)
	}
	as.pc = pc

	// Test the connection immediately after creating the client
	_, err = pc.GetWaitingRooms()
	if err != nil {
		return fmt.Errorf("gRPC server connection failed: %v", err)
	}

	// Start the notifier in a goroutine
	g.Go(func() error { return pc.StartNotifier(ctx) })

	defer as.cancel()

	p := tea.NewProgram(as)

	_, err = p.Run()
	if err != nil {
		return err
	}

	return g.Wait()
}

func main() {
	err := realMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
