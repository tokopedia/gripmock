// Code generated by protoc-gen-go. DO NOT EDIT.
// source: well-known-types.proto

package main

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import empty "github.com/golang/protobuf/ptypes/empty"
import _struct "github.com/golang/protobuf/ptypes/struct"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GripmockClient is the client API for Gripmock service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GripmockClient interface {
	HealthCheck(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*_struct.Struct, error)
}

type gripmockClient struct {
	cc *grpc.ClientConn
}

func NewGripmockClient(cc *grpc.ClientConn) GripmockClient {
	return &gripmockClient{cc}
}

func (c *gripmockClient) HealthCheck(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*_struct.Struct, error) {
	out := new(_struct.Struct)
	err := c.cc.Invoke(ctx, "/main.Gripmock/HealthCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GripmockServer is the server API for Gripmock service.
type GripmockServer interface {
	HealthCheck(context.Context, *empty.Empty) (*_struct.Struct, error)
}

func RegisterGripmockServer(s *grpc.Server, srv GripmockServer) {
	s.RegisterService(&_Gripmock_serviceDesc, srv)
}

func _Gripmock_HealthCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GripmockServer).HealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.Gripmock/HealthCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GripmockServer).HealthCheck(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gripmock_serviceDesc = grpc.ServiceDesc{
	ServiceName: "main.Gripmock",
	HandlerType: (*GripmockServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HealthCheck",
			Handler:    _Gripmock_HealthCheck_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "well-known-types.proto",
}

func init() {
	proto.RegisterFile("well-known-types.proto", fileDescriptor_well_known_types_263ba476363344a2)
}

var fileDescriptor_well_known_types_263ba476363344a2 = []byte{
	// 139 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2b, 0x4f, 0xcd, 0xc9,
	0xd1, 0xcd, 0xce, 0xcb, 0x2f, 0xcf, 0xd3, 0x2d, 0xa9, 0x2c, 0x48, 0x2d, 0xd6, 0x2b, 0x28, 0xca,
	0x2f, 0xc9, 0x17, 0x62, 0xc9, 0x4d, 0xcc, 0xcc, 0x93, 0x92, 0x4e, 0xcf, 0xcf, 0x4f, 0xcf, 0x49,
	0xd5, 0x07, 0x8b, 0x25, 0x95, 0xa6, 0xe9, 0xa7, 0xe6, 0x16, 0x94, 0x54, 0x42, 0x94, 0x48, 0xc9,
	0xa0, 0x4b, 0x16, 0x97, 0x14, 0x95, 0x26, 0x97, 0x40, 0x64, 0x8d, 0xbc, 0xb8, 0x38, 0xdc, 0x8b,
	0x32, 0x0b, 0x72, 0xf3, 0x93, 0xb3, 0x85, 0xec, 0xb8, 0xb8, 0x3d, 0x52, 0x13, 0x73, 0x4a, 0x32,
	0x9c, 0x33, 0x52, 0x93, 0xb3, 0x85, 0xc4, 0xf4, 0x20, 0x3a, 0xf5, 0x60, 0x3a, 0xf5, 0x5c, 0x41,
	0xc6, 0x4a, 0x89, 0x63, 0x88, 0x07, 0x83, 0x4d, 0x4c, 0x62, 0x03, 0x0b, 0x18, 0x03, 0x02, 0x00,
	0x00, 0xff, 0xff, 0xb5, 0x0e, 0x31, 0x36, 0xad, 0x00, 0x00, 0x00,
}
