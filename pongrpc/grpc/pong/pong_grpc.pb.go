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
	// pong game
	SendInput(ctx context.Context, in *PlayerInput, opts ...grpc.CallOption) (*GameUpdate, error)
	StartGameStream(ctx context.Context, in *StartGameStreamRequest, opts ...grpc.CallOption) (PongGame_StartGameStreamClient, error)
	StartNtfnStream(ctx context.Context, in *StartNtfnStreamRequest, opts ...grpc.CallOption) (PongGame_StartNtfnStreamClient, error)
	UnreadyGameStream(ctx context.Context, in *UnreadyGameStreamRequest, opts ...grpc.CallOption) (*UnreadyGameStreamResponse, error)
	SignalReadyToPlay(ctx context.Context, in *SignalReadyToPlayRequest, opts ...grpc.CallOption) (*SignalReadyToPlayResponse, error)
	// waiting room
	GetWaitingRoom(ctx context.Context, in *WaitingRoomRequest, opts ...grpc.CallOption) (*WaitingRoomResponse, error)
	GetWaitingRooms(ctx context.Context, in *WaitingRoomsRequest, opts ...grpc.CallOption) (*WaitingRoomsResponse, error)
	CreateWaitingRoom(ctx context.Context, in *CreateWaitingRoomRequest, opts ...grpc.CallOption) (*CreateWaitingRoomResponse, error)
	JoinWaitingRoom(ctx context.Context, in *JoinWaitingRoomRequest, opts ...grpc.CallOption) (*JoinWaitingRoomResponse, error)
	LeaveWaitingRoom(ctx context.Context, in *LeaveWaitingRoomRequest, opts ...grpc.CallOption) (*LeaveWaitingRoomResponse, error)
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

func (c *pongGameClient) StartGameStream(ctx context.Context, in *StartGameStreamRequest, opts ...grpc.CallOption) (PongGame_StartGameStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &PongGame_ServiceDesc.Streams[0], "/pong.PongGame/StartGameStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &pongGameStartGameStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PongGame_StartGameStreamClient interface {
	Recv() (*GameUpdateBytes, error)
	grpc.ClientStream
}

type pongGameStartGameStreamClient struct {
	grpc.ClientStream
}

func (x *pongGameStartGameStreamClient) Recv() (*GameUpdateBytes, error) {
	m := new(GameUpdateBytes)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pongGameClient) StartNtfnStream(ctx context.Context, in *StartNtfnStreamRequest, opts ...grpc.CallOption) (PongGame_StartNtfnStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &PongGame_ServiceDesc.Streams[1], "/pong.PongGame/StartNtfnStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &pongGameStartNtfnStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PongGame_StartNtfnStreamClient interface {
	Recv() (*NtfnStreamResponse, error)
	grpc.ClientStream
}

type pongGameStartNtfnStreamClient struct {
	grpc.ClientStream
}

