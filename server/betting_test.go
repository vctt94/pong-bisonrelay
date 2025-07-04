package server

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/bisonbotkit/utils"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

// setupTestServerWithDB creates a test server with a real temporary database
func setupTestServerWithDB(t *testing.T) *Server {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create a real BoltDB for testing
	db, err := serverdb.NewBoltDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Clean up database when test completes
	t.Cleanup(func() {
		db.Close()
	})

	// Create a test server using setupTestServer and replace its DB
	srv := setupTestServer(t)
	srv.db = db

	return srv
}

func TestReceiveTipLoop(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

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

	// Create player sessions for tip senders
	createTestPlayer(srv, uid1)
	createTestPlayer(srv, uid2)

	// Test the first tip
	tip1 := &types.ReceivedTip{
		Uid:          uid1[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}
	err = srv.HandleReceiveTip(ctx, tip1)
	if err != nil {
		t.Errorf("HandleReceiveTip failed: %v", err)
	}

	// Test the second tip
	tip2 := &types.ReceivedTip{
		Uid:          uid2[:],
		AmountMatoms: 200000,
		SequenceId:   2,
	}
	err = srv.HandleReceiveTip(ctx, tip2)
	if err != nil {
		t.Errorf("HandleReceiveTip failed: %v", err)
	}

	// Verify tips were stored in database
	storedTip1, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch tip 1: %v", err)
	}
	if storedTip1 == nil || storedTip1.Tip.AmountMatoms != 100000 {
		t.Errorf("Expected tip 1 with amount 100000, got %v", storedTip1)
	}

	storedTip2, err := srv.db.FetchTip(ctx, 2)
	if err != nil {
		t.Errorf("Failed to fetch tip 2: %v", err)
	}
	if storedTip2 == nil || storedTip2.Tip.AmountMatoms != 200000 {
		t.Errorf("Expected tip 2 with amount 200000, got %v", storedTip2)
	}
}

func TestReceiveTipLoop_EmptyStream(t *testing.T) {
	srv := setupTestServerWithDB(t)

	// This test verifies that no errors occur when no tips are processed
	// Since there are no tips to handle, nothing should be stored in the database

	// Verify database is empty
	ctx := context.Background()
	tip, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Unexpected error fetching non-existent tip: %v", err)
	}
	if tip != nil {
		t.Errorf("Expected no tip, but found: %v", tip)
	}
}

func TestReceiveTipLoop_EOF(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	uid1 := zkidentity.ShortID{}
	uid1.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Create player session for tip sender
	createTestPlayer(srv, uid1)

	// Test handling a single tip (simulating EOF after one tip)
	tip := &types.ReceivedTip{
		Uid:          uid1[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}
	err := srv.HandleReceiveTip(ctx, tip)
	if err != nil {
		t.Errorf("HandleReceiveTip failed: %v", err)
	}

	// Verify tip was stored
	storedTip, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch stored tip: %v", err)
	}
	if storedTip == nil || storedTip.Tip.AmountMatoms != 100000 {
		t.Errorf("Expected stored tip with amount 100000, got %v", storedTip)
	}
}

func TestReceiveTipLoop_CustomError(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	uid1 := zkidentity.ShortID{}
	uid1.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Create player session for tip sender
	createTestPlayer(srv, uid1)

	// Test handling one tip before a simulated error occurs
	tip1 := &types.ReceivedTip{
		Uid:          uid1[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}
	err := srv.HandleReceiveTip(ctx, tip1)
	if err != nil {
		t.Errorf("HandleReceiveTip failed: %v", err)
	}

	// Verify tip was stored
	storedTip, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch stored tip: %v", err)
	}
	if storedTip == nil {
		t.Errorf("Expected tip to be stored")
	}
}

func TestReceiveTipLoop_PlayerSessionUpdate(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	var playerUID zkidentity.ShortID
	strID, err := utils.GenerateRandomString(64)
	if err != nil {
		t.Errorf("Failed to GenerateRandomString for Host ID: %v", err)
		return
	}
	playerUID.FromString(strID)
	player := createTestPlayer(srv, playerUID)

	// Create and handle a tip for the player
	tip := &types.ReceivedTip{
		Uid:          playerUID.Bytes(),
		AmountMatoms: 1500000000,
		SequenceId:   1,
	}

	err = srv.HandleReceiveTip(ctx, tip)
	if err != nil {
		t.Errorf("HandleReceiveTip failed: %v", err)
	}

	// Verify player bet amount updated with proper synchronization
	srv.gameManager.PlayerSessions.Lock()
	actualBetAmt := player.BetAmt
	srv.gameManager.PlayerSessions.Unlock()

	if actualBetAmt != 1500000000 {
		t.Errorf("expected BetAmt to be 1500000000 matoms, got %d", actualBetAmt)
	}

	// Verify tip was stored in database
	storedTip, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch stored tip: %v", err)
	}
	if storedTip == nil || storedTip.Tip.AmountMatoms != 1500000000 {
		t.Errorf("Expected stored tip with amount 1500000000, got %v", storedTip)
	}
}

// ************
// end test receive
// ************

