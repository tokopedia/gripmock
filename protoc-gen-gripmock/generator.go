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

	"google.golang.org/protobuf/types/pluginpb"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/markbates/pkger"
	"golang.org/x/tools/imports"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	// Tip of the hat to Tim Coulson
	// https://medium.com/@tim.r.coulson/writing-a-protoc-plugin-with-google-golang-org-protobuf-cd5aa75f5777

	// Protoc passes pluginpb.CodeGeneratorRequest in via stdin
	// marshalled with Protobuf
	input, _ := ioutil.ReadAll(os.Stdin)
	var request pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(input, &request); err != nil {
		log.Fatalf("error unmarshalling [%s]: %v", string(input), err)
	}

	// Initialise our plugin with default options
	opts := protogen.Options{}
	plugin, err := opts.New(&request)
	if err != nil {
		log.Fatalf("error initializing plugin: %v", err)
	}

	protos := make([]*descriptor.FileDescriptorProto, len(plugin.Files))
	for index, file := range plugin.Files {
		protos[index] = file.Proto
	}

	params := make(map[string]string)
	for _, param := range strings.Split(request.GetParameter(), ",") {
		split := strings.Split(param, "=")
		params[split[0]] = split[1]
	}

	buf := new(bytes.Buffer)
	err = generateServer(protos, &Options{
		writer:    buf,
		adminPort: params["admin-port"],
		grpcAddr:  fmt.Sprintf("%s:%s", params["grpc-address"], params["grpc-port"]),
	})

	if err != nil {
		log.Fatalf("Failed to generate server %v", err)
	}

	file := plugin.NewGeneratedFile("server.go", ".")
	file.Write(buf.Bytes())

	// Generate a response from our plugin and marshall as protobuf
	out, err := proto.Marshal(plugin.Response())
	if err != nil {
		log.Fatalf("error marshalling plugin response: %v", err)
	}

	// Write the response to stdout, to be picked up by protoc
	os.Stdout.Write(out)
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
	f, err := pkger.Open("/protoc-gen-gripmock/server.tmpl")
	if err != nil {
		log.Fatalf("error opening server.tmpl: %s", err)
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("error reading server.tmpl: %s", err)
	}

	SERVER_TEMPLATE = string(bytes)
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
	bytProcessed, err := imports.Process("", byt, nil)
	if err != nil {
		return fmt.Errorf("formatting: %v \n%s", err, string(byt))
	}

	_, err = opt.writer.Write(bytProcessed)
	return err
}

func resolveDependencies(protos []*descriptor.FileDescriptorProto) map[string]string {
	depsFile := []string{}
	for _, proto := range protos {
		depsFile = append(depsFile, proto.GetDependency()...)
	}

	deps := map[string]string{}
	aliases := map[string]bool{}
	aliasNum := 1
	for _, dep := range depsFile {
		for _, proto := range protos {
			alias, pkg := getGoPackage(proto)

			// skip whether its not intended deps
			// or has empty Go package
			if proto.GetName() != dep || pkg == "" {
				continue
			}

			// in case of found same alias
			if ok := aliases[alias]; ok {
				alias = fmt.Sprintf("%s%d", alias, aliasNum)
				aliasNum++
			} else {
				aliases[alias] = true
			}
			deps[pkg] = alias
		}
	}

	return deps
}

func getGoPackage(proto *descriptor.FileDescriptorProto) (alias string, goPackage string) {
	goPackage = proto.GetOptions().GetGoPackage()
	if goPackage == "" {
		return
	}

	// support go_package alias declaration
	// https://github.com/golang/protobuf/issues/139
	if splits := strings.Split(goPackage, ";"); len(splits) > 1 {
		goPackage = splits[0]
		alias = splits[1]
	} else {
		splitSlash := strings.Split(proto.GetName(), "/")
		split := strings.Split(splitSlash[len(splitSlash)-1], ".")
		alias = split[0]
	}

	// Aliases can't be keywords
	if isKeyword(alias) {
		alias = fmt.Sprintf("%s_pb", alias)
	}

	return
}

// change the structure also translate method type
func extractServices(protos []*descriptor.FileDescriptorProto) []Service {
	svcTmp := []Service{}
	for _, proto := range protos {
		for _, svc := range proto.GetService() {
			var s Service
			s.Name = svc.GetName()
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
			s.Methods = methods
			svcTmp = append(svcTmp, s)
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
					alias, _ := getGoPackage(proto)
					if alias != "" {
						alias += "."
					}
					return fmt.Sprintf("%s%s", alias, msg.GetName())
				}
			}
		}
	}
	return targetType
}

func isKeyword(word string) bool {
	keywords := [...]string{
		"break",
		"case",
		"chan",
		"const",
		"continue",
		"default",
		"defer",
		"else",
		"fallthrough",
		"for",
		"func",
		"go",
		"goto",
		"if",
		"import",
		"interface",
		"map",
		"package",
		"range",
		"return",
		"select",
		"struct",
		"switch",
		"type",
		"var",
	}

	for _, keyword := range keywords {
		if strings.ToLower(word) == keyword {
			return true
		}
	}

	return false
}
