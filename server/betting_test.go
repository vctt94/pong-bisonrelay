package server

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/stretchr/testify/mock"
	"github.com/vctt94/pong-bisonrelay/server/internal/mocks"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

func TestReceiveTipLoop(t *testing.T) {
	srv := setupTestServer(t)

	// Mock Payment Client
	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)

	// Mock DB
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	// Expect that FetchTip returns nil (tip doesn't exist) for both tips
	mockDB.
		On("FetchTip", mock.Anything, uint64(1)).
		Return((*serverdb.ReceivedTipWrapper)(nil), nil).
		Once()
	mockDB.
		On("FetchTip", mock.Anything, uint64(2)).
		Return((*serverdb.ReceivedTipWrapper)(nil), nil).
		Once()

	// Expect that StoreUnprocessedTip is called at least once
	mockDB.
		On("StoreUnprocessedTip", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Times(2)

	uid1 := zkidentity.ShortID{}
	err := uid1.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")
	if err != nil {
		t.Logf("err: %v", err)
	}
	uid2 := zkidentity.ShortID{}
	err = uid2.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde2")
	if err != nil {
		t.Logf("err: %v", err)
	}

	// Provide a mock stream that returns two ReceivedTip messages.
	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{
			{Uid: uid1[:], AmountMatoms: 100000, SequenceId: 1},
			{Uid: uid2[:], AmountMatoms: 200000, SequenceId: 2},
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

func TestReceiveTipLoop_EmptyStream(t *testing.T) {
	srv := setupTestServer(t)

	// Mock Payment Client
	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)

	// Mock DB
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	// Provide a mock stream that returns no ReceivedTip messages.
	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{},
	}

	// Expect exactly one call to TipStream
	mockPayClient.
		On("TipStream", mock.Anything, mock.Anything).
		Return(tipStream, nil).
		Once()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch the loop
	go func() {
		err := srv.ReceiveTipLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Stop the loop
	cancel()

	// Assert expectations for the payment client
	mockPayClient.AssertExpectations(t)
}

func TestReceiveTipLoop_EOF(t *testing.T) {
	srv := setupTestServer(t)

	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid1 := zkidentity.ShortID{}
	uid1.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{
			{Uid: uid1[:], AmountMatoms: 100000, SequenceId: 1},
		},
		ErrorAfter: -1, // No errors, ends with EOF
	}

	// Add FetchTip expectation
	mockDB.
		On("FetchTip", mock.Anything, uint64(1)).
		Return((*serverdb.ReceivedTipWrapper)(nil), nil).
		Once()

	mockPayClient.
		On("TipStream", mock.Anything, mock.Anything).
		Return(tipStream, nil).
		Once()

	mockDB.
		On("StoreUnprocessedTip", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := srv.ReceiveTipLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestReceiveTipLoop_CustomError(t *testing.T) {
	srv := setupTestServer(t)

	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid1 := zkidentity.ShortID{}
	uid1.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")
	uid2 := zkidentity.ShortID{}
	uid2.FromString("0123456789abcdef0123456789abcde20123456789abcdef0123456789abcde1")

	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{
			{Uid: uid1[:], AmountMatoms: 100000, SequenceId: 1},
			{Uid: uid2[:], AmountMatoms: 200000, SequenceId: 2},
		},
		ErrorAfter: 1, // Simulate error after the first tip
		RecvError:  context.DeadlineExceeded,
	}

	// Add missing FetchTip expectation for first tip
	mockDB.
		On("FetchTip", mock.Anything, uint64(1)).
		Return((*serverdb.ReceivedTipWrapper)(nil), nil).
		Once()

	mockPayClient.
		On("TipStream", mock.Anything, mock.Anything).
		Return(tipStream, nil).
		Once()

	mockDB.
		On("StoreUnprocessedTip", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := srv.ReceiveTipLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestReceiveTipLoop_PlayerSessionUpdate(t *testing.T) {
	srv := setupTestServer(t)

	var playerUID zkidentity.ShortID
	strID, err := GenerateRandomString(64)
	if err != nil {
		t.Errorf("Failed to GenerateRandomString for Host ID: %v", err)
		return
	}
	playerUID.FromString(strID)
	player := srv.gameManager.PlayerSessions.CreateSession(playerUID)
	// Mock Payment Client
	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)

	// Mock DB
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	// Add FetchTip expectation
	mockDB.
		On("FetchTip", mock.Anything, uint64(1)).
		Return((*serverdb.ReceivedTipWrapper)(nil), nil).
		Once()

	// Expect StoreUnprocessedTip to be called once
	mockDB.
		On("StoreUnprocessedTip", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	// Provide a mock stream with a tip for the player
	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{
			{
				Uid:          playerUID.Bytes(),
				AmountMatoms: 1500000000,
				SequenceId:   1,
			},
		},
	}

	// Expect exactly one call to TipStream
	mockPayClient.
		On("TipStream", mock.Anything, mock.Anything).
		Return(tipStream, nil).
		Once()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch the loop
	go func() {
		err := srv.ReceiveTipLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Stop the loop
	cancel()

	// Verify player bet amount updated
	if player.BetAmt != 0.015 {
		t.Errorf("expected BetAmt to be 0.015, got %f", player.BetAmt)
	}

	// Assert expectations
	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

// ************
// end test receive
// ************

// ************
// start test send tip
// ************
func TestSendTipProgressLoop_NormalOperation(t *testing.T) {
	srv := setupTestServer(t)

	// Mock Payment Client
	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)

	// Mock DB
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Provide a mock stream that returns progress events.
	progressStream := &mocks.MockTipProgressClient{
		Events: []types.TipProgressEvent{
			{Uid: uid[:], AmountMatoms: 100000, SequenceId: 1, Completed: true},
		},
	}

	// Expect exactly one call to TipProgress
	mockPayClient.
		On("TipProgress", mock.Anything, mock.Anything).
		Return(progressStream, nil).
		Once()

	// Expect AckTipProgress call
	mockPayClient.
		On("AckTipProgress", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	// Create the mock tip that will be in the progress record
	tipInProgress := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}

	// Mock the FetchLatestUncompletedTipProgress call
	mockDB.
		On("FetchLatestUncompletedTipProgress", mock.Anything, uid[:], int64(100000)).
		Return(&serverdb.TipProgressRecord{
			ID:          1,
			WinnerUID:   uid[:],
			TotalAmount: 100000,
			Status:      serverdb.StatusSending,
			Tips:        []*types.ReceivedTip{tipInProgress},
			CreatedAt:   time.Now(),
		}, nil).
		Once()

	// Expect UpdateTipStatus call for the tip in the progress record
	tipID := make([]byte, 8)
	binary.BigEndian.PutUint64(tipID, tipInProgress.SequenceId)
	mockDB.
		On("UpdateTipStatus", mock.Anything, tipInProgress.Uid, tipID, serverdb.StatusPaid).
		Return(nil).
		Once()

	// Expect AckTipReceived call for the tip in the progress record
	mockPayClient.
		On("AckTipReceived", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	// Mock the UpdateTipProgressStatus call
	mockDB.
		On("UpdateTipProgressStatus", mock.Anything, uint64(1), serverdb.StatusPaid).
		Return(nil).
		Once()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch the loop
	go func() {
		err := srv.SendTipProgressLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Stop the loop
	cancel()

	// Assert expectations
	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestSendTipProgressLoop_StreamError(t *testing.T) {
	srv := setupTestServer(t)

	// Mock Payment Client
	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)

	// Mock DB
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Provide a mock stream that returns an error after one event.
	progressStream := &mocks.MockTipProgressClient{
		Events: []types.TipProgressEvent{
			{Uid: uid[:], AmountMatoms: 100000, SequenceId: 1, Completed: false},
		},
		RecvError: errors.New("stream error"),
	}

	mockPayClient.
		On("TipProgress", mock.Anything, mock.Anything).
		Return(progressStream, nil).
		Once()

	mockPayClient.
		On("AckTipProgress", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Maybe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch the loop
	go func() {
		err := srv.SendTipProgressLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Stop the loop
	cancel()

	// Assert expectations
	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestSendTipProgressLoop_DBError(t *testing.T) {
	srv := setupTestServer(t)

	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	progressStream := &mocks.MockTipProgressClient{
		Events: []types.TipProgressEvent{
			{Uid: uid[:], AmountMatoms: 100000, SequenceId: 1, Completed: true},
		},
	}

	mockPayClient.On("TipProgress", mock.Anything, mock.Anything).Return(progressStream, nil).Once()

	// Add expectation for AckTipReceived that might be called
	mockPayClient.On("AckTipReceived", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()
	mockPayClient.On("AckTipProgress", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	// Add expectation for FetchLatestUncompletedTipProgress
	tipInProgress := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}
	mockDB.On("FetchLatestUncompletedTipProgress", mock.Anything, uid[:], int64(100000)).
		Return(&serverdb.TipProgressRecord{
			ID:          1,
			WinnerUID:   uid[:],
			TotalAmount: 100000,
			Status:      serverdb.StatusSending,
			Tips:        []*types.ReceivedTip{tipInProgress},
			CreatedAt:   time.Now(),
		}, nil).
		Once()

	mockDB.On("UpdateTipStatus", mock.Anything, uid[:], mock.Anything, serverdb.StatusPaid).
		Return(errors.New("database error")).Once()

	// Add expectation for UpdateTipProgressStatus
	mockDB.On("UpdateTipProgressStatus", mock.Anything, uint64(1), serverdb.StatusPaid).
		Return(nil).
		Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := srv.SendTipProgressLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestSendTipProgressLoop_UnprocessedTips(t *testing.T) {
	srv := setupTestServer(t)

	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	progressStream := &mocks.MockTipProgressClient{
		Events: []types.TipProgressEvent{
			{Uid: uid[:], AmountMatoms: 100000, SequenceId: 1, Completed: true},
			{Uid: uid[:], AmountMatoms: 200000, SequenceId: 2, Completed: true},
		},
	}

	mockPayClient.On("TipProgress", mock.Anything, mock.Anything).Return(progressStream, nil).Once()

	// Replace FetchReceivedTipsByUID with FetchLatestUncompletedTipProgress expectations
	// We need one expectation for each progress event
	tipInProgress1 := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}
	tipInProgress2 := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 200000,
		SequenceId:   2,
	}

	// First progress event
	mockDB.On("FetchLatestUncompletedTipProgress", mock.Anything, uid[:], int64(100000)).
		Return(&serverdb.TipProgressRecord{
			ID:          1,
			WinnerUID:   uid[:],
			TotalAmount: 100000,
			Status:      serverdb.StatusSending,
			Tips:        []*types.ReceivedTip{tipInProgress1},
			CreatedAt:   time.Now(),
		}, nil).
		Once()

	// Second progress event
	mockDB.On("FetchLatestUncompletedTipProgress", mock.Anything, uid[:], int64(200000)).
		Return(&serverdb.TipProgressRecord{
			ID:          2,
			WinnerUID:   uid[:],
			TotalAmount: 200000,
			Status:      serverdb.StatusSending,
			Tips:        []*types.ReceivedTip{tipInProgress2},
			CreatedAt:   time.Now(),
		}, nil).
		Once()

	// We expect 2 UpdateTipStatus calls (one for each progress event's tip)
	mockDB.On("UpdateTipStatus", mock.Anything, uid[:], mock.Anything, serverdb.StatusPaid).
		Return(nil).Times(2)

	// Expect 2 UpdateTipProgressStatus calls (one for each progress record)
	mockDB.On("UpdateTipProgressStatus", mock.Anything, uint64(1), serverdb.StatusPaid).
		Return(nil).Once()
	mockDB.On("UpdateTipProgressStatus", mock.Anything, uint64(2), serverdb.StatusPaid).
		Return(nil).Once()

	// Expect 2 AckTipProgress calls (1 per progress event)
	mockPayClient.On("AckTipProgress", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Twice()

	// Expect 2 AckTipReceived calls (1 per tip)
	mockPayClient.On("AckTipReceived", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Times(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := srv.SendTipProgressLoop(ctx)
		t.Logf("loop finished: %v", err)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestReceiveTipLoop_DBStoreError(t *testing.T) {
	srv := setupTestServer(t)

	mockPayClient := srv.paymentClient.(*mocks.MockPaymentClient)
	mockDB := &mocks.MockDB{}
	srv.db = mockDB

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	tipStream := &mocks.MockTipStreamClient{
		ReceivedTips: []*types.ReceivedTip{
			{Uid: uid[:], AmountMatoms: 100000, SequenceId: 1},
		},
	}

	mockPayClient.On("TipStream", mock.Anything, mock.Anything).Return(tipStream, nil).Once()

	// Add the missing FetchTip expectation
	mockDB.On("FetchTip", mock.Anything, uint64(1)).
		Return((*serverdb.ReceivedTipWrapper)(nil), nil).
		Once()

	// Simulate database storage error
	mockDB.On("StoreUnprocessedTip", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("storage error")).Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := srv.ReceiveTipLoop(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	mockPayClient.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
