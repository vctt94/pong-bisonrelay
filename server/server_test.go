package server

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vctt94/pong-bisonrelay/botlib"
	"github.com/vctt94/pong-bisonrelay/ponggame"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"github.com/vctt94/pong-bisonrelay/server/internal/mocks"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	mockDB := &mocks.MockDB{}
	mockPaymentClient := &mocks.MockPaymentClient{}

	// Remove the TipStream On(...) from here:
	// mockPaymentClient.On("TipStream", ...).Return(...)

	tempDir := t.TempDir()
	cfg := ServerConfig{
		ServerDir:     tempDir,
		MinBetAmt:     0.1,
		PaymentClient: mockPaymentClient,
		ChatClient:    &mocks.MockChatClient{},
	}

	var serverID zkidentity.ShortID
	_ = serverID.FromString("0123456789abcdef0123456789abcdef") // Dummy ID

	srv := NewServer(&serverID, cfg)
	srv.db = mockDB
	srv.gameManager = &ponggame.GameManager{
		ID:             &serverID,
		PlayerSessions: &ponggame.PlayerSessions{Sessions: make(map[zkidentity.ShortID]*ponggame.Player)},
	}
	return srv
}

func startInProcessGRPC(t *testing.T, srv *Server) (pong.PongGameClient, func()) {
	t.Helper()

	// Create the bufconn listener
	lis := bufconn.Listen(1024 * 1024)

	// Create and register the gRPC server
	grpcServer := grpc.NewServer()
	pong.RegisterPongGameServer(grpcServer, srv)

	// Serve in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Dial the in-process server
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
	)
	require.NoError(t, err, "failed to dial bufnet")

	client := pong.NewPongGameClient(conn)

	// Return a cleanup function
	cleanup := func() {
		grpcServer.Stop()
		conn.Close()
	}
	return client, cleanup
}

func TestCreateWaitingRoom(t *testing.T) {
	srv := setupTestServer(t) // create our server

	// Add mock expectations for FetchReceivedTipsByUID with a tip of sufficient amount
	mockDB := srv.db.(*mocks.MockDB)
	var hostID zkidentity.ShortID
	_ = hostID.FromString("11111111111111111111111111111111")
	mockDB.On("FetchReceivedTipsByUID", mock.Anything, hostID, serverdb.StatusUnpaid).
		Return([]*types.ReceivedTip{
			{
				AmountMatoms: 50000000000, // Match the requested bet amount
			},
		}, nil)

	client, cleanup := startInProcessGRPC(t, srv)
	defer cleanup()

	ctx := context.Background()

	player := srv.gameManager.PlayerSessions.CreateSession(hostID)
	player.BetAmt = 0.5 // Set the bet amount to match the request

	// Now call CreateWaitingRoom from the client side
	resp, err := client.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomRequest{
		HostId: hostID.String(),
		BetAmt: 0.5,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Wr)

	// Check the results
	require.Equal(t, 0.5, resp.Wr.BetAmt)
	t.Logf("Created waiting room: %s (bet=%.2f)", resp.Wr.Id, resp.Wr.BetAmt)

	// Possibly check if the server knows about it
	require.Len(t, srv.gameManager.WaitingRooms, 1)
	wr := srv.gameManager.WaitingRooms[0]
	require.Equal(t, hostID, *wr.HostID)
}

func TestCreateWaitingRoomWithInvalidBet(t *testing.T) {
	srv := setupTestServer(t)
	client, cleanup := startInProcessGRPC(t, srv)
	defer cleanup()

	ctx := context.Background()

	var hostID zkidentity.ShortID
	_ = hostID.FromString("11111111111111111111111111111111")
	player := srv.gameManager.PlayerSessions.CreateSession(hostID)

	// Add mock expectations for FetchReceivedTipsByUID
	mockDB := srv.db.(*mocks.MockDB)
	mockDB.On("FetchReceivedTipsByUID", mock.Anything, hostID, serverdb.StatusUnpaid).
		Return([]*types.ReceivedTip{
			{
				AmountMatoms: 50000000000,
			},
		}, nil)

	// For the low bet test
	player.BetAmt = 0.01 // Set the bet amount to match the request

	// Case: Bet amount less than MinBetAmt
	_, err := client.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomRequest{
		HostId: hostID.String(),
		BetAmt: 0.01, // Below the minimum bet
	})
	require.Error(t, err, "expected an error for invalid bet amount")
	require.Contains(t, err.Error(), "bet needs to be higher than 0.1")

	// For the valid bet test
	player.BetAmt = 0.5 // Set the bet amount to match the request

	// Case: Valid Bet
	resp, err := client.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomRequest{
		HostId: hostID.String(),
		BetAmt: 0.5,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0.5, resp.Wr.BetAmt)
}

