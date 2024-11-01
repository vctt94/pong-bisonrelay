package server

import (
	"context"
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	canvas "github.com/vctt94/pong-bisonrelay/pong"
)

const maxScore = 3

type gameInstance struct {
	sync.Mutex
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
	winner      *zkidentity.ShortID
	// betAmt sum of total bets
	betAmt float64

	log slog.Logger
}

type gameManager struct {
	sync.RWMutex

	ID             *zkidentity.ShortID
	games          map[string]*gameInstance
	waitingRooms   []*WaitingRoom
	playerSessions *PlayerSessions

	debug slog.Level
	log   slog.Logger
}

func (g *gameManager) GetWaitingRoom(roomID string) *WaitingRoom {
	g.RLock()
	defer g.RUnlock()

	for _, room := range g.waitingRooms {
		if room.ID == roomID {
			return room
		}
	}
	return nil
}

func (gm *gameManager) RemoveWaitingRoom(roomID string) {
	for i, room := range gm.waitingRooms {
		if room.ID == roomID {
			// Remove the room by appending the elements before and after it
			gm.waitingRooms = append(gm.waitingRooms[:i], gm.waitingRooms[i+1:]...)
			break
		}
	}
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
			s.log.Infof("Game %s cleaned up", gameID)
			break
		}
	}
}

func (gm *gameManager) getPlayerGame(clientID zkidentity.ShortID) *gameInstance {
	gm.Lock()
	defer gm.Unlock()
	for _, game := range gm.games {
		for _, player := range game.players {
			if player.ID == clientID {
				return game
			}
		}
	}

	return nil
}

func (s *gameManager) startGame(ctx context.Context, players []*Player) (*gameInstance, error) {
	s.Lock()
	defer s.Unlock()
	gameID, err := generateRandomID()
	if err != nil {
		return nil, err
	}

	newGameInstance := s.startNewGame(ctx, players, gameID)
	s.games[gameID] = newGameInstance

	return newGameInstance, nil
}

func (s *gameManager) startNewGame(ctx context.Context, players []*Player, id string) *gameInstance {
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(1, 1),
	)

	players[0].playerNumber = 1
	players[1].playerNumber = 2

	canvasEngine := canvas.New(game)
	canvasEngine.SetDebug(s.debug).SetFPS(*fps)

	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 10)
	roundResult := make(chan int32)
	instanceCtx, cancel := context.WithCancel(ctx)
	// sum of all bets
	betAmt := players[0].BetAmt + players[1].BetAmt
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
		betAmt:      betAmt,
		log:         s.log,
	}

	return instance
}

func (g *gameInstance) Run() {
	g.running = true
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.log.Warnf("Recovered from panic in NewRound: %v", r)
			}
		}()

		// Run a new round only if the game is still running
		if g.running {
			g.engine.NewRound(g.ctx, g.framesch, g.inputch, g.roundResult)
		}
	}()

	go func() {
		for winnerNumber := range g.roundResult {
			if !g.running {
				break
			}

			// Handle the result of each round
			g.handleRoundResult(winnerNumber)

			// Check if the game should continue or end
			if g.shouldEndGame() {
				g.Stop() // Stop the game if necessary
				break
			} else {
				g.engine.NewRound(g.ctx, g.framesch, g.inputch, g.roundResult)
			}
		}
	}()
}

func (g *gameInstance) handleRoundResult(winner int32) {
	// update player score
	for _, player := range g.players {
		if player.playerNumber == winner {
			player.score++
		}
	}
}

func (g *gameInstance) Stop() {
	g.running = false
	g.cleanedUp = true
	g.cancel()
	close(g.framesch)
	close(g.inputch)
	close(g.roundResult)
}

func (g *gameInstance) shouldEndGame() bool {
	for _, player := range g.players {
		// Check if any player has reached the max score
		if player.score >= maxScore {
			g.log.Debugf("Game ending: Player %s reached the maximum score of %d", player.ID, maxScore)
			g.winner = &player.ID
			return true
		}
	}

	// Add other conditions as needed, e.g., time limit or disconnection
	if g.isTimeout() {
		g.log.Debug("Game ending: Timeout reached")
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
