package main

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
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

	r, err := c.HealthCheck(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatalf("error from grpc: %v", err)
	}
	code := r.GetFields()["code"].GetNumberValue()
	log.Println("response code: %v", code)
}
