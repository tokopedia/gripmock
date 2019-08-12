package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func main() {
	gen := generator.New()
	byt, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	err = proto.Unmarshal(byt, gen.Request)
	if err != nil {
		log.Fatalf("Failed to unmarshal proto: %v", err)
	}

	services := []Service{}
	for _, proto := range gen.Request.ProtoFile {
		services = append(services, convertTemplateService(proto.Service)...)
	}

	gen.CommandLineParameters(gen.Request.GetParameter())

	buf := new(bytes.Buffer)
	err = generateServer(services, &Options{
		writer: buf,
	})
	if err != nil {
		log.Fatalf("Failed to generate server %v", err)
	}
	gen.Response.File = []*plugin_go.CodeGeneratorResponse_File{
		{
			Name:    proto.String("server.go"),
			Content: proto.String(buf.String()),
		},
	}

	data, err := proto.Marshal(gen.Response)
	if err != nil {
		gen.Error(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		gen.Error(err, "failed to write output proto")
	}
}

func parseParam() {
	// TODO parse param
}

type generatorParam struct {
	Services  []Service
	GrpcAddr  string
	AdminPort string
	PbPath    string
}

type Service struct {
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
	tmplBox := packr.New("template", "")

	s, err := tmplBox.FindString("server.tmpl")
	if err != nil {
		log.Fatal("Can't find server.tmpl")
	}
	SERVER_TEMPLATE = s
}

func generateServer(services []Service, opt *Options) error {
	param := generatorParam{
		Services:  services,
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
func convertTemplateService(services []*descriptor.ServiceDescriptorProto) []Service {
	svcTmp := make([]Service, len(services))
	for i, svc := range services {
		svcTmp[i].Name = *svc.Name
		methods := make([]methodTemplate, len(svc.Method))
		for j, method := range svc.Method {
			tipe := methodTypeStandard
			if method.GetServerStreaming() && !method.GetClientStreaming() {
				tipe = methodTypeServerStream
			} else if !method.GetServerStreaming() && method.GetClientStreaming() {
				tipe = methodTypeClientStream
			} else if method.GetServerStreaming() && method.GetClientStreaming() {
				tipe = methodTypeBidirectional
			}

			methods[j] = methodTemplate{
				Name:        strings.Title(*method.Name),
				ServiceName: svc.GetName(),
				Input:       getMessageType(method.GetInputType()),
				Output:      getMessageType(method.GetOutputType()),
				MethodType:  tipe,
			}
		}
		svcTmp[i].Methods = methods
	}
	return svcTmp
}

func getMessageType(tipe string) string {
	tipes := strings.Split(tipe, ".")
	return tipes[len(tipes)-1]
}
