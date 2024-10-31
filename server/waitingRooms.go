package server

import (
	"fmt"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
)

type WaitingRoom struct {
	sync.RWMutex
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
	fmt.Printf("wr players: %+v\n", wr.players)
	if len(wr.players) >= 2 {
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
