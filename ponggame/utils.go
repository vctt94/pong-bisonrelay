package ponggame

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"

	"github.com/companyzero/bisonrelay/zkidentity"
	"google.golang.org/grpc/metadata"
)

func getClientIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata found in context")
	}

	clientIDs, ok := md["client-id"]
	if !ok || len(clientIDs) == 0 {
		return "", fmt.Errorf("client-id not found in metadata")
	}

	return clientIDs[0], nil
}

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// Helper function to get remaining players in the waiting room
func GetRemainingPlayersInWaitingRoom(waitingRoom *WaitingRoom, disconnectedID zkidentity.ShortID) []*Player {
	var remainingPlayers []*Player
	for _, player := range waitingRoom.Players {
		if *player.ID != disconnectedID {
			remainingPlayers = append(remainingPlayers, player)
		}
	}
	return remainingPlayers
}

// Helper function to get the remaining player in a game
func GetRemainingPlayerInGame(game *GameInstance, disconnectedID zkidentity.ShortID) *Player {
	for _, player := range game.Players {
		if *player.ID != disconnectedID {
			return player
		}
	}
	return nil
}
