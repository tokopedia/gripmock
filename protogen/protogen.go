package protogen

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// ProtoGen is a placeholder variable to ensure this package is imported
var ProtoGen = "protogen"

// Use all imports to prevent them from being removed by go mod tidy
var _ = proto.Message(nil)
var _ = grpc.NewServer()
