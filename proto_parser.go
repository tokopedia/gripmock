package main

import (
	"log"
	"regexp"

	"github.com/alecthomas/participle"
)

type Service struct {
	Name    string    `"service" @Ident "{"`
	Methods []*Method `{ @@ }`
	Closing string    `"}"`
}

type Method struct {
	Name   string `"rpc" @Ident `
	Input  string `"(" @(Ident{ "." Ident }) ")"`
	Output string `"returns" "(" "stream"? @(Ident{ "." Ident }) ")"`
	// TODO deal with body of method
	Closing string `"{"?"}"? ";"?`
}

func GetServicesFromProto(text string) ([]Service, error) {
	parser, err := participle.Build(&Service{})
	if err != nil {
		log.Fatalf("Error creating proto parser %v", err)
	}
	serviceStr := pickServiceDeclarations(text)
	if len(serviceStr) == 0 {
		return nil, nil
	}

	services := []Service{}
	for _, svc := range serviceStr {
		service := Service{}
		err := parser.ParseString(svc, &service)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

var serviceRegex = regexp.MustCompile(`service \w+\s*\{`)
var multiCommentRegex = regexp.MustCompile(`(?sU)/\*(.*)\*/`)
var inlineCommentRegex = regexp.MustCompile(`(?m)//.*$`)

// only pick service declaration from proto file
func pickServiceDeclarations(protoString string) []string {
	// clean the comment first
	protoString = multiCommentRegex.ReplaceAllString(protoString, "")
	protoString = inlineCommentRegex.ReplaceAllString(protoString, "")

	idxs := serviceRegex.FindAllStringIndex(protoString, -1)
	services := []string{}
	for _, idx := range idxs {
		header := protoString[idx[0]:idx[1]]
		body := extractBody(protoString[idx[1]:])
		services = append(services, header+body)
	}
	return services
}

func extractBody(protoString string) string {
	openBracket := 1
	for i, byt := range protoString {
		if byt == '{' {
			openBracket++
			continue
		}

		if byt == '}' {
			openBracket--
		}

		if openBracket == 0 {
			return protoString[:i+1]
		}
	}
	return ""
}
