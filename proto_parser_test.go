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
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
  rpc saySmall (HelloRequest) returns (HelloReply) {}
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
	assert.Equal(t, "Greeter", services[0].Name)
}

func TestPickServiceDeclaration(t *testing.T) {
	svcs := pickServiceDeclarations(protofile)
	assert.Len(t, svcs, 1)
}
