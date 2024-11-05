package serverdb

import (
	"context"
	"errors"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
)

var ErrAlreadyStoredRV = errors.New("already stored tip")

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
	StoreUnprocessedTip(ctx context.Context, sequenceID []byte, payload *ReceivedTipWrapper) error
	FetchUnprocessedTips(ctx context.Context) (map[zkidentity.ShortID][]ReceivedTipWrapper, error)
	FetchReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID, status TipStatus) ([]ReceivedTipWrapper, error)
	UpdateTipStatus(ctx context.Context, uid []byte, tipID []byte, status TipStatus) error
	FetchAllReceivedTipsByUID(ctx context.Context, uid zkidentity.ShortID) ([]ReceivedTipWrapper, error)
	Close() error
}
