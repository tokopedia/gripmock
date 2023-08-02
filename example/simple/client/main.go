package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/tokopedia/gripmock/protogen/example/simple"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	r, err := c.SayHello(context.Background(), &pb.Request{Name: name})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	log.Printf("Greeting: %s (return code %d)", r.Message, r.ReturnCode)

	md := metadata.New(map[string]string{"header-1": "value-1"})
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	var headers metadata.MD

	name = "world"
	r, err = c.SayHello(ctx, &pb.Request{Name: name}, grpc.Header(&headers))
	if err != nil {
		log.Fatalf("error from grpc: %s", err)
	}

	header := headers["response-header"]
	if len(header) == 0 {
		log.Fatal("gripmock did not respond with any expected header")
	}
	if header[0] != "response-value" {
		log.Fatal("gripmock did not respond with the expected header")
	}

	log.Printf("Greeting: %s (return code %d)", r.Message, r.ReturnCode)
}
