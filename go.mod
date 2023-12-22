module github.com/tokopedia/gripmock

go 1.15

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lithammer/fuzzysearch v1.1.1
	github.com/stretchr/testify v1.7.0
	github.com/tokopedia/gripmock/protogen/example v0.0.0
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201111145450-ac7456db90a6 // indirect
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0 // indirect
)

// this is for generated server to be able to run
replace github.com/tokopedia/gripmock/protogen/example v0.0.0 => ./protogen/example

// this is for example client to be able to run
replace github.com/tokopedia/gripmock/protogen v0.0.0 => ./protogen
