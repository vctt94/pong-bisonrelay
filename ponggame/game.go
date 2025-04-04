package ponggame

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/bisonbotkit/utils"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/protobuf/proto"
)

const maxScore = 3

// HandleWaitingRoomDisconnection handles player disconnection from a waiting room.
func (gm *GameManager) HandleWaitingRoomDisconnection(clientID zkidentity.ShortID, log slog.Logger) {
	wr := gm.GetWaitingRoomFromPlayer(clientID)
	if wr == nil {
		return
	}

	// If host disconnected, remove wr
	if clientID == *wr.HostID {
		remainingPlayers := GetRemainingPlayersInWaitingRoom(wr, clientID)
		for _, player := range remainingPlayers {
			if player.NotifierStream != nil {
				player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
					Message:          "Host left the waiting room. Room closed.",
					Started:          false,
				})
			}
		}

		log.Debugf("Player %s disconnected; removing waiting room %s", clientID, wr.ID)
		wr.Cancel()
		gm.RemoveWaitingRoom(wr.ID)
	} else {
		// Handle regular player disconnection
		wr.RemovePlayer(clientID)
		remainingPlayers := GetRemainingPlayersInWaitingRoom(wr, clientID)

		// Marshal updated waiting room
		pongwr, err := wr.Marshal()
		if err != nil {
			log.Errorf("Failed to marshal waiting room: %v", err)
			return
		}

		// Notify remaining players
		for _, player := range remainingPlayers {
			if player.NotifierStream != nil {
				player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
					Message:          "Player left the waiting room",
					Wr:               pongwr,
				})
			}
		}

		log.Debugf("Player %s disconnected from waiting room %s", clientID, wr.ID)
	}
}

// HandleGameDisconnection handles player disconnection from an active game.
func (gm *GameManager) HandleGameDisconnection(clientID zkidentity.ShortID, log slog.Logger) {
	game := gm.GetPlayerGame(clientID)
	if game == nil {
		return
	}

	remainingPlayer := GetRemainingPlayerInGame(game, clientID)
	if remainingPlayer != nil && remainingPlayer.NotifierStream != nil {
		remainingPlayer.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
			Message:          "Opponent disconnected. Game over.",
			Started:          false,
		})
	}

}

func (gm *GameManager) HandlePlayerInput(clientID zkidentity.ShortID, req *pong.PlayerInput) (*pong.GameUpdate, error) {
	player := gm.PlayerSessions.GetPlayer(clientID)
	if player == nil {
		return nil, fmt.Errorf("player: %s not found", clientID)
	}
	if player.PlayerNumber != 1 && player.PlayerNumber != 2 {
		return nil, fmt.Errorf("player number incorrect, it must be 1 or 2; it is: %d", player.PlayerNumber)
	}

	game := gm.GetPlayerGame(clientID)
	if game == nil {
		return nil, fmt.Errorf("game instance not found for client ID %s", clientID)
	}

	req.PlayerNumber = player.PlayerNumber
	inputBytes, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	if !game.Running {
		return nil, fmt.Errorf("game has ended for client ID %s", clientID)
	}

	// Try to send input without blocking, discard old inputs if channel is full
	select {
	case game.Inputch <- inputBytes:
		// success
	default:
		// channel is full; drop the oldest one
		select {
		case <-game.Inputch:
		default:
			// means the channel was emptied in the meantime
		}

		// now that we've popped one off, try once more
		select {
		case game.Inputch <- inputBytes:
			// success
		default:
			// still no room, just drop the input
			gm.Log.Debugf("Input channel full for game %s, dropping input", game.Id)
		}
	}

	return &pong.GameUpdate{}, nil
}

func (g *GameManager) GetWaitingRoomFromPlayer(playerID zkidentity.ShortID) *WaitingRoom {
	g.RLock()
	defer g.RUnlock()

	for _, room := range g.WaitingRooms {
		for _, p := range room.Players {
			if *p.ID == playerID {
				return room
			}
		}
	}
	return nil
}

func (g *GameManager) GetWaitingRoom(roomID string) *WaitingRoom {
	g.RLock()
	defer g.RUnlock()

	for _, room := range g.WaitingRooms {
		if room.ID == roomID {
			return room
		}
	}
	return nil
}

