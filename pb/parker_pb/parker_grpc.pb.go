// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.1
// source: parker.proto

package parker_pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Parker_Lookup_FullMethodName = "/parker_server.Parker/Lookup"
)

// ParkerClient is the client API for Parker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ParkerClient interface {
	// / Gateway => Parker
	Lookup(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error)
}

type parkerClient struct {
	cc grpc.ClientConnInterface
}

func NewParkerClient(cc grpc.ClientConnInterface) ParkerClient {
	return &parkerClient{cc}
}

func (c *parkerClient) Lookup(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LookupResponse)
	err := c.cc.Invoke(ctx, Parker_Lookup_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ParkerServer is the server API for Parker service.
// All implementations must embed UnimplementedParkerServer
// for forward compatibility.
type ParkerServer interface {
	// / Gateway => Parker
	Lookup(context.Context, *LookupRequest) (*LookupResponse, error)
	mustEmbedUnimplementedParkerServer()
}

// UnimplementedParkerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedParkerServer struct{}

func (UnimplementedParkerServer) Lookup(context.Context, *LookupRequest) (*LookupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lookup not implemented")
}
func (UnimplementedParkerServer) mustEmbedUnimplementedParkerServer() {}
func (UnimplementedParkerServer) testEmbeddedByValue()                {}

// UnsafeParkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ParkerServer will
// result in compilation errors.
type UnsafeParkerServer interface {
	mustEmbedUnimplementedParkerServer()
}

func RegisterParkerServer(s grpc.ServiceRegistrar, srv ParkerServer) {
	// If the following call pancis, it indicates UnimplementedParkerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Parker_ServiceDesc, srv)
}

func _Parker_Lookup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ParkerServer).Lookup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Parker_Lookup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ParkerServer).Lookup(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Parker_ServiceDesc is the grpc.ServiceDesc for Parker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Parker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "parker_server.Parker",
	HandlerType: (*ParkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Lookup",
			Handler:    _Parker_Lookup_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "parker.proto",
}
