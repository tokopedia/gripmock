package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/tokopedia/gripmock/stub"
)

func main() {
	outputPointer := flag.String("o", "", "directory to output server.go. Default is $GOPATH/src/grpc/")
	grpcPort := flag.String("grpc-port", "4770", "Port of gRPC tcp server")
	grpcBindAddr := flag.String("grpc-listen", "", "Address the gRPC server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	adminport := flag.String("admin-port", "4771", "Port of stub admin server")
	adminBindAddr := flag.String("admin-listen", "", "Address the admin server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	stubPath := flag.String("stub", "/stubs", "Path where the stub files are (Optional)")
	imports := flag.String("imports", "/protobuf", "comma separated imports path. default path /protobuf is where gripmock Dockerfile install WKT protos")

	if len(os.Args) == 0 {
		log.Fatal("No arguments were passed")
	}

	// for backwards compatibility
	if len(os.Args) > 1 && os.Args[1] == "gripmock" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	flag.Parse()
	fmt.Println("Starting GripMock")
	if os.Getenv("GOPATH") == "" {
		log.Fatal("$GOPATH is empty")
	}
	output := *outputPointer
	if output == "" {
		output = os.Getenv("GOPATH") + "/src/grpc"
	}

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
		protoPaths = []string{"/proto"}
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
	})

	// and run
	run, runerr := runGrpcServer(output)

	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	select {
	case err := <-runerr:
		log.Fatal(err)
	case <-term:
		fmt.Println("Stopping gRPC Server")
		_ = run.Process.Kill()
	}
}

type protocParam struct {
	protoPath   []string
	adminPort   string
	grpcAddress string
	grpcPort    string
	output      string
	imports     []string
}

func getProtodirs(protoPath string, imports []string) []string {
	// deduced protodir from protoPath
	splitpath := strings.Split(protoPath, "/")
	protodir := ""
	if len(splitpath) > 0 {
		protodir = path.Join(splitpath[:len(splitpath)-1]...)
	}

	// search protodir prefix
	protodirIdx := -1
	for i := range imports {
		dir := path.Join("protogen", imports[i])
		if strings.HasPrefix(protodir, dir) {
			protodir = dir
			protodirIdx = i
			break
		}
	}

	protodirs := make([]string, 0, len(imports)+1)
	protodirs = append(protodirs, protodir)
	// include all dir in imports, skip if it has been added before
	for i, dir := range imports {
		if i == protodirIdx {
			continue
		}
		protodirs = append(protodirs, dir)
	}
	return protodirs
}

func generateProtoc(param protocParam) {
	protoPaths := []string{}
	for _, proto := range param.protoPath {
		dir, paths := getProtoDirAndPath(proto)
		dir = "protogen/" + strings.TrimLeft(dir, "/")
		paths = fixGoPackage(paths)

		param.imports = append(param.imports, dir)
		protoPaths = append(protoPaths, paths...)
	}
	param.protoPath = protoPaths

	// estimate args length to prevent expand
	args := make([]string, 0, len(param.imports)+len(param.protoPath)+2)
	fmt.Println("Imports:", param.imports)
	for _, dir := range param.imports {
		args = append(args, "-I", dir)
	}

	// the latest go-grpc plugin will generate subfolders under $GOPATH/src based on go_package option
	pbOutput := os.Getenv("GOPATH") + "/src"
	args = append(args, "--go_out="+pbOutput)
	args = append(args, "--go-grpc_out=require_unimplemented_servers=false:"+pbOutput)
	args = append(args, fmt.Sprintf("--gripmock_out=admin-port=%s,grpc-address=%s,grpc-port=%s:%s",
		param.adminPort, param.grpcAddress, param.grpcPort, param.output))
	args = append(args, param.protoPath...)
	protoc := exec.Command("protoc", args...)
	protoc.Stdout = os.Stdout
	protoc.Stderr = os.Stderr
	if err := protoc.Run(); err != nil {
		log.Fatal("Fail on protoc ", err)
	}
}

func getProtoDirAndPath(proto string) (protoDir string, protoPaths []string) {
	stat, err := os.Stat(proto)
	if err != nil {
		fmt.Println(proto)
		log.Fatal(fmt.Errorf("fail to stat proto %s: %w", proto, err))
	}
	if stat.Mode().IsRegular() {
		protoDir = path.Dir(proto)
		protoPaths = append(protoPaths, proto)
	} else if stat.Mode().IsDir() {
		protoDir = proto
		protoPaths = append(protoPaths, readDirProto(proto)...)
	}

	return
}

func readDirProto(proto string) (protoPaths []string) {
	entries, err := os.ReadDir(proto)
	if err != nil {
		log.Fatal(fmt.Errorf("Error reading dir %s: %w", proto, err))
	}
	for _, entry := range entries {
		name := path.Join(proto, entry.Name())
		if entry.Type().IsRegular() {
			if filepath.Ext(entry.Name()) == ".proto" {
				protoPaths = append(protoPaths, name)
			}
		} else if entry.Type().IsDir() {
			protoPaths = append(protoPaths, readDirProto(name)...)
		}
	}

	return protoPaths
}

// append gopackage in proto files if doesn't have any
func fixGoPackage(protoPaths []string) []string {
	fixgopackage := exec.Command("fix_gopackage.sh", protoPaths...)
	buf := &bytes.Buffer{}
	fixgopackage.Stdout = buf
	fixgopackage.Stderr = os.Stderr
	err := fixgopackage.Run()
	if err != nil {
		log.Println("error on fixGoPackage", err)
		return protoPaths
	}

	return strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
}

func runGrpcServer(output string) (*exec.Cmd, <-chan error) {
	run := exec.Command("start_server.sh")
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
