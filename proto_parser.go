package main

import (
	"io"

	"github.com/alecthomas/participle"
)

type Proto struct {
	ToplevelProto headerProto `{ @@ }`
	Services      []*Service  `{ @@ }`
	Message       string      `{ "message" Ident "{" { Ident Ident {Ident} "=" Int ";" } "}" }`
}

type headerProto struct {
	Syntax   string     `"syntax" "=" String  ";" |`
	Imprt    string     `"import" String ";" |`
	Option   string     `"option" Ident "=" {String | Ident} ";" |`
	Package string 		`"package" Ident ";"`
}


type Service struct {
	Name    string    `"service" @Ident "{"`
	Methods []*Method `{ @@ }`
	Closing string    `"}"`
}

type Method struct {
	Name    string `"rpc" @Ident `
	Input   string `"(" @Ident ")"`
	Output  string `"returns" "(" @Ident ")"`
	Closing string `"{""}"`
}

type Input struct {
	Identifier string `@Ident`
}

type Output struct {
	Identifier string `@Ident`
}

func ParseProto(reader io.Reader) (Proto, error) {
	parser, err := participle.Build(&Proto{}, nil)
	proto := Proto{}
	err = parser.Parse(reader, &proto)
	return proto, err
}
