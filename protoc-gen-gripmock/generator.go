package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var _ = json.Marshal
var _ = protojson.Unmarshal

func main() {
	// Tip of the hat to Tim Coulson
	// https://medium.com/@tim.r.coulson/writing-a-protoc-plugin-with-google-golang-org-protobuf-cd5aa75f5777

	// Protoc passes pluginpb.CodeGeneratorRequest in via stdin
	// marshalled with Protobuf
	input, _ := io.ReadAll(os.Stdin)
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

	plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

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
	Package string
	Methods []methodTemplate
}

type methodTemplate struct {
	SvcPackage  string
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

//go:embed server.tmpl
var SERVER_TEMPLATE string

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

	deps := map[string]string{}
	for _, proto := range protos {
		alias, pkg := getGoPackage(proto)

		// fatal if go_package is not present
		if pkg == "" {
			log.Fatalf("option go_package is required. but %s doesn't have any", proto.GetName())
		}

		if _, ok := deps[pkg]; ok {
			continue
		}

		deps[pkg] = alias
	}

	return deps
}

var aliases = map[string]bool{}
var aliasNum = 1
var packages = map[string]string{}

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
		// get the alias based on the latest folder
		splitSlash := strings.Split(goPackage, "/")
		// replace - with _
		alias = strings.ReplaceAll(splitSlash[len(splitSlash)-1], "-", "_")
	}

	// if package already discovered just return
	if al, ok := packages[goPackage]; ok {
		alias = al
		return
	}

	// Aliases can't be keywords
	if isKeyword(alias) {
		alias = fmt.Sprintf("%s_pb", alias)
	}

	// in case of found same alias
	// add numbers on it
	if ok := aliases[alias]; ok {
		alias = fmt.Sprintf("%s%d", alias, aliasNum)
		aliasNum++
	}

	packages[goPackage] = alias
	aliases[alias] = true

	return
}

// change the structure also translate method type
func extractServices(protos []*descriptor.FileDescriptorProto) []Service {
	svcTmp := []Service{}
	for _, proto := range protos {
		for _, svc := range proto.GetService() {
			var s Service
			s.Name = svc.GetName()
			alias, _ := getGoPackage(proto)
			if alias != "" {
				s.Package = alias + "."
			}
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
					SvcPackage:  s.Package,
					ServiceName: svc.GetName(),
					Input:       getMessageType(protos, method.GetInputType()),
					Output:      getMessageType(protos, method.GetOutputType()),
					MethodType:  tipe,
				}
			}
			s.Methods = methods
			svcTmp = append(svcTmp, s)
		}
	}
	return svcTmp
}

func getMessageType(protos []*descriptor.FileDescriptorProto, tipe string) string {
	split := strings.Split(tipe, ".")[1:]

	for i := 1; i < len(split); i++ {
		targetPackage := strings.Join(split[:len(split)-i], ".")
		targetType := split[len(split)-i]

		for _, proto := range protos {
			if proto.GetPackage() != targetPackage {
				continue
			}

			for _, msg := range proto.GetMessageType() {
				if msg.GetName() == targetType {
					alias, _ := getGoPackage(proto)
					if alias != "" {
						alias += "."
					}

					targetType = strings.Join(split[len(split)-i:], "_")

					return fmt.Sprintf("%s%s", alias, targetType)
				}
			}
		}
	}

	return split[len(split)-1]
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
