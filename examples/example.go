package main

import (
	"context"
	"fmt"
	"log"
	"time"
	
	gripmock "github.com/Dmytro-Hladkykh/gripmock"
)

func main() {
	// Example usage of simplified gripmock
	
	// Create a server on port 9001 (no proto files for this example)
	server, err := gripmock.NewServer(9001, []string{})
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Start the server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer server.Stop()

	// Wait for server to be ready
	if err := server.WaitForReady(5 * time.Second); err != nil {
		log.Fatal("Server not ready:", err)
	}

	fmt.Printf("gRPC Mock Server running on port %d\n", server.GetPort())
	fmt.Println("Press Ctrl+C to stop")

	// Keep running
	select {}
}