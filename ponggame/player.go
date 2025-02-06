package ponggame

import (
	"fmt"
	"sync"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

func (p *Player) Marshal() (*pong.Player, error) {
	if p == nil {
		return nil, fmt.Errorf("player is nil")
	}

	if p.ID == nil {
		return nil, fmt.Errorf("player id is nil")
	}

	return &pong.Player{
		Uid:    p.ID.String(),
		Nick:   p.Nick,
		BetAmt: p.BetAmt,
		Number: p.PlayerNumber,
		Score:  int32(p.Score),
		Ready:  p.Ready,
	}, nil
}

// Unmarshal converts a PlayerProto to a Player struct.
func (p *Player) Unmarshal(proto *pong.Player) error {
	var id zkidentity.ShortID
	id.FromString(proto.Uid)

	if id.IsEmpty() {
		return fmt.Errorf("id is nil")
	}

	p.ID = &id
	p.Nick = proto.GetNick()
	p.BetAmt = proto.GetBetAmt()
	p.PlayerNumber = proto.GetNumber()
	p.Score = int(proto.GetScore())
	p.Ready = proto.GetReady()

	return nil
}

type PlayerSessions struct {
	sync.RWMutex
	Sessions map[zkidentity.ShortID]*Player
}

func (ps *PlayerSessions) RemovePlayer(clientID zkidentity.ShortID) {
	ps.Lock()
	defer ps.Unlock()
	delete(ps.Sessions, clientID)
}

func (ps *PlayerSessions) GetPlayer(clientID zkidentity.ShortID) *Player {
	ps.RLock()
	defer ps.RUnlock()
	player := ps.Sessions[clientID]
	return player
}

func (ps *PlayerSessions) CreateSession(clientID zkidentity.ShortID) *Player {
	ps.Lock()
	defer ps.Unlock()

	player := ps.Sessions[clientID]
	if player == nil {
		clientIDCopy := clientID
		player = &Player{
			ID:    &clientIDCopy,
			Score: 0,
		}
		ps.Sessions[clientID] = player
	}

	return player
}
