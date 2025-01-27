package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
	"google.golang.org/grpc/metadata"
)

type MockGameStreamServer struct {
	mock.Mock
	Ctx context.Context
}

func (m *MockGameStreamServer) Send(update *pong.GameUpdateBytes) error {
	args := m.Called(update)
	return args.Error(0)
}

func (m *MockGameStreamServer) SetHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

func (m *MockGameStreamServer) SendHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

func (m *MockGameStreamServer) SetTrailer(md metadata.MD) {
	m.Called(md)
}

func (m *MockGameStreamServer) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockGameStreamServer) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockGameStreamServer) Context() context.Context {
	return m.Ctx
}

// MockNtfnStreamServer implements pong.PongGame_StartNtfnStreamServer
type MockNtfnStreamServer struct {
	Ctx context.Context
}

func (m *MockNtfnStreamServer) Send(resp *pong.NtfnStreamResponse) error {
	return nil
}

func (m *MockNtfnStreamServer) SetHeader(metadata.MD) error  { return nil }
func (m *MockNtfnStreamServer) SendHeader(metadata.MD) error { return nil }
func (m *MockNtfnStreamServer) SetTrailer(metadata.MD)       {}
func (m *MockNtfnStreamServer) Context() context.Context     { return m.Ctx }
func (m *MockNtfnStreamServer) SendMsg(interface{}) error    { return nil }
func (m *MockNtfnStreamServer) RecvMsg(interface{}) error    { return nil }
