package ponggame

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

const maxScore = 3

// SetupPlayerSession sets up a player session and assigns the notifier stream.
func (gm *GameManager) SetupPlayerSession(clientID zkidentity.ShortID, stream pong.PongGame_StartNtfnStreamServer) *Player {
	player := gm.PlayerSessions.GetOrCreateSession(clientID)
	player.NotifierStream = stream
	return player
}

func (gm *GameManager) StartGameStream(req *StartGameStreamRequest) (*Player, error) {
	player := gm.PlayerSessions.GetPlayer(req.ClientID)
	if player == nil {
		return nil, fmt.Errorf("player not found for client ID %s", req.ClientID)
	}
	if player.NotifierStream == nil {
		return nil, fmt.Errorf("player notifier nil %s", req.ClientID)
	}
	if player.GameStream != nil {
		return nil, fmt.Errorf("game stream is already set for id %s", req.ClientID)
	}
	if !req.IsF2P && player.BetAmt < req.MinBet {
		return nil, fmt.Errorf("player needs to place bet higher or equal to: %.8f DCR", req.MinBet)
	}

	player.GameStream = req.Stream
	player.Ready = true
	player.NotifierStream.Send(&pong.NtfnStreamResponse{
		NotificationType: pong.NotificationType_ON_PLAYER_READY,
		Message:          "player ready",
		PlayerId:         player.ID.String(),
	})

	req.Log.Debugf("Player %s is now ready for the game", req.ClientID)
	return player, nil
}

// HandleWaitingRoomDisconnection handles player disconnection from a waiting room.
func (gm *GameManager) HandleWaitingRoomDisconnection(clientID zkidentity.ShortID, log slog.Logger) {
	waitingRoom := gm.GetWaitingRoomFromPlayer(clientID)
	if waitingRoom == nil {
		return
	}

	remainingPlayers := GetRemainingPlayersInWaitingRoom(waitingRoom, clientID)
	for _, player := range remainingPlayers {
		if player.NotifierStream != nil {
			player.NotifierStream.Send(&pong.NtfnStreamResponse{
				NotificationType: pong.NotificationType_OPPONENT_DISCONNECTED,
				Message:          "Opponent left the waiting room.",
				Started:          false,
			})
		}
	}

	log.Debugf("Player %s disconnected; removing waiting room %s", clientID, waitingRoom.ID)
	waitingRoom.Cancel()
	gm.RemoveWaitingRoom(waitingRoom.ID)
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

	log.Debugf("Player %s disconnected; cleaning up game", clientID)
	for gameID, g := range gm.Games {
		if g == game {
			delete(gm.Games, gameID)
			log.Debugf("Game %s cleaned up", gameID)
			break
		}
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
	inputBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input: %w", err)
	}

	game.Lock()
	defer game.Unlock()

	if !game.Running {
		return nil, fmt.Errorf("game has ended for client ID %s", clientID)
	}

	// Send inputBytes to game.inputch
	game.Inputch <- inputBytes

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
			// Remove the room by appending the elements before and after it
			gm.WaitingRooms = append(gm.WaitingRooms[:i], gm.WaitingRooms[i+1:]...)
			break
		}
	}
}

func (gm *GameManager) GetPlayerGame(clientID zkidentity.ShortID) *GameInstance {
	gm.Lock()
	defer gm.Unlock()
	for _, game := range gm.Games {
		for _, player := range game.Players {
			if *player.ID == clientID {
				return game
			}
		}
	}

	return nil
}

func (s *GameManager) StartGame(ctx context.Context, players []*Player) (*GameInstance, error) {
	s.Lock()
	defer s.Unlock()
	gameID, err := GenerateRandomString(16)
	if err != nil {
		return nil, err
	}

	newGameInstance := s.startNewGame(ctx, players, gameID)
	s.Games[gameID] = newGameInstance

	return newGameInstance, nil
}

func (s *GameManager) startNewGame(ctx context.Context, players []*Player, id string) *GameInstance {
	game := engine.NewGame(
		80, 40,
		engine.NewPlayer(1, 5),
		engine.NewPlayer(1, 5),
		engine.NewBall(1, 1),
	)

	players[0].PlayerNumber = 1
	players[1].PlayerNumber = 2

	canvasEngine := New(game)
	canvasEngine.SetDebug(s.Debug).SetFPS(DEFAULT_FPS)

	framesch := make(chan []byte, 100)
	inputch := make(chan []byte, 10)
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
		log:         s.Log,
	}

	return instance
}

func (g *GameInstance) Run() {
	g.Running = true
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.log.Warnf("Recovered from panic in NewRound: %v", r)
			}
		}()

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
			g.log.Debugf("Game ending: Player %s reached the maximum score of %d", player.ID, maxScore)
			g.Winner = player.ID
			g.Running = false
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
func (g *GameInstance) isTimeout() bool {
	// For example, a simple time limit check
	// const maxGameDuration = 10 * time.Minute
	// return time.Since(g.startTime) >= maxGameDuration
	return false
}
