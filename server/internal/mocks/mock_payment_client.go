package mocks

import (
	"context"
	"io"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

// MockPaymentClient is a mock implementation of the PaymentsServiceClient
// interface using testify's mock package.
type MockPaymentClient struct {
	mock.Mock
	ReceivedTips []*types.ReceivedTip
	CurrentIndex int
}

// ----- TipStream mock ----- //

type MockTipStreamClient struct {
	// For each Recv() call, you’ll return the next ReceivedTip in this slice
	ReceivedTips []*types.ReceivedTip
	CurrentIndex int
	ErrorAfter   int   // Return an error after this many Recv() calls
	RecvError    error // The error to return after ErrorAfter Recv() calls

	// You can embed a mock.Mock if you want to track calls:
	mock.Mock
}

// Recv implements types.PaymentsService_TipStreamClient.Recv(*ReceivedTip) error
func (m *MockTipStreamClient) Recv(tip *types.ReceivedTip) error {
	if m.ErrorAfter > 0 && m.CurrentIndex >= m.ErrorAfter {
		return m.RecvError
	}
	if m.CurrentIndex >= len(m.ReceivedTips) {
		return io.EOF
	}
	*tip = *m.ReceivedTips[m.CurrentIndex]
	m.CurrentIndex++
	return nil
}

// CloseSend (optional, if needed in your real interface)
func (m *MockTipStreamClient) CloseSend() error {
	return nil
}

// The following methods are often required by grpc.ClientStream:

func (m *MockTipStreamClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *MockTipStreamClient) Trailer() metadata.MD {
	return nil
}

func (m *MockTipStreamClient) CloseSendCtx() error {
	return nil
}

// Context returns the current context for the stream.
func (m *MockTipStreamClient) Context() context.Context {
	return context.Background()
}

// SendMsg is part of grpc.ClientStream
func (m *MockTipStreamClient) SendMsg(msg interface{}) error {
	return nil
}

// RecvMsg is part of grpc.ClientStream
func (m *MockTipStreamClient) RecvMsg(msg interface{}) error {
	return nil
}

// ----- TipProgress mock ----- //

type MockTipProgressClient struct {
	mock.Mock
	Events       []types.TipProgressEvent
	CurrentIndex int
	ErrorAfter   int   // Return an error after this many Recv() calls
	RecvError    error // The error to return after ErrorAfter Recv() calls
}

func (m *MockTipProgressClient) Recv(tip *types.TipProgressEvent) error {
	if m.ErrorAfter > 0 && m.CurrentIndex >= m.ErrorAfter {
		return m.RecvError
	}
	if m.CurrentIndex >= len(m.Events) {
		return io.EOF
	}
	*tip = m.Events[m.CurrentIndex]
	m.CurrentIndex++
	return nil
}

// You may need similar gRPC stubs for TipProgress, if your code uses them:

func (m *MockTipProgressClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *MockTipProgressClient) Trailer() metadata.MD {
	return nil
}

func (m *MockTipProgressClient) CloseSend() error {
	return nil
}

func (m *MockTipProgressClient) Context() context.Context {
	return context.Background()
}

func (m *MockTipProgressClient) SendMsg(msg interface{}) error {
	return nil
}

func (m *MockTipProgressClient) RecvMsg(msg interface{}) error {
	return nil
}

// ----- MockPaymentClient methods ----- //

// TipUser attempts to send a tip to a user.
func (m *MockPaymentClient) TipUser(
	ctx context.Context,
	in *types.TipUserRequest,
	out *types.TipUserResponse,
) error {
	args := m.Called(ctx, in, out)
	return args.Error(0)
}

// TipProgress starts a stream that receives events about the progress of
// TipUser requests.
func (m *MockPaymentClient) TipProgress(
	ctx context.Context,
	in *types.TipProgressRequest,
) (types.PaymentsService_TipProgressClient, error) {

	args := m.Called(ctx, in)
	// Return a *MockTipProgressClient (not nil!) so your code can Recv() from it.
	stream, _ := args.Get(0).(types.PaymentsService_TipProgressClient)
	return stream, args.Error(1)
}

// AckTipProgress acknowledges events received up to a given sequence_id.
func (m *MockPaymentClient) AckTipProgress(
	ctx context.Context,
	in *types.AckRequest,
	out *types.AckResponse,
) error {
	args := m.Called(ctx, in, out)
	return args.Error(0)
}

// TipStream returns a stream that gets tips received by the client.
func (m *MockPaymentClient) TipStream(
	ctx context.Context,
	in *types.TipStreamRequest,
) (types.PaymentsService_TipStreamClient, error) {

	args := m.Called(ctx, in)
	// Return a *MockTipStreamClient (not nil!) so your code can Recv() from it.
	stream, _ := args.Get(0).(types.PaymentsService_TipStreamClient)
	return stream, args.Error(1)
}

// AckTipReceived acknowledges tip(s) received up to a given sequence_id.
func (m *MockPaymentClient) AckTipReceived(
	ctx context.Context,
	in *types.AckRequest,
	out *types.AckResponse,
) error {
	args := m.Called(ctx, in, out)
	return args.Error(0)
}
