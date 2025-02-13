package server

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

// SendTipProgressLoop continuously establishes a stream for tip progress events.
// If the connection is lost, it reconnects and requests any events that have not yet been acknowledged.
func (s *Server) SendTipProgressLoop(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest

	for {
		// Create a new stream, asking for tip progress events starting after the last acknowledged event.
		streamReq := types.TipProgressRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := s.paymentClient.TipProgress(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			return err
		}
		if err != nil {
			s.log.Warn("Error while obtaining payment stream: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for {
			var tip types.TipProgressEvent
			err := stream.Recv(&tip)
			if errors.Is(err, context.Canceled) {
				// Context cancelled; exit the loop.
				return nil
			}
			if err != nil {
				s.log.Warnf("Error while receiving stream: %v", err)
				break
			}

			// Update the last acknowledged sequence ID.
			ackReq.SequenceId = tip.SequenceId
			err = s.paymentClient.AckTipProgress(ctx, &ackReq, &ackRes)
			if err != nil {
				s.log.Warnf("Error while acknowledging tip progress: %v", err)
				break
			}
			s.log.Infof("Tip amount: %.8f sent to: %s, completed: %t", float64(tip.AmountMatoms)/1e11, hex.EncodeToString(tip.Uid), tip.Completed)

			// Only process tip receipt acknowledgement after the tip send progress is completed.
			if tip.Completed {
				var uid zkidentity.ShortID
				uid.FromBytes(tip.Uid)

				// Fetch tips that are currently marked as 'sending' so they can be updated.
				tips, err := s.db.FetchReceivedTipsByUID(ctx, uid, serverdb.StatusSending)
				if err != nil {
					s.log.Warnf("Error while fetching unprocessed tips: %v", err)
					break
				}

				for _, tip := range tips {
					// Mark the tip as processed in the database.
					tipID := make([]byte, 8)
					binary.BigEndian.PutUint64(tipID, tip.SequenceId)
					err = s.db.UpdateTipStatus(ctx, tip.Uid, tipID, serverdb.StatusProcessed)
					if err != nil {
						s.log.Debugf("Failed to update tip status for player %s: %v", uid.String(), err)
					}
					// Acknowledge that the tip has been received.
					ackRes := &types.AckResponse{}
					err = s.paymentClient.AckTipReceived(ctx, &types.AckRequest{SequenceId: tip.SequenceId}, ackRes)
					if err != nil {
						s.log.Debugf("Failed to acknowledge tip for player %s: %v", uid.String(), err)
					} else {
						s.log.Debugf("Acknowledged tip with SequenceId %d for player %s", tip.SequenceId, uid.String())
					}
				}
			}
		}
		time.Sleep(time.Second)
	}
}

// ReceiveTipLoop continuously establishes a stream for incoming tips.
// It stores new tips and updates player bet amounts when appropriate.
func (s *Server) ReceiveTipLoop(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest

	for {
		// Create a new tip stream, requesting tips starting after the last acknowledged one.
		streamReq := types.TipStreamRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := s.paymentClient.TipStream(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			return err
		}
		if err != nil {
			s.log.Warn("Error while obtaining tip stream: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for {
			var tip types.ReceivedTip
			err := stream.Recv(&tip)
			if errors.Is(err, context.Canceled) {
				// Context cancelled; exit the loop.
				return err
			}
			if err != nil {
				s.log.Warnf("Error while receiving stream: %v", err)
				break
			}

			s.log.Debugf("Received tip from %s amount %d", hex.EncodeToString(tip.Uid), tip.AmountMatoms)

			// Check if the tip already exists in the database.
			dbTip, err := s.db.FetchTip(ctx, tip.SequenceId)
			if err != nil {
				s.log.Warnf("Error while fetching tip %d: %v", tip.SequenceId, err)
			}
			// If the tip is not present, store it as unprocessed.
			if dbTip == nil {
				err = s.db.StoreUnprocessedTip(ctx, &tip)
				if err != nil {
					s.log.Errorf("Error while storing unprocessed tip: %v", err)
					break
				}
			} else {
				// If the tip has already been processed, acknowledge it.
				if dbTip.Status == serverdb.StatusProcessed {
					ackReq.SequenceId = tip.SequenceId
					err = s.paymentClient.AckTipReceived(ctx, &ackReq, &ackRes)
					if err != nil {
						s.log.Warnf("Error while acknowledging received tip: %v", err)
						break
					}
					continue
				}
				// If the tip is still in the 'sending' state, do not update the player's bet amount.
				if dbTip.Status == serverdb.StatusSending {
					continue
				}
			}

			// Retrieve the player's session using the tip sender's ID.
			player := s.gameManager.PlayerSessions.GetPlayer(zkidentity.ShortID(tip.Uid))
			// If the player's session is not found, skip processing this tip.
			if player == nil {
				continue
			}

			// Update the player's bet amount with the tip value.
			s.gameManager.PlayerSessions.Lock()
			// Convert to dcr from mAtoms and add to player tip amount.
			player.BetAmt += float64(tip.AmountMatoms) / 1e11
			if player.NotifierStream != nil {
				player.NotifierStream.Send(&pong.NtfnStreamResponse{
					NotificationType: pong.NotificationType_BET_AMOUNT_UPDATE,
					BetAmt:           player.BetAmt,
					PlayerId:         player.ID.String(),
				})
			}
			s.gameManager.PlayerSessions.Unlock()
		}
		time.Sleep(time.Second)
	}
}