func TestConcurrentWaitingRoomCreation(t *testing.T) {
	srv := setupTestServer(t)

	// We need to set up mock expectations here with a tip that has enough amount
	mockDB := srv.db.(*mocks.MockDB)
	mockDB.On("FetchReceivedTipsByUID", mock.Anything, mock.Anything, serverdb.StatusUnpaid).
		Return([]*types.ReceivedTip{
			{
				AmountMatoms: 50000000000, // Match the requested bet amount
			},
		}, nil)

	client, cleanup := startInProcessGRPC(t, srv)
	defer cleanup()

	ctx := context.Background()

	var wg sync.WaitGroup
	numRooms := 10
	wg.Add(numRooms)

	for i := 0; i < numRooms; i++ {
		go func(i int) {
			defer wg.Done()

			// Generate a unique Host ID for each goroutine
			var hostID zkidentity.ShortID
			strID, err := botlib.GenerateRandomString(64)
			if err != nil {
				t.Errorf("Failed to GenerateRandomString for Host ID: %v", err)
				return
			}
			if err := hostID.FromString(strID); err != nil {
				t.Errorf("Failed to convert string to Host ID: %v", err)
				return
			}
			player := srv.gameManager.PlayerSessions.CreateSession(hostID)
			player.BetAmt = 0.5 // Set the bet amount to match the request

			// Attempt to create a waiting room
			resp, err := client.CreateWaitingRoom(ctx, &pong.CreateWaitingRoomRequest{
				HostId: hostID.String(),
				BetAmt: 0.5,
			})
			if err != nil {
				t.Errorf("Failed to create waiting room for Host ID %s: %v", hostID.String(), err)
				return
			}
			t.Logf("Created room: %s (Host ID: %s)", resp.Wr.Id, hostID.String())
		}(i) // Pass the loop variable to the goroutine
	}

	wg.Wait()

	// Validate that all rooms were created
	require.Len(t, srv.gameManager.WaitingRooms, numRooms)
}

func TestGameStreamDisconnection(t *testing.T) {
	srv := setupTestServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var clientID zkidentity.ShortID
	err := clientID.FromString("1111111111111111111111111111111111111111111111111111111111111111")
	require.NoError(t, err)

	// Properly setup mock DB expectations
	mockDB := &mocks.MockDB{}
	srv.db = mockDB
	mockDB.On("FetchReceivedTipsByUID", mock.Anything, clientID, serverdb.StatusUnpaid).
		Return([]*types.ReceivedTip{}, nil)
	mockDB.On("UpdateTipStatus", mock.Anything, mock.Anything, serverdb.StatusSending).Return(nil)

	// Create player session before starting stream
	player := srv.gameManager.PlayerSessions.CreateSession(clientID)
	player.NotifierStream = &mocks.MockNtfnStreamServer{Ctx: ctx}
	player.BetAmt = 0.5
	// Setup mock stream
	mockStream := &mocks.MockGameStreamServer{Ctx: ctx}
	req := &pong.StartGameStreamRequest{
		ClientId: clientID.String(),
	}

	// Start stream in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srv.StartGameStream(req, mockStream)
		require.ErrorIs(t, err, context.Canceled)
	}()

	// Simulate disconnection
	cancel()
	wg.Wait()

	// Verify cleanup
	srv.gameManager.Lock()
	defer srv.gameManager.Unlock()
	require.Empty(t, srv.gameManager.PlayerSessions.Sessions)
}
