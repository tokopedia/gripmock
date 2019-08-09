package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
)

type generatorParam struct {
	Services  []serviceTemplate
	GrpcAddr  string
	AdminPort string
	PbPath    string
}

type serviceTemplate struct {
	Name    string
	Methods []methodTemplate
}

type methodTemplate struct {
	Name        string
	ServiceName string
	MethodType  string
	Input       string
	Output      string
}

const (
	methodTypeStandard = "standard"
	// server to client stream
	methodTypeServerStream = "server-stream"
	// client to server stream
	methodTypeClientStream  = "client-stream"
	methodTypeBidirectional = "bidirectional"
)

type Options struct {
	writer    io.Writer
	grpcAddr  string
	adminPort string
	pbPath    string
	format    bool
}

var SERVER_TEMPLATE string

func init() {
	tmplBox := packr.New("template", "./template")

	s, err := tmplBox.FindString("server.tmpl")
	if err != nil {
		log.Fatal("Can't find server.tmpl")
	}
	SERVER_TEMPLATE = s
}

func GenerateServer(services []Service, opt *Options) error {
	param := generatorParam{
		Services:  convertTemplateService(services),
		GrpcAddr:  opt.grpcAddr,
		AdminPort: opt.adminPort,
		PbPath:    opt.pbPath,
	}

	if opt == nil {
		opt = &Options{}
	}

	if opt.writer == nil {
		opt.writer = os.Stdout
	}

	tmpl := template.New("server.tmpl")
	tmpl, err := tmpl.Parse(SERVER_TEMPLATE)
	if err != nil {
		return fmt.Errorf("template parse %v", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, param)
	if err != nil {
		return fmt.Errorf("template execute %v", err)
	}

	byt := buf.Bytes()
	if opt.format {
		byt, err = format.Source(byt)
		if err != nil {
			return fmt.Errorf("formatting %v", err)
		}
	}

	_, err = opt.writer.Write(byt)
	return err
}

// change the structure also translate method type
func convertTemplateService(services []Service) []serviceTemplate {
	svcTmp := make([]serviceTemplate, len(services))
	for i, svc := range services {
		svcTmp[i].Name = svc.Name
		methods := make([]methodTemplate, len(svc.Methods))
		for j, method := range svc.Methods {
			tipe := methodTypeStandard
			if !method.StreamInput && method.StreamOutput {
				tipe = methodTypeServerStream
			} else if method.StreamInput && !method.StreamOutput {
				tipe = methodTypeClientStream
			} else if method.StreamInput && method.StreamOutput {
				tipe = methodTypeBidirectional
			}

			methods[j] = methodTemplate{
				Name:        strings.Title(method.Name),
				ServiceName: svc.Name,
				Input:       method.Input,
				Output:      method.Output,
				MethodType:  tipe,
			}
		}
		svcTmp[i].Methods = methods
	}
	return svcTmp
}
