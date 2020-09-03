package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/quintans/gripmock/example/upload/client/bar"
	"github.com/quintans/gripmock/example/upload/client/foo"
	"github.com/quintans/gripmock/servers"
	"github.com/quintans/gripmock/tool"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := tool.UploadAsJson("http://localhost:4772/reset", servers.Reset{
		ImportSubDirs: false,
	})
	if err != nil {
		log.Fatalf("did not reset upload server: %v", err)
	}
	// upload proto files
	_, err = tool.ZipFolderAndUpload("http://localhost:4772/upload", "example/upload/proto")
	if err != nil {
		log.Fatalf("did not upload proto: %v", err)
	}

	_, err = tool.UploadJsonFile("http://localhost:4771/add", "example/upload/stub/simple.json")
	if err != nil {
		log.Fatalf("did not upload json: %v", err)
	}

	// Set up a connection to the server.
	conn, err := grpc.DialContext(ctx, "localhost:4770", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := foo.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := "tokopedia"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.Greet(context.Background(), &bar.Request{Name: name})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	log.Printf("Greeting: %s (return code %d)", r.Message, r.ReturnCode)

}
