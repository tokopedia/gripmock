package main

import (
	"context"
	"log"
	"time"

	pb "github.com/quintans/gripmock/example/multi-files"
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

	c := pb.NewGripmock1Client(conn)

	// Contact the server and print out its response.
	r, err := c.SayHello(context.Background(), &pb.Request1{Name: "tokopedia"})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)

	c2 := pb.NewGripmock2Client(conn)

	// Contact the server and print out its response.
	r2, err := c2.SayHello(context.Background(), &pb.Request2{Name: "tokopedia"})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	log.Printf("Greeting: %s", r2.Message)
}
