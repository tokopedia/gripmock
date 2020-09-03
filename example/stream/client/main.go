package main

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/quintans/gripmock/example/stream"
	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Set up a connection to the server.
	conn, err := grpc.DialContext(ctx, "localhost:4770", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGripmockClient(conn)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go serverStream(c, wg)

	wg.Add(1)
	go clientStream(c, wg)

	wg.Add(1)
	go bidirectionalStream(c, wg)

	wg.Wait()

}

// server to client streaming
func serverStream(c pb.GripmockClient, wg *sync.WaitGroup) {
	defer wg.Done()
	req := &pb.Request{
		Name: "server-to-client-streaming",
	}
	stream, err := c.ServerStream(context.Background(), req)
	if err != nil {
		log.Fatalf("server stream error: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("s2c error: %v", err)
		}

		log.Printf("s2c message: %s\n", reply.Message)
	}
}

// client to server streaming
func clientStream(c pb.GripmockClient, wg *sync.WaitGroup) {
	defer wg.Done()
	stream, err := c.ClientStream(context.Background())
	if err != nil {
		log.Fatalf("c2s error: %v", err)
	}

	requests := []pb.Request{
		{
			Name: "c2s-1",
		}, {
			Name: "c2s-2",
		},
	}
	for _, req := range requests {
		err := stream.Send(&req)
		if err != nil {
			log.Fatalf("c2s error: %v", err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("c2s error: %v", err)
	}

	log.Printf("c2s message: %v", reply.Message)
}

// bidirectional stream
func bidirectionalStream(c pb.GripmockClient, wg *sync.WaitGroup) {
	stream, err := c.Bidirectional(context.Background())
	if err != nil {
		log.Fatalf("2ds error: %v", err)
	}

	requests := []pb.Request{
		{
			Name: "2ds-message1",
		}, {
			Name: "2ds-message2",
		},
	}

	go func() {
		defer wg.Done()
		for {
			reply, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatalf("2ds error %v", err)
			}

			log.Printf("2ds message: %s\n", reply.Message)
		}
	}()

	for _, request := range requests {
		if err := stream.Send(&request); err != nil {
			log.Fatalf("2ds error: %v", err)
		}
	}
	stream.CloseSend()
}
