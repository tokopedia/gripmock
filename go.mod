module github.com/tokopedia/gripmock

go 1.23.0

toolchain go1.23.4

require (
	github.com/go-chi/chi/v5 v5.2.1
	github.com/lithammer/fuzzysearch v1.1.8
	github.com/stretchr/testify v1.10.0
	github.com/tokopedia/gripmock/protogen v0.0.0
	golang.org/x/text v0.24.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