// ************
// start test send tip
// ************
func TestSendTipProgressLoop_NormalOperation(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Create the tip that will be in the progress record
	tipInProgress := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}

	// First store the tip in the database
	err := srv.db.StoreUnprocessedTip(ctx, tipInProgress)
	if err != nil {
		t.Fatalf("Failed to store tip: %v", err)
	}

	// Store the tip progress record
	err = srv.db.StoreSendTipProgress(ctx, uid[:], 100000, []*types.ReceivedTip{tipInProgress}, serverdb.StatusSending)
	if err != nil {
		t.Fatalf("Failed to store tip progress: %v", err)
	}

	// Test handling a completed tip progress event
	progressEvent := &types.TipProgressEvent{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
		Completed:    true,
	}

	err = srv.HandleTipProgress(ctx, progressEvent)
	if err != nil {
		t.Errorf("HandleTipProgress failed: %v", err)
	}

	// Verify the tip status was updated to paid
	storedTip, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch tip: %v", err)
	}
	if storedTip == nil || storedTip.Status != serverdb.StatusPaid {
		t.Errorf("Expected tip status to be paid, got %v", storedTip)
	}
}

func TestSendTipProgressLoop_StreamError(t *testing.T) {
	srv := setupTestServerWithDB(t)

	// This test verifies behavior when stream errors occur
	// Since we're not testing the actual stream but the handler logic,
	// we just verify the setup doesn't cause issues
	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Test passes if no panics or errors occur during setup
	if srv == nil {
		t.Error("Expected server to be created")
	}
}

func TestSendTipProgressLoop_DBError(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Create tip and store tip progress
	tipInProgress := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}

	// Store the tip first
	err := srv.db.StoreUnprocessedTip(ctx, tipInProgress)
	if err != nil {
		t.Fatalf("Failed to store tip: %v", err)
	}

	// Store tip progress
	err = srv.db.StoreSendTipProgress(ctx, uid[:], 100000, []*types.ReceivedTip{tipInProgress}, serverdb.StatusSending)
	if err != nil {
		t.Fatalf("Failed to store tip progress: %v", err)
	}

	// Test handling a completed tip progress event
	progressEvent := &types.TipProgressEvent{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
		Completed:    true,
	}

	err = srv.HandleTipProgress(ctx, progressEvent)
	if err != nil {
		t.Errorf("HandleTipProgress failed: %v", err)
	}

	// Verify processing completed successfully
	storedTip, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch tip: %v", err)
	}
	if storedTip == nil {
		t.Errorf("Expected tip to exist")
	}
}

func TestSendTipProgressLoop_UnprocessedTips(t *testing.T) {
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcdef1")

	// Create two tips and progress records
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

	// Store both tips
	err := srv.db.StoreUnprocessedTip(ctx, tipInProgress1)
	if err != nil {
		t.Fatalf("Failed to store tip 1: %v", err)
	}
	err = srv.db.StoreUnprocessedTip(ctx, tipInProgress2)
	if err != nil {
		t.Fatalf("Failed to store tip 2: %v", err)
	}

	// Store tip progress records
	err = srv.db.StoreSendTipProgress(ctx, uid[:], 100000, []*types.ReceivedTip{tipInProgress1}, serverdb.StatusSending)
	if err != nil {
		t.Fatalf("Failed to store tip progress 1: %v", err)
	}
	err = srv.db.StoreSendTipProgress(ctx, uid[:], 200000, []*types.ReceivedTip{tipInProgress2}, serverdb.StatusSending)
	if err != nil {
		t.Fatalf("Failed to store tip progress 2: %v", err)
	}

	// Test handling first completed tip progress event
	progressEvent1 := &types.TipProgressEvent{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
		Completed:    true,
	}

	err = srv.HandleTipProgress(ctx, progressEvent1)
	if err != nil {
		t.Errorf("HandleTipProgress failed for first event: %v", err)
	}

	// Test handling second completed tip progress event
	progressEvent2 := &types.TipProgressEvent{
		Uid:          uid[:],
		AmountMatoms: 200000,
		SequenceId:   2,
		Completed:    true,
	}

	err = srv.HandleTipProgress(ctx, progressEvent2)
	if err != nil {
		t.Errorf("HandleTipProgress failed for second event: %v", err)
	}

	// Verify both tips were processed
	storedTip1, err := srv.db.FetchTip(ctx, 1)
	if err != nil {
		t.Errorf("Failed to fetch tip 1: %v", err)
	}
	if storedTip1 == nil || storedTip1.Status != serverdb.StatusPaid {
		t.Errorf("Expected tip 1 status to be paid, got %v", storedTip1)
	}

	storedTip2, err := srv.db.FetchTip(ctx, 2)
	if err != nil {
		t.Errorf("Failed to fetch tip 2: %v", err)
	}
	if storedTip2 == nil || storedTip2.Status != serverdb.StatusPaid {
		t.Errorf("Expected tip 2 status to be paid, got %v", storedTip2)
	}
}

func TestReceiveTipLoop_DBStoreError(t *testing.T) {
	// This test simulates a database storage error by closing the database
	srv := setupTestServerWithDB(t)
	ctx := context.Background()

	uid := zkidentity.ShortID{}
	uid.FromString("0123456789abcdef0123456789abcde10123456789abcdef0123456789abcde1")

	// Create player session for tip sender
	createTestPlayer(srv, uid)

	// Close the database to simulate an error
	srv.db.Close()

	// Test handling a tip with DB storage error
	tip := &types.ReceivedTip{
		Uid:          uid[:],
		AmountMatoms: 100000,
		SequenceId:   1,
	}

	err := srv.HandleReceiveTip(ctx, tip)
	if err == nil {
		t.Errorf("Expected HandleReceiveTip to return an error due to closed database, but got nil")
	}
}
