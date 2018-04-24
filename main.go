package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ahmadmuzakki/grpcmock/stub"
)

func main() {
	port := flag.String("http-port", ":4771", "Port of http server")
	fmt.Println("Starting gRPC Mock")

	stub.RunStubServer(*port)

	var term = make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	select {
	case <-term:
		fmt.Println("Stopping gRPC Mock")
	}
}
