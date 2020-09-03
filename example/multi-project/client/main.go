package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/quintans/gripmock/example/multi-project/proto/prj-bar/bar"
	"github.com/quintans/gripmock/example/multi-project/proto/prj-foo/foo"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
