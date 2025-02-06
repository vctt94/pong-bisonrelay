package serverdb

import (
	"context"
	"errors"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
)

var (
	ErrDuplicateEntry     = errors.New("tip already stored")
	ErrMainBucketNotFound = errors.New("main bucket not found")
	ErrUserBucketNotFound = errors.New("user bucket not found")
	ErrTipNotFound        = errors.New("tip not found")
	ErrTipBucketNotFound  = errors.New("tip bucket not found")
)

type TipStatus string

const (
	StatusUnprocessed TipStatus = "unprocessed"
	StatusSending     TipStatus = "sending"
	StatusProcessed   TipStatus = "processed"
)

type ReceivedTipWrapper struct {
	Tip    *types.ReceivedTip
	Status TipStatus
}

type FetchUnprocessedTipsResult struct {
	UnprocessedTips map[zkidentity.ShortID][]types.ReceivedTip
}

type ServerDB interface {
	StoreUnprocessedTip(ctx context.Context, tip *types.ReceivedTip) error
	FetchUnprocessedTips(ctx context.Context) (map[zkidentity.ShortID][]*types.ReceivedTip, error)
	FetchTip(ctx context.Context, tipID uint64) (*ReceivedTipWrapper, error)
	FetchReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID, status TipStatus) ([]*types.ReceivedTip, error)
	UpdateTipStatus(ctx context.Context, uid []byte, tipID []byte, status TipStatus) error
	FetchAllReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID) ([]ReceivedTipWrapper, error)
	Close() error
}
