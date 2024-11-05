package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vctt94/pong-bisonrelay/client"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"golang.org/x/sync/errgroup"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/companyzero/bisonrelay/clientrpc/jsonrpc"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
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
	serverAddr = flag.String("server_addr", "localhost:50051", "The server address in the format of host:port")
	// brdatadir          = flag.String("brdatadir", "", "Directory containing the certificates and keys")
	flagURL            = flag.String("url", "wss://127.0.0.1:7777/ws", "URL of the websocket endpoint")
	flagServerCertPath = flag.String("servercert", "/home/vctt/brclientdir/rpc.cert", "Path to rpc.cert file")
	flagClientCertPath = flag.String("clientcert", "/home/vctt/brclientdir/rpc-client.cert", "Path to rpc-client.cert file")
	flagClientKeyPath  = flag.String("clientkey", "/home/vctt/brclientdir/rpc-client.key", "Path to rpc-client.key file")
	rpcUser            = flag.String("rpcuser", "rpcuser", "RPC user for basic authentication")
	rpcPass            = flag.String("rpcpass", "rpcpass", "RPC password for basic authentication")
)

type appstate struct {
	sync.Mutex
	mode              appMode
	isReady           bool
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

	log     slog.Logger
	players []*pong.Player

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

func (m *appstate) Init() tea.Cmd {
	m.msgCh = make(chan tea.Msg)

	// Initialize the viewport with zero dimensions; we'll set them upon receiving the window size
	m.viewport = viewport.New(0, 0)

	return tea.Batch(
		m.listenForUpdates(),
		tea.EnterAltScreen, // Optional: Use the alternate screen buffer
	)
}

func (m *appstate) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Lock()
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.viewport.SetContent(m.View())
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
				m.createRoom()
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
					m.log.Errorf("Error joining room: %v", err)
				}
			}
			return m, nil
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
			m.mode = gameMode
			m.makeClientReady()
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
		// fmt.Printf("game update: %+v\n", gameUpdate)

		return m, m.waitForMsg()

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
		m.log.Errorf("Error fetching waiting rooms: %v", err)
		return err
	}
	m.waitingRooms = wr
	return nil
}

func (m *appstate) createRoom() error {
	var err error
	_, err = m.pc.CreatewaitingRoom(m.ctx)
	if err != nil {
		return err
	}

	return nil
}
func (m *appstate) joinRoom(roomID string) error {

	// Send request to join the specified room
	res, err := m.pc.JoinWaitingRoom(m.ctx, roomID)
	if err != nil {
		m.log.Errorf("Error joining room %s: %v", roomID, err)
		return err
	}

	m.currentWR = res.Wr
	m.mode = gameMode

	return nil
}

func (m *appstate) makeClientReady() tea.Cmd {
	m.log.Debugf("Client signaling readiness")
	m.isReady = true
	err := m.pc.SignalReady()
	if err != nil {
		m.log.Errorf("Error signaling readiness: %v", err)
	}
	return nil
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
				m.log.Errorf("Error sending game input: %v", err)
			}
		}
		return nil
	}
}

