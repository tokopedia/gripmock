package main

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/quintans/gripmock/example/well_known_types"
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

	r, err := c.ApiInfo(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}

	if r.Name != "Gripmock" {
		log.Fatalf("expecting api name: Gripmock, but got '%v' instead", r.Name)
	}

	log.Printf("Api Name: %v", r.Name)
}
