package mocks

import (
	"context"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/stretchr/testify/mock"
	"github.com/vctt94/pong-bisonrelay/server/serverdb"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) StoreUnprocessedTip(ctx context.Context, tip *types.ReceivedTip) error {
	// If you want to track calls or return an error, do it here.
	m.Called(ctx, tip)
	return nil
}

func (m *MockDB) FetchUnprocessedTips(ctx context.Context) (map[zkidentity.ShortID][]*types.ReceivedTip, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[zkidentity.ShortID][]*types.ReceivedTip), args.Error(1)
}

// You also need to mock whatever else your code calls, for example:
func (m *MockDB) FetchReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID, status serverdb.TipStatus) ([]*types.ReceivedTip, error) {
	args := m.Called(ctx, uid, status)
	// The first return value is the list of tips. If you need to return a slice, use:
	if raw := args.Get(0); raw != nil {
		return raw.([]*types.ReceivedTip), args.Error(1)
	}
	return nil, args.Error(1)
}

// If you do UpdateTipStatus, you need that too, etc.
func (m *MockDB) UpdateTipStatus(ctx context.Context, uid, tipID []byte, status serverdb.TipStatus) error {
	m.Called(ctx, uid, tipID, status)
	return nil
}

func (m *MockDB) Close() error {
	// If you need special behavior, do it here. Otherwise:
	return nil
}

func (m *MockDB) FetchAllReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID) ([]serverdb.ReceivedTipWrapper, error) {
	args := m.Called(ctx, uid)
	// Return the values from the `mock.Called(...)`
	return args.Get(0).([]serverdb.ReceivedTipWrapper), args.Error(1)
}

func (m *MockDB) FetchTip(ctx context.Context, tipID uint64) (*serverdb.ReceivedTipWrapper, error) {
	args := m.Called(ctx, tipID)
	return args.Get(0).(*serverdb.ReceivedTipWrapper), args.Error(1)
}

func (m *MockDB) FetchLatestUncompletedTipProgress(ctx context.Context, winnerUID []byte, totalAmount int64) (*serverdb.TipProgressRecord, error) {
	args := m.Called(ctx, winnerUID, totalAmount)
	if args.Get(0) != nil {
		return args.Get(0).(*serverdb.TipProgressRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDB) FetchSendTipProgressByClient(ctx context.Context, clientID []byte) ([]*serverdb.TipProgressRecord, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) != nil {
		return args.Get(0).([]*serverdb.TipProgressRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDB) StoreSendTipProgress(ctx context.Context, winnerUID []byte, totalAmount int64, tips []*types.ReceivedTip, status serverdb.TipStatus) error {
	args := m.Called(ctx, winnerUID, totalAmount, tips, status)
	return args.Error(0)
}

func (m *MockDB) UpdateTipProgressStatus(ctx context.Context, recordID uint64, status serverdb.TipStatus) error {
	args := m.Called(ctx, recordID, status)
	return args.Error(0)
}
