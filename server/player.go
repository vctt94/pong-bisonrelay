package server

import (
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

type Player struct {
	ID zkidentity.ShortID

	playerNumber int32 // 1 for player 1, 2 for player 2
	score        int
	betAmt       float64
	stream       pong.PongGame_StartGameStreamServer
	notifier     pong.PongGame_StartNtfnStreamServer
}

type PlayerSessions struct {
	mu       sync.Mutex
	sessions map[zkidentity.ShortID]*Player
}

func (ps *PlayerSessions) RemovePlayer(clientID zkidentity.ShortID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.sessions, clientID)
}

func (ps *PlayerSessions) GetPlayer(clientID zkidentity.ShortID) *Player {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	player := ps.sessions[clientID]
	return player
}

func (ps *PlayerSessions) GetOrCreateSession(clientID zkidentity.ShortID) *Player {
	player := ps.GetPlayer(clientID)
	if player == nil {
		p := &Player{
			ID:    clientID,
			score: 0,
		}
		ps.mu.Lock()
		ps.sessions[clientID] = p
		ps.mu.Unlock()

		return p
	}

	return player
}

type WaitingRoom struct {
	mu    sync.Mutex
	queue []*Player
}

func (wr *WaitingRoom) AddPlayer(player *Player) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	for _, p := range wr.queue {
		// don't add repeated players
		if p.ID == player.ID {
			return
		}
	}
	wr.queue = append(wr.queue, player)
}

func (wr *WaitingRoom) ReadyPlayers() ([]*Player, bool) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	if len(wr.queue) >= 2 {
		players := wr.queue[:2]
		wr.queue = wr.queue[2:]
		return players, true
	}
	return nil, false
}

func (wr *WaitingRoom) GetPlayer(clientID zkidentity.ShortID) (*Player, bool) {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	for _, player := range wr.queue {
		if player.ID == clientID {
			return player, true
		}
	}
	return nil, false
}

func (wr *WaitingRoom) RemovePlayer(clientID zkidentity.ShortID) {
	wr.mu.Lock()
	defer wr.mu.Unlock()

	for i, player := range wr.queue {
		if player.ID == clientID {
			wr.queue = append(wr.queue[:i], wr.queue[i+1:]...)
			break
		}
	}
}

func (wr *WaitingRoom) getWaitingRoom() *WaitingRoom {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	return wr
}

func (wr *WaitingRoom) length() int {
	wr.mu.Lock()
	defer wr.mu.Unlock()
	return len(wr.queue)
}
