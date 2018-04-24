package main

import (
	"testing"

	"github.com/alecthomas/participle"
	"github.com/stretchr/testify/assert"
)

func TestGenerateServerFromProto(t *testing.T) {
	parser, err := participle.Build(&Proto{}, nil)
	assert.NoError(t, err)
	ast := Proto{}
	err = parser.ParseString(protofile, &ast)
	assert.NoError(t, err)
	err = GenerateServerFromProto(ast)
	assert.NoError(t, err)
}
