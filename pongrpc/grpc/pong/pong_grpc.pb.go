// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: pong.proto

package pong

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PongGameClient is the client API for PongGame service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PongGameClient interface {
	SendInput(ctx context.Context, in *PlayerInput, opts ...grpc.CallOption) (*GameUpdate, error)
	SignalReady(ctx context.Context, in *SignalReadyRequest, opts ...grpc.CallOption) (PongGame_SignalReadyClient, error)
	Init(ctx context.Context, in *GameStartedStreamRequest, opts ...grpc.CallOption) (PongGame_InitClient, error)
}

type pongGameClient struct {
	cc grpc.ClientConnInterface
}

func NewPongGameClient(cc grpc.ClientConnInterface) PongGameClient {
	return &pongGameClient{cc}
}

func (c *pongGameClient) SendInput(ctx context.Context, in *PlayerInput, opts ...grpc.CallOption) (*GameUpdate, error) {
	out := new(GameUpdate)
	err := c.cc.Invoke(ctx, "/pong.PongGame/SendInput", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) SignalReady(ctx context.Context, in *SignalReadyRequest, opts ...grpc.CallOption) (PongGame_SignalReadyClient, error) {
	stream, err := c.cc.NewStream(ctx, &PongGame_ServiceDesc.Streams[0], "/pong.PongGame/SignalReady", opts...)
	if err != nil {
		return nil, err
	}
	x := &pongGameSignalReadyClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PongGame_SignalReadyClient interface {
	Recv() (*GameUpdateBytes, error)
	grpc.ClientStream
}

type pongGameSignalReadyClient struct {
	grpc.ClientStream
}

func (x *pongGameSignalReadyClient) Recv() (*GameUpdateBytes, error) {
	m := new(GameUpdateBytes)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pongGameClient) Init(ctx context.Context, in *GameStartedStreamRequest, opts ...grpc.CallOption) (PongGame_InitClient, error) {
	stream, err := c.cc.NewStream(ctx, &PongGame_ServiceDesc.Streams[1], "/pong.PongGame/Init", opts...)
	if err != nil {
		return nil, err
	}
	x := &pongGameInitClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PongGame_InitClient interface {
	Recv() (*GameStartedStreamResponse, error)
	grpc.ClientStream
}

type pongGameInitClient struct {
	grpc.ClientStream
}

func (x *pongGameInitClient) Recv() (*GameStartedStreamResponse, error) {
	m := new(GameStartedStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PongGameServer is the server API for PongGame service.
// All implementations must embed UnimplementedPongGameServer
// for forward compatibility
type PongGameServer interface {
	SendInput(context.Context, *PlayerInput) (*GameUpdate, error)
	SignalReady(*SignalReadyRequest, PongGame_SignalReadyServer) error
	Init(*GameStartedStreamRequest, PongGame_InitServer) error
	mustEmbedUnimplementedPongGameServer()
}

// UnimplementedPongGameServer must be embedded to have forward compatible implementations.
type UnimplementedPongGameServer struct {
}

func (UnimplementedPongGameServer) SendInput(context.Context, *PlayerInput) (*GameUpdate, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendInput not implemented")
}
func (UnimplementedPongGameServer) SignalReady(*SignalReadyRequest, PongGame_SignalReadyServer) error {
	return status.Errorf(codes.Unimplemented, "method SignalReady not implemented")
}
func (UnimplementedPongGameServer) Init(*GameStartedStreamRequest, PongGame_InitServer) error {
	return status.Errorf(codes.Unimplemented, "method Init not implemented")
}
func (UnimplementedPongGameServer) mustEmbedUnimplementedPongGameServer() {}

// UnsafePongGameServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PongGameServer will
// result in compilation errors.
type UnsafePongGameServer interface {
	mustEmbedUnimplementedPongGameServer()
}

func RegisterPongGameServer(s grpc.ServiceRegistrar, srv PongGameServer) {
	s.RegisterService(&PongGame_ServiceDesc, srv)
}

func _PongGame_SendInput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PlayerInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).SendInput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/SendInput",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).SendInput(ctx, req.(*PlayerInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_SignalReady_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SignalReadyRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PongGameServer).SignalReady(m, &pongGameSignalReadyServer{stream})
}

type PongGame_SignalReadyServer interface {
	Send(*GameUpdateBytes) error
	grpc.ServerStream
}

type pongGameSignalReadyServer struct {
	grpc.ServerStream
}

func (x *pongGameSignalReadyServer) Send(m *GameUpdateBytes) error {
	return x.ServerStream.SendMsg(m)
}

func _PongGame_Init_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GameStartedStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PongGameServer).Init(m, &pongGameInitServer{stream})
}

type PongGame_InitServer interface {
	Send(*GameStartedStreamResponse) error
	grpc.ServerStream
}

type pongGameInitServer struct {
	grpc.ServerStream
}

func (x *pongGameInitServer) Send(m *GameStartedStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// PongGame_ServiceDesc is the grpc.ServiceDesc for PongGame service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PongGame_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pong.PongGame",
	HandlerType: (*PongGameServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendInput",
			Handler:    _PongGame_SendInput_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SignalReady",
			Handler:       _PongGame_SignalReady_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Init",
			Handler:       _PongGame_Init_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pong.proto",
}
