package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
)

func TestGenerateServerFromProto(t *testing.T) {
	services, err := GetServicesFromProto(protofile)
	assert.NoError(t, err)
	err = GenerateServer(services, &Options{
		writer: os.Stdout,
	})
	assert.NoError(t, err)
}
