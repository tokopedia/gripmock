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

	"github.com/ahmadmuzakki/grpcmock/stub"
)

func main() {
	protoPathPointer := flag.String("proto", "", "Proto file to generate gRPC server")
	outputPointer := flag.String("o", "", "directory to output server.go. default is $GOPATH/src/grpc/")
	grpcPort := flag.String("grpc-port", "4770", "Port of gRPC tcp server")
	adminport := flag.String("admin-port", "4771", "Port of stub admin server")

	flag.Parse()
	fmt.Println("Starting gRPC Mock")

	// run admin stub server
	stub.RunStubServer(*adminport)

	// parse proto files
	protoPath := *protoPathPointer
	if protoPath == "" {
		log.Fatal("proto can't be empty")
	}

	proto, err := parseProto(protoPath)
	if err != nil {
		log.Fatal("can't parse proto ", err)
	}

	// generate grpc server based on proto
	output := *outputPointer + "/"
	if output == "" {
		if os.Getenv("GOPATH") == "" {
			log.Fatal("output is not provided and GOPATH is empty")
		}
		output = os.Getenv("GOPATH") + "src/grpc/"
	}

	file, err := os.Create(output + "server.go")
	if err != nil {
		log.Fatal(err)
	}
	GenerateServerFromProto(proto, &Options{
		writer:    file,
		grpcPort:  *grpcPort,
		adminPort: *adminport,
	})

	// build the server
	protoname := getProtoName(protoPath)

	build := exec.Command("go", "build", "-o", output+"grpcserver", output+"server.go", output+protoname+".pb.go")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	err = build.Run()
	if err != nil {
		log.Fatal(err)
	}

	// and run
	run := exec.Command(output + "grpcserver")
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	err = run.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("grpc server pid: %d", run.Process.Pid)
	runerr := make(chan error)
	go func() {
		runerr <- run.Wait()
	}()

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

func parseProto(protoPath string) (Proto, error) {
	byt, err := ioutil.ReadFile(protoPath)
	if err != nil {
		log.Fatal("Error on reading proto " + err.Error())
	}

	return ParseProto(bytes.NewReader(byt))
}
