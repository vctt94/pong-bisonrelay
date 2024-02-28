package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"strings"

	"pingpongexample/pongrpc/grpc/types/pong"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/decred/slog"
	"google.golang.org/grpc"
)

type PongClientCfg struct {
	ServerAddr string      // Address of the Pong server
	Log        slog.Logger // Application's logger
}

type pongClient struct {
	cfg        *PongClientCfg
	conn       *grpc.ClientConn
	pongClient pong.PongGameClient
	pc         *pongClient // Add this line
}

func (pc *pongClient) SendInput(input string) error {
	// Example input could be "ArrowUp" or "ArrowDown"
	_, err := pc.pongClient.SendInput(context.Background(), &pong.PlayerInput{Input: input})
	if err != nil {
		pc.cfg.Log.Errorf("Error sending input: %v", err)
		return fmt.Errorf("error sending input: %v", err)
	}
	return nil
}

func (pc *pongClient) StreamUpdates() error {
	stream, err := pc.pongClient.StreamUpdates(context.Background(), &pong.GameStreamRequest{})
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
	gameState *pong.GameUpdate
	err       error
	ctx       context.Context
	cancel    context.CancelFunc
	pc        *pongClient // Add this line
}

func initialModel(pc *pongClient) model {
	ctx, cancel := context.WithCancel(context.Background())
	return model{ctx: ctx, cancel: cancel, pc: pc}
}

func (m model) Init() tea.Cmd {
	return nil // Return nil if no initial command is needed
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GameUpdateMsg:
		m.gameState = msg
		return m, nil // No command to return

	case tea.KeyMsg:
		// Handle quitting the program
		if msg.String() == "q" {
			m.cancel() // Cancel the context to stop receiving updates
			return m, tea.Quit
		}

		// Example: handle up and down key presses for player 1
		var cmd tea.Cmd
		switch msg.String() {
		case "w", "up":
			// Send "ArrowUp" to the server
			cmd = func() tea.Msg {
				err := m.pc.SendInput("ArrowUp")
				if err != nil {
					log.Println("Error sending input:", err)
				}
				return nil // You might want to define a message type for handling errors
			}
		case "s", "down":
			// Send "ArrowDown" to the server
			cmd = func() tea.Msg {
				err := m.pc.SendInput("ArrowDown")
				if err != nil {
					log.Println("Error sending input:", err)
				}
				return nil // Similar to above, handle errors appropriately
			}
		}

		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.gameState == nil {
		return "Waiting for game to start..."
	}

	var gameView strings.Builder
	for y := 0; y < int(m.gameState.GameHeight); y++ {
		for x := 0; x < int(m.gameState.GameWidth); x++ {
			// Round ball's position before comparing
			ballX := int(math.Round(float64(m.gameState.BallX)))
			ballY := int(math.Round(float64(m.gameState.BallY)))
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

func (pc *pongClient) receiveUpdates(m *model, client pong.PongGameClient, playerID string, p *tea.Program) {
	stream, err := client.StreamUpdates(m.ctx, &pong.GameStreamRequest{PlayerId: playerID})
	if err != nil {
		log.Fatalf("could not subscribe to game updates: %v", err)
	}

	for {
		update, err := stream.Recv()
		if err == io.EOF || m.ctx.Err() != nil {
			break // Stream closed by server or context canceled
		}
		if err != nil {
			log.Fatalf("error receiving game update: %v", err)
		}

		p.Send(GameUpdateMsg(update)) // Send update to Bubble Tea program
	}
}

func main() {

	cfg := &PongClientCfg{
		ServerAddr: "localhost:50051",
	}
	conn, err := grpc.Dial(cfg.ServerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := pong.NewPongGameClient(conn)
	pc := &pongClient{
		cfg:        cfg,
		conn:       conn,
		pongClient: client,
	}

	m := initialModel(pc)
	p := tea.NewProgram(m)

	// Start receiving game updates in a separate goroutine
	go pc.receiveUpdates(&m, client, "player1", p)

	if err := p.Start(); err != nil {
		log.Fatalf("Could not start Bubble Tea program: %v", err)
	}
}
