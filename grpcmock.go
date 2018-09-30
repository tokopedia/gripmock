package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jekiapp/gripmock/stub"
)

func main() {
	outputPointer := flag.String("o", "", "directory to output server.go. Default is $GOPATH/src/grpc/")
	grpcPort := flag.String("grpc-port", "4770", "Port of gRPC tcp server")
	grpcBindAddr := flag.String("grpc-listen", "", "Adress the gRPC server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	adminport := flag.String("admin-port", "4771", "Port of stub admin server")
	adminBindAddr := flag.String("admin-listen", "", "Adress the admin server will bind to. Default to localhost, set to 0.0.0.0 to use from another machine")
	stubPath := flag.String("stub", "", "Path where the stub files are (Optional)")

	flag.Parse()
	fmt.Println("Starting GripMock")

	output := *outputPointer
	if output == "" {
		if os.Getenv("GOPATH") == "" {
			log.Fatal("output is not provided and GOPATH is empty")
		}
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
	protos, err := parseProto(protoPaths)
	if err != nil {
		log.Fatal("can't parse proto ", err)
	}

	// generate pb.go using protoc
	generateProtoc(protoPaths, output)

	// generate grpc server based on proto
	generateGrpcServer(output, fmt.Sprintf("%s:%s", *grpcBindAddr, *grpcPort), *adminport, protos)

	// build the server
	buildServer(output, protoPaths)

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

func getProtoName(path string) string {
	paths := strings.Split(path, "/")
	filename := paths[len(paths)-1]
	return strings.Split(filename, ".")[0]
}

func parseProto(protoPath []string) ([]Proto, error) {
	protos := []Proto{}
	for _, path := range protoPath {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatal(fmt.Sprintf("Proto file '%s' not found", protoPath))
		}
		byt, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("Error on reading proto " + err.Error())
		}

		proto, err := ParseProto(bytes.NewReader(byt))
		if err != nil {
			return nil, err
		}
		protos = append(protos, proto)
	}
	return protos, nil
}

func generateProtoc(protoPath []string, output string) {
	protodirs := strings.Split(protoPath[0], "/")
	protodir := ""
	if len(protodirs) > 0 {
		protodir = strings.Join(protodirs[:len(protodirs)-1], "/") + "/"
	}

	args := []string{"-I", protodir}
	args = append(args, protoPath...)
	args = append(args, "--go_out=plugins=grpc:"+output)
	protoc := exec.Command("protoc", args...)
	protoc.Stdout = os.Stdout
	protoc.Stderr = os.Stderr
	err := protoc.Run()
	if err != nil {
		log.Fatal("Fail on protoc ", err)
	}

	// change package to "main" on generated code
	for _, proto := range protoPath {
		protoname := getProtoName(proto)
		sed := exec.Command("sed", "-i", `s/^package \w*$/package main/`, output+protoname+".pb.go")
		sed.Stderr = os.Stderr
		sed.Stdout = os.Stdout
		err = sed.Run()
		if err != nil {
			log.Fatal("Fail on sed")
		}
	}
}

func generateGrpcServer(output, grpcAddr, adminPort string, proto []Proto) {
	file, err := os.Create(output + "server.go")
	if err != nil {
		log.Fatal(err)
	}
	GenerateServerFromProto(proto, &Options{
		writer:    file,
		grpcAddr:  grpcAddr,
		adminPort: adminPort,
	})

}

func buildServer(output string, protoPaths []string) {
	args := []string{"build", "-o", output + "grpcserver", output + "server.go"}
	for _, path := range protoPaths {
		args = append(args, output+getProtoName(path)+".pb.go")
	}
	build := exec.Command("go", args...)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	err := build.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func runGrpcServer(output string) (*exec.Cmd, <-chan error) {
	run := exec.Command(output + "grpcserver")
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
