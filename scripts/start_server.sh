#!/bin/bash

# Change to scripts directory
cd "/go/src/grpc"

# Run go mod tidy to ensure dependencies are up to date
echo "Running go mod tidy..."
go get google.golang.org/grpc@v1.47.0
go mod tidy

# Run the server.go file
echo "Running server.go..."
go run server.go 