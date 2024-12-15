package ponggame

import (
	"context"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

type WaitingRoom struct {
	sync.RWMutex
	Ctx       context.Context
	Cancel    context.CancelFunc
	ID        string
	HostID    *clientintf.UserID
	Players   []*Player
	BetAmount float64
}

// Marshal converts a WaitingRoom struct to a WaitingRoomProto.
func (wr *WaitingRoom) Marshal() (*pong.WaitingRoom, error) {
	wr.RLock()
	defer wr.RUnlock()

	// Marshal Players
	players := make([]*pong.Player, len(wr.Players))
	for i, player := range wr.Players {
		protoPlayer, err := player.Marshal()
		if err != nil {
			return nil, err
		}
		players[i] = protoPlayer
	}

	return &pong.WaitingRoom{
		Id:      wr.ID,
		HostId:  wr.HostID.String(), // Assuming HostID has a String() method
		Players: players,
		BetAmt:  wr.BetAmount,
	}, nil
}

// Unmarshal converts a WaitingRoomProto to a WaitingRoom struct.
func (wr *WaitingRoom) Unmarshal(proto *pong.WaitingRoom) error {
	wr.Lock()
	defer wr.Unlock()

	// Unmarshal Players
	players := make([]*Player, len(proto.GetPlayers()))
	for i, protoPlayer := range proto.GetPlayers() {
		player := &Player{}
		player.Unmarshal(protoPlayer)
		players[i] = player
	}

	wr.ID = proto.GetId()

	var hostID zkidentity.ShortID
	hostID.FromString(proto.GetHostId())

	wr.HostID = &hostID
	wr.Players = players
	wr.BetAmount = proto.GetBetAmt()
	return nil
}

func (wr *WaitingRoom) AddPlayer(player *Player) {
	wr.Lock()
	defer wr.Unlock()
	for _, p := range wr.Players {
		// don't add repeated players
		if p.ID == player.ID {
			return
		}
	}
	wr.Players = append(wr.Players, player)
}

func (wr *WaitingRoom) ReadyPlayers() ([]*Player, bool) {
	wr.Lock()
	defer wr.Unlock()
	if len(wr.Players) >= 2 {
		for i := range wr.Players {
			if !wr.Players[i].Ready {
				return nil, false
			}
		}
		players := wr.Players[:2]
		wr.Players = wr.Players[2:]
		return players, true
	}
	return nil, false
}

func (wr *WaitingRoom) GetPlayer(clientID *zkidentity.ShortID) *Player {
	wr.RLock()
	defer wr.RUnlock()
	for _, player := range wr.Players {
		if player.ID == clientID {
			return player
		}
	}
	return nil
}

func (wr *WaitingRoom) GetPlayers() []*Player {
	wr.RLock()
	defer wr.RUnlock()
	return wr.Players
}

func (wr *WaitingRoom) RemovePlayer(clientID *zkidentity.ShortID) {
	wr.Lock()
	defer wr.Unlock()

	for i, player := range wr.Players {
		if player.ID == clientID {
			wr.Players = append(wr.Players[:i], wr.Players[i+1:]...)
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
	return len(wr.Players)
}