func (gm *GameManager) RemoveWaitingRoom(roomID string) {
	gm.Lock()
	defer gm.Unlock()

	for i, room := range gm.WaitingRooms {
		if room.ID == roomID {

			if gm.OnWaitingRoomRemoved != nil {
				pongwr, err := room.Marshal()
				if err != nil {
					gm.Log.Errorf("Failed to Marshal waiting room %v", err)
				}
				gm.OnWaitingRoomRemoved(pongwr)
			}
			// Remove the room by appending the elements before and after it
			gm.WaitingRooms = append(gm.WaitingRooms[:i], gm.WaitingRooms[i+1:]...)
			gm.Log.Debugf("Waiting room %s removed successfully", roomID)

			break
		}
	}
}

func (gm *GameManager) GetPlayerGame(clientID zkidentity.ShortID) *GameInstance {
	gm.RLock()
	defer gm.RUnlock()
	return gm.PlayerGameMap[clientID]
}

func (s *GameManager) StartGame(ctx context.Context, players []*Player) (*GameInstance, error) {
	s.Lock()
	defer s.Unlock()
	gameID, err := utils.GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	newGameInstance := s.startNewGame(ctx, players, gameID)
	s.Games[gameID] = newGameInstance

	return newGameInstance, nil
}

func (gm *GameManager) startNewGame(ctx context.Context, players []*Player, id string) *GameInstance {
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(1, 1),
	)

	players[0].PlayerNumber = 1
	players[1].PlayerNumber = 2

	canvasEngine := New(game)
	canvasEngine.SetLogger(gm.Log).SetFPS(DEFAULT_FPS)

	framesch := make(chan []byte, INPUT_BUF_SIZE)
	inputch := make(chan []byte, INPUT_BUF_SIZE)
	roundResult := make(chan int32)
	instanceCtx, cancel := context.WithCancel(ctx)
	// sum of all bets
	betAmt := players[0].BetAmt + players[1].BetAmt
	instance := &GameInstance{
		Id:          id,
		engine:      canvasEngine,
		Framesch:    framesch,
		Inputch:     inputch,
		roundResult: roundResult,
		Running:     true,
		ctx:         instanceCtx,
		cancel:      cancel,
		Players:     players,
		betAmt:      betAmt,
		log:         gm.Log,
	}

	gm.PlayerGameMap[*players[0].ID] = instance
	gm.PlayerGameMap[*players[1].ID] = instance
	return instance
}

func (g *GameInstance) Run() {
	g.Running = true
	go func() {
		// Run a new round only if the game is still running
		if g.Running {
			g.engine.NewRound(g.ctx, g.Framesch, g.Inputch, g.roundResult)
		}
	}()

	go func() {
		for winnerNumber := range g.roundResult {
			if !g.Running {
				break
			}

			// Handle the result of each round
			g.handleRoundResult(winnerNumber)

			// Check if the game should continue or end
			if g.shouldEndGame() {
				// clean up the game after ending
				g.Cleanup()
				break
			} else {
				g.engine.NewRound(g.ctx, g.Framesch, g.Inputch, g.roundResult)
			}
		}
	}()
}

func (g *GameInstance) handleRoundResult(winner int32) {
	// update player score
	for _, player := range g.Players {
		if player.PlayerNumber == winner {
			player.Score++
		}
	}
}

func (g *GameInstance) Cleanup() {
	g.cleanedUp = true
	g.cancel()
	close(g.Framesch)
	close(g.Inputch)
	close(g.roundResult)
}

func (g *GameInstance) shouldEndGame() bool {
	for _, player := range g.Players {
		// Check if any player has reached the max score
		if player.Score >= maxScore {
			g.log.Infof("Game ending: Player %s reached the maximum score of %d", player.ID, player.Score)
			g.Winner = player.ID
			g.Running = false
			return true
		}
	}

	// Add other conditions as needed, e.g., time limit or disconnection
	if g.isTimeout() {
		g.log.Info("Game ending: Timeout reached")
		return true
	}

	// Return false if none of the end conditions are met
	return false
}

// isTimeout checks if the game duration has exceeded a set limit
func (g *GameInstance) isTimeout() bool {
	// For example, a simple time limit check
	// const maxGameDuration = 10 * time.Minute
	// return time.Since(g.startTime) >= maxGameDuration
	return false
}
