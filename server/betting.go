package server

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
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
			// Program is done.
			return err
		}
		if err != nil {
			s.log.Warn("Error while obtaining payment stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}

		for {
			var tip types.TipProgressEvent
			err := stream.Recv(&tip)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
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
			s.log.Infof("tip amount: %.8f send to: %s, completed: %t", float64(tip.AmountMatoms)/1e11, hex.EncodeToString(tip.Uid), tip.Completed)
		}

		time.Sleep(time.Second)
	}
}

func (s *Server) ReceiveTipLoop(ctx context.Context) error {
	var ackReq types.AckRequest
	// var ackRes types.AckResponse
	for {
		streamReq := types.TipStreamRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := s.paymentClient.TipStream(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			s.log.Warn("Error while obtaining tip stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
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

			s.log.Debugf("Received tip from '%s' amount %d",
				hex.EncodeToString(tip.Uid), tip.AmountMatoms)

			s.Lock()
			player := s.gameManager.playerSessions.GetPlayer(zkidentity.ShortID(tip.Uid))
			if player != nil {
				player.betAmt += float64(tip.AmountMatoms) / 1e11 // Add the tip amount to existing betAmt
			} else {
				// If the player is not connected, append the tip to unprocessed tips
				if _, exists := s.unprocessedTips[zkidentity.ShortID(tip.Uid)]; !exists {
					s.unprocessedTips[zkidentity.ShortID(tip.Uid)] = []*types.ReceivedTip{}
				}
				s.unprocessedTips[zkidentity.ShortID(tip.Uid)] = append(s.unprocessedTips[zkidentity.ShortID(tip.Uid)], &tip)
			}
			s.Unlock()

		}

		time.Sleep(time.Second)
	}
}
