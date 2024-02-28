package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	ball        position
	velocity    position
	paddleLeft  int
	paddleRight int
	scoreLeft   int
	scoreRight  int
}

type position struct {
	x, y int
}

type tickMsg struct{}

const (
	paddleHeight = 4
	winWidth     = 80
	winHeight    = 20
)

func initialModel() model {
	return model{
		ball: position{winWidth / 2, winHeight / 2},
		velocity: position{
			x: 2*rand.Intn(2) - 1, // Ensure non-zero initial x velocity
			y: 2*rand.Intn(2) - 1, // Ensure non-zero initial y velocity
		},
		paddleLeft:  winHeight/2 - paddleHeight/2,
		paddleRight: winHeight/2 - paddleHeight/2,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "w", "s", "o", "l":
			m = handleInput(m, msg.String())
		}
	case tea.WindowSizeMsg:
		// Can handle window resizing here if needed
	case tickMsg:
		m.updateGame()
	}
	return m, tickCmd()
}

func handleInput(m model, key string) model {
	switch key {
	case "w":
		if m.paddleLeft > 0 {
			m.paddleLeft--
		}
	case "s":
		if m.paddleLeft < winHeight-paddleHeight {
			m.paddleLeft++
		}
	case "o":
		if m.paddleRight > 0 {
			m.paddleRight--
		}
	case "l":
		if m.paddleRight < winHeight-paddleHeight {
			m.paddleRight++
		}
	}
	return m
}

func (m *model) updateGame() {
	// Update ball position
	m.ball.x += m.velocity.x
	m.ball.y += m.velocity.y

	// Collision with top and bottom
	if m.ball.y <= 0 || m.ball.y >= winHeight-1 {
		m.velocity.y *= -1
	}

	// Check collision with left paddle
	if m.ball.x == 2 && m.ball.y >= m.paddleLeft && m.ball.y <= m.paddleLeft+paddleHeight {
		m.velocity.x = -m.velocity.x              // Invert X direction
		adjustBallVelocityAfterPaddleHit(m, true) // Adjust velocity based on hit position
	}

	// Check collision with right paddle
	if m.ball.x == winWidth-3 && m.ball.y >= m.paddleRight && m.ball.y <= m.paddleRight+paddleHeight {
		m.velocity.x = -m.velocity.x               // Invert X direction
		adjustBallVelocityAfterPaddleHit(m, false) // Adjust velocity based on hit position
	}

	// Limit the velocity to prevent excessive speed
	limitBallVelocity(m)

	// Scoring
	if m.ball.x <= 0 {
		m.scoreRight++
		m.resetBall()
	} else if m.ball.x >= winWidth-1 {
		m.scoreLeft++
		m.resetBall()
	}
}

func adjustBallVelocityAfterPaddleHit(m *model, isLeftPaddle bool) {
	paddleCenter := 0
	if isLeftPaddle {
		paddleCenter = m.paddleLeft + paddleHeight/2
	} else {
		paddleCenter = m.paddleRight + paddleHeight/2
	}
	hitPos := m.ball.y - paddleCenter
	// Adjust Y velocity based on hit position
	// This calculation can be refined to change the ball's trajectory more significantly
	m.velocity.y += hitPos / 2
}

func limitBallVelocity(m *model) {
	maxVelocityX := 3 // Adjusted maximum speed in the x direction
	maxVelocityY := 4 // Adjusted maximum speed in the y direction

	if m.velocity.x > maxVelocityX {
		m.velocity.x = maxVelocityX
	} else if m.velocity.x < -maxVelocityX {
		m.velocity.x = -maxVelocityX
	}

	if m.velocity.y > maxVelocityY {
		m.velocity.y = maxVelocityY
	} else if m.velocity.y < -maxVelocityY {
		m.velocity.y = -maxVelocityY
	}
}

func (m *model) resetBall() {
	m.ball = position{winWidth / 2, winHeight / 2}
	m.velocity = position{
		x: 2*rand.Intn(2) - 1, // This will ensure x is either -1 or 1
		y: 2*rand.Intn(2) - 1, // Similarly, ensures y is either -1 or 1
	}
}

func (m model) View() string {
	// Call a separate rendering function
	return renderGame(m)
}

func renderGame(m model) string {
	// This function now contains what used to be in View
	var gameView string
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidth; x++ {
			switch {
			case x == m.ball.x && y == m.ball.y:
				gameView += "O"
			case x == 0 && y >= m.paddleLeft && y < m.paddleLeft+paddleHeight:
				gameView += "|"
			case x == winWidth-1 && y >= m.paddleRight && y < m.paddleRight+paddleHeight:
				gameView += "|"
			default:
				gameView += " "
			}
		}
		gameView += "\n"
	}
	gameView += fmt.Sprintf("Score: %d - %d\n", m.scoreLeft, m.scoreRight)
	gameView += "Controls: W/S and O/L - Q to quit"
	return gameView
}

func tickCmd() tea.Cmd {
	return tea.Every(time.Second/20, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Init() tea.Cmd {
	// Start the ticking process as soon as the program starts
	return tickCmd()
}

// func main() {
// 	rand.Seed(time.Now().UnixNano())
// 	p := tea.NewProgram(initialModel())
// 	if err := p.Start(); err != nil {
// 		fmt.Fprintf(os.Stderr, "Could not start the program: %v", err)
// 		os.Exit(1)
// 	}
// }
