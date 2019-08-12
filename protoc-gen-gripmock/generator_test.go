package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
)

func TestGenerateServerFromProto(t *testing.T) {
	services, err := GetServicesFromProto(protofile)
	assert.NoError(t, err)
	f, _ := os.Create("./example/server/generated_server.go")

	err = GenerateServer(services, &Options{
		writer: f,
	})

	assert.NoError(t, err)
}
