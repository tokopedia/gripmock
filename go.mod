module github.com/tokopedia/gripmock

go 1.22

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.5.3
	github.com/lithammer/fuzzysearch v1.1.8
	github.com/stretchr/testify v1.8.4
	github.com/tokopedia/gripmock/protogen/example v0.0.0
	google.golang.org/grpc v1.61.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// this is for generated server to be able to run
replace github.com/tokopedia/gripmock/protogen/example v0.0.0 => ./protogen/example

// this is for example client to be able to run
replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
