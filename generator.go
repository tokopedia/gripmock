package main

import (
	"os"
	"text/template"
)

type generatorParam struct {
	Proto
	Port   string
	PbPath string
}

func GenerateServerFromProto(proto Proto) error {
	param := generatorParam{
		Proto:  proto,
		Port:   ":9000",
		PbPath: "asdf",
	}

	tmpl := template.New("server.tmpl")
	tmpl, err := tmpl.ParseFiles("server.tmpl")
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, param)
}
