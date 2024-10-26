package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/ndabAP/ping-pong/engine"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
)

type gameInstance struct {
	id          string
	engine      *canvas.CanvasEngine
	framesch    chan []byte
	inputch     chan []byte
	roundResult chan int32
	players     []*Player
	cleanedUp   bool
	running     bool
	ctx         context.Context
	cancel      context.CancelFunc
}

type gameManager struct {
	ID             *zkidentity.ShortID
	mu             sync.Mutex
	games          map[string]*gameInstance
	waitingRoom    *WaitingRoom
	playerSessions *PlayerSessions
	debug          bool
}

func (s *gameManager) cleanupGameInstance(instance *gameInstance) {
	if !instance.cleanedUp {
		instance.cleanedUp = true
		instance.cancel()
		close(instance.framesch)
		close(instance.inputch)
		close(instance.roundResult)
	}

	for gameID, game := range s.games {
		if game == instance {
			delete(s.games, gameID)
			serverLogger.Printf("[SERVER] Game %s cleaned up", gameID)
			break
		}
	}
}

func (gm *gameManager) getPlayerGame(clientID zkidentity.ShortID) *gameInstance {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for _, game := range gm.games {
		for _, player := range game.players {
			if player.ID == clientID {
				return game
			}
		}
	}

	return nil
}

func (s *gameManager) startGame(ctx context.Context, players []*Player) *gameInstance {
	s.mu.Lock()
	defer s.mu.Unlock()
	gameID := generateGameID()

	newGameInstance := s.startNewGame(ctx, players, gameID)
	s.games[gameID] = newGameInstance

	return newGameInstance
}

func (s *gameManager) startNewGame(ctx context.Context, players []*Player, id string) *gameInstance {
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(3, 3),
	)

	players[0].playerNumber = 1
	players[1].playerNumber = 2

	canvasEngine := canvas.New(game)
	canvasEngine.SetDebug(s.debug).SetFPS(*fps)

	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 10)
	roundResult := make(chan int32)
	instanceCtx, cancel := context.WithCancel(ctx)
	instance := &gameInstance{
		id:          id,
		engine:      canvasEngine,
		framesch:    framesch,
		inputch:     inputch,
		roundResult: roundResult,
		running:     true,
		ctx:         instanceCtx,
		cancel:      cancel,
		players:     players,
	}

	return instance
}

func (g *gameInstance) Run() {
	g.running = true
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serverLogger.Printf("Recovered from panic in NewRound: %v", r)
			}
		}()

		// Run a new round only if the game is still running
		if g.running {
			g.engine.NewRound(g.ctx, g.framesch, g.inputch, g.roundResult)
		}
	}()

	go func() {
		for winnerID := range g.roundResult {
			if !g.running {
				break
			}

			// Handle the result of each round
			player := g.handleRoundResult(winnerID)
			fmt.Printf("player: %+v\n", player)

			// Check if the game should continue or end
			if g.shouldEndGame() {
				g.Stop() // Stop the game if necessary
				break
			}
		}
	}()
}

func (g *gameInstance) handleRoundResult(winner int32) *Player {
	for _, player := range g.players {
		if player.playerNumber == winner {
			player.score++
			return player
		}
	}

	return nil
}

func (g *gameInstance) Stop() {
	g.running = false
}

const maxScore = 10

func (g *gameInstance) shouldEndGame() bool {
	for _, player := range g.players {
		// Check if any player has reached the max score
		if player.score >= maxScore {
			serverLogger.Printf("Game ending: Player %s reached the maximum score of %d", player.ID, maxScore)
			return true
		}
	}

	// Add other conditions as needed, e.g., time limit or disconnection
	if g.isTimeout() {
		serverLogger.Printf("Game ending: Timeout reached")
		return true
	}

	// Return false if none of the end conditions are met
	return false
}

// isTimeout checks if the game duration has exceeded a set limit
func (g *gameInstance) isTimeout() bool {
	// For example, a simple time limit check
	// const maxGameDuration = 10 * time.Minute
	// return time.Since(g.startTime) >= maxGameDuration
	return false
}
