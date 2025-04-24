package main

import (
	"log"

	"github.com/tokopedia/gripmock/protogen"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func main() {
	// Use all imports to prevent them from being removed by go mod tidy
	_ = proto.Message(nil)
	_ = grpc.NewServer()
	_ = protogen.ProtoGen
	log.Println("Dummy server with imports loaded successfully")
}
