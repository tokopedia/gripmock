package foo

import (
	context "context"
	reflect "reflect"

	proto "github.com/golang/protobuf/proto"
	"github.com/quintans/gripmock/example/multi-project/proto/prj-bar/bar"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

var File_foo_hello_proto protoreflect.FileDescriptor

var file_foo_hello_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x66, 0x6f, 0x6f, 0x2f, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x03, 0x66, 0x6f, 0x6f, 0x1a, 0x0d, 0x62, 0x61, 0x72, 0x2f, 0x62, 0x61, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x66, 0x6f, 0x6f, 0x2f, 0x66, 0x6f, 0x6f, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x32, 0x2f, 0x0a, 0x07, 0x47, 0x72, 0x65, 0x65, 0x74, 0x65, 0x72, 0x12,
	0x24, 0x0a, 0x05, 0x47, 0x72, 0x65, 0x65, 0x74, 0x12, 0x0c, 0x2e, 0x62, 0x61, 0x72, 0x2e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x66, 0x6f, 0x6f, 0x2e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x16, 0x5a, 0x14, 0x67, 0x72, 0x69, 0x70, 0x6d, 0x6f, 0x63,
	0x6b, 0x2f, 0x70, 0x72, 0x6a, 0x2d, 0x66, 0x6f, 0x6f, 0x2f, 0x66, 0x6f, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_foo_hello_proto_goTypes = []interface{}{
	(*bar.Request)(nil), // 0: bar.Request
	(*Response)(nil),    // 1: foo.Response
}
var file_foo_hello_proto_depIdxs = []int32{
	0, // 0: foo.Greeter.Greet:input_type -> bar.Request
	1, // 1: foo.Greeter.Greet:output_type -> foo.Response
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_foo_hello_proto_init() }
func file_foo_hello_proto_init() {
	if File_foo_hello_proto != nil {
		return
	}
	file_foo_foo_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_foo_hello_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_foo_hello_proto_goTypes,
		DependencyIndexes: file_foo_hello_proto_depIdxs,
	}.Build()
	File_foo_hello_proto = out.File
	file_foo_hello_proto_rawDesc = nil
	file_foo_hello_proto_goTypes = nil
	file_foo_hello_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// GreeterClient is the client API for Greeter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GreeterClient interface {
	Greet(ctx context.Context, in *bar.Request, opts ...grpc.CallOption) (*Response, error)
}

type greeterClient struct {
	cc grpc.ClientConnInterface
}

func NewGreeterClient(cc grpc.ClientConnInterface) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) Greet(ctx context.Context, in *bar.Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/foo.Greeter/Greet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GreeterServer is the server API for Greeter service.
type GreeterServer interface {
	Greet(context.Context, *bar.Request) (*Response, error)
}

// UnimplementedGreeterServer can be embedded to have forward compatible implementations.
type UnimplementedGreeterServer struct {
}

func (*UnimplementedGreeterServer) Greet(context.Context, *bar.Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Greet not implemented")
}

func RegisterGreeterServer(s *grpc.Server, srv GreeterServer) {
	s.RegisterService(&_Greeter_serviceDesc, srv)
}

func _Greeter_Greet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(bar.Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).Greet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/foo.Greeter/Greet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).Greet(ctx, req.(*bar.Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _Greeter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "foo.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Greet",
			Handler:    _Greeter_Greet_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "foo/hello.proto",
}
