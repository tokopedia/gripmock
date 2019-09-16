// Code generated by protoc-gen-go. DO NOT EDIT.
// source: simple.proto

package simple

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

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

// The request message containing the user's name.
type Request struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_simple_7ef306436b49ce3d, []int{0}
}
func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (dst *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(dst, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// The response message containing the greetings
type Reply struct {
	Message              string   `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Reply) Reset()         { *m = Reply{} }
func (m *Reply) String() string { return proto.CompactTextString(m) }
func (*Reply) ProtoMessage()    {}
func (*Reply) Descriptor() ([]byte, []int) {
	return fileDescriptor_simple_7ef306436b49ce3d, []int{1}
}
func (m *Reply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Reply.Unmarshal(m, b)
}
func (m *Reply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Reply.Marshal(b, m, deterministic)
}
func (dst *Reply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Reply.Merge(dst, src)
}
func (m *Reply) XXX_Size() int {
	return xxx_messageInfo_Reply.Size(m)
}
func (m *Reply) XXX_DiscardUnknown() {
	xxx_messageInfo_Reply.DiscardUnknown(m)
}

var xxx_messageInfo_Reply proto.InternalMessageInfo

func (m *Reply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*Request)(nil), "simple.Request")
	proto.RegisterType((*Reply)(nil), "simple.Reply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SimpleClient is the client API for Simple service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SimpleClient interface {
	// simple unary method
	SayHello(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Reply, error)
}

type simpleClient struct {
	cc *grpc.ClientConn
}

func NewSimpleClient(cc *grpc.ClientConn) SimpleClient {
	return &simpleClient{cc}
}

func (c *simpleClient) SayHello(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Reply, error) {
	out := new(Reply)
	err := c.cc.Invoke(ctx, "/simple.Simple/SayHello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SimpleServer is the server API for Simple service.
type SimpleServer interface {
	// simple unary method
	SayHello(context.Context, *Request) (*Reply, error)
}

func RegisterSimpleServer(s *grpc.Server, srv SimpleServer) {
	s.RegisterService(&_Simple_serviceDesc, srv)
}

func _Simple_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/simple.Simple/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServer).SayHello(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Simple_serviceDesc = grpc.ServiceDesc{
	ServiceName: "simple.Simple",
	HandlerType: (*SimpleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Simple_SayHello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "simple.proto",
}

func init() { proto.RegisterFile("simple.proto", fileDescriptor_simple_7ef306436b49ce3d) }

var fileDescriptor_simple_7ef306436b49ce3d = []byte{
	// 131 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x29, 0xce, 0xcc, 0x2d,
	0xc8, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x83, 0xf0, 0x94, 0x64, 0xb9, 0xd8,
	0x83, 0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0x84, 0xb8, 0x58, 0xf2, 0x12, 0x73, 0x53, 0x25,
	0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xc0, 0x6c, 0x25, 0x45, 0x2e, 0xd6, 0xa0, 0xd4, 0x82, 0x9c,
	0x4a, 0x21, 0x09, 0x2e, 0xf6, 0xdc, 0xd4, 0xe2, 0xe2, 0xc4, 0x74, 0x98, 0x3c, 0x8c, 0x6b, 0x64,
	0xc2, 0xc5, 0x16, 0x0c, 0x36, 0x4b, 0x48, 0x8b, 0x8b, 0x23, 0x38, 0xb1, 0xd2, 0x23, 0x35, 0x27,
	0x27, 0x5f, 0x88, 0x5f, 0x0f, 0x6a, 0x1d, 0xd4, 0x74, 0x29, 0x5e, 0x84, 0x40, 0x41, 0x4e, 0x65,
	0x12, 0x1b, 0xd8, 0x19, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x19, 0x8a, 0x14, 0x59, 0x96,
	0x00, 0x00, 0x00,
}
