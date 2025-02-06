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

func (s *Server) SendTipProgressLoop(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest

	for {
		// Keep requesting a new stream if the connection breaks. Also
		// request any messages received since the last one we acked.
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
				// Program is done.
				return nil
			}
			if err != nil {
				s.log.Warnf("Error while receiving stream: %v", err)
				break
			}
			ackReq.SequenceId = tip.SequenceId
			err = s.paymentClient.AckTipProgress(ctx, &ackReq, &ackRes)
			if err != nil {
				s.log.Warnf("Error while ack'ing received pm: %v", err)
				break
			}
			s.log.Infof("tip amount: %.8f sent to: %s, completed: %t", float64(tip.AmountMatoms)/1e11, hex.EncodeToString(tip.Uid), tip.Completed)

			// we only Acknowledge tip received after its send progress is completed.
			if tip.Completed {
				var uid zkidentity.ShortID
				uid.FromBytes(tip.Uid)

				// get sending tips, so we can mark them as processed
				tips, err := s.db.FetchReceivedTipsByUID(ctx, uid, serverdb.StatusSending)
				if err != nil {
					s.log.Warnf("Error while fetching unprocessed tips: %v", err)
					break
				}

				for _, tip := range tips {
					// update tip status to processed
					tipID := make([]byte, 8)
					binary.BigEndian.PutUint64(tipID, tip.SequenceId)
					err = s.db.UpdateTipStatus(ctx, tip.Uid, tipID, serverdb.StatusProcessed)
					if err != nil {
						s.log.Debugf("Failed to UpdateTipStatus player %s: %v", uid.String(), err)
					}
					// ack tip received
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

func (s *Server) ReceiveTipLoop(ctx context.Context) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest

	for {
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
				return err
			}
			if err != nil {
				s.log.Warnf("Error while receiving stream: %v", err)
				break
			}

			s.log.Debugf("Received tip from %s amount %d", hex.EncodeToString(tip.Uid), tip.AmountMatoms)

			dbTip, err := s.db.FetchTip(ctx, tip.SequenceId)
			if err != nil {
				s.log.Warnf("Error while fetching tip %d: %v", tip.SequenceId, err)
			}
			// if tip is not in the db, we store it
			if dbTip == nil {
				err = s.db.StoreUnprocessedTip(ctx, &tip)
				if err != nil {
					s.log.Errorf("Error while storing unprocessed tip: %v", err)
					break
				}
			} else {
				// if tip already processed, we can ack it
				if dbTip.Status == serverdb.StatusProcessed {
					ackReq.SequenceId = tip.SequenceId
					err = s.paymentClient.AckTipReceived(ctx, &ackReq, &ackRes)
					if err != nil {
						s.log.Warnf("Error while ack'ing received tip: %v", err)
						break
					}
					continue
				}
				// if tip is sending, we do not update the player's betAmt
				if dbTip.Status == serverdb.StatusSending {
					continue
				}
			}

			player := s.gameManager.PlayerSessions.GetPlayer(zkidentity.ShortID(tip.Uid))
			// if player is nil, we can skip it.
			if player == nil {
				continue
			}
			s.gameManager.PlayerSessions.Lock()
			player.BetAmt += float64(tip.AmountMatoms) / 1e11 // Add the tip amount to existing betAmt
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