func (x *pongGameStartNtfnStreamClient) Recv() (*NtfnStreamResponse, error) {
	m := new(NtfnStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pongGameClient) UnreadyGameStream(ctx context.Context, in *UnreadyGameStreamRequest, opts ...grpc.CallOption) (*UnreadyGameStreamResponse, error) {
	out := new(UnreadyGameStreamResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/UnreadyGameStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) SignalReadyToPlay(ctx context.Context, in *SignalReadyToPlayRequest, opts ...grpc.CallOption) (*SignalReadyToPlayResponse, error) {
	out := new(SignalReadyToPlayResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/SignalReadyToPlay", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) GetWaitingRoom(ctx context.Context, in *WaitingRoomRequest, opts ...grpc.CallOption) (*WaitingRoomResponse, error) {
	out := new(WaitingRoomResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/GetWaitingRoom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) GetWaitingRooms(ctx context.Context, in *WaitingRoomsRequest, opts ...grpc.CallOption) (*WaitingRoomsResponse, error) {
	out := new(WaitingRoomsResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/GetWaitingRooms", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) CreateWaitingRoom(ctx context.Context, in *CreateWaitingRoomRequest, opts ...grpc.CallOption) (*CreateWaitingRoomResponse, error) {
	out := new(CreateWaitingRoomResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/CreateWaitingRoom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) JoinWaitingRoom(ctx context.Context, in *JoinWaitingRoomRequest, opts ...grpc.CallOption) (*JoinWaitingRoomResponse, error) {
	out := new(JoinWaitingRoomResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/JoinWaitingRoom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pongGameClient) LeaveWaitingRoom(ctx context.Context, in *LeaveWaitingRoomRequest, opts ...grpc.CallOption) (*LeaveWaitingRoomResponse, error) {
	out := new(LeaveWaitingRoomResponse)
	err := c.cc.Invoke(ctx, "/pong.PongGame/LeaveWaitingRoom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PongGameServer is the server API for PongGame service.
// All implementations must embed UnimplementedPongGameServer
// for forward compatibility
type PongGameServer interface {
	// pong game
	SendInput(context.Context, *PlayerInput) (*GameUpdate, error)
	StartGameStream(*StartGameStreamRequest, PongGame_StartGameStreamServer) error
	StartNtfnStream(*StartNtfnStreamRequest, PongGame_StartNtfnStreamServer) error
	UnreadyGameStream(context.Context, *UnreadyGameStreamRequest) (*UnreadyGameStreamResponse, error)
	SignalReadyToPlay(context.Context, *SignalReadyToPlayRequest) (*SignalReadyToPlayResponse, error)
	// waiting room
	GetWaitingRoom(context.Context, *WaitingRoomRequest) (*WaitingRoomResponse, error)
	GetWaitingRooms(context.Context, *WaitingRoomsRequest) (*WaitingRoomsResponse, error)
	CreateWaitingRoom(context.Context, *CreateWaitingRoomRequest) (*CreateWaitingRoomResponse, error)
	JoinWaitingRoom(context.Context, *JoinWaitingRoomRequest) (*JoinWaitingRoomResponse, error)
	LeaveWaitingRoom(context.Context, *LeaveWaitingRoomRequest) (*LeaveWaitingRoomResponse, error)
	mustEmbedUnimplementedPongGameServer()
}

// UnimplementedPongGameServer must be embedded to have forward compatible implementations.
type UnimplementedPongGameServer struct {
}

func (UnimplementedPongGameServer) SendInput(context.Context, *PlayerInput) (*GameUpdate, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendInput not implemented")
}
func (UnimplementedPongGameServer) StartGameStream(*StartGameStreamRequest, PongGame_StartGameStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method StartGameStream not implemented")
}
func (UnimplementedPongGameServer) StartNtfnStream(*StartNtfnStreamRequest, PongGame_StartNtfnStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method StartNtfnStream not implemented")
}
func (UnimplementedPongGameServer) UnreadyGameStream(context.Context, *UnreadyGameStreamRequest) (*UnreadyGameStreamResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnreadyGameStream not implemented")
}
func (UnimplementedPongGameServer) SignalReadyToPlay(context.Context, *SignalReadyToPlayRequest) (*SignalReadyToPlayResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignalReadyToPlay not implemented")
}
func (UnimplementedPongGameServer) GetWaitingRoom(context.Context, *WaitingRoomRequest) (*WaitingRoomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWaitingRoom not implemented")
}
func (UnimplementedPongGameServer) GetWaitingRooms(context.Context, *WaitingRoomsRequest) (*WaitingRoomsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWaitingRooms not implemented")
}
func (UnimplementedPongGameServer) CreateWaitingRoom(context.Context, *CreateWaitingRoomRequest) (*CreateWaitingRoomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWaitingRoom not implemented")
}
func (UnimplementedPongGameServer) JoinWaitingRoom(context.Context, *JoinWaitingRoomRequest) (*JoinWaitingRoomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinWaitingRoom not implemented")
}
func (UnimplementedPongGameServer) LeaveWaitingRoom(context.Context, *LeaveWaitingRoomRequest) (*LeaveWaitingRoomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LeaveWaitingRoom not implemented")
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

func _PongGame_StartGameStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StartGameStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PongGameServer).StartGameStream(m, &pongGameStartGameStreamServer{stream})
}

type PongGame_StartGameStreamServer interface {
	Send(*GameUpdateBytes) error
	grpc.ServerStream
}

type pongGameStartGameStreamServer struct {
	grpc.ServerStream
}

func (x *pongGameStartGameStreamServer) Send(m *GameUpdateBytes) error {
	return x.ServerStream.SendMsg(m)
}

func _PongGame_StartNtfnStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StartNtfnStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PongGameServer).StartNtfnStream(m, &pongGameStartNtfnStreamServer{stream})
}

type PongGame_StartNtfnStreamServer interface {
	Send(*NtfnStreamResponse) error
	grpc.ServerStream
}

type pongGameStartNtfnStreamServer struct {
	grpc.ServerStream
}

func (x *pongGameStartNtfnStreamServer) Send(m *NtfnStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _PongGame_UnreadyGameStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnreadyGameStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).UnreadyGameStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/UnreadyGameStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).UnreadyGameStream(ctx, req.(*UnreadyGameStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_SignalReadyToPlay_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignalReadyToPlayRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).SignalReadyToPlay(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/SignalReadyToPlay",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).SignalReadyToPlay(ctx, req.(*SignalReadyToPlayRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_GetWaitingRoom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitingRoomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).GetWaitingRoom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/GetWaitingRoom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).GetWaitingRoom(ctx, req.(*WaitingRoomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_GetWaitingRooms_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitingRoomsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).GetWaitingRooms(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/GetWaitingRooms",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).GetWaitingRooms(ctx, req.(*WaitingRoomsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_CreateWaitingRoom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateWaitingRoomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).CreateWaitingRoom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/CreateWaitingRoom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).CreateWaitingRoom(ctx, req.(*CreateWaitingRoomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_JoinWaitingRoom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinWaitingRoomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).JoinWaitingRoom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/JoinWaitingRoom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).JoinWaitingRoom(ctx, req.(*JoinWaitingRoomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PongGame_LeaveWaitingRoom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LeaveWaitingRoomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PongGameServer).LeaveWaitingRoom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pong.PongGame/LeaveWaitingRoom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PongGameServer).LeaveWaitingRoom(ctx, req.(*LeaveWaitingRoomRequest))
	}
	return interceptor(ctx, in, info, handler)
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
		{
			MethodName: "UnreadyGameStream",
			Handler:    _PongGame_UnreadyGameStream_Handler,
		},
		{
			MethodName: "SignalReadyToPlay",
			Handler:    _PongGame_SignalReadyToPlay_Handler,
		},
		{
			MethodName: "GetWaitingRoom",
			Handler:    _PongGame_GetWaitingRoom_Handler,
		},
		{
			MethodName: "GetWaitingRooms",
			Handler:    _PongGame_GetWaitingRooms_Handler,
		},
		{
			MethodName: "CreateWaitingRoom",
			Handler:    _PongGame_CreateWaitingRoom_Handler,
		},
		{
			MethodName: "JoinWaitingRoom",
			Handler:    _PongGame_JoinWaitingRoom_Handler,
		},
		{
			MethodName: "LeaveWaitingRoom",
			Handler:    _PongGame_LeaveWaitingRoom_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StartGameStream",
			Handler:       _PongGame_StartGameStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StartNtfnStream",
			Handler:       _PongGame_StartNtfnStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pong.proto",
}
