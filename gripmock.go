package main

/*
 * "gripmock" is a wrapper to generate the golang protocol files and server
 * implementation from the input .proto files.
 *
 * It invokes protoc to generate the regular golang protobuf client/server
 * packages and a custom "protoc-gen-gripmock" plugin to generate the server
 * implementation.
 * 
 * The server implementation is based on a "server.tmpl" file that's populated
 * with setup code based on the protocol(s) it should support and linked with
 * the stub loading support code.
 *
 * Once the files are all generated, gripmock compiles them to generate a
 * server binary and by default invokes the server binary. The main gripmock
 * binary runs to serve stubs for the API server(s).
 */

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/tokopedia/gripmock/stub"
)

const (
	GENERATED_MODULE_NAME="gripmock/generated"
)

func main() {
	outputPointer := flag.String("o", "generated", "directory to output generated files and binaries. Default is \"generated\"")
	templateDir := flag.String("template-dir", "server_template", "path to directory containing server.tmpl and its go.mod, default \"server_tmpl\"")
	grpcPort := flag.String("grpc-port", "4770", "Port of gRPC tcp server")
	grpcBindAddr := flag.String("grpc-listen", "", "Adress the gRPC server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	adminport := flag.String("admin-port", "4771", "Port of stub admin server")
	adminBindAddr := flag.String("admin-listen", "", "Adress the admin server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	stubPath := flag.String("stub", "", "Path where the stub files are (Optional)")
	imports := flag.String("imports", "/protobuf", "comma separated imports path. default path /protobuf is where gripmock Dockerfile install WKT protos")
	// for backwards compatibility
	if os.Args[1] == "gripmock" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	flag.Parse()
	fmt.Println("Starting GripMock")
	output := *outputPointer

	// for safety
	output += "/"
	if _, err := os.Stat(output); os.IsNotExist(err) {
		os.Mkdir(output, os.ModePerm)
	}

	// run admin stub server
	stub.RunStubServer(stub.Options{
		StubPath: *stubPath,
		Port:     *adminport,
		BindAddr: *adminBindAddr,
	})

	// parse proto files
	protoPaths := flag.Args()

	if len(protoPaths) == 0 {
		log.Fatal("Need at least one proto file")
	}

	importDirs := strings.Split(*imports, ",")

	// generate pb.go and grpc server based on proto
	generateProtoc(protocParam{
		protoPath:   protoPaths,
		adminPort:   *adminport,
		grpcAddress: *grpcBindAddr,
		grpcPort:    *grpcPort,
		output:      output,
		imports:     importDirs,
		templateDir:    *templateDir,
	})

	// Build the server binary
	buildServer(output)

	// and run
	run, runerr := runGrpcServer(output)

	var term = make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	select {
	case err := <-runerr:
		log.Fatal(err)
	case <-term:
		fmt.Println("Stopping gRPC Server")
		run.Process.Kill()
	}
}

type protocParam struct {
	protoPath   []string
	adminPort   string
	grpcAddress string
	grpcPort    string
	output      string
	imports     []string
	templateDir string
}

func generateProtoc(param protocParam) {
	log.Printf("Generating server protocol %s to %s...", param.protoPath, param.output)
	param.protoPath = fixGoPackage(param.protoPath)

	args := []string{
		"-I", "protogen",
	}
	for _, imp := range param.imports {
		args = append(args, "-I", imp)
	}
	args = append(args, param.protoPath...)
	args = append(args,
		"--go_out="+param.output,
		"--go_opt=module="+GENERATED_MODULE_NAME,
		"--go-grpc_out="+param.output,
		"--go-grpc_opt=module="+GENERATED_MODULE_NAME,
	)
	args = append(args,
		"--gripmock_out="+param.output,
		"--gripmock_opt=paths=source_relative",
		"--gripmock_opt=admin-port="+param.adminPort,
		"--gripmock_opt=grpc-address="+param.grpcAddress,
		"--gripmock_opt=grpc-port="+param.grpcPort,
		"--gripmock_opt=template-dir="+param.templateDir,
	)
	log.Printf("invoking \"protoc\" with args %v", args)
	protoc := exec.Command("protoc", args...)
	protoc.Stdout = os.Stdout
	protoc.Stderr = os.Stderr
	err := protoc.Run()
	if err != nil {
		log.Fatal("Fail on protoc ", err)
	}

	log.Print("Generated protocol")
}

// Rewrite the .proto file to replace any go_package directive with one based
// on our local package path for generated servers, GENERATED_MODULE_NAME
//
// Currently delegated to a hacky shell script
//
func fixGoPackage(protoPaths []string) []string {
	fixgopackage := exec.Command("fix_gopackage.sh", protoPaths...)
	fixgopackage.Env = append(fixgopackage.Environ(),
		"GENERATED_MODULE_NAME="+GENERATED_MODULE_NAME)
	buf := &bytes.Buffer{}
	fixgopackage.Stdout = buf
	fixgopackage.Stderr = os.Stderr
	err := fixgopackage.Run()
	if err != nil {
		log.Fatal("error on fixGoPackage", err)
	}

	return strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
}

func runGrpcServer(output string) (*exec.Cmd, <-chan error) {
	run := exec.Command(path.Join(output,"server"))
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	err := run.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("grpc server pid: %d\n", run.Process.Pid)
	runerr := make(chan error)
	go func() {
		runerr <- run.Wait()
	}()
	return run, runerr
}

func buildServer(output string) {
	log.Print("Building server...")
	oldCwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(output); err != nil {
		log.Fatal(err)
	}

	log.Printf("setting module name")
	run := exec.Command("go", "mod", "edit", "-module", GENERATED_MODULE_NAME)
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		log.Fatal("go mod edit: ", err)
	}

	log.Printf("go mod tidy")
	run = exec.Command("go", "mod", "tidy")
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		log.Fatal("go mod tidy: ", err)
	}

	log.Printf("go build")
	run = exec.Command("go", "build", "-o", "server", ".")
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		log.Fatal("go build -o server .: ", err)
	}

	if err := os.Chdir(oldCwd); err != nil {
		log.Fatal(err)
	}
	log.Print("Built ", path.Join(output,"server"))
}
