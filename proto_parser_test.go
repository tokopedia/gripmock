package main

import (
	"log"
	"testing"

	"github.com/alecthomas/participle"
	"github.com/stretchr/testify/assert"
)

var protofile = `
syntax = "proto3";
import "adsf";
option java_multiple_files = true;
option java_package = "io.grpc.examples.helloworld";
option java_outer_classname = "HelloWorldProto";

package gripmock;

// The greeting service definition.
service Greeter {
// h
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// Stock Check
// Check Stock protocol. Old version of check stock
message StockCheckRequest {
    repeated StockCheckData requestData = 1;
}

message StockCheckData {
    int64 productID = 1;
    int64 shopID = 2;
    int64 userID = 3;
    int64 quantity = 4;
    bool partial = 5;
}

message StockCheckResponse {
    repeated StockCheckResponseData data = 1;
    bool allowAll = 2;
    repeated string errorMessages = 3;
}

message StockCheckResponseData {
    int64 productID = 1;
    int64 userID = 2;
    int64 shopID = 3;
    int64 quantity = 4;
    bool allow = 5;
    string reason = 6;
    int64 minQty = 7;
    int64 maxQty = 8;
    int64 status = 9;
}
`

func TestProtoParser(t *testing.T) {
	parser, err := participle.Build(&Proto{}, nil)
	assert.NoError(t, err)
	ast := Proto{}
	err = parser.ParseString(protofile, &ast)
	assert.NoError(t, err)
	log.Println(ast.Services[0].Methods[0].Output)

}
