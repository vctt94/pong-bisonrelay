package server

import (
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

type Player struct {
	ID zkidentity.ShortID

	Nick         string
	BetAmt       float64
	playerNumber int32 // 1 for player 1, 2 for player 2
	score        int
	stream       pong.PongGame_StartGameStreamServer
	notifier     pong.PongGame_StartNtfnStreamServer
}

type PlayerSessions struct {
	sync.RWMutex
	sessions map[zkidentity.ShortID]*Player
}

func (ps *PlayerSessions) RemovePlayer(clientID zkidentity.ShortID) {
	ps.Lock()
	defer ps.Unlock()
	delete(ps.sessions, clientID)
}

func (ps *PlayerSessions) GetPlayer(clientID zkidentity.ShortID) *Player {
	ps.RLock()
	defer ps.RUnlock()
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
		ps.Lock()
		ps.sessions[clientID] = p
		ps.Unlock()

		return p
	}

	return player
}

type WaitingRoom struct {
	sync.RWMutex
	queue []*Player
}

func (wr *WaitingRoom) AddPlayer(player *Player) {
	wr.Lock()
	defer wr.Unlock()
	for _, p := range wr.queue {
		// don't add repeated players
		if p.ID == player.ID {
			return
		}
	}
	wr.queue = append(wr.queue, player)
}

func (wr *WaitingRoom) ReadyPlayers() ([]*Player, bool) {
	wr.Lock()
	defer wr.Unlock()
	if len(wr.queue) >= 2 {
		players := wr.queue[:2]
		wr.queue = wr.queue[2:]
		return players, true
	}
	return nil, false
}

func (wr *WaitingRoom) GetPlayer(clientID zkidentity.ShortID) *Player {
	wr.RLock()
	defer wr.RUnlock()
	for _, player := range wr.queue {
		if player.ID == clientID {
			return player
		}
	}
	return nil
}

func (wr *WaitingRoom) GetPlayers() []*Player {
	wr.RLock()
	defer wr.RUnlock()
	return wr.queue
}

func (wr *WaitingRoom) RemovePlayer(clientID zkidentity.ShortID) {
	wr.Lock()
	defer wr.Unlock()

	for i, player := range wr.queue {
		if player.ID == clientID {
			wr.queue = append(wr.queue[:i], wr.queue[i+1:]...)
			break
		}
	}
}

func (wr *WaitingRoom) getWaitingRoom() *WaitingRoom {
	wr.RLock()
	defer wr.RUnlock()
	return wr
}

func (wr *WaitingRoom) length() int {
	wr.RLock()
	defer wr.RUnlock()
	return len(wr.queue)
}
