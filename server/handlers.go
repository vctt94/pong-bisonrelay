package server

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

// handleFetchTipsByClientIDHandler fetches tips for a specific client ID.
func (s *Server) handleFetchTipsByClientIDHandler(w http.ResponseWriter, r *http.Request) {
	clientIDStr := r.URL.Query().Get("clientID")
	if clientIDStr == "" {
		http.Error(w, "clientID parameter is required", http.StatusBadRequest)
		return
	}

	var clientID zkidentity.ShortID
	if err := clientID.FromString(clientIDStr); err != nil {
		http.Error(w, fmt.Sprintf("invalid client ID: %v", err), http.StatusBadRequest)
		return
	}

	tips, err := s.db.FetchAllReceivedTipsByUID(context.Background(), clientID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching tips: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tips)
}

// handleFetchAllUnprocessedTipsHandler fetches all unprocessed tips for all clients.
func (s *Server) handleFetchAllUnprocessedTipsHandler(w http.ResponseWriter, r *http.Request) {
	tips, err := s.db.FetchUnprocessedTips(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching unprocessed tips: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert the map[zkidentity.ShortID][]serverdb.ReceivedTipWrapper to map[string][]serverdb.ReceivedTipWrapper
	response := make(map[string][]serverdb.ReceivedTipWrapper)
	for clientID, clientTips := range tips {
		response[clientID.String()] = clientTips // Convert clientID to string
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handleReturnUnprocessedTips(ctx context.Context, clientID zkidentity.ShortID, paymentClient types.PaymentsServiceClient, log slog.Logger) error {
	tips, err := s.db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnprocessed)
	if err != nil {
		log.Errorf("Failed to fetch unprocessed tips for client %s: %v", clientID.String(), err)
		return err
	}

	if len(tips) == 0 {
		return nil
	}

	totalDcrAmount := 0.0
	for _, tip := range tips {
		totalDcrAmount += float64(tip.Tip.AmountMatoms) / 1e11 // Convert matoms to DCR
	}

	paymentReq := &types.TipUserRequest{
		User:        clientID.String(),
		DcrAmount:   totalDcrAmount,
		MaxAttempts: 3,
	}
	resp := &types.TipUserResponse{}
	if err := paymentClient.TipUser(ctx, paymentReq, resp); err != nil {
		log.Errorf("Failed to return unprocessed tips to client %s: %v", clientID.String(), err)
		return err
	}

	log.Infof("Returned unprocessed tips to client %s: %.8f", clientID.String(), totalDcrAmount)
	for _, tip := range tips {
		tipID := make([]byte, 8)
		binary.BigEndian.PutUint64(tipID, tip.Tip.SequenceId)
		if err := s.db.UpdateTipStatus(ctx, clientID.Bytes(), tipID, serverdb.StatusSending); err != nil {
			log.Errorf("Failed to update tip status for client %s: %v", clientID.String(), err)
		}
	}

	return nil
}

func (s *Server) handleFetchTotalUnprocessedTips(ctx context.Context, clientID zkidentity.ShortID) (float64, []serverdb.ReceivedTipWrapper, error) {
	// Fetch unprocessed tips from the database
	tips, err := s.db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnprocessed)
	if err != nil {
		s.log.Errorf("Failed to fetch unprocessed tips for client %s: %v", clientID.String(), err)
		return 0, nil, err
	}

	// Calculate total DCR amount
	totalDcrAmount := 0.0
	for _, tip := range tips {
		totalDcrAmount += float64(tip.Tip.AmountMatoms) / 1e11 // Convert matoms to DCR
	}

	s.log.Infof("Fetched %d unprocessed tips for client %s, total amount: %.8f", len(tips), clientID.String(), totalDcrAmount)
	return totalDcrAmount, tips, nil
}

func (s *Server) handleGameLifecycle(ctx context.Context, players []*ponggame.Player, betAmt float64) {
	game, err := s.gameManager.StartGame(ctx, players)
	if err != nil {
		s.log.Errorf("Failed to start game: %v", err)
		return
	}
	defer func() {
		// reset player status
		for _, g := range s.gameManager.Games {
			if g == game {
				for _, player := range game.Players {
					player.Score = 0
					player.PlayerNumber = 0
					player.BetAmt = 0
				}
			}
		}
		// remove game from gameManager after it ended
		for gameID, g := range s.gameManager.Games {
			if g == game {
				delete(s.gameManager.Games, gameID)
				s.log.Infof("Game %s cleaned up", gameID)
				break
			}
		}
	}()

	game.Run()

	var wg sync.WaitGroup
	for _, player := range players {
		wg.Add(1)
		go func(player *ponggame.Player) {
			defer wg.Done()
			if player.NotifierStream != nil {
				err := player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_GAME_START,
					Message:          "Game started with ID: " + game.Id,
					Started:          true,
					GameId:           game.Id,
				})
				if err != nil {
					s.log.Warnf("Failed to notify player %s: %v", player.ID, err)
				}
			}
			s.sendGameUpdates(ctx, player, game)
		}(player)
	}

	wg.Wait() // Wait for both players' streams to finish

	s.handleGameEnd(ctx, game, players, betAmt)
}

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance, players []*ponggame.Player, betAmt float64) {
	winner := game.Winner
	var winnerID string
	if winner != nil {
		winnerID = winner.String()
		s.log.Infof("Game ended. Winner: %s", winnerID)
	} else {
		s.log.Infof("Game ended in a draw.")
	}

	totalBet := betAmt * 2
	// Notify players of game outcome
	for _, player := range players {
		message := "Game ended in a draw."
		if player.ID == winner {
			message = fmt.Sprintf("Congratulations, you won and received: %.8f", totalBet)
		} else {
			message = fmt.Sprintf("Sorry, you lost and lose: %.8f", betAmt)
		}
		player.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.Id,
		})
	}

	// Transfer bet amount to winner
	if winner != nil {
		resp := &types.TipUserResponse{}
		err := s.paymentClient.TipUser(ctx, &types.TipUserRequest{
			User:        winner.String(),
			DcrAmount:   totalBet,
			MaxAttempts: 3,
		}, resp)
		if err != nil {
			s.log.Errorf("Failed to transfer bet amount to winner %s: %v", winner.String(), err)
			return
		}

		s.log.Infof("transfering %.8f to winner %s", totalBet, winner.String())
		for _, player := range players {
			unprocessedTips, err := s.db.FetchReceivedTipsByUID(ctx, *player.ID, serverdb.StatusUnprocessed)
			if err != nil {
				s.log.Errorf("Failed to fetch unprocessed tips for player %s: %v", player.ID, err)
			}
			for _, w := range unprocessedTips {
				tipID := make([]byte, 8)
				binary.BigEndian.PutUint64(tipID, w.Tip.SequenceId)
				err := s.db.UpdateTipStatus(ctx, player.ID.Bytes(), tipID, serverdb.StatusSending)
				if err != nil {
					s.log.Errorf("Failed to update tip status for player %s: %v", player.ID, err)
				}
			}
		}
	}
}

func (s *Server) handleWaitingRoomRemoved(wr *pong.WaitingRoom) {
	s.log.Infof("Waiting room %s removed", wr.Id)

	// Notify all users about the waiting room removal
	for _, user := range s.users {
		user.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_ON_WR_REMOVED,
			Message:          fmt.Sprintf("Waiting room %s was removed", wr.Id),
			RoomId:           wr.Id,
		})
	}
}
