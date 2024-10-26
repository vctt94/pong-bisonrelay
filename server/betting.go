package server

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/decred/slog"
)

func (s *Server) SendTipProgressLoop(ctx context.Context, payment types.PaymentsServiceClient, log slog.Logger) error {
	var ackRes types.AckResponse
	var ackReq types.AckRequest
	for {
		// Keep requesting a new stream if the connection breaks. Also
		// request any messages received since the last one we acked.
		streamReq := types.TipProgressRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := payment.TipProgress(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			log.Warn("Error while obtaining payment stream: %v", err)
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
				log.Warnf("Error while receiving stream: %v", err)
				break
			}

			ruid := hex.EncodeToString(tip.Uid)
			fmt.Printf("tip progress to id: %+v\n\n", ruid)
			// Ack to client that message is processed.
			ackReq.SequenceId = tip.SequenceId
			err = payment.AckTipProgress(ctx, &ackReq, &ackRes)
			if err != nil {
				log.Warnf("Error while ack'ing received pm: %v", err)
				break
			}
		}

		time.Sleep(time.Second)
	}
}

func (s *Server) ReceiveTipLoop(ctx context.Context, payment types.PaymentsServiceClient, log slog.Logger) error {
	var ackReq types.AckRequest
	var ackRes types.AckResponse
	for {
		streamReq := types.TipStreamRequest{UnackedFrom: ackReq.SequenceId}
		stream, err := payment.TipStream(ctx, &streamReq)
		if errors.Is(err, context.Canceled) {
			// Program is done.
			return err
		}
		if err != nil {
			log.Warn("Error while obtaining tip stream: %v", err)
			time.Sleep(time.Second) // Wait to try again.
			continue
		}

		for {
			var tip types.ReceivedTip
			err := stream.Recv(&tip)
			if errors.Is(err, context.Canceled) {
				// Program is done.
				return err
			}
			if err != nil {
				log.Warnf("Error while receiving stream: %v", err)
				break
			}

			log.Debugf("Received tip from '%s' amount %d",
				hex.EncodeToString(tip.Uid), tip.AmountMatoms)

			dcrAmount := float64(tip.AmountMatoms) / 1e11

			fmt.Printf("<- %v %.8f\n", hex.EncodeToString(tip.Uid), dcrAmount)
			// Ack to client that message is processed.
			log.Debugf("tip sequenceId: %+v\n", tip.SequenceId)
			ackReq.SequenceId = tip.SequenceId
			err = payment.AckTipReceived(ctx, &ackReq, &ackRes)
			if err != nil {
				log.Warnf("Error while ack'ing received pm: %v", err)
				break
			}
		}

		time.Sleep(time.Second)
	}
}
