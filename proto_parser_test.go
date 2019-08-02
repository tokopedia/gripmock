package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var protofile string

func TestMain(m *testing.M) {
	byt, err := ioutil.ReadFile("./example/pb/hello.proto")
	if err != nil {
		log.Fatal(err)
	}
	protofile = string(byt)
	os.Exit(m.Run())
}

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
