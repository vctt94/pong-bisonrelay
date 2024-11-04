package server

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

type WaitingRoom struct {
	sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	ID        string
	hostID    clientintf.UserID
	players   []*Player
	BetAmount float64
}

func (wr *WaitingRoom) AddPlayer(player *Player) {
	wr.Lock()
	defer wr.Unlock()
	for _, p := range wr.players {
		// don't add repeated players
		if p.ID == player.ID {
			return
		}
	}
	wr.players = append(wr.players, player)
}

func (wr *WaitingRoom) ReadyPlayers() ([]*Player, bool) {
	wr.Lock()
	defer wr.Unlock()
	if len(wr.players) >= 2 {
		for i := range wr.players {
			if !wr.players[i].ready {
				return nil, false
			}
		}
		players := wr.players[:2]
		wr.players = wr.players[2:]
		return players, true
	}
	return nil, false
}

func (wr *WaitingRoom) GetPlayer(clientID zkidentity.ShortID) *Player {
	wr.RLock()
	defer wr.RUnlock()
	for _, player := range wr.players {
		if player.ID == clientID {
			return player
		}
	}
	return nil
}

func (wr *WaitingRoom) GetPlayers() []*Player {
	wr.RLock()
	defer wr.RUnlock()
	return wr.players
}

func (wr *WaitingRoom) RemovePlayer(clientID zkidentity.ShortID) {
	wr.Lock()
	defer wr.Unlock()

	for i, player := range wr.players {
		if player.ID == clientID {
			wr.players = append(wr.players[:i], wr.players[i+1:]...)
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
	return len(wr.players)
}

// ToPongWaitingRoom converts a WaitingRoom instance to a pong.WaitingRoom instance.
func (wr *WaitingRoom) ToPongWaitingRoom() (*pong.WaitingRoom, error) {
	wr.Lock()
	defer wr.Unlock()

	hostIDStr := hex.EncodeToString(wr.hostID[:])

	// Prepare players for pong.WaitingRoom
	var players []*pong.Player
	for _, player := range wr.players {
		p := &pong.Player{
			Uid:       player.ID.String(),
			Nick:      player.Nick,
			BetAmount: player.BetAmt,
		}
		players = append(players, p)
	}

	return &pong.WaitingRoom{
		Id:      wr.ID,
		HostId:  hostIDStr,
		Players: players,
		BetAmt:  wr.BetAmount,
	}, nil
}
