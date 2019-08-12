package main

import (
	"bytes"
	"fmt"
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
	"golang.org/x/tools/imports"
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

	gen.CommandLineParameters(gen.Request.GetParameter())

	buf := new(bytes.Buffer)
	err = generateServer(gen.Request.ProtoFile, &Options{
		writer:    buf,
		adminPort: gen.Param["admin-port"],
		grpcAddr:  fmt.Sprintf("%s:%s", gen.Param["grpc-address"], gen.Param["grpc-port"]),
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

type generatorParam struct {
	Services     []Service
	Dependencies map[string]string
	GrpcAddr     string
	AdminPort    string
	PbPath       string
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

func generateServer(protos []*descriptor.FileDescriptorProto, opt *Options) error {
	services := extractServices(protos)
	deps := resolveDependencies(protos)

	param := generatorParam{
		Services:     services,
		Dependencies: deps,
		GrpcAddr:     opt.grpcAddr,
		AdminPort:    opt.adminPort,
		PbPath:       opt.pbPath,
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
	byt, err = imports.Process("", byt, nil)
	if err != nil {
		return fmt.Errorf("formatting %v", err)
	}

	_, err = opt.writer.Write(byt)
	return err
}

func resolveDependencies(protos []*descriptor.FileDescriptorProto) map[string]string {
	depsFile := []string{}
	for _, proto := range protos {
		depsFile = append(depsFile, proto.GetDependency()...)
	}

	deps := map[string]string{}
	for _, dep := range depsFile {
		for _, proto := range protos {
			if proto.GetName() != dep {
				continue
			}
			pkg := proto.GetOptions().GetGoPackage()
			deps[pkg] = getAlias(proto.GetName())
		}
	}

	return deps
}

func getAlias(protoName string) string {
	splitSlash := strings.Split(protoName, "/")
	split := strings.Split(splitSlash[len(splitSlash)-1], ".")
	return split[0]
}

// change the structure also translate method type
func extractServices(protos []*descriptor.FileDescriptorProto) []Service {
	svcTmp := []Service{}
	for _, proto := range protos {
		for i, svc := range proto.GetService() {
			svcTmp = append(svcTmp, Service{})
			svcTmp[i].Name = svc.GetName()
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
					Input:       getMessageType(protos, proto.GetDependency(), method.GetInputType()),
					Output:      getMessageType(protos, proto.GetDependency(), method.GetOutputType()),
					MethodType:  tipe,
				}
			}
			svcTmp[i].Methods = methods
		}
	}
	return svcTmp
}

func getMessageType(protos []*descriptor.FileDescriptorProto, deps []string, tipe string) string {
	split := strings.Split(tipe, ".")[1:]
	targetPackage := strings.Join(split[:len(split)-1], ".")
	targetType := split[len(split)-1]
	for _, dep := range deps {
		for _, proto := range protos {
			if proto.GetName() != dep || proto.GetPackage() != targetPackage {
				continue
			}

			for _, msg := range proto.GetMessageType() {
				if msg.GetName() == targetType {
					alias := getAlias(proto.GetName())
					return fmt.Sprintf("%s.%s", alias, msg.GetName())
				}
			}
		}
	}
	return targetType
}