func (m *appstate) View() string {
	var b strings.Builder

	// Build the header
	b.WriteString("=== Pong Game Client ===\n")
	if m.notification != "" {
		b.WriteString(fmt.Sprintf("Notification: %s\n", m.notification))
	} else {
		b.WriteString("No new notifications.\n")
	}
	b.WriteString(fmt.Sprintf("Player: %s\nBet Amount: %.8f\nStatus Ready: %t\n\nCurrent Room: %s\n", m.pc.ID, m.betAmount, m.isReady, m.currentWR))
	b.WriteString("Use the following keys to navigate:\n")

	// Display different modes based on the current app mode
	switch m.mode {
	case gameIdle:
		b.WriteString("\n[Idle Mode]\n")
		b.WriteString("Press 'space' to get ready for the game.\n")
		b.WriteString("Press 'l' to list available rooms.\n")
		b.WriteString("Press 'c' to create a room (requires bet > 0).\n")
		b.WriteString("Press 'j' to join an existing room.\n")
		b.WriteString("Press 'esc' to quit.\n")

	case gameMode:
		b.WriteString("\n[Game Mode]\n")
		b.WriteString("Press 'esc' to return to chat.\n")
		b.WriteString("Use W/S or Arrow Keys to move.\n")

		// Display game state if available
		if m.gameState != nil {
			var gameView strings.Builder
			for y := 0; y < int(m.gameState.GameHeight); y++ {
				for x := 0; x < int(m.gameState.GameWidth); x++ {
					ballX := int(math.Round(m.gameState.BallX))
					ballY := int(math.Round(m.gameState.BallY))
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
			b.WriteString(gameView.String())
		} else {
			b.WriteString("Waiting for game to start... Not all players ready\nHit [space] to get ready\n")
		}

	case listRooms:
		b.WriteString("\n[List Rooms Mode]\n")
		if len(m.waitingRooms) > 0 {
			for i, room := range m.waitingRooms {
				b.WriteString(fmt.Sprintf("%d: Room ID %s - Bet Price: %.8f\n", i+1, room.Id, room.BetAmt))
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
				b.WriteString(fmt.Sprintf("%s %d: Room ID %s - Bet Price: %.8f\n", indicator, i+1, room.Id, room.BetAmt))
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bknd := slog.NewBackend(os.Stderr)
	log := bknd.Logger("EXMP")
	log.SetLevel(slog.LevelDebug)

	g, gctx := errgroup.WithContext(ctx)

	log.SetLevel(slog.LevelInfo)

	c, err := jsonrpc.NewWSClient(
		jsonrpc.WithWebsocketURL(*flagURL),
		jsonrpc.WithServerTLSCertPath(*flagServerCertPath),
		jsonrpc.WithClientTLSCert(*flagClientCertPath, *flagClientKeyPath),
		jsonrpc.WithClientLog(log),
		jsonrpc.WithClientBasicAuth(*rpcUser, *rpcPass),
	)
	if err != nil {
		return err
	}
	g.Go(func() error { return c.Run(gctx) })

	var zkShortID zkidentity.ShortID
	chat := types.NewChatServiceClient(c)
	req := &types.PublicIdentityReq{}
	var publicIdentity types.PublicIdentity
	err = chat.UserPublicIdentity(ctx, req, &publicIdentity)
	if err != nil {
		return fmt.Errorf("failed to get user public identity: %v", err)
	}

	clientID := hex.EncodeToString(publicIdentity.Identity[:])
	copy(zkShortID[:], clientID)
	as := &appstate{
		ctx:    ctx,
		cancel: cancel,
		log:    log,
	}
	// Setup notification handlers.
	ntfns := client.NewNotificationManager()
	ntfns.RegisterSync(client.OnWRCreatedNtfn(func(wr *pong.WaitingRoom, ts time.Time) {
		as.Lock()
		as.waitingRooms = append(as.waitingRooms, wr)
		as.currentWR = wr
		as.betAmount = wr.BetAmt
		as.Unlock()
		as.notification = fmt.Sprintf("New waiting room created with ID: %s", wr.Id)

		go func() {
			select {
			case as.createdWRChan <- struct{}{}:
			case <-as.ctx.Done():
			}
		}()
	}))

	ntfns.Register(client.OnBetAmtChangedNtfn(func(playerID string, betAmt float64, ts time.Time) {
		// Update bet amount for the player in the local state (e.g., as.Players).
		if clientID == playerID {
			as.notification = "bet amount updated"
			as.betAmount = betAmt
			as.msgCh <- client.UpdatedMsg{}
		}
		for i, p := range as.players {
			if p.Uid == playerID {
				as.Lock()
				as.players[i].BetAmount = betAmt
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

	pc, err := client.NewPongClient(clientID, &client.PongClientCfg{
		ServerAddr:    *serverAddr,
		ChatClient:    chat,
		Notifications: ntfns,
		Log:           log,
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
