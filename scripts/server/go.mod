module grpc

go 1.21.13

replace github.com/tokopedia/gripmock/protogen => /go/src/github.com/tokopedia/gripmock/protogen

require (
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.27.1
)
