package server

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/stretchr/testify/mock"
	"github.com/vctt94/pong-bisonrelay/server/mocks"
)

func TestReceiveTipLoop(t *testing.T) {
	srv := setupTestServer(t)

	// Mock Payment Client
	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)

	// Mock DB
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	// Expect that StoreUnprocessedTip is called at least once
	mockDB.
		On("StoreUnprocessedTip", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Times(2)

	uid1, _ := hex.DecodeString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")
	uid2, _ := hex.DecodeString("0123456789abcdef0123456789abcde20123456789abcdef0123456789abcde1")

	// Provide a mock stream that returns two ReceivedTip messages.
	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{
			{
				Uid:          uid1,
				AmountMatoms: 100000,
				SequenceId:   1,
			},
			{
				Uid:          uid2,
				AmountMatoms: 200000,
				SequenceId:   2,
			},
		},
	}

	// Expect exactly one call to TipStream
	mockPayClient.
		On("TipStream", mock.Anything, mock.Anything).
		Return(tipStream, nil).
		Once()

	// Create a cancellable context so we can stop the loop after it processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch the loop
	go func() {
		err := srv.ReceiveTipLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	// Wait a bit...
	time.Sleep(50 * time.Millisecond)

	// Stop the loop
	cancel()

	// Assert expectations for both the payment client and DB
	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
