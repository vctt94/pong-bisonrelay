package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

func (s *Server) handleReturnUnprocessedTips(ctx context.Context, clientID zkidentity.ShortID, paymentClient types.PaymentsServiceClient, log slog.Logger) error {
	tips, err := s.db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnpaid)
	if err != nil {
		log.Errorf("Failed to fetch unprocessed tips for client %s: %v", clientID.String(), err)
		return err
	}

	if len(tips) == 0 {
		return nil
	}

	totalDcrAmount := 0.0
	for _, tip := range tips {
		totalDcrAmount += float64(tip.AmountMatoms) / 1e11 // Convert matoms to DCR
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

	// Convert total back to matoms for storage
	totalMatoms := int64(totalDcrAmount * 1e11)

	// Store send progress with all tips being returned
	err = s.db.StoreSendTipProgress(ctx, clientID.Bytes(), totalMatoms, tips, serverdb.StatusSending)
	if err != nil {
		log.Errorf("Failed to store return tip progress: %v", err)
		return err
	}

	for _, tip := range tips {
		tipID := make([]byte, 8)
		binary.BigEndian.PutUint64(tipID, tip.SequenceId)
		if err := s.db.UpdateTipStatus(ctx, clientID.Bytes(), tipID, serverdb.StatusSending); err != nil {
			log.Errorf("Failed to update tip status for client %s: %v", clientID.String(), err)
		}
	}

	return nil
}

func (s *Server) handleFetchTotalUnprocessedTips(ctx context.Context, clientID zkidentity.ShortID) (int64, []*types.ReceivedTip, error) {
	// Fetch unprocessed tips from the database
	tips, err := s.db.FetchReceivedTipsByUID(ctx, clientID, serverdb.StatusUnpaid)
	if err != nil {
		s.log.Errorf("Failed to fetch unprocessed tips for client %s: %v", clientID.String(), err)
		return 0, nil, err
	}

	// Calculate total DCR amount
	totalDcrAmount := int64(0)
	for _, tip := range tips {
		totalDcrAmount += tip.AmountMatoms
	}

	s.log.Infof("Fetched %d unprocessed tips for client %s, total amount: %.8f", len(tips), clientID.String(), totalDcrAmount)
	return totalDcrAmount, tips, nil
}

func (s *Server) handleGameLifecycle(ctx context.Context, players []*ponggame.Player, tips []*types.ReceivedTip) {
	game, err := s.gameManager.StartGame(ctx, players)
	if err != nil {
		s.log.Errorf("Failed to start game: %v", err)
		return
	}

	defer func() {
		// reset player status
		for _, player := range game.Players {
			player.ResetPlayer()
			// Fetch latest unprocessed tips and update bet amount
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			totalDcrAmount, _, err := s.handleFetchTotalUnprocessedTips(ctx, *player.ID)
			cancel()
			if err != nil {
				s.log.Errorf("Error fetching tips for player %s: %v", player.ID, err)
				continue
			}
			playerSession := s.gameManager.PlayerSessions.GetPlayer(*player.ID)
			if playerSession == nil {
				s.log.Errorf("Error finding player session %s", player.ID)
				continue
			}
			playerSession.BetAmt = totalDcrAmount
			s.log.Debugf("Reset player %s with updated bet amount: %.8f", player.ID, totalDcrAmount)
		}
		// remove game from gameManager after it ended
		delete(s.gameManager.Games, game.Id)
		s.log.Infof("Game %s cleaned up", game.Id)
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

	s.handleGameEnd(ctx, game, players, tips)
}

func (s *Server) handleGameEnd(ctx context.Context, game *ponggame.GameInstance, players []*ponggame.Player, tips []*types.ReceivedTip) {
	winner := game.Winner
	var winnerID string
	if winner != nil {
		winnerID = winner.String()
		s.log.Infof("Game ended. Winner: %s", winnerID)
	} else {
		s.log.Infof("Game ended in a draw.")
	}

	// Calculate total from actual reserved tips
	totalAmountMatoms := int64(0)
	for _, tip := range tips {
		totalAmountMatoms += tip.AmountMatoms
	}

	// Convert total to matoms for storage
	totalDcrAmount := float64(totalAmountMatoms) / 1e11

	// Store send progress with ALL tips (both players')
	err := s.db.StoreSendTipProgress(ctx, winner[:], totalAmountMatoms, tips, serverdb.StatusSending)
	if err != nil {
		s.log.Errorf("Failed to store send progress: %v", err)
		return
	}

	// Process the reserved tips
	for _, tip := range tips {
		tipID := make([]byte, 8)
		binary.BigEndian.PutUint64(tipID, tip.SequenceId)
		err := s.db.UpdateTipStatus(ctx, tip.Uid, tipID, serverdb.StatusSending)
		if err != nil {
			s.log.Errorf("Failed to update tip status for player %s: %v", tip.Uid, err)
		}
	}

	// Notify players of game outcome
	for _, player := range players {
		message := "Game ended in a draw."
		if player.ID == winner {
			message = fmt.Sprintf("Congratulations, you won and received: %.8f", totalDcrAmount)
		} else {
			// Calculate lost amount for this player
			lostAmount := 0.0
			for _, tip := range tips {
				// Compare player ID with tip UID
				if bytes.Equal(player.ID[:], tip.Uid) {
					lostAmount += float64(tip.AmountMatoms) / 1e11
				}
			}
			message = fmt.Sprintf("Sorry, you lost and lose: %.8f", lostAmount)
		}
		player.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_GAME_END,
			Message:          message,
			GameId:           game.Id,
		})
	}

	// Transfer actual reserved tip amounts to winner
	if winner != nil {
		// Process the reserved tips
		for _, tip := range tips {
			tipID := make([]byte, 8)
			binary.BigEndian.PutUint64(tipID, tip.SequenceId)
			err := s.db.UpdateTipStatus(ctx, tip.Uid, tipID, serverdb.StatusSending)
			if err != nil {
				s.log.Errorf("Failed to update tip status for player %s: %v", tip.Uid, err)
			}
		}
		resp := &types.TipUserResponse{}
		err := s.paymentClient.TipUser(ctx, &types.TipUserRequest{
			User:        winner.String(),
			DcrAmount:   totalDcrAmount,
			MaxAttempts: 3,
		}, resp)
		if err != nil {
			s.log.Errorf("Failed to transfer bet amount to winner %s: %v", winner.String(), err)
			return
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
