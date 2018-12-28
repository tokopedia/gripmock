package main

import (
	"testing"

	"github.com/alecthomas/participle"
	"github.com/stretchr/testify/assert"
	"os"
)

func TestGenerateServerFromProto(t *testing.T) {
	parser, err := participle.Build(&Proto{})
	assert.NoError(t, err)
	ast := Proto{}
	err = parser.ParseString(protofile, &ast)
	assert.NoError(t, err)
	err = GenerateServerFromProto([]Proto{ast}, &Options{
		writer: os.Stdout,
	})
	assert.NoError(t, err)
}
