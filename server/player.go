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
	ready        bool
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
