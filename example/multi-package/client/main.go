package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/quintans/gripmock/example/multi-package"
	"github.com/quintans/gripmock/example/multi-package/bar"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Set up a connection to the server.
	conn, err := grpc.DialContext(ctx, "localhost:4770", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGripmockClient(conn)

	// Contact the server and print out its response.
	name := "tokopedia"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.Greet(context.Background(), &bar.Bar{Name: name})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	log.Printf("Greeting: %s", r.Response)
}
