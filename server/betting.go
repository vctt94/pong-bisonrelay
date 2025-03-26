package server

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

// SendTipProgressLoop continuously establishes a stream for tip progress events.
// If the connection is lost, it reconnects and requests any events that have not yet been acknowledged.
func (s *Server) SendTipProgressLoop(ctx context.Context, tip *types.TipProgressEvent) error {
	// Only process tip receipt acknowledgement after the tip send progress is completed.
	var err error
	if tip.Completed {
		// ack tip progress if completed
		err = s.bot.AckTipProgress(ctx, tip.SequenceId)
		if err != nil {
			s.log.Warnf("Error while acknowledging tip progress: %v", err)
			return err
		}
		// Convert winner UID and amount to match stored progress
		winnerUID := tip.Uid
		totalMatoms := tip.AmountMatoms

		// Fetch latest tip progress for this winner and amount
		record, err := s.db.FetchLatestUncompletedTipProgress(ctx, winnerUID, totalMatoms)
		if err != nil {
			s.log.Errorf("Error fetching tip progress records: %v", err)
			return err
		}

		// Skip processing if no record was found
		if record == nil {
			err = fmt.Errorf("no matching tip progress record found for UID %s and amount %.8f",
				hex.EncodeToString(winnerUID[:]), float64(totalMatoms)/1e11)
			s.log.Infof(err.Error())
			return err
		}

		// Mark all associated tips as paid
		for _, rt := range record.Tips {
			tipID := make([]byte, 8)
			binary.BigEndian.PutUint64(tipID, rt.SequenceId)

			// Update tip status to processed
			err = s.db.UpdateTipStatus(ctx, rt.Uid, tipID, serverdb.StatusPaid)
			if err != nil {
				s.log.Warnf("Error updating tip %d status: %v", rt.SequenceId, err)
				continue
			}

			// Ack the received tip
			err = s.bot.AckTipReceived(ctx, rt.SequenceId)
			if err != nil {
				s.log.Warnf("Error acknowledging tip %d: %v", rt.SequenceId, err)
			}
		}

		// Update the tip progress record status to processed
		err = s.db.UpdateTipProgressStatus(ctx, record.ID, serverdb.StatusPaid)
		if err != nil {
			s.log.Errorf("Error updating tip progress record status: %v", err)
		}
	}
	return nil
}

// ReceiveTipLoop continuously establishes a stream for incoming tips.
// It stores new tips and updates player bet amounts when appropriate.
func (s *Server) ReceiveTipLoop(ctx context.Context, tip *types.ReceivedTip) error {
	// Check if the tip already exists in the database.
	dbTip, err := s.db.FetchTip(ctx, tip.SequenceId)
	if err != nil {
		s.log.Warnf("Error while fetching tip %d: %v", tip.SequenceId, err)
	}
	// If the tip is not present, store it as unprocessed.
	if dbTip == nil {
		err = s.db.StoreUnprocessedTip(ctx, tip)
		if err != nil {
			s.log.Errorf("Error while storing unprocessed tip: %v", err)
			return err
		}
	} else {
		// If the tip has already been processed, acknowledge it.
		if dbTip.Status == serverdb.StatusPaid {
			err = s.bot.AckTipReceived(ctx, tip.SequenceId)
			if err != nil {
				s.log.Warnf("Error while acknowledging received tip: %v", err)
				return err
			}
			return nil
		}
		// If the tip is still in the 'sending' state, do not update the player's bet amount.
		if dbTip.Status == serverdb.StatusSending {
			return nil
		}
	}

	// Retrieve the player's session using the tip sender's ID.
	player := s.gameManager.PlayerSessions.GetPlayer(zkidentity.ShortID(tip.Uid))
	// If the player's session is not found, skip processing this tip.
	if player == nil {
		return fmt.Errorf("player not found")
	}

	// Update the player's bet amount with the tip value.
	s.gameManager.PlayerSessions.Lock()
	// Convert to dcr from mAtoms and add to player tip amount.
	player.BetAmt += tip.AmountMatoms
	s.log.Debugf("Player %s bet amount updated to %.8f", player.ID.String(), float64(player.BetAmt)/1e11)
	if player.NotifierStream != nil {
		player.NotifierStream.Send(&pong.NtfnStreamResponse{
			NotificationType: pong.NotificationType_BET_AMOUNT_UPDATE,
			BetAmt:           player.BetAmt,
			PlayerId:         player.ID.String(),
		})
	}
	s.gameManager.PlayerSessions.Unlock()
	return nil
}
