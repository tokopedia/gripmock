package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var protofile = `

syntax = "proto3";

option java_multiple_files = true;
option java_package = "io.grpc.examples.helloworld";
option java_outer_classname = "HelloWorldProto";

package gripmock;

import "dummy";
import "anotherdummy";

// The greeting service definition.
service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply);
  rpc serverStream (stream HelloRequest) returns (HelloReply);
  rpc clientStream (HelloRequest) returns (stream HelloReply);
  rpc bidirectional (stream HelloRequest) returns (stream HelloReply);
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
`

func TestProtoParser(t *testing.T) {
	services, err := GetServicesFromProto(protofile)
	assert.NoError(t, err)

	assert.Len(t, services, 1)

	service := services[0]

	assert.Equal(t, "Greeter", services[0].Name)

	assert.Equal(t, 4, len(service.Methods))

	assert.Equal(t, Method{
		Name:         "SayHello",
		StreamInput:  false,
		Input:        "HelloRequest",
		StreamOutput: false,
		Output:       "HelloReply",
	}, *service.Methods[0])

	assert.Equal(t, Method{
		Name:         "serverStream",
		StreamInput:  true,
		Input:        "HelloRequest",
		StreamOutput: false,
		Output:       "HelloReply",
	}, *service.Methods[1])

	assert.Equal(t, Method{
		Name:         "clientStream",
		StreamInput:  false,
		Input:        "HelloRequest",
		StreamOutput: true,
		Output:       "HelloReply",
	}, *service.Methods[2])

	assert.Equal(t, Method{
		Name:         "bidirectional",
		StreamInput:  true,
		Input:        "HelloRequest",
		StreamOutput: true,
		Output:       "HelloReply",
	}, *service.Methods[3])
}

func TestPickServiceDeclaration(t *testing.T) {
	svcs := pickServiceDeclarations(protofile)
	assert.Len(t, svcs, 1)
}
