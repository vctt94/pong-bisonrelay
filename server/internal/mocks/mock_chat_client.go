package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/companyzero/bisonrelay/clientrpc/types"
)

// MockChatClient is a mock implementation of types.ChatServiceServer
// (even though it's named "Server" in the proto, you can use it as a client).
type MockChatClient struct {
	mock.Mock
}

// UserPublicIdentity mocks the corresponding RPC.
func (m *MockChatClient) UserPublicIdentity(
	ctx context.Context,
	req *types.PublicIdentityReq,
	resp *types.PublicIdentity,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// PM mocks the corresponding RPC.
func (m *MockChatClient) PM(
	ctx context.Context,
	req *types.PMRequest,
	resp *types.PMResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// PMStream mocks the PM streaming RPC.
func (m *MockChatClient) PMStream(
	ctx context.Context,
	req *types.PMStreamRequest,
) (types.ChatService_PMStreamClient, error) {
	args := m.Called(ctx, req)
	stream, _ := args.Get(0).(types.ChatService_PMStreamClient)
	// The second return is an error
	return stream, args.Error(1)
}

// AckReceivedPM mocks the corresponding RPC.
func (m *MockChatClient) AckReceivedPM(
	ctx context.Context,
	req *types.AckRequest,
	resp *types.AckResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// GCM mocks the corresponding RPC.
func (m *MockChatClient) GCM(
	ctx context.Context,
	req *types.GCMRequest,
	resp *types.GCMResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// Example method: GCMStream
// Notice it returns (ChatService_GCMStreamClient, error) as a client interface
func (m *MockChatClient) GCMStream(
	ctx context.Context,
	req *types.GCMStreamRequest,
) (types.ChatService_GCMStreamClient, error) {
	// Use testify's .Called(...) to record arguments and obtain return values
	args := m.Called(ctx, req)
	// The first return is a ChatService_GCMStreamClient
	stream, _ := args.Get(0).(types.ChatService_GCMStreamClient)
	// The second return is an error
	return stream, args.Error(1)
}

// AckReceivedGCM mocks the corresponding RPC.
func (m *MockChatClient) AckReceivedGCM(
	ctx context.Context,
	req *types.AckRequest,
	resp *types.AckResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// MediateKX mocks the corresponding RPC.
func (m *MockChatClient) MediateKX(
	ctx context.Context,
	req *types.MediateKXRequest,
	resp *types.MediateKXResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// KXStream mocks the KX streaming RPC.
func (m *MockChatClient) KXStream(
	ctx context.Context,
	req *types.KXStreamRequest,
) (types.ChatService_KXStreamClient, error) {
	args := m.Called(ctx, req)
	// The first return is a ChatService_GCMStreamClient
	stream, _ := args.Get(0).(types.ChatService_KXStreamClient)
	// The second return is an error
	return stream, args.Error(1)
}

// AckKXCompleted mocks the corresponding RPC.
func (m *MockChatClient) AckKXCompleted(
	ctx context.Context,
	req *types.AckRequest,
	resp *types.AckResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// WriteNewInvite mocks the corresponding RPC.
func (m *MockChatClient) WriteNewInvite(
	ctx context.Context,
	req *types.WriteNewInviteRequest,
	resp *types.WriteNewInviteResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// AcceptInvite mocks the corresponding RPC.
func (m *MockChatClient) AcceptInvite(
	ctx context.Context,
	req *types.AcceptInviteRequest,
	resp *types.AcceptInviteResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// SendFile mocks the corresponding RPC.
func (m *MockChatClient) SendFile(
	ctx context.Context,
	req *types.SendFileRequest,
	resp *types.SendFileResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}

// UserNick mocks the corresponding RPC.
func (m *MockChatClient) UserNick(
	ctx context.Context,
	req *types.UserNickRequest,
	resp *types.UserNickResponse,
) error {
	args := m.Called(ctx, req, resp)
	return args.Error(0)
}
